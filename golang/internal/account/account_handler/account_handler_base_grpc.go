// ============================================================================
// REQUEST TYPES FOR GRPC METHODS
// ============================================================================
package account_handler

// CreateUserRequest represents user creation request

// ============================================================================
// MISSING GRPC METHODS FOR BUSINESS LOGIC
// ============================================================================package account_handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ============================================================================
// MISSING GRPC METHODS FOR BUSINESS LOGIC
// ============================================================================

// createUserViaGRPC creates a new user account
func (h *BaseAccountHandler) createUserViaGRPC(ctx context.Context, req account_dto.CreateUserRequest) (*pb.Account, error) {
	operation := "create_user_grpc"
	
	h.logger.Debug("Creating user via gRPC", map[string]interface{}{
		"email": req.Email,
		"role":  req.Role,
	})
	
	// Create user request matching your protobuf schema
	grpcReq := &pb.AccountReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		Title:    req.Title,
		Avatar:   req.Avatar,
		BranchId: req.BranchID,
		OwnerId:  req.OwnerID,
	}
	
	// Call gRPC service
	user, err := h.userClient.CreateUser(ctx, grpcReq)
	if err != nil {
		return nil, h.handleGRPCError(err, operation, map[string]interface{}{
			"email": req.Email,
			"role":  req.Role,
		})
	}
	
	if user == nil {
		return nil, errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			"Empty response from create user service",
			nil,
		)
	}
	
	h.logger.Info("User created successfully via gRPC", map[string]interface{}{
		"user_id": user.Id,
		"email":   user.Email,
		"role":    user.Role,
	})
	
	return user, nil
}

// authenticateUserViaGRPC authenticates user and returns user info
func (h *BaseAccountHandler) authenticateUserViaGRPC(ctx context.Context, email, password string) (*pb.Account, error) {
	operation := "authenticate_user_grpc"
	
	h.logger.Debug("Authenticating user via gRPC", map[string]interface{}{
		"email": email,
	})
	
	// Create login request
	loginReq := &pb.LoginReq{
		Email:    email,
		Password: password,
	}
	
	// Call gRPC service
	resp, err := h.userClient.Login(ctx, loginReq)
	if err != nil {
		return nil, h.handleGRPCError(err, operation, map[string]interface{}{
			"email": email,
		})
	}
	
	if resp == nil || resp.Account == nil {
		return nil, errorcustom.NewAuthenticationError(h.domain, "Invalid credentials")
	}
	
	h.logger.Debug("User authenticated successfully via gRPC", map[string]interface{}{
		"user_id": resp.Account.Id,
		"email":   resp.Account.Email,
	})
	
	return resp.Account, nil
}

// getUserByID retrieves user by ID via gRPC
func (h *BaseAccountHandler) getUserByID(userID int64) (*pb.Account, error) {
	operation := "get_user_by_id_grpc"
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	h.logger.Debug("Getting user by ID via gRPC", map[string]interface{}{
		"user_id": userID,
	})
	
	// Create find by ID request
	req := &pb.FindByIDReq{
		Id: userID,
	}
	
	// Call gRPC service
	resp, err := h.userClient.FindByID(ctx, req)
	if err != nil {
		return nil, h.handleGRPCError(err, operation, map[string]interface{}{
			"user_id": userID,
		})
	}
	
	if resp == nil || resp.Account == nil {
		return nil, errorcustom.NewNotFoundError(h.domain, "user", fmt.Sprintf("User with ID %d not found", userID))
	}
	
	h.logger.Debug("User retrieved successfully via gRPC", map[string]interface{}{
		"user_id": resp.Account.Id,
		"email":   resp.Account.Email,
	})
	
	return resp.Account, nil
}

