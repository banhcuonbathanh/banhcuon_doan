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

// VerifyEmail handles email verification requests
func (h *AccountHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token, apiErr := utils.GetStringParam(r, "token", 1)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.userClient.VerifyEmail(ctx, &pb.VerifyEmailReq{
		VerificationToken: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("verification", "invalid or expired verification token"))
			return
		}
		log.Printf("Error verifying email: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Email verification failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

// ResendVerification handles resend verification email requests
func (h *AccountHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.userClient.ResendVerification(ctx, &pb.ResendVerificationReq{
		Email: req.Email,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{
				"message": "If the email exists and is unverified, a verification email has been sent",
			})
			return
		}
		log.Printf("Error resending verification: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to resend verification email",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}