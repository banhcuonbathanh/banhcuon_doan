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

// CreateAccount handles user creation
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := errorcustom.DecodeJSON(r.Body, &req, "create_account", false); err != nil {
		errorcustom.HandleError(w, err, "create_account")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorcustom.HandleValidationErrors(w, validationErrors, "create_account")
		} else {
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "create_account")
		}
		return
	}

	if err := errorcustom.ValidatePasswordWithDetails(req.Password, "create_account"); err != nil {
		errorcustom.HandleError(w, err, "create_account")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		), "create_account")
		return
	}

	userRes, err := h.userClient.CreateUser(r.Context(), &pb.AccountReq{
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	if err != nil {
		log.Printf("User creation error: %v", err)

		if strings.Contains(err.Error(), "already exists") {
			errorcustom.HandleError(w, errorcustom.NewDuplicateEmailError(req.Email), "create_account")
			return
		}

		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"User creation failed",
			http.StatusInternalServerError,
		), "create_account")
		return
	}

	errorcustom.RespondWithJSON(w, http.StatusCreated, dto.CreateUserResponse{
		BranchID: userRes.BranchId,
		Name:     userRes.Name,
		Email:    userRes.Email,
		Avatar:   userRes.Avatar,
		Title:    userRes.Title,
		Role:     userRes.Role,
		OwnerID:  userRes.OwnerId,
	}, "create_account")
}

// UpdateUserByID handles user updates
func (h *AccountHandler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	id, apiErr := errorcustom.ParseIDParam(r, "id")
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	var req dto.UpdateUserRequest
	if err := errorcustom.DecodeJSON(r.Body, &req, "update_user", false); err != nil {
		errorcustom.HandleError(w, err, "update_user")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorcustom.HandleValidationErrors(w, validationErrors, "update_user")
		} else {
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "update_user")
		}
		return
	}

	res, err := h.userClient.UpdateUser(r.Context(), &pb.UpdateUserReq{
		Id:       id,
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	if err != nil {
		log.Printf("Update user error: %v", err)

		if strings.Contains(err.Error(), "not found") {
			errorcustom.HandleError(w, errorcustom.NewUserNotFoundByID(id), "update_user")
			return
		}

		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"User update failed",
			http.StatusInternalServerError,
		), "update_user")
		return
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.UpdateUserResponse{
		User: dto.UserProfile{
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
		},
		Success: true,
		Message: "User updated successfully",
	}, "update_user")
}

// DeleteUser handles user deletion
func (h *AccountHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, apiErr := errorcustom.ParseIDParam(r, "id")
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.userClient.DeleteUser(r.Context(), &pb.DeleteAccountReq{UserID: id})
	if err != nil {
		log.Printf("Delete user error: %v", err)

		if strings.Contains(err.Error(), "not found") {
			errorcustom.HandleError(w, errorcustom.NewUserNotFoundByID(id), "delete_user")
			return
		}

		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"User deletion failed",
			http.StatusInternalServerError,
		), "delete_user")
		return
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.DeleteUserResponse{
		Success: res.Success,
		Message: "User deleted successfully",
	}, "delete_user")
}