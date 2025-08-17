package account_handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
)

// ============================================================================
// CONSTANTS FOR GRPC OPERATIONS
// ============================================================================

const (
	DefaultGRPCTimeout       = 5 * time.Second
	CreateUpdateGRPCTimeout  = 10 * time.Second
	MaxRetryAttempts         = 3
	BaseRetryDelay          = 100 * time.Millisecond
)

// ============================================================================
// GRPC ERROR HANDLING
// ============================================================================

// handleGRPCError converts gRPC errors to domain-aware API errors
func (h *BaseAccountHandler) handleGRPCError(err error, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	
	// Use the enhanced gRPC error parser with domain context
	domainErr := errorcustom.ParseGRPCError(err, h.domain, operation, context)
	
	// Apply domain-specific error handling through configuration
	processedErr := h.errorHandler.HandleError(h.domain, domainErr)
	
	return processedErr
}

// HandleGRPCError processes gRPC errors using the unified error handler
func (h *BaseAccountHandler) HandleGRPCError(err error, operation string, context map[string]interface{}) error {
	return h.errorHandler.ParseGRPCError(err, h.domain, operation, context)
}

// ============================================================================
// USER OPERATIONS VIA GRPC
// ============================================================================

// getUserByID fetches user with domain-aware error handling
func (h *BaseAccountHandler) getUserByID(userID int64) (*pb.Account, error) {
	operation := "get_user_by_id"
	
	return h.measureGRPCOperation(operation, func() (*pb.Account, error) {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultGRPCTimeout)
		defer cancel()
		
		req := &pb.FindByIDReq{Id: userID}
		resp, err := h.userClient.FindByID(ctx, req)
		
		if err != nil {
			return nil, h.handleGRPCError(err, operation, map[string]interface{}{
				"user_id": userID,
			})
		}
		
		// Create sanitized user response (exclude sensitive data)
		user := &pb.Account{
			Id:        resp.Account.Id,
			BranchId:  resp.Account.BranchId,
			Name:      resp.Account.Name,
			Email:     resp.Account.Email,
			// Password:  "", // Exclude password for security
			Avatar:    resp.Account.Avatar,
			Title:     resp.Account.Title,
			Role:      resp.Account.Role,
			OwnerId:   resp.Account.OwnerId,
			CreatedAt: resp.Account.CreatedAt,
			UpdatedAt: resp.Account.UpdatedAt,
		}
		
		return user, nil
	})
}

// getUserByEmail fetches user by email with domain-aware error handling
func (h *BaseAccountHandler) getUserByEmail(email string) (*pb.Account, error) {
	operation := "get_user_by_email"
	
	return h.measureGRPCOperation(operation, func() (*pb.Account, error) {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultGRPCTimeout)
		defer cancel()
		
		req := &pb.FindByEmailReq{Email: email}
		resp, err := h.userClient.FindByEmail(ctx, req)
		
		if err != nil {
			return nil, h.handleGRPCError(err, operation, map[string]interface{}{
				"email": email,
			})
		}
		
		// Return sanitized account
		return h.sanitizeAccount(resp.Account), nil
	})
}

// createUserViaGRPC creates a new user via gRPC
func (h *BaseAccountHandler) createUserViaGRPC(ctx context.Context, req CreateUserRequest) (*pb.Account, error) {
	operation := "create_user"
	
	return h.measureGRPCOperation(operation, func() (*pb.Account, error) {
		grpcCtx, cancel := context.WithTimeout(ctx, CreateUpdateGRPCTimeout)
		defer cancel()
		
		grpcReq := &pb.CreateReq{
			Name:     req.Name,
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
		}
		
		resp, err := h.userClient.Create(grpcCtx, grpcReq)
		if err != nil {
			return nil, h.handleGRPCError(err, operation, map[string]interface{}{
				"email": req.Email,
				"role":  req.Role,
			})
		}
		
		return h.sanitizeAccount(resp.Account), nil
	})
}

