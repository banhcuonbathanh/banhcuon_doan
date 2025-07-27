package account_handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	dto "english-ai-full/internal/account/account_dto"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

)

// FindAccountByID handles finding user by ID
func (h *AccountHandler) FindAccountByID(w http.ResponseWriter, r *http.Request) {
	id, apiErr := utils.ParseIDParam(r, "id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.userClient.FindByID(r.Context(), &pb.FindByIDReq{Id: id})
	if err != nil {
		log.Printf("Find user error: %v", err)

		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, errorcustom.NewUserNotFoundByID(id))
			return
		}

		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve user",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, dto.FindAccountByIDResponse{
		ID:        id,
		BranchID:  res.Account.BranchId,
		Name:      res.Account.Name,
		Email:     res.Account.Email,
		Avatar:    res.Account.Avatar,
		Title:     res.Account.Title,
		Role:      res.Account.Role,
		OwnerID:   res.Account.OwnerId,
		CreatedAt: res.Account.CreatedAt.AsTime(),
		UpdatedAt: res.Account.UpdatedAt.AsTime(),
	})
}




// GetUserProfile handles getting user profile
func (h *AccountHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr, _ := utils.GetStringParam(r, "id", 0)

	var id int64
	var err error

	if idStr == "" {
		userID, err := h.getUserIDFromContext(ctx)
		if err != nil {
			utils.HandleError(w, err)
			return
		}
		id = userID
	} else {
		id, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid user ID format",
				http.StatusBadRequest,
			).WithDetail("provided_id", idStr))
			return
		}
	}

	res, err := h.userClient.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			utils.HandleError(w, errorcustom.NewUserNotFoundByID(id))
			return
		}

		log.Printf("Error finding user by ID %d: %v", id, err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve user profile",
			http.StatusInternalServerError,
		))
		return
	}

	userProfile := dto.UserProfile{
		ID:       res.Account.Id,
		BranchID: res.Account.BranchId,
		Name:     res.Account.Name,
		Email:    res.Account.Email,
		Avatar:   res.Account.Avatar,
		Title:    res.Account.Title,
		Role:     res.Account.Role,
		OwnerID:  res.Account.OwnerId,
	}

	utils.RespondWithJSON(w, http.StatusOK, dto.UserProfileResponse{
		User: userProfile,
	})
}