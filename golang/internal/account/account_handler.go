package account

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	dto "english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/mapping"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
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
	var req dto.LoginRequest
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




// Helper function to extract user ID from JWT token context
func (h Handler) getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// new1 asdlifjlasdjfl;ajsd;fljas;ldfj;alsjfd;lasjfljasl;fj

// Add these missing methods to your Handler struct to implement AccountHandlerInterface

// Authentication endpoints
func (h Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
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

	// Call the gRPC service to refresh token
	res, err := h.user.RefreshToken(ctx, &pb.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("error refreshing token: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_at":    res.ExpiresAt,
		
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Call the gRPC service to validate token
	res, err := h.user.ValidateToken(ctx, &pb.ValidateTokenReq{
		Token: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("error validating token: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"valid":      res.Valid,
		"expires_at": res.ExpiresAt,
		"message": res.Message,
	"id": res.UserId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// User management endpoints
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
// Password management endpoints
func (h Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// Call the gRPC service to initiate forgot password
	res, err := h.user.ForgotPassword(ctx, &pb.ForgotPasswordReq{
		Email: req.Email,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			// For security, don't reveal whether email exists or not
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "If the email exists, a password reset link has been sent",
			})
			return
		}
		http.Error(w, fmt.Sprintf("error processing forgot password: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
		"token":   res.ResetToken, // Only for testing - remove in production
	})
}


func (h Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "error processing password", http.StatusInternalServerError)
		return
	}

	// Call the gRPC service to reset password
	res, err := h.user.ResetPassword(ctx, &pb.ResetPasswordReq{
		Token:       req.Token,
		NewPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "invalid or expired reset token", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("error resetting password: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}
// Account verification endpoints
func (h Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "missing verification token", http.StatusBadRequest)
		return
	}

	// Call the gRPC service to verify email
	res, err := h.user.VerifyEmail(ctx, &pb.VerifyEmailReq{
		VerificationToken: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "invalid or expired verification token", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("error verifying email: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,

	})
}

func (h Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// Call the gRPC service to resend verification
	res, err := h.user.ResendVerification(ctx, &pb.ResendVerificationReq{
		Email: req.Email,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			// For security, don't reveal whether email exists or not
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "If the email exists and is unverified, a verification email has been sent",
			})
			return
		}
		http.Error(w, fmt.Sprintf("error resending verification: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

func (h Handler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive suspended pending"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// Call the gRPC service to update account status
	res, err := h.user.UpdateAccountStatus(ctx, &pb.UpdateAccountStatusReq{
		UserId: id,
		Status: req.Status,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating account status: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}

	w.Header().Set("Content-Type", "application/json")
	if res.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}



func (h Handler) FindByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	branchIDStr := chi.URLParam(r, "branch_id")
	if branchIDStr == "" {
		http.Error(w, "missing branch_id parameter", http.StatusBadRequest)
		return
	}

	branchID, err := strconv.ParseInt(branchIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid branch_id parameter", http.StatusBadRequest)
		return
	}

	// Call the gRPC service to find users by branch
	res, err := h.user.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error finding users by branch: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "missing search query parameter 'q'", http.StatusBadRequest)
		return
	}

	// Parse optional filter parameters
	role := r.URL.Query().Get("role")
	branchIDStr := r.URL.Query().Get("branch_id")
	statusFilters := r.URL.Query()["status"] // Get multiple status values
	
	var branchID int64
	if branchIDStr != "" {
		if id, err := strconv.ParseInt(branchIDStr, 10, 64); err == nil {
			branchID = id
		}
	}

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	var page, pageSize int32 = 1, 10 // defaults
	
	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 32); err == nil && p > 0 {
			page = int32(p)
		}
	}
	
	if pageSizeStr != "" {
		if ps, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil && ps > 0 && ps <= 100 {
			pageSize = int32(ps)
		}
	}

	// Parse sorting parameters
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")
	
	// Default sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Build the search request
	searchReq := &pb.SearchUsersReq{
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
		StatusFilter: statusFilters,
	}

	// Call the gRPC service to search users
	res, err := h.user.SearchUsers(ctx, searchReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("error searching users: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// new1 doneeeeeee


// mew 2 -------

// Enhanced search and filtering endpoints
func (h Handler) FindByRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	role := chi.URLParam(r, "role")
	if role == "" {
		http.Error(w, "missing role parameter", http.StatusBadRequest)
		return
	}

	// Parse optional pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	var page, pageSize int32 = 1, 10 // defaults
	
	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 32); err == nil && p > 0 {
			page = int32(p)
		}
	}
	
	if pageSizeStr != "" {
		if ps, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil && ps > 0 && ps <= 100 {
			pageSize = int32(ps)
		}
	}

	// Call the gRPC service to find users by role
	res, err := h.user.FindByRole(ctx, &pb.FindByRoleReq{
		Role: role,
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error finding users by role: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h Handler) GetUsersByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get branch ID from URL parameter
	branchIDStr := chi.URLParam(r, "branch_id")
	if branchIDStr == "" {
		http.Error(w, "missing branch_id parameter", http.StatusBadRequest)
		return
	}

	branchID, err := strconv.ParseInt(branchIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid branch_id parameter", http.StatusBadRequest)
		return
	}

	// Parse optional pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	var page, pageSize int32 = 1, 10 // defaults
	
	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 32); err == nil && p > 0 {
			page = int32(p)
		}
	}
	
	if pageSizeStr != "" {
		if ps, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil && ps > 0 && ps <= 100 {
			pageSize = int32(ps)
		}
	}

	// Call the gRPC service to get users by branch
	res, err := h.user.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting users by branch: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Fix the ChangePassword method to properly hash the new password
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

	// Hash new password
	hashedNewPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "error hashing new password", http.StatusInternalServerError)
		return
	}

	// Call the gRPC service to change password
	res, err := h.user.ChangePassword(ctx, &pb.ChangePasswordReq{
		UserId:      id,
		CurrentPassword: req.OldPassword,
		NewPassword: hashedNewPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "wrong") {
			http.Error(w, "invalid old password", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error changing password: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}
// new 2 done