// updateUserViaGRPC updates user information via gRPC
func (h *BaseAccountHandler) updateUserViaGRPC(ctx context.Context, userID int64, updates map[string]interface{}) (*pb.Account, error) {
	operation := "update_user"
	
	return h.measureGRPCOperation(operation, func() (*pb.Account, error) {
		grpcCtx, cancel := context.WithTimeout(ctx, CreateUpdateGRPCTimeout)
		defer cancel()
		
		// Build update request based on provided fields
		updateReq := &pb.UpdateReq{
			Id: userID,
		}
		
		// Map updates to gRPC request fields
		if name, ok := updates["name"].(string); ok {
			updateReq.Name = name
		}
		if email, ok := updates["email"].(string); ok {
			updateReq.Email = email
		}
		if role, ok := updates["role"].(string); ok {
			updateReq.Role = role
		}
		if avatar, ok := updates["avatar"].(string); ok {
			updateReq.Avatar = avatar
		}
		if title, ok := updates["title"].(string); ok {
			updateReq.Title = title
		}
		
		resp, err := h.userClient.Update(grpcCtx, updateReq)
		if err != nil {
			return nil, h.handleGRPCError(err, operation, map[string]interface{}{
				"user_id": userID,
				"updates": updates,
			})
		}
		
		return h.sanitizeAccount(resp.Account), nil
	})
}

// deleteUserViaGRPC deletes a user via gRPC
func (h *BaseAccountHandler) deleteUserViaGRPC(ctx context.Context, userID int64) error {
	operation := "delete_user"
	
	return h.measureGRPCOperationError(operation, func() error {
		grpcCtx, cancel := context.WithTimeout(ctx, DefaultGRPCTimeout)
		defer cancel()
		
		req := &pb.DeleteReq{Id: userID}
		_, err := h.userClient.Delete(grpcCtx, req)
		
		if err != nil {
			return h.handleGRPCError(err, operation, map[string]interface{}{
				"user_id": userID,
			})
		}
		
		return nil
	})
}

// listUsersViaGRPC retrieves a list of users via gRPC
func (h *BaseAccountHandler) listUsersViaGRPC(ctx context.Context, page, pageSize int32, filters map[string]interface{}) (*pb.ListResponse, error) {
	operation := "list_users"
	
	return h.measureGRPCOperation(operation, func() (*pb.ListResponse, error) {
		grpcCtx, cancel := context.WithTimeout(ctx, CreateUpdateGRPCTimeout)
		defer cancel()
		
		// Build list request
		listReq := &pb.ListReq{
			Page:     page,
			PageSize: pageSize,
		}
		
		// Apply filters if provided
		if role, ok := filters["role"].(string); ok {
			listReq.Role = role
		}
		if branchID, ok := filters["branch_id"].(int64); ok {
			listReq.BranchId = branchID
		}
		if status, ok := filters["status"].(string); ok {
			listReq.Status = status
		}
		
		resp, err := h.userClient.List(grpcCtx, listReq)
		if err != nil {
			return nil, h.handleGRPCError(err, operation, map[string]interface{}{
				"page":      page,
				"page_size": pageSize,
				"filters":   filters,
			})
		}
		
		// Sanitize all accounts in the response
		for i, account := range resp.Accounts {
			resp.Accounts[i] = h.sanitizeAccount(account)
		}
		
		return resp, nil
	})
}

