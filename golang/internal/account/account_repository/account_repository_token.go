package account_repository

import (
	"context"
	"database/sql"
	"errors"

	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	"english-ai-full/orm"

	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8" // Use this instead of volatiletech/null
	"github.com/aarondl/sqlboiler/v4/boil"
	"google.golang.org/protobuf/types/known/timestamppb"

    	pb "english-ai-full/internal/proto_qr/account"
)

// StoreRefreshToken stores a new refresh token in the database using ORM
func (r *Repository) StoreRefreshToken(ctx context.Context, accountID int64, token string, expiresAt time.Time) error {
    const operation = "store_refresh_token"
    const table = "refresh_tokens"
    const function = "StoreRefreshToken"

    startTime := time.Now()

    // Extract request ID from context
    requestID := r.getRequestIDFromContext(ctx)

    // Build operation context
    operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
        "request_id": requestID,
        "account_id": accountID,
    })

    // Set request ID in logger
    r.logger.SetOperation(operation)

    // Log the start with request ID
    r.logger.Info(core.MsgDatabaseOperationStarted, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldTable:     table,
        core.FieldFunction:  function,
        "account_id":        accountID,
        "expires_at":        expiresAt.Format(time.RFC3339),
    }))

    // Context timeout check
    if err := ctx.Err(); err != nil {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)

        r.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled, core.LayerRepository, operation,
            r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":         requestID,
                core.FieldError:      err.Error(),
                core.FieldDurationMS: duration.Milliseconds(),
                core.FieldTable:      table,
                core.FieldFunction:   function,
                "source_method":      "StoreRefreshToken",
                "error_location":     "context_check",
            }))

        return r.errorHandler.Handle(err, operation, table, operationCtx)
    }

    // Validate token data
    if accountID <= 0 {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)

        r.logger.Error("Validation failed", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":         requestID,
            "error_message":      "invalid account_id",
            core.FieldDurationMS: duration.Milliseconds(),
            core.FieldTable:      table,
            core.FieldFunction:   function,
            "source_method":      "StoreRefreshToken",
            "error_location":     "validation",
            "account_id":         accountID,
        }))

        return fmt.Errorf("invalid account_id: must be greater than 0")
    }

    if token == "" {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)

        r.logger.Error("Validation failed", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":         requestID,
            "error_message":      "empty token",
            core.FieldDurationMS: duration.Milliseconds(),
            core.FieldTable:      table,
            core.FieldFunction:   function,
            "source_method":      "StoreRefreshToken",
            "error_location":     "validation",
        }))

        return fmt.Errorf("token cannot be empty")
    }

    if expiresAt.Before(time.Now()) {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)

        r.logger.Error("Validation failed", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":         requestID,
            "error_message":      "expired token",
            core.FieldDurationMS: duration.Milliseconds(),
            core.FieldTable:      table,
            core.FieldFunction:   function,
            "source_method":      "StoreRefreshToken",
            "error_location":     "validation",
            "expires_at":         expiresAt.Format(time.RFC3339),
        }))

        return fmt.Errorf("token expiration time is in the past")
    }

    // Log database insert attempt
    r.logger.Info("Attempting to insert refresh token", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldTable:     table,
        core.FieldFunction:  function,
        "account_id":        accountID,
        "source_method":     "StoreRefreshToken",
        "insert_step":       "database_insert",
        "expires_at":        expiresAt.Format(time.RFC3339),
    }))

    // Build ORM refresh token model - Use orm.RefreshToken instead of account_dto.RefreshToken
    ormRefreshToken := &orm.RefreshToken{
        AccountID: accountID,
        Token:     token,
        ExpiresAt: expiresAt,
        IsRevoked: null.BoolFrom(false), // Use null.BoolFrom since IsRevoked is null.Bool
        // CreatedAt will be set automatically by the database if it has a default
        // RevokedAt will be null by default
    }

    // Database insert using ORM
    if err := ormRefreshToken.Insert(ctx, r.db, boil.Infer()); err != nil {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)

        // Check for unique constraint violation (duplicate token)
        if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
            strings.Contains(err.Error(), "refresh_tokens_token_key") {
            r.logger.Warn("Duplicate token detected", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":         requestID,
                core.FieldDurationMS: duration.Milliseconds(),
                core.FieldTable:      table,
                core.FieldFunction:   function,
                "source_method":      "StoreRefreshToken",
                "error_location":     "duplicate_token",
                "account_id":         accountID,
                "security_event":     "duplicate_token_attempt",
            }))

            return fmt.Errorf("duplicate token: token already exists")
        }

        // Check for foreign key violation (account doesn't exist)
        if strings.Contains(err.Error(), "violates foreign key constraint") &&
            strings.Contains(err.Error(), "fk_account") {
            r.logger.Error("Foreign key violation", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":         requestID,
                core.FieldError:      err.Error(),
                core.FieldDurationMS: duration.Milliseconds(),
                core.FieldTable:      table,
                core.FieldFunction:   function,
                "source_method":      "StoreRefreshToken",
                "error_location":     "foreign_key_violation",
                "account_id":         accountID,
            }))

            return fmt.Errorf("account not found: account_id %d does not exist", accountID)
        }

        // General database error
        r.logger.ErrorWithCause("Database error during token storage", "database_error", core.LayerRepository, operation,
            r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":         requestID,
                core.FieldError:      err.Error(),
                core.FieldDurationMS: duration.Milliseconds(),
                core.FieldTable:      table,
                core.FieldFunction:   function,
                "source_method":      "StoreRefreshToken",
                "error_location":     "database_insert",
                "account_id":         accountID,
            }))

        return r.errorHandler.Handle(err, operation, table, operationCtx)
    }

    duration := time.Since(startTime)
    r.logDatabaseOperation(operation, table, duration, true, 1)

    // Log successful storage
    r.logger.Info("Refresh token stored successfully", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":           requestID,
        "token_id":             ormRefreshToken.ID,
        "account_id":           accountID,
        core.FieldDurationMS:   duration.Milliseconds(),
        core.FieldRowsAffected: int64(1),
        core.FieldOperation:    operation,
        core.FieldTable:        table,
        core.FieldFunction:     function,
        "source_method":        "StoreRefreshToken",
        "success_step":         "token_stored",
        "expires_at":           expiresAt.Format(time.RFC3339),
        "created_at": func() string {
            if ormRefreshToken.CreatedAt.Valid {
                return ormRefreshToken.CreatedAt.Time.Format(time.RFC3339)
            }
            return "not_set"
        }(),
    }))

    return nil
}



