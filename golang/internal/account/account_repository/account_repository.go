package account_repository

import (
	"context"
	"database/sql"


	"time"

	"english-ai-full/error_system"
	common "english-ai-full/internal"
	"english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	"english-ai-full/orm"
	"english-ai-full/utils"
	utils_config "english-ai-full/utils/config"


	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

var _ account.AccountRepositoryInterface = (*Repository)(nil)

type Repository struct {
	db           *sql.DB
	logger       *core.CoreLogger
	errorHandler *error_system.RepositoryErrorHandler
	config       *utils_config.Config
	layerContext *common.RepositoryLayerContext
}

func NewAccountRepository(db *sql.DB) *Repository {
	// Create layer context first
	layerContext := common.NewRepositoryLayerContext("account", "account-service")
	
	// Create logger using layer context - it will be pre-configured
	logger := layerContext.NewRepositoryLogger()
	errorHandler := error_system.NewRepositoryErrorHandler(logger, "account")
	
	// Enable enhanced error tracking with stack traces
	logger.SetStackCapture(true, 3) // Capture 3 stack frames for better context
	
	return &Repository{
		db:           db,
		logger:       logger,
		errorHandler: errorHandler,
		config:       utils_config.GetConfig(),
		layerContext: layerContext,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user account_dto.Account) (account_dto.Account, error) {
	const operation = core.OperationCreateUser
	const table = core.TableAccounts
	const function = core.FuncCreateUser
	
	startTime := time.Now()
	
	// Build operation context using layer context
	operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
		"email": utils.MaskEmail(user.Email),
		"role":  string(user.Role),
	})
	
	// Set operation in logger
	r.logger.SetOperation(operation)
	
	// Log the start with layer context
	r.logger.Info(core.MsgDatabaseOperationStarted, r.layerContext.MergeWithContext(map[string]interface{}{
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
				core.FieldError:      err.Error(),
				core.FieldDurationMS: duration.Milliseconds(),
				core.FieldTable:      table,
				core.FieldFunction:   function,
				"source_method":      "CreateUser",
				"error_location":     "context_check",
			}))
		
		return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
	}

	// Log validation start
	r.logger.Debug("Starting user validation", r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation: operation,
		core.FieldFunction:  function,
		core.FieldEmail:     utils.MaskEmail(user.Email),
		core.FieldRole:      string(user.Role),
		"source_method":     "CreateUser",
		"validation_step":   "build_orm_account",
	}))

	// Build ORM account - simplified error handling
	ormAccount, err := r.buildORMAccount(user)
	if err != nil {
		duration := time.Since(startTime)
		r.logDatabaseOperation(operation, table, duration, false, 0)
		
		// Check if it's already an AppError - don't double wrap
		if appErr, ok := error_system.IsAppError(err); ok {
			r.logger.Error("Failed to build ORM account with AppError", r.layerContext.MergeWithContext(map[string]interface{}{
				"error_code":        appErr.Code,
				"error_message":     appErr.Message,
				core.FieldDurationMS: duration.Milliseconds(),
				core.FieldTable:      table,
				core.FieldFunction:   function,
				"source_method":      "CreateUser",
				"error_location":     "build_orm_account",
			}))
			return account_dto.Account{}, appErr
		}
		
		// Log and handle non-AppError
		r.logger.ErrorWithCause("Failed to build ORM account", core.CauseValidationFailed, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: duration.Milliseconds(),
				core.FieldTable:      table,
				core.FieldFunction:   function,
				"source_method":      "CreateUser",
				"error_location":     "build_orm_account",
			}))
		
		return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
	}

	// Log insert attempt
	r.logger.Info(core.MsgDatabaseInsertAttempt, r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation: operation,
		core.FieldTable:     table,
		core.FieldFunction:  function,
		core.FieldEmail:     utils.MaskEmail(user.Email),
		core.FieldRole:      string(user.Role),
		"source_method":     "CreateUser",
		"insert_step":       "database_insert",
	}))

	// Attempt database insert
	if err := ormAccount.Insert(ctx, r.db, boil.Infer()); err != nil {
		duration := time.Since(startTime)
		r.logDatabaseOperation(operation, table, duration, false, 0)
		
		// Enhanced error logging with cause detection
		cause := r.determineCause(err)
		r.logger.ErrorWithCause(core.MsgDatabaseInsertFailed, cause, core.LayerRepository, operation,
			r.layerContext.MergeWithContext(map[string]interface{}{
				core.FieldError:         err.Error(),
				core.FieldDurationMS:    duration.Milliseconds(),
				core.FieldAttemptedEmail: utils.MaskEmail(user.Email),
				core.FieldAttemptedRole: string(user.Role),
				core.FieldTable:         table,
				core.FieldFunction:      function,
				"source_method":         "CreateUser",
				"error_location":        "database_insert",
				"sql_operation":         "INSERT",
			}))
		
		return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
	}
	
	duration := time.Since(startTime)
	r.logDatabaseOperation(operation, table, duration, true, 1)
	
	r.logger.Info(core.MsgDatabaseInsertSuccess, r.layerContext.MergeWithContext(map[string]interface{}{
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

// ExistsByEmail checks if an email already exists in the database
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const operation = "exists_by_email"
	const table = core.TableAccounts
	
	r.logger.Debug("Checking if email exists", r.layerContext.MergeWithContext(map[string]interface{}{
		"email":     utils.MaskEmail(email),
		"operation": operation,
		"table":     table,
	}))
	
	// Context timeout check
	if err := ctx.Err(); err != nil {
		return false, r.errorHandler.Handle(err, operation, table, map[string]interface{}{
			"email": utils.MaskEmail(email),
		})
	}
	
	// Query database for existing email
	exists, err := orm.Accounts(qm.Where("email = ?", email)).Exists(ctx, r.db)
	if err != nil {
		r.logger.Error("Failed to check email existence", r.layerContext.MergeWithContext(map[string]interface{}{
			"email": utils.MaskEmail(email),
			"error": err.Error(),
			"table": table,
		}))
		return false, r.errorHandler.Handle(err, operation, table, map[string]interface{}{
			"email": utils.MaskEmail(email),
		})
	}
	
	r.logger.Debug("Email existence check completed", r.layerContext.MergeWithContext(map[string]interface{}{
		"email":  utils.MaskEmail(email),
		"exists": exists,
		"table":  table,
	}))
	
	return exists, nil
}
