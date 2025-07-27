 package account_handler

// import (
// 	"log"
// 	"net/http"
// 	"strings"

// 	errorcustom "english-ai-full/internal/error_custom"
// 	pb "english-ai-full/internal/proto_qr/account"
// 	"english-ai-full/utils"

// 	"github.com/go-playground/validator/v10"
// )

// // RefreshToken handles token refresh requests
// // func (h *AccountHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
// // 	ctx := r.Context()

// // 	var req struct {
// // 		RefreshToken string `json:"refresh_token" validate:"required"`
// // 	}

// // 	if err := utils.DecodeJSON(r.Body, &req); err != nil {
// // 		utils.HandleError(w, err)
// // 		return
// // 	}

// // 	if err := h.validator.Struct(&req); err != nil {
// // 		if validationErrors, ok := err.(validator.ValidationErrors); ok {
// // 			utils.HandleValidationErrors(w, validationErrors)
// // 		} else {
// // 			utils.HandleError(w, errorcustom.NewAPIError(
// // 				errorcustom.ErrCodeValidationError,
// // 				"Validation failed",
// // 				http.StatusBadRequest,
// // 			))
// // 		}
// // 		return
// // 	}

// // 	res, err := h.userClient.RefreshToken(ctx, &pb.RefreshTokenReq{
// // 		RefreshToken: req.RefreshToken,
// // 	})
// // 	if err != nil {
// // 		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
// // 			utils.HandleError(w, errorcustom.NewInvalidTokenError("refresh", "invalid or expired"))
// // 			return
// // 		}
// // 		log.Printf("Error refreshing token: %v", err)
// // 		utils.HandleError(w, errorcustom.NewAPIError(
// // 			errorcustom.ErrCodeServiceError,
// // 			"Token refresh failed",
// // 			http.StatusInternalServerError,
// // 		))
// // 		return
// // 	}

// // 	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
// // 		"access_token":  res.AccessToken,
// // 		"refresh_token": res.RefreshToken,
// // 		"expires_at":    res.ExpiresAt,
// // 	})
// // }

// // ValidateToken handles token validation requests
// func (h *AccountHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	authHeader := r.Header.Get("Authorization")
// 	if authHeader == "" {
// 		utils.HandleError(w, errorcustom.NewAPIError(
// 			errorcustom.ErrCodeInvalidInput,
// 			"Missing authorization header",
// 			http.StatusUnauthorized,
// 		))
// 		return
// 	}

// 	token := strings.TrimPrefix(authHeader, "Bearer ")
// 	if token == authHeader {
// 		utils.HandleError(w, errorcustom.NewAPIError(
// 			errorcustom.ErrCodeInvalidInput,
// 			"Invalid authorization header format",
// 			http.StatusUnauthorized,
// 		))
// 		return
// 	}

// 	res, err := h.userClient.ValidateToken(ctx, &pb.ValidateTokenReq{
// 		Token: token,
// 	})
// 	if err != nil {
// 		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
// 			utils.HandleError(w, errorcustom.NewInvalidTokenError("access", "invalid or expired"))
// 			return
// 		}
// 		log.Printf("Error validating token: %v", err)
// 		utils.HandleError(w, errorcustom.NewAPIError(
// 			errorcustom.ErrCodeServiceError,
// 			"Token validation failed",
// 			http.StatusInternalServerError,
// 		))
// 		return
// 	}

// 	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
// 		"valid":      res.Valid,
// 		"expires_at": res.ExpiresAt,
// 		"message":    res.Message,
// 		"id":         res.UserId,
// 	})
// }