// authenticateUserViaGRPC authenticates a user via gRPC
func (h *BaseAccountHandler) authenticateUserViaGRPC(ctx context.Context, email, password string) (*pb.Account, error) {
	operation := "authenticate_user"
	
	return h.measureGRPCOperation(operation, func() (*pb.Account, error) {
		grpcCtx, cancel := context.WithTimeout(ctx, CreateUpdateGRPCTimeout)
		defer cancel()
		
		req := &pb.LoginReq{
			Email:    email,
			Password: password,
		}
		
		resp, err := h.userClient.Login(grpcCtx, req)
		if err != nil {
			// Log failed authentication attempt
			h.logSecurityEvent(
				"authentication_failure",
				"Failed login attempt",
				"medium",
				map[string]interface{}{
					"email":     email,
					"operation": operation,
				},
			)
			
			return nil, h.handleGRPCError(err, operation, map[string]interface{}{
				"email": email,
			})
		}
		
		// Log successful authentication
		h.logger.Info("User authenticated successfully", map[string]interface{}{
			"user_id":   resp.Account.Id,
			"email":     resp.Account.Email,
			"role":      resp.Account.Role,
			"operation": operation,
		})
		
		return h.sanitizeAccount(resp.Account), nil
	})
}

// ============================================================================
// GRPC RETRY AND RESILIENCE
// ============================================================================

// withGRPCRetry executes gRPC operations with retry logic
func (h *BaseAccountHandler) withGRPCRetry(operation string, fn func() error) error {
	for attempt := 0; attempt < MaxRetryAttempts; attempt++ {
		err := fn()
		
		if err == nil {
			if attempt > 0 {
				h.logger.Info("gRPC operation succeeded after retry", map[string]interface{}{
					"operation": operation,
					"attempt":   attempt + 1,
					"domain":    h.domain,
				})
			}
			return nil
		}
		
		// Check if error is retryable
		if !errorcustom.IsRetryableError(err) {
			h.logger.Warning("gRPC operation failed with non-retryable error", map[string]interface{}{
				"operation": operation,
				"attempt":   attempt + 1,
				"error":     err.Error(),
				"domain":    h.domain,
			})
			return err
		}
		
		if attempt < MaxRetryAttempts-1 {
			delay := time.Duration(attempt+1) * BaseRetryDelay
			h.logger.Warning("gRPC operation failed, retrying", map[string]interface{}{
				"operation":    operation,
				"attempt":      attempt + 1,
				"max_retries":  MaxRetryAttempts,
				"retry_delay":  delay.String(),
				"error":        err.Error(),
				"domain":       h.domain,
			})
			
			time.Sleep(delay)
		} else {
			h.logger.Error("gRPC operation failed after all retries", map[string]interface{}{
				"operation":   operation,
				"attempts":    MaxRetryAttempts,
				"final_error": err.Error(),
				"domain":      h.domain,
			})
		}
	}
	
	return fmt.Errorf("operation %s failed after %d attempts", operation, MaxRetryAttempts)
}

// ============================================================================
// GRPC OPERATION MEASUREMENT HELPERS
// ============================================================================

// measureGRPCOperation measures gRPC operation performance and handles errors
func (h *BaseAccountHandler) measureGRPCOperation(operation string, fn func() (*pb.Account, error)) (*pb.Account, error) {
	start := time.Now()
	
	h.logger.Debug("Starting gRPC operation", map[string]interface{}{
		"operation": operation,
		"domain":    h.domain,
	})
	
	result, err := fn()
	duration := time.Since(start)
	
	// Log performance metrics
	logger.LogPerformance(fmt.Sprintf("grpc_%s_%s", h.domain, operation), duration, map[string]interface{}{
		"operation":   operation,
		"domain":      h.domain,
		"success":     err == nil,
		"duration_ms": duration.Milliseconds(),
	})
	
	if err != nil {
		h.logger.Warning("gRPC operation completed with error", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
	} else {
		h.logger.Debug("gRPC operation completed successfully", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"duration_ms": duration.Milliseconds(),
		})
	}
	
	return result, err
}