// RevokeAllUserTokens marks all active refresh tokens for a user as revoked
func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
    const operation = "RevokeAllUserTokens"
    startTime := time.Now()

    r.logger.Debug("Revoking all user tokens", map[string]interface{}{
        "operation": operation,
        "user_id":   userID,
    })

    // SQL query to revoke all active tokens
    query := `
        UPDATE refresh_tokens 
        SET is_revoked = true, 
            revoked_at = NOW()
        WHERE account_id = $1 
            AND is_revoked = false
    `

    // Execute the update query
    result, err := r.db.ExecContext(ctx, query, userID)
    if err != nil {
        r.logger.Error("Failed to execute revoke tokens query", map[string]interface{}{
            "operation":   operation,
            "user_id":     userID,
            "duration_ms": time.Since(startTime).Milliseconds(),
            "error":       err.Error(),
        })
        return fmt.Errorf("failed to revoke user tokens: %w", err)
    }

    // Get number of affected rows
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        r.logger.Warn("Could not get rows affected count", map[string]interface{}{
            "operation": operation,
            "user_id":   userID,
            "error":     err.Error(),
        })
        // Don't return error - the revoke operation was successful
    } else {
        r.logger.Info("Successfully revoked user tokens", map[string]interface{}{
            "operation":      operation,
            "user_id":        userID,
            "tokens_revoked": rowsAffected,
            "duration_ms":    time.Since(startTime).Milliseconds(),
        })
    }

    return nil
}

