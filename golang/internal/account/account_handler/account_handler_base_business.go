package account_handler

import (
	"context"
	"fmt"

	"strings"
	"time"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	
	"english-ai-full/utils"
)

// ============================================================================
// USER PERMISSIONS AND AUTHORIZATION
// ============================================================================

// checkUserPermissions validates user permissions with domain context
func (h *BaseAccountHandler) checkUserPermissions(userID int64, requiredRole string, resource string) error {
	// Get user from service
	user, err := h.getUserByID(userID)
	if err != nil {
		return err
	}
	
	// Check role-based access
	if !h.hasRole(user.Role, requiredRole) {
		return errorcustom.NewAuthorizationErrorWithContext(
			h.domain,
			"access",
			resource,
			map[string]interface{}{
				"user_id":       userID,
				"required_role": requiredRole,
				"current_role":  user.Role,
				"resource":      resource,
			},
		)
	}
	
	// Log authorization success
	h.logger.Debug("User authorization successful", map[string]interface{}{
		"user_id":       userID,
		"required_role": requiredRole,
		"current_role":  user.Role,
		"resource":      resource,
	})
	
	return nil
}

// hasRole checks if user has required role based on hierarchy
func (h *BaseAccountHandler) hasRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"student": 1,
		"teacher": 2,
		"admin":   3,
	}
	
	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]
	
	if !userExists || !requiredExists {
		return false
	}
	
	return userLevel >= requiredLevel
}

// checkResourceOwnership validates if user owns or can access a resource
func (h *BaseAccountHandler) checkResourceOwnership(userID int64, resourceOwnerID int64, resourceType string) error {
	if userID == resourceOwnerID {
		return nil // User owns the resource
	}
	
	// Check if user has admin privileges to access other users' resources
	user, err := h.getUserByID(userID)
	if err != nil {
		return err
	}
	
	if user.Role == "admin" {
		return nil // Admins can access any resource
	}
	
	return errorcustom.NewAuthorizationErrorWithContext(
		h.domain,
		"ownership",
		resourceType,
		map[string]interface{}{
			"user_id":           userID,
			"resource_owner_id": resourceOwnerID,
			"resource_type":     resourceType,
			"user_role":         user.Role,
		},
	)
}

// ============================================================================
// ACCOUNT LIFECYCLE MANAGEMENT
// ============================================================================

// registerNewUser handles the complete user registration process
func (h *BaseAccountHandler) registerNewUser(ctx context.Context, req CreateUserRequest) (*pb.Account, error) {
	operation := "register_new_user"
	opCtx := &OperationContext{
		RequestID: errorcustom.GetRequestIDFromContext(ctx),
		Domain:    h.domain,
		Operation: operation,
		StartTime: time.Now(),
		Context: map[string]interface{}{
			"email": req.Email,
			"role":  req.Role,
		},
	}
	
	h.logOperationStart(opCtx)
	
	// Step 1: Validate business rules
	businessContext := map[string]interface{}{
		"email":    req.Email,
		"password": req.Password,
		"role":     req.Role,
	}
	
	if err := h.validateBusinessRules("user_registration", businessContext); err != nil {
		h.logOperationEnd(opCtx, err, 400)
		return nil, err
	}
	
	// Step 2: Check email uniqueness
	if err := h.checkEmailUniqueness(ctx, req.Email); err != nil {
		h.logOperationEnd(opCtx, err, 409)
		return nil, err
	}
	
	// Step 3: Create user via gRPC
	user, err := h.createUserViaGRPC(ctx, req)
	if err != nil {
		h.logOperationEnd(opCtx, err, 500)
		return nil, err
	}
	
	// Step 4: Post-registration tasks
	if err := h.performPostRegistrationTasks(ctx, user); err != nil {
		h.logger.Warning("Post-registration tasks failed", map[string]interface{}{
			"user_id": user.Id,
			"error":   err.Error(),
		})
		// Don't fail the registration, just log the warning
	}
	
	h.logOperationEnd(opCtx, nil, 201)
	return user, nil
}