// measureGRPCOperationError measures gRPC operations that return only errors
func (h *BaseAccountHandler) measureGRPCOperationError(operation string, fn func() error) error {
	start := time.Now()
	
	h.logger.Debug("Starting gRPC operation", map[string]interface{}{
		"operation": operation,
		"domain":    h.domain,
	})
	
	err := fn()
	duration := time.Since(start)
	
	// Log performance metrics
	logger.LogPerformance(fmt.Sprintf("grpc_%s_%s", h.domain, operation), duration, map[string]interface{}{
		"operation":   operation,
		"domain":      h.domain,
		"success":     err == nil,
		"duration_ms": duration.Milliseconds(),
	})
	
	if err != nil {
		h.logger.Warning("gRPC operation completed with error", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
	} else {
		h.logger.Debug("gRPC operation completed successfully", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"duration_ms": duration.Milliseconds(),
		})
	}
	
	return err
}

// measureGRPCListOperation measures gRPC list operations with specific return type
func (h *BaseAccountHandler) measureGRPCListOperation(operation string, fn func() (*pb.ListResponse, error)) (*pb.ListResponse, error) {
	start := time.Now()
	
	h.logger.Debug("Starting gRPC list operation", map[string]interface{}{
		"operation": operation,
		"domain":    h.domain,
	})
	
	result, err := fn()
	duration := time.Since(start)
	
	// Log performance metrics with additional list-specific context
	context := map[string]interface{}{
		"operation":   operation,
		"domain":      h.domain,
		"success":     err == nil,
		"duration_ms": duration.Milliseconds(),
	}
	
	if result != nil && err == nil {
		context["total_count"] = result.TotalCount
		context["page_count"] = len(result.Accounts)
	}
	
	logger.LogPerformance(fmt.Sprintf("grpc_%s_%s", h.domain, operation), duration, context)
	
	if err != nil {
		h.logger.Warning("gRPC list operation completed with error", context)
	} else {
		h.logger.Debug("gRPC list operation completed successfully", context)
	}
	
	return result, err
}

// ============================================================================
// ACCOUNT SANITIZATION AND SECURITY
// ============================================================================

