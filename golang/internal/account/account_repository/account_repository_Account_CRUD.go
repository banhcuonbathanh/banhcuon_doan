package account_repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"english-ai-full/integration"
	"english-ai-full/internal/account"
	"english-ai-full/internal/model"
	"english-ai-full/logger"
)

var _ account.AccountRepositoryInterface = (*account.Repository)(nil)

// Repository struct using SpecializedLogger
type Repository struct {
	db     *sql.DB
	logger *logger.SpecializedLogger
}

// NewAccountRepository constructor using integration utilities
func NewAccountRepository(db *sql.DB) *Repository {
	repoUtils := integration.NewRepositoryUtilities("account")
	
	repoUtils.RepoLogger.LogBusinessEvent("account", "repository", "initialization", "started", map[string]interface{}{
		"database_connected": db != nil,
	})
	
	return &Repository{
		db:     db,
		logger: repoUtils.RepoLogger,
	}
}

// CreateUser creates a new user account in the database
func (r *Repository) CreateUser(ctx context.Context, user model.Account) (model.Account, error) {
	// Create domain-specific utilities for repository layer
	repoUtils := integration.NewRepositoryUtilities("account")
	
	// Start timing the operation
	startTime := time.Now()
	
	// Validate input using integration error handling
	if user.Email == "" {
		err := fmt.Errorf("email is required")
		repoUtils.HandleError(err, "create_user_validation", map[string]interface{}{
			"validation_field": "email",
			"error_type":      "validation_error",
		})
		return model.Account{}, err
	}

	// Log the operation start
	repoUtils.RepoLogger.LogBusinessEvent("account", "repository", "create_user", "started", map[string]interface{}{
		"email":     user.Email,
		"role":      user.Role,
		"branch_id": user.BranchID,
	})

	// Check if user already exists
	exists, err := r.ExistsByEmail(ctx, user.Email)
	if err != nil {
		duration := time.Since(startTime)
		repoUtils.LogOperation("create_user_exists_check", "failed", false, duration, map[string]interface{}{
			"email": user.Email,
			"error": err.Error(),
		})
		return model.Account{}, repoUtils.HandleError(err, "create_user_exists_check", nil)
	}

	if exists {
		duration := time.Since(startTime)
		err := fmt.Errorf("user with email %s already exists", user.Email)
		repoUtils.LogOperation("create_user", "user_exists", false, duration, map[string]interface{}{
			"email": user.Email,
		})
		return model.Account{}, repoUtils.HandleError(err, "create_user", map[string]interface{}{
			"error_type": "business_logic_error",
		})
	}

	// Prepare SQL query
	query := `
		INSERT INTO accounts (
			email, password_hash, first_name, last_name, phone, 
			role, status, branch_id, owner_id, created_at, updated_at,
			email_verified
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id, created_at, updated_at`

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	// Set default status if not provided
	if user.Status == "" {
		user.Status = "active"
	}

	// Execute the query
	var createdUser model.Account = user // Copy all fields first
	
	err = r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.Role,
		user.Status,
		user.BranchID,
		user.OwnerID,
		user.CreatedAt,
		user.UpdatedAt,
		user.EmailVerified,
	).Scan(
		&createdUser.ID,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	duration := time.Since(startTime)

	if err != nil {
		// Handle different types of database errors using integration
		var contextData = map[string]interface{}{
			"email":        user.Email,
			"error":        err.Error(),
			"error_type":   getErrorType(err),
			"query_params": map[string]interface{}{
				"email":     user.Email,
				"role":      user.Role,
				"branch_id": user.BranchID,
			},
		}

		// Add specific error type context
		switch {
		case err == sql.ErrNoRows:
			contextData["error_category"] = "database_error"
			contextData["error_subtype"] = "create_failed"
		case isDuplicateKeyError(err):
			contextData["error_category"] = "business_logic_error"
			contextData["error_subtype"] = "duplicate_email"
		case isConstraintViolationError(err):
			contextData["error_category"] = "validation_error"
			contextData["error_subtype"] = "constraint_violation"
		default:
			contextData["error_category"] = "database_error"
			contextData["error_subtype"] = "unknown"
		}

		repoUtils.LogOperation("create_user", "failed", false, duration, contextData)

		return model.Account{}, repoUtils.HandleError(err, "create_user", contextData)
	}

	// Success logging
	repoUtils.LogOperation("create_user", "success", true, duration, map[string]interface{}{
		"user_id":   createdUser.ID,
		"email":     createdUser.Email,
		"role":      createdUser.Role,
		"branch_id": createdUser.BranchID,
		"status":    createdUser.Status,
	})

	repoUtils.RepoLogger.LogBusinessEvent("account", "repository", "create_user", "completed", map[string]interface{}{
		"user_id":      createdUser.ID,
		"email":        createdUser.Email,
		"role":         createdUser.Role,
		"branch_id":    createdUser.BranchID,
		"duration_ms":  duration.Milliseconds(),
		"success":      true,
	})

	// Clear sensitive data before returning
	createdUser.PasswordHash = "" // Don't return password hash

	return createdUser, nil
}

// Helper functions for error handling

// isDuplicateKeyError checks if the error is a duplicate key violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL duplicate key error code is 23505
	return strings.Contains(err.Error(), "duplicate key") || 
		   strings.Contains(err.Error(), "23505") ||
		   strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// isConstraintViolationError checks if the error is a constraint violation
func isConstraintViolationError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL constraint violation error codes start with 23
	return strings.Contains(err.Error(), "violates") ||
		   strings.Contains(err.Error(), "constraint") ||
		   strings.Contains(err.Error(), "23")
}

// getErrorType returns a string representation of the error type for logging
func getErrorType(err error) string {
	if err == nil {
		return "none"
	}
	
	switch {
	case err == sql.ErrNoRows:
		return "no_rows"
	case isDuplicateKeyError(err):
		return "duplicate_key"
	case isConstraintViolationError(err):
		return "constraint_violation"
	case strings.Contains(err.Error(), "timeout"):
		return "timeout"
	case strings.Contains(err.Error(), "connection"):
		return "connection_error"
	default:
		return "unknown_database_error"
	}
}