// RevokeToken marks a specific refresh token as revoked
func (r *Repository) RevokeToken(ctx context.Context, token string) error {
    const operation = "RevokeToken"
    startTime := time.Now()

    r.logger.Debug("Revoking specific token", map[string]interface{}{
        "operation": operation,
    })

    query := `
        UPDATE refresh_tokens 
        SET is_revoked = true, 
            revoked_at = NOW()
        WHERE token = $1 
            AND is_revoked = false
    `

    result, err := r.db.ExecContext(ctx, query, token)
    if err != nil {
        r.logger.Error("Failed to revoke token", map[string]interface{}{
            "operation":   operation,
            "duration_ms": time.Since(startTime).Milliseconds(),
            "error":       err.Error(),
        })
        return fmt.Errorf("failed to revoke token: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        r.logger.Warn("Could not get rows affected count", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
    }

    if rowsAffected == 0 {
        r.logger.Warn("No token was revoked - token may not exist or already revoked", map[string]interface{}{
            "operation":   operation,
            "duration_ms": time.Since(startTime).Milliseconds(),
        })
        return fmt.Errorf("token not found or already revoked")
    }

    r.logger.Info("Successfully revoked token", map[string]interface{}{
        "operation":   operation,
        "duration_ms": time.Since(startTime).Milliseconds(),
    })

    return nil
}

// CountActiveTokens returns the number of active (non-revoked, non-expired) tokens for a user
func (r *Repository) CountActiveTokens(ctx context.Context, userID int64) (int, error) {
    const operation = "CountActiveTokens"
    startTime := time.Now()

    r.logger.Debug("Counting active tokens", map[string]interface{}{
        "operation": operation,
        "user_id":   userID,
    })

    query := `
        SELECT COUNT(*) 
        FROM refresh_tokens 
        WHERE account_id = $1 
            AND is_revoked = false 
            AND expires_at > NOW()
    `

    var count int
    err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
    if err != nil {
        r.logger.Error("Failed to count active tokens", map[string]interface{}{
            "operation":   operation,
            "user_id":     userID,
            "duration_ms": time.Since(startTime).Milliseconds(),
            "error":       err.Error(),
        })
        return 0, fmt.Errorf("failed to count active tokens: %w", err)
    }

    r.logger.Debug("Active tokens counted", map[string]interface{}{
        "operation":   operation,
        "user_id":     userID,
        "count":       count,
        "duration_ms": time.Since(startTime).Milliseconds(),
    })

    return count, nil
}

// RevokeOldestTokens revokes the N oldest active tokens for a user
func (r *Repository) RevokeOldestTokens(ctx context.Context, userID int64, count int) error {
    const operation = "RevokeOldestTokens"
    startTime := time.Now()

    if count <= 0 {
        return nil // Nothing to revoke
    }

    r.logger.Debug("Revoking oldest tokens", map[string]interface{}{
        "operation":      operation,
        "user_id":        userID,
        "tokens_to_revoke": count,
    })

    query := `
        UPDATE refresh_tokens 
        SET is_revoked = true, 
            revoked_at = NOW()
        WHERE id IN (
            SELECT id 
            FROM refresh_tokens 
            WHERE account_id = $1 
                AND is_revoked = false
            ORDER BY created_at ASC 
            LIMIT $2
        )
    `

    result, err := r.db.ExecContext(ctx, query, userID, count)
    if err != nil {
        r.logger.Error("Failed to revoke oldest tokens", map[string]interface{}{
            "operation":   operation,
            "user_id":     userID,
            "duration_ms": time.Since(startTime).Milliseconds(),
            "error":       err.Error(),
        })
        return fmt.Errorf("failed to revoke oldest tokens: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        r.logger.Warn("Could not get rows affected count", map[string]interface{}{
            "operation": operation,
            "user_id":   userID,
            "error":     err.Error(),
        })
    } else {
        r.logger.Info("Successfully revoked oldest tokens", map[string]interface{}{
            "operation":      operation,
            "user_id":        userID,
            "tokens_revoked": rowsAffected,
            "duration_ms":    time.Since(startTime).Milliseconds(),
        })
    }

    return nil
}