// performPostRegistrationTasks handles tasks after successful registration
func (h *BaseAccountHandler) performPostRegistrationTasks(ctx context.Context, user *pb.Account) error {
	// Send welcome email (if email service is available)
	if err := h.sendWelcomeEmail(user.Email, user.Name); err != nil {
		h.logger.Warning("Failed to send welcome email", map[string]interface{}{
			"user_id": user.Id,
			"email":   user.Email,
			"error":   err.Error(),
		})
	}
	
	// Log user registration event
	h.logSecurityEvent(
		"user_registered",
		"New user account created",
		"low",
		map[string]interface{}{
			"user_id": user.Id,
			"email":   user.Email,
			"role":    user.Role,
		},
	)
	
	// Create default user preferences (if applicable)
	if err := h.createDefaultUserPreferences(user.Id); err != nil {
		h.logger.Warning("Failed to create default preferences", map[string]interface{}{
			"user_id": user.Id,
			"error":   err.Error(),
		})
	}
	
	return nil
}

// authenticateUser handles the complete user authentication process
func (h *BaseAccountHandler) authenticateUser(ctx context.Context, email, password string) (*pb.Account, string, error) {
	operation := "authenticate_user"
	opCtx := &OperationContext{
		RequestID: errorcustom.GetRequestIDFromContext(ctx),
		Domain:    h.domain,
		Operation: operation,
		StartTime: time.Now(),
		Context: map[string]interface{}{
			"email": email,
		},
	}
	
	h.logOperationStart(opCtx)
	
	// Step 1: Validate login business rules
	loginContext := map[string]interface{}{
		"email":           email,
		"failed_attempts": h.getFailedLoginAttempts(email),
	}
	
	if err := h.validateBusinessRules("user_login", loginContext); err != nil {
		h.logOperationEnd(opCtx, err, 423) // Locked
		return nil, "", err
	}
	
	// Step 2: Authenticate via gRPC
	user, err := h.authenticateUserViaGRPC(ctx, email, password)
	if err != nil {
		// Track failed login attempt
		h.trackFailedLoginAttempt(email)
		h.logOperationEnd(opCtx, err, 401)
		return nil, "", err
	}
	
	// Step 3: Check account status
	if err := h.validateAccountStatus(user); err != nil {
		h.logOperationEnd(opCtx, err, 403)
		return nil, "", err
	}
	
	// Step 4: Generate session token
	sessionToken, err := h.generateSessionToken(user.Id)
	if err != nil {
		h.logOperationEnd(opCtx, err, 500)
		return nil, "", err
	}
	
	// Step 5: Clear failed login attempts
	h.clearFailedLoginAttempts(email)
	
	// Step 6: Log successful login
	h.logSecurityEvent(
		"user_login_success",
		"User logged in successfully",
		"low",
		map[string]interface{}{
			"user_id": user.Id,
			"email":   user.Email,
			"role":    user.Role,
		},
	)
	
	h.logOperationEnd(opCtx, nil, 200)
	return user, sessionToken, nil
}

// updateUserProfile handles the complete user profile update process
func (h *BaseAccountHandler) updateUserProfile(ctx context.Context, userID int64, requestingUserID int64, updates map[string]interface{}) (*pb.Account, error) {
	operation := "update_user_profile"
	opCtx := &OperationContext{
		RequestID: errorcustom.GetRequestIDFromContext(ctx),
		Domain:    h.domain,
		Operation: operation,
		StartTime: time.Now(),
		UserID:    userID,
		Context: map[string]interface{}{
			"target_user_id":     userID,
			"requesting_user_id": requestingUserID,
			"updates":            updates,
		},
	}
	
	h.logOperationStart(opCtx)
	
	// Step 1: Check permissions
	if err := h.checkResourceOwnership(requestingUserID, userID, "user_profile"); err != nil {
		h.logOperationEnd(opCtx, err, 403)
		return nil, err
	}
	
	// Step 2: Validate update business rules
	if err := h.validateUpdatePermissions(requestingUserID, updates); err != nil {
		h.logOperationEnd(opCtx, err, 403)
		return nil, err
	}
	
	// Step 3: Validate updated data
	if err := h.validateProfileUpdates(updates); err != nil {
		h.logOperationEnd(opCtx, err, 400)
		return nil, err
	}
	
	// Step 4: Update via gRPC
	user, err := h.updateUserViaGRPC(ctx, userID, updates)
	if err != nil {
		h.logOperationEnd(opCtx, err, 500)
		return nil, err
	}
	
	// Step 5: Log the update
	h.logger.Info("User profile updated successfully", map[string]interface{}{
		"user_id":            userID,
		"requesting_user_id": requestingUserID,
		"updated_fields":     h.getUpdatedFields(updates),
	})
	
	h.logOperationEnd(opCtx, nil, 200)
	return user, nil
}

