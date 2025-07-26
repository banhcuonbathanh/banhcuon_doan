package account

import (
	"context"
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
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Custom validations ---------------------------------------------------------
func validatePassword(fl validator.FieldLevel) bool {
	return utils.ValidatePassword(fl.Field().String())
}

func validateRole(fl validator.FieldLevel) bool {
	validRoles := map[string]bool{
		"admin":   true,
		"user":    true,
		"manager": true,
	}
	return validRoles[fl.Field().String()]
}

func validateEmailUnique(user pb.AccountServiceClient) validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := user.FindByEmail(context.Background(), &pb.FindByEmailReq{
			Email: fl.Field().String(),
		})
		return err != nil // true if email not found
	}
}

type Handler struct {
	user      pb.AccountServiceClient
	validator *validator.Validate
}

func New(user pb.AccountServiceClient) Handler {
	v := validator.New()
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("role", validateRole)
	v.RegisterValidation("uniqueemail", validateEmailUnique(user))
	
	return Handler{
		validator: v,
		user:      user,
	}
}

// Helper function to extract user ID from JWT token context
func (h Handler) getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// Auth Handlers --------------------------------------------------------------
func (h Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter or JWT token
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		userID, err := h.getUserIDFromContext(ctx)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "missing user identification")
			return
		}
		idStr = strconv.FormatInt(userID, 10)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}

	res, err := h.user.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			utils.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error getting user profile: %v", err))
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

func (h Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	res, err := h.user.RefreshToken(ctx, &pb.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error refreshing token: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_at":    res.ExpiresAt,
	})
}

func (h Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "missing authorization header")
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid authorization header format")
		return
	}

	res, err := h.user.ValidateToken(ctx, &pb.ValidateTokenReq{
		Token: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error validating token: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      res.Valid,
		"expires_at": res.ExpiresAt,
		"message":    res.Message,
		"id":         res.UserId,
	})
}

// User management endpoints
func (h Handler) FindAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, offset := utils.GetPaginationParams(r)

	res, err := h.user.FindAllUsers(ctx, &emptypb.Empty{})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error finding users: %v", err))
		return
	}

	accounts := res.GetAccounts()
	total := len(accounts)
	start, end := utils.CalculatePaginationBounds(int(offset), int(offset+limit), total)
	paginatedAccounts := accounts[start:end]

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"accounts": paginatedAccounts,
		"total":    total,
		"pagination": map[string]interface{}{
			"page":       (offset / limit) + 1,
			"page_size":  limit,
			"total_pages": (int64(total) + limit - 1) / limit,
			"has_next":   end < total,
			"has_prev":   start > 0,
		},
	})
}

// Password management endpoints
func (h Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	res, err := h.user.ForgotPassword(ctx, &pb.ForgotPasswordReq{
		Email: req.Email,
	})
	if err != nil {
		// Security: Don't reveal if email exists
		if strings.Contains(err.Error(), "user not found") {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{
				"message": "If the email exists, a password reset link has been sent",
			})
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error processing forgot password: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

func (h Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "error processing password")
		return
	}

	res, err := h.user.ResetPassword(ctx, &pb.ResetPasswordReq{
		Token:       req.Token,
		NewPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid or expired reset token")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error resetting password: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

// Account verification endpoints
func (h Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	token := chi.URLParam(r, "token")
	if token == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "missing verification token")
		return
	}

	res, err := h.user.VerifyEmail(ctx, &pb.VerifyEmailReq{
		VerificationToken: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid or expired verification token")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error verifying email: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

func (h Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	res, err := h.user.ResendVerification(ctx, &pb.ResendVerificationReq{
		Email: req.Email,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{
				"message": "If the email exists and is unverified, a verification email has been sent",
			})
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error resending verification: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

func (h Handler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := utils.ParseIDParam(r, "id")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive suspended pending"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	res, err := h.user.UpdateAccountStatus(ctx, &pb.UpdateAccountStatusReq{
		UserId: id,
		Status: req.Status,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error updating account status: %v", err))
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

func (h Handler) FindByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	branchID, err := utils.ParseIDParam(r, "branch_id")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.user.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error finding users by branch: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	})
}

func (h Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	query := r.URL.Query().Get("q")
	if query == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "missing search query parameter 'q'")
		return
	}

	// Parse optional parameters
	role := r.URL.Query().Get("role")
	branchID, _ := utils.ParseIDParam(r, "branch_id") // Optional
	statusFilters := r.URL.Query()["status"]
	page, pageSize := utils.GetPaginationParams(r)
	sortBy, sortOrder := utils.GetSortParams(r)

	// Build request
	searchReq := &pb.SearchUsersReq{
		Query:    query,
		Role:     role,
		BranchId: branchID,
		Pagination: &pb.PaginationInfo{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
		Sort: &pb.SortInfo{
			SortBy:    sortBy,
			SortOrder: sortOrder,
		},
		StatusFilter: statusFilters,
	}

	res, err := h.user.SearchUsers(ctx, searchReq)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error searching users: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	})
}

// Enhanced search and filtering
func (h Handler) FindByRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	role := chi.URLParam(r, "role")
	if role == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "missing role parameter")
		return
	}

	page, pageSize := utils.GetPaginationParams(r)

	res, err := h.user.FindByRole(ctx, &pb.FindByRoleReq{
		Role: role,
		Pagination: &pb.PaginationInfo{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error finding users by role: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	})
}

