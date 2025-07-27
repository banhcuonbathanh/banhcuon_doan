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

// UpdateAccountStatus handles account status updates
func (h *AccountHandler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, apiErr := utils.ParseIDParam(r, "id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive suspended pending"`
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

	res, err := h.userClient.UpdateAccountStatus(ctx, &pb.UpdateAccountStatusReq{
		UserId: id,
		Status: req.Status,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, errorcustom.NewUserNotFoundByID(id))
			return
		}
		log.Printf("Error updating account status: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to update account status",
			http.StatusInternalServerError,
		))
		return
	}

	status := http.StatusOK
	if !res.Success {
		status = http.StatusBadRequest
	}

	utils.RespondWithJSON(w, status, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}