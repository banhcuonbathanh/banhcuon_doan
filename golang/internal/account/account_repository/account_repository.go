package account_repository

import (
	"context"
	"database/sql"
	
	"strings"
	"time"

	error_custom "english-ai-full/error_custom"
	"english-ai-full/internal"
	"english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	

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
		
		// Enhanced error logging using layer context
		r.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: duration.Milliseconds(),
				core.FieldTable:      table,
				core.FieldFunction:   function,
			}))
		
		return account_dto.Account{}, r.handleContextError(ctx, operation, table, operationCtx)
	}

	// Log validation start with layer context
	r.logger.Debug("Starting user validation", r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation: operation,
		core.FieldFunction:  function,
		core.FieldEmail:     maskEmail(user.Email),
		core.FieldRole:      string(user.Role),
	}))

	// Build ORM account
	ormAccount, err := r.buildORMAccount(user)
	if err != nil {
		duration := time.Since(startTime)
		
		// Log database operation failure
		r.logDatabaseOperation(operation, table, duration, false, 0)
		
		// Enhanced error logging with layer context
		r.logger.ErrorWithCause("Failed to build ORM account", core.CauseValidationFailed, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: duration.Milliseconds(),
				core.FieldTable:      table,
				core.FieldFunction:   function,
			}))
		
		return account_dto.Account{}, r.wrapError(err, operation, table, operationCtx, &startTime)
	}

	// Log insert attempt with layer context
	r.logger.Info(core.MsgDatabaseInsertAttempt, r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation: operation,
		core.FieldTable:     table,
		core.FieldFunction:  function,
		core.FieldEmail:     maskEmail(user.Email),
		core.FieldRole:      string(user.Role),
	}))

	// Attempt database insert
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
			}))
		
		return account_dto.Account{}, r.wrapError(err, operation, table, operationCtx, &startTime)
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