// getUserByEmail retrieves user by email via gRPC
func (h *BaseAccountHandler) getUserByEmail(email string) (*pb.Account, error) {
	operation := "get_user_by_email_grpc"
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	h.logger.Debug("Getting user by email via gRPC", map[string]interface{}{
		"email": email,
	})
	
	// Create find by email request
	req := &pb.FindByEmailReq{
		Email: email,
	}
	
	// Call gRPC service
	resp, err := h.userClient.FindByEmail(ctx, req)
	if err != nil {
		return nil, h.handleGRPCError(err, operation, map[string]interface{}{
			"email": email,
		})
	}
	
	if resp == nil || resp.Account == nil {
		return nil, errorcustom.NewNotFoundError(h.domain, "user", fmt.Sprintf("User with email %s not found", email))
	}
	
	h.logger.Debug("User retrieved successfully via gRPC", map[string]interface{}{
		"user_id": resp.Account.Id,
		"email":   resp.Account.Email,
	})
	
	return resp.Account, nil
}

// updateUserViaGRPC updates user profile via gRPC
func (h *BaseAccountHandler) updateUserViaGRPC(ctx context.Context, userID int64, updates map[string]interface{}) (*pb.Account, error) {
	operation := "update_user_grpc"
	
	h.logger.Debug("Updating user via gRPC", map[string]interface{}{
		"user_id": userID,
		"updates": updates,
	})
	
	// Create update request
	updateReq := &pb.UpdateUserReq{
		Id: userID,
	}
	
	// Map the updates to the protobuf fields
	if name, ok := updates["name"]; ok {
		updateReq.Name = fmt.Sprintf("%v", name)
	}
	if email, ok := updates["email"]; ok {
		updateReq.Email = fmt.Sprintf("%v", email)
	}
	if role, ok := updates["role"]; ok {
		updateReq.Role = fmt.Sprintf("%v", role)
	}
	if title, ok := updates["title"]; ok {
		updateReq.Title = fmt.Sprintf("%v", title)
	}
	if avatar, ok := updates["avatar"]; ok {
		updateReq.Avatar = fmt.Sprintf("%v", avatar)
	}
	if branchID, ok := updates["branch_id"]; ok {
		if id, parseErr := strconv.ParseInt(fmt.Sprintf("%v", branchID), 10, 64); parseErr == nil {
			updateReq.BranchId = id
		}
	}
	if ownerID, ok := updates["owner_id"]; ok {
		if id, parseErr := strconv.ParseInt(fmt.Sprintf("%v", ownerID), 10, 64); parseErr == nil {
			updateReq.OwnerId = id
		}
	}
	
	// Call gRPC service
	resp, err := h.userClient.UpdateUser(ctx, updateReq)
	if err != nil {
		return nil, h.handleGRPCError(err, operation, map[string]interface{}{
			"user_id": userID,
			"updates": updates,
		})
	}
	
	if resp == nil || resp.Account == nil {
		return nil, errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			"Empty response from update user service",
			nil,
		)
	}
	
	h.logger.Info("User updated successfully via gRPC", map[string]interface{}{
		"user_id": resp.Account.Id,
		"updates": h.getUpdatedFields(updates),
	})
	
	return resp.Account, nil
}

// checkEmailUniqueness checks if email is already in use
func (h *BaseAccountHandler) checkEmailUniqueness(ctx context.Context, email string) error {
	// Try to find user by email
	user, err := h.getUserByEmail(email)
	if err != nil {
		// If error is "not found", then email is unique
		if errorcustom.IsNotFoundError(err) {
			return nil
		}
		return err
	}
	
	if user != nil {
		// Use AccountDomainErrors instance to create proper duplicate error
		accountErrors := errorcustom.NewAccountDomainErrors()
		return accountErrors.NewDuplicateEmailError(email)
	}
	
	return nil
}


// resetPasswordViaGRPC initiates password reset and returns reset token
func (h *BaseAccountHandler) resetPasswordViaGRPC(ctx context.Context, email string) (string, error) {
	operation := "reset_password_grpc"
	
	h.logger.Debug("Initiating password reset via gRPC", map[string]interface{}{
		"email": email,
	})
	
	// First call ForgotPassword to get reset token
	forgotReq := &pb.ForgotPasswordReq{
		Email: email,
	}
	
	// Call gRPC service
	resp, err := h.userClient.ForgotPassword(ctx, forgotReq)
	if err != nil {
		return "", h.handleGRPCError(err, operation, map[string]interface{}{
			"email": email,
		})
	}
	
	if resp == nil {
		return "", errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			"Empty response from password reset service",
			nil,
		)
	}
	
	h.logger.Debug("Password reset initiated successfully via gRPC", map[string]interface{}{
		"email": email,
	})
	
	return resp.ResetToken, nil
}

