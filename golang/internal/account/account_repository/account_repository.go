package account_repository

import (
	"context"
	"database/sql"
	"fmt"

	"strings"
	"time"

	error_custom "english-ai-full/error_custom"
	common "english-ai-full/internal"
	"english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	"english-ai-full/orm"

	utils_config "english-ai-full/utils/config"

	"github.com/aarondl/sqlboiler/v4/boil"
)

var _ account.AccountRepositoryInterface = (*Repository)(nil)

type Repository struct {
	db           *sql.DB
	logger       *core.CoreLogger
	errorHandler *error_custom.RepositoryErrorManager
	config       *utils_config.Config
	layerContext *common.RepositoryLayerContext
}

func NewAccountRepository(db *sql.DB) *Repository {
	// Create layer context first
	layerContext := common.NewRepositoryLayerContext("account", "account-service")
	
	// Create logger using layer context - it will be pre-configured
	logger := layerContext.NewRepositoryLogger()
	
	// Enable enhanced error tracking with stack traces
	logger.SetStackCapture(true, 3) // Capture 3 stack frames for better context
	
	return &Repository{
		db:           db,
		logger:       logger,
		errorHandler: error_custom.NewRepositoryErrorManager(),
		config:       utils_config.GetConfig(),
		layerContext: layerContext,
	}
}


// ============================================================================
// IMPROVED IMPLEMENTATION WITH LAYER CONTEXT
// ============================================================================
func (r *Repository) CreateUser(ctx context.Context, user account_dto.Account) (account_dto.Account, error) {
	const operation = core.OperationCreateUser
	const table = core.TableAccounts
	const function = core.FuncCreateUser
	
	startTime := time.Now()
	
	// Build operation context using layer context
	operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
		"email": maskEmail(user.Email),
		"role":  string(user.Role),
	})
	
	// Set operation in logger
	r.logger.SetOperation(operation)
	
	// Log the start with layer context
	r.logger.Info(core.MsgDatabaseOperationStarted, r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation: operation,
		core.FieldTable:     table,
		core.FieldFunction:  function,
		core.FieldEmail:     maskEmail(user.Email),
	}))

	// Context timeout check
	if err := ctx.Err(); err != nil {
		duration := time.Since(startTime)
		
		// Log database operation failure with layer context
		r.logDatabaseOperation(operation, table, duration, false, 0)
		
		// Enhanced error logging using layer context with caller info
		r.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: duration.Milliseconds(),
				core.FieldTable:      table,
				core.FieldFunction:   function,
				"source_method":      "CreateUser", // Explicit method identification
				"error_location":     "context_check",
			}))
		
		return account_dto.Account{}, r.handleContextError(ctx, operation, table, operationCtx)
	}

	// Log validation start with layer context
	r.logger.Debug("Starting user validation", r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation: operation,
		core.FieldFunction:  function,
		core.FieldEmail:     maskEmail(user.Email),
		core.FieldRole:      string(user.Role),
		"source_method":     "CreateUser",
		"validation_step":   "build_orm_account",
	}))

	// Build ORM account with enhanced error context
	ormAccount, err := r.buildORMAccountWithContext(user, operationCtx)
	if err != nil {
		duration := time.Since(startTime)
		
		// Log database operation failure
		r.logDatabaseOperation(operation, table, duration, false, 0)
		
		// Enhanced error logging with detailed context
		r.logger.ErrorWithCause("Failed to build ORM account", core.CauseValidationFailed, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: duration.Milliseconds(),
				core.FieldTable:      table,
				core.FieldFunction:   function,
				"source_method":      "CreateUser",
				"error_location":     "build_orm_account",
				"validation_target":  "user_struct_to_orm",
			}))
		
		return account_dto.Account{}, r.wrapErrorWithContext(err, operation, table, operationCtx, &startTime, "build_orm_account")
	}

	// Log insert attempt with enhanced context
	r.logger.Info(core.MsgDatabaseInsertAttempt, r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation: operation,
		core.FieldTable:     table,
		core.FieldFunction:  function,
		core.FieldEmail:     maskEmail(user.Email),
		core.FieldRole:      string(user.Role),
		"source_method":     "CreateUser",
		"insert_step":       "database_insert",
	}))

	// Attempt database insert with detailed error tracking
	if err := ormAccount.Insert(ctx, r.db, boil.Infer()); err != nil {
		duration := time.Since(startTime)
		
		// Log database operation failure
		r.logDatabaseOperation(operation, table, duration, false, 0)
		
		// Enhanced error logging with layer context and cause detection
		cause := r.determineCause(err)
		r.logger.ErrorWithCause(core.MsgDatabaseInsertFailed, cause, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(map[string]interface{}{
				core.FieldError:          err.Error(),
				core.FieldDurationMS:     duration.Milliseconds(),
				core.FieldAttemptedEmail: maskEmail(user.Email),
				core.FieldAttemptedRole:  string(user.Role),
				core.FieldTable:          table,
				core.FieldFunction:       function,
				"source_method":          "CreateUser",
				"error_location":         "database_insert",
				"sql_operation":          "INSERT",
				"database_table":         table,
				"boil_operation":         "Insert",
			}))
		
		return account_dto.Account{}, r.wrapErrorWithContext(err, operation, table, operationCtx, &startTime, "database_insert")
	}
	
	duration := time.Since(startTime)
	
	// Log successful operation with layer context
	r.logDatabaseOperation(operation, table, duration, true, 1)
	
	r.logger.Info(core.MsgDatabaseInsertSuccess, r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldUserID:       ormAccount.ID,
		core.FieldEmail:        maskEmail(user.Email),
		core.FieldDurationMS:   duration.Milliseconds(),
		core.FieldRowsAffected: int64(1),
		core.FieldOperation:    operation,
		core.FieldTable:        table,
		core.FieldFunction:     function,
		"source_method":        "CreateUser",
		"success_step":         "database_insert_complete",
	}))

	return r.handleInsertSuccess(ormAccount, operation, table, operationCtx, startTime), nil
}