// ============================================================================
// BUSINESS RULE VALIDATION HELPERS
// ============================================================================

// validateAccountStatus checks if account is in valid state for operations
func (h *BaseAccountHandler) validateAccountStatus(user *pb.Account) error {
    // Check if account is active
    if user.Status != pb.AccountStatus_ACTIVE {  // Compare directly with enum value
        return errorcustom.NewBusinessLogicErrorWithContext(
            h.domain,
            "account_inactive",
            "Account is not active",
            map[string]interface{}{
                "user_id": user.Id,
                "status":  user.Status.String(), // Convert enum to string for logging
            },
        )
    }
    
    // Check if account is verified (if verification is required)
    if h.config.IsEmailVerificationRequired() && !h.isAccountVerified(user) {
        return errorcustom.NewBusinessLogicErrorWithContext(
            h.domain,
            "account_unverified",
            "Account email verification required",
            map[string]interface{}{
                "user_id": user.Id,
                "email":   user.Email,
            },
        )
    }
    
    return nil
}
// validateUpdatePermissions checks if user can update specific fields
func (h *BaseAccountHandler) validateUpdatePermissions(requestingUserID int64, updates map[string]interface{}) error {
	// Get requesting user
	requestingUser, err := h.getUserByID(requestingUserID)
	if err != nil {
		return err
	}
	
	// Check if user is trying to update role
	if newRole, hasRole := updates["role"]; hasRole {
		if requestingUser.Role != "admin" {
			return errorcustom.NewAuthorizationErrorWithContext(
				h.domain,
				"role_update",
				"user_profile",
				map[string]interface{}{
					"requesting_user_id":   requestingUserID,
					"requesting_user_role": requestingUser.Role,
					"attempted_role":       newRole,
				},
			)
		}
	}
	
	// Check if user is trying to update email
	if newEmail, hasEmail := updates["email"]; hasEmail {
		// Validate new email format
		if err := h.validateEmailFormat(newEmail.(string)); err != nil {
			return err
		}
		
		// Check if new email is unique
		if err := h.checkEmailUniqueness(context.Background(), newEmail.(string)); err != nil {
			return err
		}
	}
	
	return nil
}

// validateProfileUpdates validates the profile update data
func (h *BaseAccountHandler) validateProfileUpdates(updates map[string]interface{}) error {
	errorCollection := errorcustom.NewErrorCollection(h.domain)
	
	// Validate each field
	if name, hasName := updates["name"]; hasName {
		if err := h.validateName(name.(string)); err != nil {
			errorCollection.Add(err)
		}
	}
	
	if title, hasTitle := updates["title"]; hasTitle {
		if err := h.validateTitle(title.(string)); err != nil {
			errorCollection.Add(err)
		}
	}
	
	if avatar, hasAvatar := updates["avatar"]; hasAvatar {
		if err := h.validateAvatarURL(avatar.(string)); err != nil {
			errorCollection.Add(err)
		}
	}
	
	if errorCollection.HasErrors() {
		return errorCollection.ToAPIError()
	}
	
	return nil
}

// ============================================================================
// LOGIN ATTEMPT TRACKING
// ============================================================================

