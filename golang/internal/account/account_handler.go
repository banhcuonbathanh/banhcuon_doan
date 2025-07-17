package account

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"english-ai-full/internal/mapping"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
)

var ctx = context.Background()

type Handler struct {
	ctx       context.Context
	user      pb.AccountServiceClient
	validator *validator.Validate
}

// Ensure Handler implements AccountHandlerInterface
var _ AccountHandlerInterface = (*Handler)(nil)

func New(user pb.AccountServiceClient) Handler {
	return Handler{
		validator: validator.New(),
		ctx:       context.Background(),
		user:      user,
	}
}

// Authentication endpoints
func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	req.Password = hashedPassword

	userRes, err := h.user.Register(ctx, &pb.RegisterReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Println("error registering user", err)
		http.Error(w, "error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		ID:     userRes.Id,
		Name:   userRes.Name,
		Email:  userRes.Email,
		Status: userRes.Success,
	})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	userRes, err := h.user.Login(ctx, &pb.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		http.Error(w, fmt.Errorf("invalid email or password %v", err).Error(), http.StatusUnauthorized)
		return
	}

	user := mapping.ToPBUserRes(userRes)

	accessToken, err := utils.GenerateJWTToken(user)
	if err != nil {
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		http.Error(w, "error creating refresh token", http.StatusInternalServerError)
		return
	}

	res := model.LoginUserRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: model.AccountLoginResponse{
			ID:       user.ID,
			BranchID: user.BranchID,
			Name:     user.Name,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Title:    user.Title,
			Role:     string(user.Role),
			OwnerID:  user.OwnerID,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logout successful"))
}

// User management endpoints
func (h Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateUserRequest
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
	var req model.UpdateUserRequest
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
	json.NewEncoder(w).Encode(UpdateUserResponse{
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

// Additional endpoints - NEW IMPLEMENTATIONS
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

	userProfile := UserProfile{
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
	json.NewEncoder(w).Encode(UserProfileResponse{
		User: userProfile,
	})
}

func (h Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter
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

	// Define request structure for password change
	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// First, verify the old password by attempting login
	userRes, err := h.user.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error finding user", http.StatusInternalServerError)
		return
	}

	// Verify old password
	loginRes, err := h.user.Login(ctx, &pb.LoginReq{
		Email:    userRes.Account.Email,
		Password: req.OldPassword,
	})
	if err != nil {
		http.Error(w, "invalid old password", http.StatusUnauthorized)
		return
	}

	// Hash new password
	// hashedNewPassword, err := utils.HashPassword(req.NewPassword)
	// if err != nil {
	// 	http.Error(w, "error hashing new password", http.StatusInternalServerError)
	// 	return
	// }

	// Update user with new password
	_, err = h.user.UpdateUser(ctx, &pb.UpdateUserReq{
		Id:       id,
		BranchId: loginRes.Account.BranchId,
		Name:     loginRes.Account.Name,
		Email:    loginRes.Account.Email,
		Avatar:   loginRes.Account.Avatar,
		Title:    loginRes.Account.Title,
		Role:     loginRes.Account.Role,
		OwnerId:  loginRes.Account.OwnerId,
		// Password: hashedNewPassword, // This assumes your proto has password field
	})
	if err != nil {
		http.Error(w, "error updating password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

func (h Handler) GetUsersByBranch(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()

	// Get branch ID from URL parameter
	branchIDStr := chi.URLParam(r, "branch_id")
	if branchIDStr == "" {
		http.Error(w, "missing branch_id parameter", http.StatusBadRequest)
		return
	}

	// branchID, err := strconv.ParseInt(branchIDStr, 10, 64)
	// if err != nil {
	// 	http.Error(w, "invalid branch_id parameter", http.StatusBadRequest)
	// 	return
	// }

	// Note: This assumes you have a service method to get users by branch
	// You'll need to implement this in your gRPC service
	// For now, I'll provide a placeholder implementation
	
	// This is a placeholder - you need to implement GetUsersByBranch in your service
	// res, err := h.user.GetUsersByBranch(ctx, &pb.GetUsersByBranchReq{BranchId: branchID})
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("error getting users by branch: %v", err), http.StatusInternalServerError)
	// 	return
	// }

	// For now, return a not implemented response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "GetUsersByBranch endpoint not yet implemented in service layer",
	})
}

// Helper function to extract user ID from JWT token context
func (h Handler) getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}