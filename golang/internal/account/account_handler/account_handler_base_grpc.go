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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		grpcCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
		grpcCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
		grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
		grpcCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
		grpcCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
	maxRetries := 3
	baseDelay := 100 * time.Millisecond
	
	for attempt := 0; attempt < maxRetries; attempt++ {
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
		
		if attempt < maxRetries-1 {
			delay := time.Duration(attempt+1) * baseDelay
			h.logger.Warning("gRPC operation failed, retrying", map[string]interface{}{
				"operation":    operation,
				"attempt":      attempt + 1,
				"max_retries":  maxRetries,
				"retry_delay":  delay.String(),
				"error":        err.Error(),
				"domain":       h.domain,
			})
			
			time.Sleep(delay)
		} else {
			h.logger.Error("gRPC operation failed after all retries", map[string]interface{}{
				"operation":   operation,
				"attempts":    maxRetries,
				"final_error": err.Error(),
				"domain":      h.domain,
			})
		}
	}
	
	return fmt.Errorf("operation %s failed after %d attempts", operation, maxRetries)
}

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
		h.logger.Debug("gRPC operation completed successfully", map[string]