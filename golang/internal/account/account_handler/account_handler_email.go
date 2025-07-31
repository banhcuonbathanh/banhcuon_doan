package account_handler

import (
	"log"
	"net/http"
	"strings"

	dto "english-ai-full/internal/account/account_dto"
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
			utils.HandleError(w, errorcustom.NewInvalidTokenError("verification", "invalid or expired verification token"), "verify_email")
			return
		}
		log.Printf("Error verifying email: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Email verification failed",
			http.StatusInternalServerError,
		), "verify_email")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "verify_email")
}

// ResendVerification handles resend verification email requests
func (h *AccountHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := utils.DecodeJSON(r.Body, &req, "resend_verification", false); err != nil {
		utils.HandleError(w, err, "resend_verification")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors, "resend_verification")
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "resend_verification")
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
			}, "resend_verification")
			return
		}
		log.Printf("Error resending verification: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to resend verification email",
			http.StatusInternalServerError,
		), "resend_verification")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "resend_verification")
}

func (h *AccountHandler) FindByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email, apiErr := utils.GetStringParam(r, "email", 1)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	// Validate email format
	var req struct {
		Email string `validate:"required,email"`
	}
	req.Email = email

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors, "find_by_email")
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Invalid email format",
				http.StatusBadRequest,
			), "find_by_email")
		}
		return
	}

	res, err := h.userClient.FindByEmail(ctx, &pb.FindByEmailReq{
		Email: email,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeNotFound,
				"User not found with provided email",
				http.StatusNotFound,
			), "find_by_email")
			return
		}
		log.Printf("Find user by email error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve user",
			http.StatusInternalServerError,
		), "find_by_email")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, dto.FindAccountByIDResponse{
		ID:        res.Account.Id,
		BranchID:  res.Account.BranchId,
		Name:      res.Account.Name,
		Email:     res.Account.Email,
		Avatar:    res.Account.Avatar,
		Title:     res.Account.Title,
		Role:      res.Account.Role,
		OwnerID:   res.Account.OwnerId,
		CreatedAt: res.Account.CreatedAt.AsTime(),
		UpdatedAt: res.Account.UpdatedAt.AsTime(),
	}, "find_by_email")
}