// changePasswordViaGRPC changes user password
func (h *BaseAccountHandler) changePasswordViaGRPC(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	operation := "change_password_grpc"
	
	h.logger.Debug("Changing password via gRPC", map[string]interface{}{
		"user_id": userID,
	})
	
	// Validate password strength first
	if err := h.validatePasswordStrength(newPassword); err != nil {
		return err
	}
	
	// Create change password request
	req := &pb.ChangePasswordReq{
		UserId:          userID,
		CurrentPassword: oldPassword,
		NewPassword:     newPassword,
	}
	
	// Call gRPC service
	_, err := h.userClient.ChangePassword(ctx, req)
	if err != nil {
		return h.handleGRPCError(err, operation, map[string]interface{}{
			"user_id": userID,
		})
	}
	
	h.logger.Info("Password changed successfully via gRPC", map[string]interface{}{
		"user_id": userID,
	})
	
	// Log security event
	h.logSecurityEvent(
		"password_changed",
		"User password changed successfully",
		"medium",
		map[string]interface{}{
			"user_id": userID,
		},
	)
	
	return nil
}

// deleteUserViaGRPC deletes user account
func (h *BaseAccountHandler) deleteUserViaGRPC(ctx context.Context, userID int64) error {
	operation := "delete_user_grpc"
	
	h.logger.Debug("Deleting user via gRPC", map[string]interface{}{
		"user_id": userID,
	})
	
	// Create delete user request
	req := &pb.DeleteAccountReq{
		UserID: userID,
	}
	
	// Call gRPC service
	_, err := h.userClient.DeleteUser(ctx, req)
	if err != nil {
		return h.handleGRPCError(err, operation, map[string]interface{}{
			"user_id": userID,
		})
	}
	
	h.logger.Info("User deleted successfully via gRPC", map[string]interface{}{
		"user_id": userID,
	})
	
	return nil
}

// listUsersViaGRPC retrieves users with pagination and filters
func (h *BaseAccountHandler) listUsersViaGRPC(ctx context.Context, page, pageSize int32, filters map[string]interface{}) (*pb.SearchUsersRes, error) {
	operation := "list_users_grpc"
	
	h.logger.Debug("Listing users via gRPC", map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
		"filters":   filters,
	})
	
	// Build search request
	req := &pb.SearchUsersReq{
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
	}
	
	// Apply filters
	if role, ok := filters["role"]; ok && role != "" {
		req.Role = fmt.Sprintf("%v", role)
	}
	
	if branchID, ok := filters["branch_id"]; ok && branchID != "" {
		if id, parseErr := strconv.ParseInt(fmt.Sprintf("%v", branchID), 10, 64); parseErr == nil {
			req.BranchId = id
		}
	}
	
	if status, ok := filters["status"]; ok && status != "" {
		req.StatusFilter = []string{fmt.Sprintf("%v", status)}
	}
	
	// Call gRPC service
	resp, err := h.userClient.SearchUsers(ctx, req)
	if err != nil {
		return nil, h.handleGRPCError(err, operation, map[string]interface{}{
			"page":      page,
			"page_size": pageSize,
			"filters":   filters,
		})
	}
	
	if resp == nil {
		return nil, errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			"Empty response from search users service",
			nil,
		)
	}
	
	h.logger.Debug("Users listed successfully via gRPC", map[string]interface{}{
		"page":        page,
		"page_size":   pageSize,
		"total_count": resp.Total,
		"returned":    len(resp.Accounts),
	})
	
	return resp, nil
}

