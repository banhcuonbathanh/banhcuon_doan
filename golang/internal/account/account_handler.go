
package account

import (
	"context"

	"log"
	"net/http"
	"strconv"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/mapping"

		dto "english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Validation functions
func validatePassword(fl validator.FieldLevel) bool {
	return utils.ValidatePassword(fl.Field().String()) == nil
}

func validateRole(fl validator.FieldLevel) bool {
	validRoles := map[string]bool{
		"admin":   true,
		"user":    true,
		"manager": true,
	}
	return validRoles[fl.Field().String()]
}

// Improved version with better error handling
func validateEmailUnique(user pb.AccountServiceClient) validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := user.FindByEmail(context.Background(), &pb.FindByEmailReq{
			Email: fl.Field().String(),
		})

		if err == nil {
			return false // Email exists
		}

		// Check for "not found" error specifically
		if status.Code(err) == codes.NotFound {
			return true // Email is unique
		}

		// Log other errors but still consider as valid to avoid blocking registration
		log.Printf("Email uniqueness check error: %v", err)
		return true
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
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"User ID not found in context",
			http.StatusUnauthorized,
		)
	}
	return userID, nil
}

// Token management endpoints
func (h Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
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

	res, err := h.user.RefreshToken(ctx, &pb.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("refresh", "invalid or expired"))
			return
		}
		log.Printf("Error refreshing token: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Token refresh failed",
			http.StatusInternalServerError,
		))
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
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Missing authorization header",
			http.StatusUnauthorized,
		))
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Invalid authorization header format",
			http.StatusUnauthorized,
		))
		return
	}

	res, err := h.user.ValidateToken(ctx, &pb.ValidateTokenReq{
		Token: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("access", "invalid or expired"))
			return
		}
		log.Printf("Error validating token: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Token validation failed",
			http.StatusInternalServerError,
		))
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

	limit, offset, apiErr := utils.GetPaginationParams(r)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.user.FindAllUsers(ctx, &emptypb.Empty{})
	if err != nil {
		log.Printf("Error finding users: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve users",
			http.StatusInternalServerError,
		))
		return
	}

	accounts := res.GetAccounts()
	total := len(accounts)
	start, end := utils.CalculatePaginationBounds(int(offset), int(offset+limit), total)
	paginatedAccounts := accounts[start:end]
	currentPage, totalPages := utils.CalculatePagination(total, int(limit), int(offset))

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"accounts": paginatedAccounts,
		"total":    total,
		"pagination": map[string]interface{}{
			"page":        currentPage,
			"page_size":   limit,
			"total_pages": totalPages,
			"has_next":    end < total,
			"has_prev":    start > 0,
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
		log.Printf("Error processing forgot password: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to process password reset request",
			http.StatusInternalServerError,
		))
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

	// Enhanced password validation
	if err := utils.ValidatePasswordWithDetails(req.NewPassword); err != nil {
		utils.HandleError(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	res, err := h.user.ResetPassword(ctx, &pb.ResetPasswordReq{
		Token:       req.Token,
		NewPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.HandleError(w, errorcustom.NewInvalidTokenError("reset", "invalid or expired reset token"))
			return
		}
		log.Printf("Error resetting password: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Password reset failed",
			http.StatusInternalServerError,
		))
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

	token, apiErr := utils.GetStringParam(r, "token", 1)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.user.VerifyEmail(ctx, &pb.VerifyEmailReq{
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

func (h Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
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

func (h Handler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.user.UpdateAccountStatus(ctx, &pb.UpdateAccountStatusReq{
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

func (h Handler) FindByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	branchID, apiErr := utils.ParseIDParam(r, "branch_id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.user.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
	})
	if err != nil {
		log.Printf("Error finding users by branch: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to find users by branch",
			http.StatusInternalServerError,
		))
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
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Missing search query parameter 'q'",
			http.StatusBadRequest,
		))
		return
	}

	// Parse optional parameters
	role := r.URL.Query().Get("role")
	branchID, _ := utils.ParseIDParam(r, "branch_id") // Optional
	statusFilters := r.URL.Query()["status"]
	
	page, pageSize, apiErr := utils.GetPaginationParams(r)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	allowedSortFields := []string{"name", "email", "created_at", "updated_at", "role"}
	sortBy, sortOrder, apiErr := utils.GetSortParams(r, allowedSortFields)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

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
		log.Printf("Error searching users: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Search failed",
			http.StatusInternalServerError,
		))
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

	role, apiErr := utils.GetStringParam(r, "role", 1)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	page, pageSize, apiErr := utils.GetPaginationParams(r)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.user.FindByRole(ctx, &pb.FindByRoleReq{
		Role: role,
		Pagination: &pb.PaginationInfo{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	})
	if err != nil {
		log.Printf("Error finding users by role: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to find users by role",
			http.StatusInternalServerError,
		))
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

	id, apiErr := utils.ParseIDParam(r, "id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
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
	// Enhanced password validation
	if err := utils.ValidatePasswordWithDetails(req.NewPassword); err != nil {
		utils.HandleError(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing new password: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	res, err := h.user.ChangePassword(ctx, &pb.ChangePasswordReq{
		UserId:          id,
		CurrentPassword: req.OldPassword,
		NewPassword:     hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "wrong") {
			utils.HandleError(w, errorcustom.NewAuthenticationError("invalid old password"))
			return
		}
		if strings.Contains(err.Error(), "user not found") {
			utils.HandleError(w, errorcustom.NewUserNotFoundByID(id))
			return
		}
		log.Printf("Error changing password: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Password change failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
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
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
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

	// Enhanced password validation
	if err := utils.ValidatePasswordWithDetails(req.Password); err != nil {
		utils.HandleError(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		))
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
			utils.HandleError(w, errorcustom.NewDuplicateEmailError(req.Email))
			return
		}

		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"User creation failed",
			http.StatusInternalServerError,
		))
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
	id, apiErr := utils.ParseIDParam(r, "id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.user.FindByID(r.Context(), &pb.FindByIDReq{Id: id})
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

func (h Handler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	id, apiErr := utils.ParseIDParam(r, "id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	var req dto.UpdateUserRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
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
			utils.HandleError(w, errorcustom.NewUserNotFoundByID(id))
			return
		}

		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"User update failed",
			http.StatusInternalServerError,
		))
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
	id, apiErr := utils.ParseIDParam(r, "id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	res, err := h.user.DeleteUser(r.Context(), &pb.DeleteAccountReq{UserID: id})
	if err != nil {
		log.Printf("Delete user error: %v", err)

		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, errorcustom.NewUserNotFoundByID(id))
			return
		}

		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"User deletion failed",
			http.StatusInternalServerError,
		))
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
	branchID, apiErr := utils.ParseIDParam(r, "branch_id")
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	// Parse pagination parameters
	page, pageSize, apiErr := h.getPaginationParams(r)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
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
		log.Printf("Error getting users by branch: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to get users by branch",
			http.StatusInternalServerError,
		))
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

// Helper method to parse pagination parameters with improved error handling
func (h Handler) getPaginationParams(r *http.Request) (page, pageSize int32, apiErr *errorcustom.APIError) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	// Set defaults
	page, pageSize = 1, 10

	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 32); err != nil || p < 1 {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page parameter: must be a positive integer",
				http.StatusBadRequest,
			)
		} else {
			page = int32(p)
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.ParseInt(pageSizeStr, 10, 32); err != nil || ps < 1 || ps > 100 {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page_size parameter: must be between 1 and 100",
				http.StatusBadRequest,
			)
		} else {
			pageSize = int32(ps)
		}
	}

	return page, pageSize, nil
}