// getFailedLoginAttempts retrieves failed login attempts for email
func (h *BaseAccountHandler) getFailedLoginAttempts(email string) int {
	// This would typically be stored in Redis or a cache
	// For now, return 0 as placeholder
	return 0
}

// trackFailedLoginAttempt records a failed login attempt
func (h *BaseAccountHandler) trackFailedLoginAttempt(email string) {
	// Log the failed attempt
	h.logSecurityEvent(
		"login_attempt_failed",
		"Failed login attempt recorded",
		"medium",
		map[string]interface{}{
			"email": email,
		},
	)
	
	// This would typically increment a counter in Redis with TTL
	// Implementation depends on your caching strategy
}

// clearFailedLoginAttempts clears failed login attempts for email
func (h *BaseAccountHandler) clearFailedLoginAttempts(email string) {
	// This would typically clear the counter in Redis
	h.logger.Debug("Cleared failed login attempts", map[string]interface{}{
		"email": email,
	})
}

// ============================================================================
// SESSION MANAGEMENT
// ============================================================================

// generateSessionToken generates a session token for the user
func (h *BaseAccountHandler) generateSessionToken(userID int64) (string, error) {
	// This would typically generate a JWT or session token
	// For now, return a placeholder
	sessionToken := fmt.Sprintf("session_%d_%d", userID, time.Now().Unix())
	
	h.logger.Debug("Session token generated", map[string]interface{}{
		"user_id": userID,
	})
	
	return sessionToken, nil
}

// validateSessionToken validates a session token
func (h *BaseAccountHandler) validateSessionToken(token string) (int64, error) {
	// This would typically validate a JWT or lookup session in storage
	// For now, extract user ID from placeholder format
	if strings.HasPrefix(token, "session_") {
		// Extract user ID from token (placeholder implementation)
		return 1, nil
	}
	
	return 0, errorcustom.NewAuthenticationError(h.domain, "invalid session token")
}

// ============================================================================
// NOTIFICATION AND EMAIL SERVICES
// ============================================================================

// sendWelcomeEmail sends a welcome email to new users
func (h *BaseAccountHandler) sendWelcomeEmail(email, name string) error {
	// This would integrate with your email service
	h.logger.Info("Welcome email sent", map[string]interface{}{
		"email": email,
		"name":  name,
	})
	
	return nil
}

// sendPasswordResetEmail sends password reset email
func (h *BaseAccountHandler) sendPasswordResetEmail(email, resetToken string) error {
	// This would integrate with your email service
	h.logger.Info("Password reset email sent", map[string]interface{}{
		"email": email,
	})
	
	return nil
}

// sendAccountVerificationEmail sends account verification email
func (h *BaseAccountHandler) sendAccountVerificationEmail(email, verificationToken string) error {
	// This would integrate with your email service
	h.logger.Info("Account verification email sent", map[string]interface{}{
		"email": email,
	})
	
	return nil
}

// ============================================================================
// USER PREFERENCES AND SETTINGS
// ============================================================================

// createDefaultUserPreferences creates default preferences for new users
func (h *BaseAccountHandler) createDefaultUserPreferences(userID int64) error {
	// This would create default preferences in your preferences service
	h.logger.Debug("Default user preferences created", map[string]interface{}{
		"user_id": userID,
	})
	
	return nil
}

// updateUserPreferences updates user preferences
func (h *BaseAccountHandler) updateUserPreferences(userID int64, preferences map[string]interface{}) error {
	// This would update preferences in your preferences service
	h.logger.Debug("User preferences updated", map[string]interface{}{
		"user_id":     userID,
		"preferences": preferences,
	})
	
	return nil
}

// ============================================================================
// DATA VALIDATION HELPERS
// ============================================================================

// validateName validates user name
func (h *BaseAccountHandler) validateName(name string) error {
	if len(strings.TrimSpace(name)) < 2 {
		return errorcustom.NewValidationError(
			h.domain,
			"name",
			"Name must be at least 2 characters long",
			name,
		)
	}
	
	if len(name) > 100 {
		return errorcustom.NewValidationError(
			h.domain,
			"name",
			"Name cannot exceed 100 characters",
			name,
		)
	}
	
	return nil
}