// bulkUpdateUsersViaGRPC performs bulk user updates
func (h *BaseAccountHandler) bulkUpdateUsersViaGRPC(ctx context.Context, updates map[int64]map[string]interface{}) error {
	operation := "bulk_update_users_grpc"
	
	h.logger.Debug("Bulk updating users via gRPC", map[string]interface{}{
		"update_count": len(updates),
	})
	
	// Since the existing protobuf doesn't have bulk update, we'll update users one by one
	successCount := 0
	errorCount := 0
	
	for userID, userUpdates := range updates {
		updateReq := &pb.UpdateUserReq{
			Id: userID,
		}
		
		// Map the updates to the protobuf fields
		if name, ok := userUpdates["name"]; ok {
			updateReq.Name = fmt.Sprintf("%v", name)
		}
		if email, ok := userUpdates["email"]; ok {
			updateReq.Email = fmt.Sprintf("%v", email)
		}
		if role, ok := userUpdates["role"]; ok {
			updateReq.Role = fmt.Sprintf("%v", role)
		}
		if title, ok := userUpdates["title"]; ok {
			updateReq.Title = fmt.Sprintf("%v", title)
		}
		if avatar, ok := userUpdates["avatar"]; ok {
			updateReq.Avatar = fmt.Sprintf("%v", avatar)
		}
		if branchID, ok := userUpdates["branch_id"]; ok {
			if id, parseErr := strconv.ParseInt(fmt.Sprintf("%v", branchID), 10, 64); parseErr == nil {
				updateReq.BranchId = id
			}
		}
		if ownerID, ok := userUpdates["owner_id"]; ok {
			if id, parseErr := strconv.ParseInt(fmt.Sprintf("%v", ownerID), 10, 64); parseErr == nil {
				updateReq.OwnerId = id
			}
		}
		
		// Call update for each user
		_, err := h.userClient.UpdateUser(ctx, updateReq)
		if err != nil {
			h.logger.Warning("Failed to update user in bulk operation", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			errorCount++
		} else {
			successCount++
		}
	}
	
	h.logger.Info("Bulk user update completed", map[string]interface{}{
		"total_requested": len(updates),
		"successful":      successCount,
		"failed":          errorCount,
	})
	
	// If all updates failed, return error
	if errorCount == len(updates) {
		return errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			"All bulk updates failed",
			nil,
		)
	}
	
	return nil
}

// checkGRPCHealth checks if gRPC service is healthy
func (h *BaseAccountHandler) checkGRPCHealth() error {
	operation := "grpc_health_check"
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Since there's no health check method in the proto, we'll use FindAllUsers as a health check
	_, err := h.userClient.FindAllUsers(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGRPCError(err, operation, map[string]interface{}{
			"service": "account_service",
		})
	}
	
	return nil
}

// ============================================================================
// CONVERSION HELPERS FOR PROTOBUF
// ============================================================================

// convertFiltersToProto converts filter map to protobuf format
func (h *BaseAccountHandler) convertFiltersToProto(filters map[string]interface{}) map[string]string {
	protoFilters := make(map[string]string)
	
	for key, value := range filters {
		if value != nil && fmt.Sprintf("%v", value) != "" {
			protoFilters[key] = fmt.Sprintf("%v", value)
		}
	}
	
	return protoFilters
}

// convertUpdatesToProto converts updates map to protobuf format
func (h *BaseAccountHandler) convertUpdatesToProto(updates map[string]interface{}) map[string]string {
	protoUpdates := make(map[string]string)
	
	for key, value := range updates {
		if value != nil && fmt.Sprintf("%v", value) != "" {
			protoUpdates[key] = fmt.Sprintf("%v", value)
		}
	}
	
	return protoUpdates
}

// convertAccountStatusToProto converts string status to protobuf enum
func (h *BaseAccountHandler) convertAccountStatusToProto(status string) pb.AccountStatus {
	switch status {
	case "active":
		return pb.AccountStatus_ACTIVE
	case "inactive":
		return pb.AccountStatus_INACTIVE
	case "suspended":
		return pb.AccountStatus_SUSPENDED
	default:
		return pb.AccountStatus_UNKNOWN
	}
}

// convertProtoStatusToString converts protobuf enum to string
func (h *BaseAccountHandler) convertProtoStatusToString(status pb.AccountStatus) string {
	switch status {
	case pb.AccountStatus_ACTIVE:
		return "active"
	case pb.AccountStatus_INACTIVE:
		return "inactive"
	case pb.AccountStatus_SUSPENDED:
		return "suspended"
	default:
		return "unknown"
	}
}

// ============================================================================
// VALIDATION HELPERS
// ============================================================================

