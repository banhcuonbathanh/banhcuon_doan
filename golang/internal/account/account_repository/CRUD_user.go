package account_repository

import (
	"context"
	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	"english-ai-full/utils"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
)


func (r *Repository) CreateUser(ctx context.Context, user account_dto.Account) (account_dto.Account, error) {
    const operation = core.OperationCreateUser
    const table = core.TableAccounts
    const function = core.FuncCreateUser
    
    startTime := time.Now()
    
    // Extract request ID from context
    requestID := r.getRequestIDFromContext(ctx)
    
    // Build operation context with request ID
    operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
        "request_id": requestID,
        "email":      utils.MaskEmail(user.Email),
        "role":       string(user.Role),
    })
    
    // Set request ID in logger
  
    r.logger.SetOperation(operation)
    
    // Log the start with request ID
    r.logger.Info(core.MsgDatabaseOperationStarted, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldTable:     table,
        core.FieldFunction:  function,
        core.FieldEmail:     utils.MaskEmail(user.Email),
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
                "source_method":      "CreateUser",
                "error_location":     "context_check",
            }))
        
        return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
    }

    // Log validation start with request ID
    r.logger.Debug("Starting user validation", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldFunction:  function,
        core.FieldEmail:     utils.MaskEmail(user.Email),
        core.FieldRole:      string(user.Role),
        "source_method":     "CreateUser",
        "validation_step":   "build_orm_account",
    }))

    // Build ORM account
    ormAccount, err := r.buildORMAccount(user)
    if err != nil {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)
        
        if appErr, ok := error_system.IsAppError(err); ok {
            r.logger.Error("Failed to build ORM account with AppError", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":         requestID,
                "error_code":         appErr.Code,
                "error_message":      appErr.Message,
                core.FieldDurationMS: duration.Milliseconds(),
                core.FieldTable:      table,
                core.FieldFunction:   function,
                "source_method":      "CreateUser",
                "error_location":     "build_orm_account",
            }))
            return account_dto.Account{}, appErr
        }
        
        return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
    }

    // Log insert attempt with request ID
    r.logger.Info(core.MsgDatabaseInsertAttempt, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldTable:     table,
        core.FieldFunction:  function,
        core.FieldEmail:     utils.MaskEmail(user.Email),
        core.FieldRole:      string(user.Role),
        "source_method":     "CreateUser",
        "insert_step":       "database_insert",
    }))

    // Database insert
    if err := ormAccount.Insert(ctx, r.db, boil.Infer()); err != nil {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)
        
        return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
    }
    
    duration := time.Since(startTime)
    r.logDatabaseOperation(operation, table, duration, true, 1)
    
    // Log success with request ID
    r.logger.Info(core.MsgDatabaseInsertSuccess, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":           requestID,
        core.FieldUserID:       ormAccount.ID,
        core.FieldEmail:        utils.MaskEmail(user.Email),
        core.FieldDurationMS:   duration.Milliseconds(),
        core.FieldRowsAffected: int64(1),
        core.FieldOperation:    operation,
        core.FieldTable:        table,
        core.FieldFunction:     function,
        "source_method":        "CreateUser",
        "success_step":         "database_insert_complete",
    }))

    createdUser := r.mapORMToDTO(ormAccount)
    return createdUser, nil
}