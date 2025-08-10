package account_handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	dto "english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"


	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FindAccountByID handles finding user by ID
func (h *AccountHandler) FindAccountByID(w http.ResponseWriter, r *http.Request) {
	id, apiErr := errorcustom.ParseIDParam(r, "id")
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.userClient.FindByID(r.Context(), &pb.FindByIDReq{Id: id})
	if err != nil {
		log.Printf("Find user error: %v", err)

		if strings.Contains(err.Error(), "not found") {
			errorcustom.HandleError(w, errorcustom.NewUserNotFoundByID(id), "find_account_by_id")
			return
		}

		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve user",
			http.StatusInternalServerError,
		), "find_account_by_id")
		return
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.FindAccountByIDResponse{
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
	}, "find_account_by_id")
}

// GetUserProfile handles getting user profile
func (h *AccountHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr, _ := errorcustom.GetStringParam(r, "id", 0)

	var id int64
	var err error

	if idStr == "" {
		userID, err := h.getUserIDFromContext(ctx)
		if err != nil {
			errorcustom.HandleError(w, err, "get_user_profile")
			return
		}
		id = userID
	} else {
		id, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid user ID format",
				http.StatusBadRequest,
			).WithDetail("provided_id", idStr), "get_user_profile")
			return
		}
	}

	res, err := h.userClient.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			errorcustom.HandleError(w, errorcustom.NewUserNotFoundByID(id), "get_user_profile")
			return
		}

		log.Printf("Error finding user by ID %d: %v", id, err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve user profile",
			http.StatusInternalServerError,
		), "get_user_profile")
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

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.UserProfileResponse{
		User: userProfile,
	}, "get_user_profile")
}

// FindAllUsers handles getting all users with pagination
func (h *AccountHandler) FindAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, pageSize, apiErr := h.getPaginationParams(r)
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.userClient.FindAllUsers(ctx, &emptypb.Empty{})
	if err != nil {
		log.Printf("Find all users error: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve users",
			http.StatusInternalServerError,
		), "find_all_users")
		return
	}

	// Apply pagination manually since the proto doesn't support it
	users := res.Accounts
	totalCount := int64(len(users))
	
	start := (page - 1) * pageSize
	end := start + pageSize
	
	if start >= int32(len(users)) {
		users = []*pb.Account{}
	} else {
		if end > int32(len(users)) {
			end = int32(len(users))
		}
		users = users[start:end]
	}

	// Convert to response format
	var userResponses []dto.UserProfile
	for _, user := range users {
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users":       userResponses,
		"total_count": totalCount,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (totalCount + int64(pageSize) - 1) / int64(pageSize),
	}, "find_all_users")
}

// FindByRole handles finding users by role
func (h *AccountHandler) FindByRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	role, apiErr := errorcustom.GetStringParam(r, "role", 1)
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Validate role
	var req struct {
		Role string `validate:"required,role"`
	}
	req.Role = role

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorcustom.HandleValidationErrors(w, validationErrors, "find_by_role")
		} else {
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Invalid role",
				http.StatusBadRequest,
			), "find_by_role")
		}
		return
	}

	res, err := h.userClient.FindByRole(ctx, &pb.FindByRoleReq{
		Role: role,
	})
	if err != nil {
		log.Printf("Find by role error: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve users by role",
			http.StatusInternalServerError,
		), "find_by_role")
		return
	}

	var userResponses []dto.UserProfile
	for _, user := range res.Accounts {
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users": userResponses,
		"count": len(userResponses),
	}, "find_by_role")
}

// FindByBranch handles finding users by branch
func (h *AccountHandler) FindByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	branchID, apiErr := errorcustom.ParseIDParam(r, "branch_id")
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.userClient.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
	})
	if err != nil {
		log.Printf("Find by branch error: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve users by branch",
			http.StatusInternalServerError,
		), "find_by_branch")
		return
	}

	var userResponses []dto.UserProfile
	for _, user := range res.Accounts {
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users": userResponses,
		"count": len(userResponses),
	}, "find_by_branch")
}

// SearchUsers handles advanced user search
func (h *AccountHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query().Get("q")
	role := r.URL.Query().Get("role")
	branchIDStr := r.URL.Query().Get("branch_id")
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	var branchID int64
	if branchIDStr != "" {
		var err error
		branchID, err = strconv.ParseInt(branchIDStr, 10, 64)
		if err != nil {
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid branch_id parameter",
				http.StatusBadRequest,
			), "search_users")
			return
		}
	}

	page, pageSize, apiErr := h.getPaginationParams(r)
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Set defaults
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	statusFilter := []string{}
	if status != "" {
		statusFilter = strings.Split(status, ",")
	}

	// Create the request according to your protobuf definition
	res, err := h.userClient.SearchUsers(ctx, &pb.SearchUsersReq{
		Query:    query,
		Role:     role,
		BranchId: branchID,
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
		Sort: &pb.SortInfo{
			SortBy:    sortBy,
			SortOrder: sortOrder,
		},
		StatusFilter: statusFilter,
	})
	if err != nil {
		log.Printf("Search users error: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Search failed",
			http.StatusInternalServerError,
		), "search_users")
		return
	}

	// Convert response according to your protobuf definition
	var userResponses []dto.UserProfile
	for _, user := range res.Accounts { // Use 'Accounts' field from SearchUsersRes
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	// Calculate total pages
	totalPages := int32(0)
	if res.Pagination != nil && res.Pagination.PageSize > 0 {
		totalPages = (res.Total + res.Pagination.PageSize - 1) / res.Pagination.PageSize
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users":       userResponses,
		"total_count": res.Total, // Use 'Total' field from SearchUsersRes
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"pagination": map[string]interface{}{
			"has_next": res.Pagination != nil && res.Pagination.HasNext,
			"has_prev": res.Pagination != nil && res.Pagination.HasPrev,
		},
	}, "search_users")
}

func (h *AccountHandler) GetUsersByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get branch ID from URL parameter using utility function
	branchID, apiErr := errorcustom.ParseIDParam(r, "branch_id")
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Parse pagination parameters
	page, pageSize, apiErr := h.getPaginationParams(r)
	if apiErr != nil {
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Call the gRPC service to get users by branch
	res, err := h.userClient.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
	})
	if err != nil {
		log.Printf("Error getting users by branch: %v", err)
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to get users by branch",
			http.StatusInternalServerError,
		), "get_users_by_branch")
		return
	}

	// Prepare and send response
	response := map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	}
	errorcustom.RespondWithJSON(w, http.StatusOK, response, "get_users_by_branch")
}