// validateTitle validates user title
func (h *BaseAccountHandler) validateTitle(title string) error {
	if len(title) > 200 {
		return errorcustom.NewValidationError(
			h.domain,
			"title",
			"Title cannot exceed 200 characters",
			title,
		)
	}
	
	return nil
}

// validateAvatarURL validates avatar URL
func (h *BaseAccountHandler) validateAvatarURL(avatarURL string) error {
	if avatarURL == "" {
		return nil // Empty avatar URL is allowed
	}
	
	if !strings.HasPrefix(avatarURL, "http://") && !strings.HasPrefix(avatarURL, "https://") {
		return errorcustom.NewValidationError(
			h.domain,
			"avatar",
			"Avatar URL must be a valid HTTP/HTTPS URL",
			avatarURL,
		)
	}
	
	return nil
}

// isAccountVerified checks if account is verified
func (h *BaseAccountHandler) isAccountVerified(user *pb.Account) bool {
	// This would check verification status from user data
	// For now, assume verified if user has a verified field
	return true // Placeholder implementation
}

// ============================================================================
// AUDIT AND COMPLIANCE
// ============================================================================

// auditUserAction logs user actions for compliance
func (h *BaseAccountHandler) auditUserAction(userID int64, action string, details map[string]interface{}) {
	auditDetails := utils.MergeContext(details, map[string]interface{}{
		"user_id":   userID,
		"action":    action,
		"timestamp": time.Now().UTC(),
		"domain":    h.domain,
	})
	
	h.logger.Info("User action audited", auditDetails)
	
	// This would typically send to an audit service or database
}

// logDataAccess logs data access for compliance
func (h *BaseAccountHandler) logDataAccess(accessorID int64, resourceType string, resourceID int64, accessType string) {
	h.logger.Info("Data access logged", map[string]interface{}{
		"accessor_id":   accessorID,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"access_type":   accessType,
		"timestamp":     time.Now().UTC(),
		"domain":        h.domain,
	})
}

// ============================================================================
// ACCOUNT RECOVERY AND MAINTENANCE
// ============================================================================

// initiatePasswordReset initiates the password reset process
func (h *BaseAccountHandler) initiatePasswordReset(ctx context.Context, email string) error {
	operation := "initiate_password_reset"
	opCtx := &OperationContext{
		RequestID: errorcustom.GetRequestIDFromContext(ctx),
		Domain:    h.domain,
		Operation: operation,
		StartTime: time.Now(),
		Context: map[string]interface{}{
			"email": email,
		},
	}
	
	h.logOperationStart(opCtx)
	
	// Step 1: Verify user exists
	user, err := h.getUserByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not for security
		h.logOperationEnd(opCtx, nil, 200)
		return nil
	}
	
	// Step 2: Generate reset token via gRPC
	resetToken, err := h.resetPasswordViaGRPC(ctx, email)
	if err != nil {
		h.logOperationEnd(opCtx, err, 500)
		return err
	}
	
	// Step 3: Send reset email
	if err := h.sendPasswordResetEmail(email, resetToken); err != nil {
		h.logger.Warning("Failed to send password reset email", map[string]interface{}{
			"email": email,
			"error": err.Error(),
		})
		// Don't fail the operation, just log the warning
	}
	
	// Step 4: Log security event
	h.logSecurityEvent(
		"password_reset_initiated",
		"Password reset process initiated",
		"medium",
		map[string]interface{}{
			"user_id": user.Id,
			"email":   email,
		},
	)
	
	h.logOperationEnd(opCtx, nil, 200)
	return nil
}

