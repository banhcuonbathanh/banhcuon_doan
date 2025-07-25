package account_handler



import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	dto "english-ai-full/internal/account/account_dto"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := h.validator.Struct(&req); err != nil {
		// Handle validation errors
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				validationErrors = append(validationErrors, fmt.Sprintf("%s is required", err.Field()))
			case "min":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
			case "max":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must not exceed %s characters", err.Field(), err.Param()))
			case "email":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be a valid email address", err.Field()))
			case "gt":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param()))
			case "url":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be a valid URL", err.Field()))
			case "oneof":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param()))
			default:
				validationErrors = append(validationErrors, fmt.Sprintf("%s is invalid", err.Field()))
			}
		}
		
		response := map[string]interface{}{
			"error": "validation failed",
			"details": validationErrors,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	req.Password = hashedPassword

	// Call the user service to create the user
	userRes, err := h.user.CreateUser(ctx, &pb.AccountReq{
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateUserResponse{
		BranchID: userRes.BranchId,
		Name:     userRes.Name,
		Email:    userRes.Email,
		Avatar:   userRes.Avatar,
		Title:    userRes.Title,
		Role:     userRes.Role,
		OwnerID:  userRes.OwnerId,
	})
}

func (h Handler) FindAccountByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	res, err := h.user.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, ErrorUserNotFound.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("error finding user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FindAccountByIDResponse{
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

func (h Handler) FindByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	user, err := h.user.FindByEmail(ctx, &pb.FindByEmailReq{Email: email})
	if err != nil {
		http.Error(w, "error getting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FindByEmailResponse{
		ID:        user.Account.Id,
		BranchID:  user.Account.BranchId,
		Name:      user.Account.Name,
		Email:     user.Account.Email,
		Avatar:    user.Account.Avatar,
		Title:     user.Account.Title,
		Role:      user.Account.Role,
		OwnerID:   user.Account.OwnerId,
		CreatedAt: user.Account.CreatedAt.AsTime(),
		UpdatedAt: user.Account.UpdatedAt.AsTime(),
	})
}

func (h Handler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, ErrMissingParameter.Error(), http.StatusBadRequest)
		return
	}

	// Parse user ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidParameter.Error(), http.StatusBadRequest)
		return
	}

	// Decode request body
	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ErrDecodeFailed.Error(), http.StatusBadRequest)
		return
	}

	// Set the ID from URL parameter
	req.ID = id

	// Call the service to update the user
	res, err := h.user.UpdateUser(ctx, &pb.UpdateUserReq{
		Id:       req.ID,
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, ErrorUserNotFound.Error(), http.StatusNotFound)
			return
		}

		log.Printf("failed to update user: %v", err)
		http.Error(w, ErrUpdateUserFailed.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.UpdateUserResponse{
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
	})
}

func (h Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	// Parse user ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	// Call the service to delete the user
	res, err := h.user.DeleteUser(ctx, &pb.DeleteAccountReq{UserID: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, ErrorUserNotFound.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("error deleting user: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeleteUserResponse{
		Success: res.Success,
		Message: "User deleted successfully",
	})
}

func (h Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter or JWT token
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		// If no ID in URL, try to get from JWT token in context
		// This assumes you have middleware that adds user info to context
		userID, ok := ctx.Value("user_id").(int64)
		if !ok {
			http.Error(w, "missing user identification", http.StatusBadRequest)
			return
		}
		idStr = strconv.FormatInt(userID, 10)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	res, err := h.user.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error getting user profile: %v", err), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.UserProfileResponse{
		User: userProfile,
	})
}

func (h Handler) FindAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var limit, offset int64 = 10, 0 // defaults
	
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil {
			offset = o
		}
	}

	// Use google.protobuf.Empty as defined in the proto
	res, err := h.user.FindAllUsers(ctx, &emptypb.Empty{})
	if err != nil {
		http.Error(w, fmt.Sprintf("error finding users: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply pagination to the response if needed
	// Note: This is client-side pagination. Ideally, pagination should be handled in the gRPC service
	accounts := res.GetAccounts()
	total := len(accounts)
	
	// Calculate pagination
	start := int(offset)
	end := int(offset + limit)
	
	if start >= total {
		start = total
	}
	if end >= total {
		end = total
	}
	
	paginatedAccounts := accounts[start:end]
	
	// Create response with pagination info
	response := map[string]interface{}{
		"accounts": paginatedAccounts,
		"total":    total,
		"pagination": map[string]interface{}{
			"page":      (offset / limit) + 1,
			"page_size": limit,
			"total_pages": (int64(total) + limit - 1) / limit, // ceiling division
			"has_next":  end < total,
			"has_prev":  start > 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}