// sanitizeAccount removes sensitive information from account data
func (h *BaseAccountHandler) sanitizeAccount(account *pb.Account) *pb.Account {
	if account == nil {
		return nil
	}
	
	// Create a new account without sensitive data
	sanitized := &pb.Account{
		Id:        account.Id,
		BranchId:  account.BranchId,
		Name:      account.Name,
		Email:     account.Email,
		// Password:  "", // Never include password in responses
		Avatar:    account.Avatar,
		Title:     account.Title,
		Role:      account.Role,
		OwnerId:   account.OwnerId,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
	
	return sanitized
}

// ============================================================================
// SPECIALIZED GRPC OPERATIONS
// ============================================================================

// checkEmailUniqueness validates email uniqueness via gRPC
func (h *BaseAccountHandler) checkEmailUniqueness(ctx context.Context, email string) error {
	operation := "check_email_uniqueness"
	
	err := h.measureGRPCOperationError(operation, func() error {
		req := &pb.FindByEmailReq{Email: email}
		resp, err := h.userClient.FindByEmail(ctx, req)
		
		if err != nil {
			// If user not found, email is unique (good)
			if strings.Contains(err.Error(), "not found") {
				return nil
			}
			
			// Handle other gRPC errors using UnifiedErrorHandler
			return h.errorHandler.HandleExternalServiceError(
				err, h.domain, "user_service", "check_email_uniqueness", true,
			)
		}
		
		// If we found a user, email is not unique
		if resp != nil && resp.Account != nil {
			return h.errorHandler.HandleBusinessRuleViolation(
				h.domain,
				"email_uniqueness",
				"An account with this email already exists",
				map[string]interface{}{
					"email":              email,
					"existing_account_id": resp.Account.Id,
				},
			)
		}
		
		return nil
	})
	
	return err
}

// validateUserRegistrationRulesViaGRPC validates registration business rules using gRPC
func (h *BaseAccountHandler) validateUserRegistrationRulesViaGRPC(ctx context.Context, email, password string) error {
	// Use ValidateBusinessRules to check multiple conditions
	validations := map[string]func() error{
		"email_format": func() error {
			if !h.isValidEmail(email) {
				return fmt.Errorf("invalid email format: %s", email)
			}
			return nil
		},
		"password_strength": func() error {
			if len(password) < 8 {
				return fmt.Errorf("password must be at least 8 characters")
			}
			return nil
		},
		"email_uniqueness": func() error {
			return h.checkEmailUniqueness(ctx, email)
		},
		"domain_restrictions": func() error {
			if h.config.IsEmailDomainRestricted() && !h.isAllowedDomain(email) {
				return fmt.Errorf("email domain not allowed: %s", email)
			}
			return nil
		},
	}
	
	return h.errorHandler.ValidateBusinessRules(h.domain, validations)
}

// getUserWithPermissionCheck gets user and validates permissions in one operation
func (h *BaseAccountHandler) getUserWithPermissionCheck(ctx context.Context, userID int64, requestingUserID int64, requiredRole string) (*pb.Account, error) {
	// First get the user
	user, err := h.getUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	// Check permissions
	if err := h.checkUserPermissions(requestingUserID, requiredRole, "user_read"); err != nil {
		return nil, err
	}
	
	return user, nil
}

// ============================================================================
// GRPC HEALTH AND DIAGNOSTICS
// ============================================================================

// checkGRPCHealth verifies the gRPC client connection health
func (h *BaseAccountHandler) checkGRPCHealth(ctx context.Context) error {
	operation := "health_check"
	
	return h.measureGRPCOperationError(operation, func() error {
		healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		// Use a simple operation to test connectivity
		_, err := h.userClient.List(healthCtx, &pb.ListReq{
			Page:     1,
			PageSize: 1,
		})
		
		if err != nil {
			return h.handleGRPCError(err, operation, map[string]interface{}{
				"health_check": true,
			})
		}
		
		return nil
	})
}

// getGRPCDiagnostics returns diagnostic information about gRPC operations
func (h *BaseAccountHandler) getGRPCDiagnostics() map[string]interface{} {
	diagnostics := map[string]interface{}{
		"grpc_client_ready":       h.userClient != nil,
		"default_timeout":         DefaultGRPCTimeout.String(),
		"create_update_timeout":   CreateUpdateGRPCTimeout.String(),
		"max_retry_attempts":      MaxRetryAttempts,
		"base_retry_delay":        BaseRetryDelay.String(),
	}
	
	// Test connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := h.checkGRPCHealth(ctx); err != nil {
		diagnostics["connection_status"] = "unhealthy"
		diagnostics["connection_error"] = err.Error()
	} else {
		diagnostics["connection_status"] = "healthy"
	}
	
	return diagnostics
}

// ============================================================================
// UTILITY METHODS FOR GRPC OPERATIONS
// ============================================================================

// CreateUserRequest represents the structure for creating a user
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin teacher student"`
}

// isValidEmail performs basic email validation
func (h *BaseAccountHandler) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// isAllowedDomain checks if email domain is in the allowed list
func (h *BaseAccountHandler) isAllowedDomain(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	
	allowedDomains := h.config.GetAllowedEmailDomains()
	if len(allowedDomains) == 0 {
		return true // No restrictions
	}
	
	domain := strings.Split(email, "@")[1]
	
	for _, allowed := range allowedDomains {
		if domain == allowed {
			return true
		}
	}
	return false
}

// logSecurityEvent logs security-related events with domain context
func (h *BaseAccountHandler) logSecurityEvent(eventType string, description string, severity string, context map[string]interface{}) {
	securityContext := map[string]interface{}{
		"domain":    h.domain,
		"component": "account_handler",
		"timestamp": time.Now().UTC(),
	}
	
	// Merge additional context
	for k, v := range context {
		securityContext[k] = v
	}
	
	logger.LogSecurityEvent(eventType, description, severity, securityContext)
}