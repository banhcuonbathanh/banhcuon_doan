 package account_handler
import (
	"log"
	"net/http"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"


	"github.com/go-playground/validator/v10"
)


// RefreshToken handles token refresh requests
func (h *AccountHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := errorcustom.DecodeJSON(r.Body, &req, "refresh_token", false); err != nil {
		errorcustom.HandleError(w, err, "refresh_token")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorcustom.HandleValidationErrors(w, validationErrors, "refresh_token")
		} else {
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "refresh_token")
		}
		return
	}

	res, err := h.userClient.RefreshToken(ctx, &pb.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			errorcustom.HandleError(w, errorcustom.NewInvalidTokenError("refresh", "invalid or expired"), "refresh_token")
			return
		}
		log.Printf("Error refreshing token: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Token refresh failed",
			http.StatusInternalServerError,
		), "refresh_token")
		return
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_at":    res.ExpiresAt,
	}, "refresh_token")
}

// ValidateToken handles token validation requests
func (h *AccountHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Missing authorization header",
			http.StatusUnauthorized,
		), "validate_token")
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Invalid authorization header format",
			http.StatusUnauthorized,
		), "validate_token")
		return
	}

	res, err := h.userClient.ValidateToken(ctx, &pb.ValidateTokenReq{
		Token: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			errorcustom.HandleError(w, errorcustom.NewInvalidTokenError("access", "invalid or expired"), "validate_token")
			return
		}
		log.Printf("Error validating token: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Token validation failed",
			http.StatusInternalServerError,
		), "validate_token")
		return
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      res.Valid,
		"expires_at": res.ExpiresAt,
		"message":    res.Message,
		"id":         res.UserId,
	}, "validate_token")
}