func (h Handler) FindByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email, apiErr := utils.GetStringParam(r, "email", 1)
	if apiErr != nil {
		utils.RespondWithAPIError(w, apiErr)
		return
	}

	// Call the gRPC service to find user by email
	user, err := h.user.FindByEmail(ctx, &pb.FindByEmailReq{Email: email})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, errorcustom.NewUserNotFoundByEmail(email))
			return
		}
		log.Printf("Error getting user by email: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve user",
			http.StatusInternalServerError,
		))
		return
	}

	// Prepare response
	response := dto.FindByEmailResponse{
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

func (h Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr, _ := utils.GetStringParam(r, "id", 0) // Allow empty for context extraction

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

	res, err := h.user.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		// Convert gRPC errors to domain errors
		if strings.Contains(err.Error(), "user not found") {
			utils.HandleError(w, errorcustom.NewUserNotFoundByID(id))
			return
		}

		// Log the actual error but don't expose it
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

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
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

	// Use the enhanced password validation with details
	if err := utils.ValidatePasswordWithDetails(req.Password); err != nil {
		utils.HandleError(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		))
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
			utils.HandleError(w, errorcustom.NewDuplicateEmailError(req.Email))
			return
		}

		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Registration failed",
			http.StatusInternalServerError,
		))
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

	userRes, err := h.user.Login(r.Context(), &pb.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Printf("Login error for email %s: %v", req.Email, err)

		// Don't reveal whether user exists or password is wrong
		authErr := errorcustom.NewAuthenticationError("invalid credentials")
		utils.HandleError(w, authErr)
		return
	}

	user := mapping.ToPBUserRes(userRes)
	accessToken, err := utils.GenerateJWTToken(user)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		log.Printf("Refresh token error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed",
			http.StatusInternalServerError,
		))
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