// validatePasswordStrength validates password meets security requirements
func (h *BaseAccountHandler) validatePasswordStrength(password string) error {
	policy := h.getPasswordPolicy()
	
	errorCollection := errorcustom.NewErrorCollection(h.domain)
	
	// Check minimum length
	if minLength, ok := policy["min_length"].(int); ok {
		if len(password) < minLength {
			errorCollection.Add(errorcustom.NewValidationError(
				h.domain,
				"password",
				fmt.Sprintf("Password must be at least %d characters long", minLength),
				"[MASKED]",
			))
		}
	}
	
	// Check for uppercase letters
	if requireUpper, ok := policy["require_upper"].(bool); ok && requireUpper {
		if !utils.ContainsUppercase(password) {
			errorCollection.Add(errorcustom.NewValidationError(
				h.domain,
				"password",
				"Password must contain at least one uppercase letter",
				"[MASKED]",
			))
		}
	}
	
	// Check for lowercase letters
	if requireLower, ok := policy["require_lower"].(bool); ok && requireLower {
		if !utils.ContainsLowercase(password) {
			errorCollection.Add(errorcustom.NewValidationError(
				h.domain,
				"password",
				"Password must contain at least one lowercase letter",
				"[MASKED]",
			))
		}
	}
	
	// Check for numbers
	if requireNumbers, ok := policy["require_numbers"].(bool); ok && requireNumbers {
		if !utils.ContainsNumbers(password) {
			errorCollection.Add(errorcustom.NewValidationError(
				h.domain,
				"password",
				"Password must contain at least one number",
				"[MASKED]",
			))
		}
	}
	
	// Check for special characters
	if requireSpecial, ok := policy["require_special"].(bool); ok && requireSpecial {
		if !utils.ContainsSpecialChars(password) {
			errorCollection.Add(errorcustom.NewValidationError(
				h.domain,
				"password",
				"Password must contain at least one special character",
				"[MASKED]",
			))
		}
	}
	
	if errorCollection.HasErrors() {
		return errorCollection.ToAPIError()
	}
	
	return nil
}

// ============================================================================
// ENHANCED ERROR HANDLING
// ============================================================================

// handleGRPCError converts gRPC errors to custom errors with context
func (h *BaseAccountHandler) handleGRPCError(err error, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	
	// Get gRPC status
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			"Unknown gRPC error",
			err,
		)
	}
	
	// Map gRPC codes to custom errors
	switch grpcStatus.Code() {
case codes.InvalidArgument:
    return errorcustom.NewValidationErrorWithContext(
        h.domain,
        "grpc_request",
        grpcStatus.Message(),
        nil,      // 👈 placeholder for invalid value
        context,  // 👈 now in correct position
    )
	case codes.Unauthenticated:
		return errorcustom.NewAuthenticationErrorWithContext(
			h.domain,
			grpcStatus.Message(),
			context,
		)
		
	case codes.PermissionDenied:
		return errorcustom.NewAuthorizationErrorWithContext(
			h.domain,
			"grpc_permission",
			operation,
			context,
		)
		
	case codes.NotFound:
		return errorcustom.NewNotFoundErrorWithContext(
			h.domain,
			"resource",
			context,
		)
		
	case codes.AlreadyExists:
		return errorcustom.NewConflictErrorWithContext(
			h.domain,
			"resource_conflict",
			grpcStatus.Message(),
			context,
		)
		
case codes.ResourceExhausted:
    return errorcustom.NewRateLimitErrorWithContext(
        h.domain,
        operation,
        grpcStatus.Message(),
        context,
    )
		
	case codes.FailedPrecondition:
		return errorcustom.NewBusinessLogicErrorWithContext(
			h.domain,
			"precondition_failed",
			grpcStatus.Message(),
			context,
		)
		
	case codes.Unavailable, codes.DeadlineExceeded:
		return errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			fmt.Sprintf("Service unavailable: %s", grpcStatus.Message()),
			err,
		)
		
	default:
		return errorcustom.NewSystemError(
			h.domain,
			"grpc_client",
			operation,
			fmt.Sprintf("gRPC error (%s): %s", grpcStatus.Code().String(), grpcStatus.Message()),
			err,
		)
	}
}