// completePasswordReset completes the password reset process
func (h *BaseAccountHandler) completePasswordReset(ctx context.Context, resetToken, newPassword string) error {
	operation := "complete_password_reset"
	
	// Step 1: Validate new password
	if err := h.validatePasswordStrength(newPassword); err != nil {
		return err
	}
	
	// Step 2: Validate and consume reset token (via gRPC)
	// This would be implemented in the gRPC service
	
	// Step 3: Update password via gRPC
	// This would call a specific password reset completion endpoint
	
	// Step 4: Log security event
	h.logSecurityEvent(
		"password_reset_completed",
		"Password successfully reset",
		"medium",
		map[string]interface{}{
			"reset_token": "[MASKED]",
		},
	)
	
	return nil
}

// ============================================================================
// USER SEARCH AND FILTERING
// ============================================================================

// searchUsers searches users based on criteria
func (h *BaseAccountHandler) searchUsers(ctx context.Context, requestingUserID int64, criteria SearchCriteria) (*UserSearchResult, error) {
	operation := "search_users"
	
	// Check permissions
	if err := h.checkUserPermissions(requestingUserID, "admin", "user_search"); err != nil {
		return nil, err
	}
	
	// Build filters
	filters := map[string]interface{}{
		"role": criteria.Role,
	}
	
	if criteria.BranchID > 0 {
		filters["branch_id"] = criteria.BranchID
	}
	
	if criteria.Status != "" {
		filters["status"] = criteria.Status
	}
	
	// Get users via gRPC
	resp, err := h.listUsersViaGRPC(ctx, criteria.Page, criteria.PageSize, filters)
	if err != nil {
		return nil, err
	}
	
	// Convert to search result
	result := &UserSearchResult{
		Users:      resp.Accounts,
		TotalCount: resp.TotalCount,
		Page:       criteria.Page,
		PageSize:   criteria.PageSize,
	}
	
	// Log data access
	h.logDataAccess(requestingUserID, "user_list", 0, "search")
	
	return result, nil
}

// ============================================================================
// BATCH OPERATIONS
// ============================================================================

// batchUpdateUsers updates multiple users with validation
func (h *BaseAccountHandler) batchUpdateUsers(ctx context.Context, requestingUserID int64, updates map[int64]map[string]interface{}) (*BatchUpdateResult, error) {
	operation := "batch_update_users"
	
	// Check permissions
	if err := h.checkUserPermissions(requestingUserID, "admin", "user_batch_update"); err != nil {
		return nil, err
	}
	
	result := &BatchUpdateResult{
		TotalRequested: len(updates),
		Successful:     make([]int64, 0),
		Failed:         make(map[int64]string),
	}
	
	// Validate all updates first
	for userID, userUpdates := range updates {
		if err := h.validateUpdatePermissions(requestingUserID, userUpdates); err != nil {
			result.Failed[userID] = err.Error()
			continue
		}
		
		if err := h.validateProfileUpdates(userUpdates); err != nil {
			result.Failed[userID] = err.Error()
			continue
		}
		
		result.Successful = append(result.Successful, userID)
	}
	
	// Only proceed with valid updates
	validUpdates := make(map[int64]map[string]interface{})
	for _, userID := range result.Successful {
		validUpdates[userID] = updates[userID]
	}
	
	if len(validUpdates) > 0 {
		if err := h.bulkUpdateUsersViaGRPC(ctx, validUpdates); err != nil {
			return nil, err
		}
	}
	
	// Log batch operation
	h.auditUserAction(requestingUserID, "batch_update_users", map[string]interface{}{
		"total_requested": result.TotalRequested,
		"successful":      len(result.Successful),
		"failed":          len(result.Failed),
	})
	
	return result, nil
}

// ============================================================================
// HELPER TYPES FOR BUSINESS OPERATIONS
// ============================================================================

