package account_repository

import (
	"context"

	"english-ai-full/logger/core"
	"english-ai-full/orm"

	"fmt"
	"strings"
	"time"


	    "github.com/aarondl/null/v8"  // Use this instead of volatiletech/null
    "github.com/aarondl/sqlboiler/v4/boil"
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