// ============================================
// Bonus: Additional Utility Methods
// ============================================

// IsTokenValid checks if a token exists and is valid (not revoked, not expired)
func (r *Repository) IsTokenValid(ctx context.Context, token string) (bool, error) {
    const operation = "IsTokenValid"
    
    query := `
        SELECT EXISTS(
            SELECT 1 
            FROM refresh_tokens 
            WHERE token = $1 
                AND is_revoked = false 
                AND expires_at > NOW()
        )
    `

    var isValid bool
    err := r.db.QueryRowContext(ctx, query, token).Scan(&isValid)
    if err != nil {
        r.logger.Error("Failed to check token validity", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
        return false, fmt.Errorf("failed to check token validity: %w", err)
    }

    return isValid, nil
}

// GetUserIDFromToken retrieves the user ID associated with a valid token
func (r *Repository) GetUserIDFromToken(ctx context.Context, token string) (int64, error) {
    const operation = "GetUserIDFromToken"
    
    query := `
        SELECT account_id 
        FROM refresh_tokens 
        WHERE token = $1 
            AND is_revoked = false 
            AND expires_at > NOW()
    `

    var userID int64
    err := r.db.QueryRowContext(ctx, query, token).Scan(&userID)
    if err == sql.ErrNoRows {
        return 0, fmt.Errorf("token not found or expired")
    }
    if err != nil {
        r.logger.Error("Failed to get user ID from token", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
        return 0, fmt.Errorf("failed to get user ID from token: %w", err)
    }

    return userID, nil
}

// CleanupExpiredTokens deletes tokens that expired more than retentionDays ago
// This should be called by a background job/cron
func (r *Repository) CleanupExpiredTokens(ctx context.Context, retentionDays int) (int64, error) {
    const operation = "CleanupExpiredTokens"
    startTime := time.Now()

    r.logger.Info("Starting cleanup of expired tokens", map[string]interface{}{
        "operation":       operation,
        "retention_days":  retentionDays,
    })

    query := `
        DELETE FROM refresh_tokens 
        WHERE expires_at < NOW() - INTERVAL '1 day' * $1
    `

    result, err := r.db.ExecContext(ctx, query, retentionDays)
    if err != nil {
        r.logger.Error("Failed to cleanup expired tokens", map[string]interface{}{
            "operation":   operation,
            "duration_ms": time.Since(startTime).Milliseconds(),
            "error":       err.Error(),
        })
        return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
    }

    deletedCount, err := result.RowsAffected()
    if err != nil {
        r.logger.Warn("Could not get deleted count", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
        return 0, nil
    }

    r.logger.Info("Successfully cleaned up expired tokens", map[string]interface{}{
        "operation":      operation,
        "deleted_count":  deletedCount,
        "retention_days": retentionDays,
        "duration_ms":    time.Since(startTime).Milliseconds(),
    })

    return deletedCount, nil
}


func (r *Repository) CleanupRevokedTokens(ctx context.Context, retentionDays int) (int64, error) {
    const operation = "CleanupRevokedTokens"
    startTime := time.Now()

    r.logger.Info("Starting cleanup of old revoked tokens", map[string]interface{}{
        "operation":      operation,
        "retention_days": retentionDays,
    })

    query := `
        DELETE FROM refresh_tokens 
        WHERE is_revoked = true 
            AND revoked_at < NOW() - INTERVAL '1 day' * $1
    `

    result, err := r.db.ExecContext(ctx, query, retentionDays)
    if err != nil {
        r.logger.Error("Failed to cleanup revoked tokens", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
        return 0, err
    }

    deletedCount, _ := result.RowsAffected()
    
    r.logger.Info("Cleaned up revoked tokens", map[string]interface{}{
        "operation":      operation,
        "deleted_count":  deletedCount,
        "retention_days": retentionDays,
        "duration_ms":    time.Since(startTime).Milliseconds(),
    })

    return deletedCount, nil
}

// ComprehensiveTokenCleanup combines both cleanup strategies
// This is what you should call in your cron job
func (r *Repository) ComprehensiveTokenCleanup(ctx context.Context) error {
    const operation = "ComprehensiveTokenCleanup"
    startTime := time.Now()

    r.logger.Info("Starting comprehensive token cleanup", map[string]interface{}{
        "operation": operation,
    })

    // 1. Delete revoked tokens older than 30 days
    revokedDeleted, err := r.CleanupRevokedTokens(ctx, 30)
    if err != nil {
        r.logger.Error("Failed to cleanup revoked tokens", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
        // Continue with other cleanup even if this fails
    }

    // 2. Delete expired tokens older than 90 days (for audit trail)
    expiredDeleted, err := r.CleanupExpiredTokens(ctx, 90)
    if err != nil {
        r.logger.Error("Failed to cleanup expired tokens", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
    }

    totalDeleted := revokedDeleted + expiredDeleted

    r.logger.Info("Comprehensive token cleanup completed", map[string]interface{}{
        "operation":        operation,
        "revoked_deleted":  revokedDeleted,
        "expired_deleted":  expiredDeleted,
        "total_deleted":    totalDeleted,
        "duration_ms":      time.Since(startTime).Milliseconds(),
    })

    return nil
}

// ============================================
// ALTERNATIVE: Delete immediately (not recommended)
// ============================================

// RevokeAllUserTokensWithDelete - Deletes tokens immediately (no audit trail)
// Use this ONLY if you don't need audit logs
func (r *Repository) RevokeAllUserTokensWithDelete(ctx context.Context, userID int64) error {
    const operation = "RevokeAllUserTokensWithDelete"
    
    query := `
        DELETE FROM refresh_tokens 
        WHERE account_id = $1 
            AND is_revoked = false
    `

    result, err := r.db.ExecContext(ctx, query, userID)
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    r.logger.Info("Deleted user tokens", map[string]interface{}{
        "operation":      operation,
        "user_id":        userID,
        "tokens_deleted": rowsAffected,
    })

    return nil
}

// ============================================
// CRON JOB SERVICE
// ============================================



// RunDailyCleanup should be called by your cron job scheduler
// Example: Run every day at 2 AM
// newe asdfasdfasdfds


func (r *Repository) GetRefreshTokenByUserIDAndToken(ctx context.Context, userID int64, token string) (*account_dto.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, is_revoked, created_at 
		FROM refresh_tokens 
		WHERE user_id = $1 AND token = $2 AND is_revoked = false
	`
	
	var refreshToken account_dto.RefreshToken
	err := r.db.QueryRowContext(ctx, query, userID, token).Scan(
		&refreshToken.ID,
		&refreshToken.AccountID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
        &refreshToken.CreatedAt,
        		&refreshToken.RevokedAt,
		&refreshToken.IsRevoked,
	
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("refresh token not found or has been revoked")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	
	return &refreshToken, nil
}



func (r *Repository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = true 
		WHERE token = $1
	`
	
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID int64) (*pb.Account, error) {
	query := `
		SELECT id, branch_id, name, email, avatar, title, role, owner_id, status, created_at, updated_at
		FROM accounts 
		WHERE id = $1
	`
	
	var account pb.Account
	var createdAt, updatedAt time.Time
	var status string
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.Id,
		&account.BranchId,
		&account.Name,
		&account.Email,
		&account.Avatar,
		&account.Title,
		&account.Role,
		&account.OwnerId,
		&status,
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Convert status string to enum
	account.Status = parseAccountStatus(status)
	account.CreatedAt = timestamppb.New(createdAt)
	account.UpdatedAt = timestamppb.New(updatedAt)
	
	return &account, nil
}

func parseAccountStatus(status string) pb.AccountStatus {
	switch status {
	case "ACTIVE":
		return pb.AccountStatus_ACTIVE
	case "INACTIVE":
		return pb.AccountStatus_INACTIVE
	case "SUSPENDED":
		return pb.AccountStatus_SUSPENDED
	default:
		return pb.AccountStatus_UNKNOWN
	}
}
// new asdfasdfadsfsd