func (h Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := utils.ParseIDParam(r, "id")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "error hashing new password")
		return
	}

	res, err := h.user.ChangePassword(ctx, &pb.ChangePasswordReq{
		UserId:      id,
		CurrentPassword: req.OldPassword,
		NewPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "wrong") {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid old password")
			return
		}
		if strings.Contains(err.Error(), "user not found") {
			utils.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error changing password: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

// Base handlers --------------------------------------------------------------
func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	userRes, err := h.user.Register(r.Context(), &pb.RegisterReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	})
	if err != nil {
		log.Printf("Registration error: %v", err)
		
		if strings.Contains(err.Error(), "already exists") {
			utils.RespondWithError(w, http.StatusConflict, "Email already registered")
			return
		}
		
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, dto.RegisterResponse{
		ID:     userRes.Id,
		Name:   userRes.Name,
		Email:  userRes.Email,
		Status: userRes.Success,
	})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userRes, err := h.user.Login(r.Context(), &pb.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Printf("Login error: %v", err)
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	user := mapping.ToPBUserRes(userRes)
	accessToken, err := utils.GenerateJWTToken(user)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		log.Printf("Refresh token error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, model.LoginUserRes{
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
	})
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}

// User Management Handlers ---------------------------------------------------
func (h Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	userRes, err := h.user.CreateUser(r.Context(), &pb.AccountReq{
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
			utils.RespondWithError(w, http.StatusConflict, "Email already registered")
			return
		}
		
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, dto.CreateUserResponse{
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
	id, err := utils.ParseIDParam(r, "id")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.user.FindByID(r.Context(), &pb.FindByIDReq{Id: id})
	if err != nil {
		log.Printf("Find user error: %v", err)
		
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not retrieve user")
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

func (h Handler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ParseIDParam(r, "id")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.UpdateUserRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.HandleValidationErrors(w, err)
		return
	}

	res, err := h.user.UpdateUser(r.Context(), &pb.UpdateUserReq{
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
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not update user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, dto.UpdateUserResponse{
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
	id, err := utils.ParseIDParam(r, "id")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.user.DeleteUser(r.Context(), &pb.DeleteAccountReq{UserID: id})
	if err != nil {
		log.Printf("Delete user error: %v", err)
		
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not delete user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, dto.DeleteUserResponse{
		Success: res.Success,
		Message: "User deleted successfully",
	})
}

func (h Handler) GetUsersByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get branch ID from URL parameter using utility function
	branchID, err := utils.ParseIDParam(r, "branch_id")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Parse pagination parameters
	page, pageSize := h.getPaginationParams(r)
	
	// Call the gRPC service to get users by branch
	res, err := h.user.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error getting users by branch: %v", err))
		return
	}
	
	// Prepare and send response
	response := map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}

// Helper method to parse pagination parameters specific to your API
func (h Handler) getPaginationParams(r *http.Request) (page, pageSize int32) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	
	// Set defaults
	page, pageSize = 1, 10
	
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
	
	return page, pageSize
}
func (h Handler) FindByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get email from URL parameter
	email := chi.URLParam(r, "email")
	if email == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "missing email parameter")
		return
	}
	
	// Call the gRPC service to find user by email
	user, err := h.user.FindByEmail(ctx, &pb.FindByEmailReq{Email: email})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "error getting user")
		return
	}
	
	// Prepare response
	response := FindByEmailResponse{
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
	}
	
	utils.RespondWithJSON(w, http.StatusOK, response)
}