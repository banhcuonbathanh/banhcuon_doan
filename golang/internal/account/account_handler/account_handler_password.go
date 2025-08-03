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



// ChangePassword handles password change requests
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.getUserIDFromContext(ctx)
	if err != nil {
		utils.HandleError(w, err, "change_password")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,password"`
	}

	if err := utils.DecodeJSON(r.Body, &req, "change_password", false); err != nil {
		utils.HandleError(w, err, "change_password")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors, "change_password")
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "change_password")
		}
		return
	}

	if err := utils.ValidatePasswordWithDetails(req.NewPassword, "change_password"); err != nil {
		utils.HandleError(w, err, "change_password")
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		), "change_password")
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
			), "change_password")
			return
		}
		log.Printf("Change password error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Password change failed",
			http.StatusInternalServerError,
		), "change_password")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "change_password")
}

// ForgotPassword handles forgot password requests
func (h *AccountHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := utils.DecodeJSON(r.Body, &req, "forgot_password", false); err != nil {
		utils.HandleError(w, err, "forgot_password")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors, "forgot_password")
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "forgot_password")
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
		}, "forgot_password")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "forgot_password")
}

// ResetPassword handles password reset requests
func (h *AccountHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,password"`
	}

	if err := utils.DecodeJSON(r.Body, &req, "reset_password", false); err != nil {
		utils.HandleError(w, err, "reset_password")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors, "reset_password")
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "reset_password")
		}
		return
	}

	if err := utils.ValidatePasswordWithDetails(req.NewPassword, "reset_password"); err != nil {
		utils.HandleError(w, err, "reset_password")
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		), "reset_password")
		return
	}

	res, err := h.userClient.ResetPassword(ctx, &pb.ResetPasswordReq{
		Token:       req.Token,
		NewPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("reset", "invalid or expired reset token"), "reset_password")
			return
		}
		log.Printf("Reset password error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Password reset failed",
			http.StatusInternalServerError,
		), "reset_password")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "reset_password")
}