// Helper method to log database operations with enhanced context
func (r *Repository) logDatabaseOperation(operation, table string, duration time.Duration, success bool, rowsAffected int64) {
	// Build database context with performance metrics
	dbCtx := r.layerContext.BuildDatabaseContext(operation, table, core.FuncCreateUser, time.Now().Add(-duration), success, rowsAffected)
	
	if success {
		r.logger.Info(core.MsgDatabaseOperationSuccess, dbCtx)
	} else {
		r.logger.Error(core.MsgDatabaseOperationFailed, dbCtx)
	}
}

// Helper method to determine error cause from database errors
func (r *Repository) determineCause(err error) string {
	errStr := strings.ToLower(err.Error())
	
	// Check for common database error patterns
	switch {
	case strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique"):
		if strings.Contains(errStr, "email") {
			return core.CauseDuplicateEmail
		}
		return "duplicate_entry"
	case strings.Contains(errStr, "timeout"):
		return core.CauseTimeout
	case strings.Contains(errStr, "connection"):
		return core.CauseNetworkError
	case strings.Contains(errStr, "constraint"):
		return "constraint_violation"
	default:
		return core.CauseDatabaseError
	}
}

// Enhanced error handling with layer context
func (r *Repository) handleContextError(ctx context.Context, operation, table string, operationCtx map[string]interface{}) error {
	if ctx.Err() == context.Canceled {
		r.logger.ErrorWithCause(core.MsgContextCancelled, core.CauseContextCancelled, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(operationCtx))
	} else if ctx.Err() == context.DeadlineExceeded {
		r.logger.ErrorWithCause(core.MsgContextTimeout, core.CauseContextTimeout, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(operationCtx))
	}
	
	// Use HandleDatabaseError instead of WrapRepositoryError
	return r.errorHandler.HandleDatabaseError(
		ctx.Err(),           // the original context error
		r.layerContext.Domain, // domain from layer context
		table,               // table name
		operation,           // operation name
		operationCtx,        // operation context
	)
}


// new stary

func (r *Repository) buildORMAccountWithContext(user account_dto.Account, operationCtx map[string]interface{}) (*orm.Account, error) {
	// Add validation step logging
	r.logger.Debug("Building ORM account from DTO", r.layerContext.MergeWithContext(map[string]interface{}{
		"email":            maskEmail(user.Email),
		"role":             string(user.Role),
		"source_method":    "buildORMAccountWithContext", 
		"validation_step":  "dto_to_orm_conversion",
	}))
	
	// Perform the actual ORM building (your existing logic here)
	ormAccount, err := r.buildORMAccount(user)
	if err != nil {
		// Log detailed validation error
		r.logger.Error("ORM account building failed", r.layerContext.MergeWithContext(map[string]interface{}{
			"error":            err.Error(),
			"email":            maskEmail(user.Email),
			"role":             string(user.Role),
			"source_method":    "buildORMAccountWithContext",
			"error_location":   "orm_conversion",
		}))
		return nil, fmt.Errorf("failed to build ORM account: %w", err)
	}
	
	r.logger.Debug("ORM account built successfully", r.layerContext.MergeWithContext(map[string]interface{}{
		"email":           maskEmail(user.Email),
		"role":            string(user.Role),
		"source_method":   "buildORMAccountWithContext",
		"success_step":    "orm_conversion_complete",
	}))
	
	return ormAccount, nil
}


// new done



// wrapErrorWithContext wraps errors with enhanced context information
func (r *Repository) wrapErrorWithContext(err error, operation, table string, operationCtx map[string]interface{}, startTime *time.Time, errorLocation string) error {
	duration := time.Since(*startTime)
	
	// Create enhanced error context
	errorCtx := make(map[string]interface{})
	for k, v := range operationCtx {
		errorCtx[k] = v
	}
	
	// Add error-specific context
	errorCtx["error_duration_ms"] = duration.Milliseconds()
	errorCtx["error_location"] = errorLocation
	errorCtx["error_timestamp"] = time.Now().Format(time.RFC3339)
	
	// Log the wrapped error with full context
	r.logger.Error("Wrapping repository error with context", r.layerContext.MergeWithContext(map[string]interface{}{
		"original_error":   err.Error(),
		"error_location":   errorLocation,
		"duration_ms":      duration.Milliseconds(),
		"operation":        operation,
		"table":            table,
		"source_method":    "wrapErrorWithContext",
	}))
	
	return r.errorHandler.HandleDatabaseError(
		err,                    // the original error
		r.layerContext.Domain,  // domain from layer context
		table,                  // table name
		operation,              // operation name
		errorCtx,              // enhanced operation context
	)
}