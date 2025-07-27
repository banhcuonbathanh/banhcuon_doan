package account_handler

import (
	"log"
	"net/http"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

// RefreshToken handles token refresh requests
func (h *AccountHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
		}
		return
	}

	res, err := h.userClient.RefreshToken(ctx, &pb.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("refresh", "invalid or expired"))
			return
		}
		log.Printf("Error refreshing token: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Token refresh failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_at":    res.ExpiresAt,
	})
}

// ValidateToken handles token validation requests
func (h *AccountHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Missing authorization header",
			http.StatusUnauthorized,
		))
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Invalid authorization header format",
			http.StatusUnauthorized,
		))
		return
	}

	res, err := h.userClient.ValidateToken(ctx, &pb.ValidateTokenReq{
		Token: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("access", "invalid or expired"))
			return
		}
		log.Printf("Error validating token: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Token validation failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      res.Valid,
		"expires_at": res.ExpiresAt,
		"message":    res.Message,
		"id":         res.UserId,
	})
}

//

// ChangePassword handles password change requests
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.getUserIDFromContext(ctx)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,password"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
		}
		return
	}

	if err := utils.ValidatePasswordWithDetails(req.NewPassword); err != nil {
		utils.HandleError(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	res, err := h.userClient.ChangePassword(ctx, &pb.ChangePasswordReq{
		UserId:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid password") {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Current password is incorrect",
				http.StatusBadRequest,
			))
			return
		}
		log.Printf("Change password error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Password change failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}
//



// ForgotPassword handles forgot password requests
func (h *AccountHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
		}
		return
	}

	res, err := h.userClient.ForgotPassword(ctx, &pb.ForgotPasswordReq{
		Email: req.Email,
	})
	if err != nil {
		log.Printf("Forgot password error: %v", err)
		// Always return success for security reasons
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

// ResetPassword handles password reset requests
func (h *AccountHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,password"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
		}
		return
	}

	if err := utils.ValidatePasswordWithDetails(req.NewPassword); err != nil {
		utils.HandleError(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	res, err := h.userClient.ResetPassword(ctx, &pb.ResetPasswordReq{
		Token:       req.Token,
		NewPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("reset", "invalid or expired reset token"))
			return
		}
		log.Printf("Reset password error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Password reset failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}