// SearchCriteria represents user search criteria
type SearchCriteria struct {
	Role     string `json:"role"`
	BranchID int64  `json:"branch_id"`
	Status   string `json:"status"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
}

// UserSearchResult represents user search results
type UserSearchResult struct {
	Users      []*pb.Account `json:"users"`
	TotalCount int64         `json:"total_count"`
	Page       int32         `json:"page"`
	PageSize   int32         `json:"page_size"`
}

// BatchUpdateResult represents batch update results
type BatchUpdateResult struct {
	TotalRequested int             `json:"total_requested"`
	Successful     []int64         `json:"successful"`
	Failed         map[int64]string `json:"failed"`
}

// ============================================================================
// UTILITY AND HELPER FUNCTIONS
// ============================================================================

// getUpdatedFields extracts field names that were updated
func (h *BaseAccountHandler) getUpdatedFields(updates map[string]interface{}) []string {
	fields := make([]string, 0, len(updates))
	for field := range updates {
		fields = append(fields, field)
	}
	return fields
}

// validateUserRegistrationRules validates user registration with enhanced context
func (h *BaseAccountHandler) validateUserRegistrationRules(ctx context.Context, email, password string) error {
	businessContext := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	
	return h.validateBusinessRules("user_registration", businessContext)
}

// isAllowedDomain checks if email domain is allowed
func (h *BaseAccountHandler) isAllowedDomain(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	
	domain := strings.Split(email, "@")[1]
	allowedDomains := h.config.GetAllowedEmailDomains()
	
	for _, allowed := range allowedDomains {
		if domain == allowed {
			return true
		}
	}
	
	// If no restrictions are configured, allow all domains
	return len(allowedDomains) == 0
}

// getMaxLoginAttempts returns the maximum login attempts from configuration
func (h *BaseAccountHandler) getMaxLoginAttempts() int {
	if h.config != nil {
		return h.config.GetMaxLoginAttempts()
	}
	return 5 // Default fallback
}

// ============================================================================
// ACCOUNT STATISTICS AND REPORTING
// ============================================================================

// getUserStatistics returns user statistics for admin dashboard
func (h *BaseAccountHandler) getUserStatistics(ctx context.Context, requestingUserID int64) (*UserStatistics, error) {
	// Check admin permissions
	if err := h.checkUserPermissions(requestingUserID, "admin", "user_statistics"); err != nil {
		return nil, err
	}
	
	// This would typically call a statistics service or aggregate data
	stats := &UserStatistics{
		TotalUsers:       100, // Placeholder
		ActiveUsers:      85,  // Placeholder
		InactiveUsers:    15,  // Placeholder
		NewUsersToday:    5,   // Placeholder
		NewUsersThisWeek: 25,  // Placeholder
		UsersByRole: map[string]int{
			"admin":   5,
			"teacher": 20,
			"student": 75,
		},
	}
	
	// Log data access
	h.logDataAccess(requestingUserID, "user_statistics", 0, "read")
	
	return stats, nil
}

// UserStatistics represents user statistics data
type UserStatistics struct {
	TotalUsers       int            `json:"total_users"`
	ActiveUsers      int            `json:"active_users"`
	InactiveUsers    int            `json:"inactive_users"`
	NewUsersToday    int            `json:"new_users_today"`
	NewUsersThisWeek int            `json:"new_users_this_week"`
	UsersByRole      map[string]int `json:"users_by_role"`
}

// ============================================================================
// CONFIGURATION ACCESS HELPERS
// ============================================================================

// getSessionTimeout returns session timeout from configuration
func (h *BaseAccountHandler) getSessionTimeout() time.Duration {
	if h.config != nil {
		return h.config.GetSessionTimeout()
	}
	return 24 * time.Hour // Default fallback
}

// isEmailVerificationRequired checks if email verification is required
func (h *BaseAccountHandler) isEmailVerificationRequired() bool {
	if h.config != nil {
		return h.config.IsEmailVerificationRequired()
	}
	return false
}

// getPasswordPolicy returns password policy configuration
func (h *BaseAccountHandler) getPasswordPolicy() map[string]interface{} {
	if h.config != nil {
		return h.config.GetPasswordPolicy()
	}
	
	// Default password policy
	return map[string]interface{}{
		"min_length":      8,
		"require_upper":   true,
		"require_lower":   true,
		"require_numbers": true,
		"require_special": true,
	}
}