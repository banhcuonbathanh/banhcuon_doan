package account_repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	error_custom "english-ai-full/error_custom"
	"english-ai-full/internal"
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
	
	return r.errorHandler.WrapRepositoryError(ctx.Err(), operation, table)
}

// Enhanced success handling with layer context
func (r *Repository) handleInsertSuccess(ormAccount interface{}, operation, table string, operationCtx map[string]interface{}, startTime time.Time) account_dto.Account {
	// Convert ORM account to DTO (implementation depends on your ORM setup)
	// This is a placeholder - implement according to your ORM model
	
	// Log final success with enhanced context
	r.logger.Info(core.MsgOperationCompleted, r.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldOperation:  operation,
		core.FieldTable:      table,
		core.FieldDurationMS: time.Since(startTime).Milliseconds(),
		core.FieldSuccess:    true,
	}))
	
	// Return converted DTO (implement based on your ORM model)
	return account_dto.Account{} // Placeholder
}

// Enhanced error wrapping with layer context
func (r *Repository) wrapError(err error, operation, table string, operationCtx map[string]interface{}, startTime *time.Time) error {
	if startTime != nil {
		operationCtx[core.FieldDurationMS] = time.Since(*startTime).Milliseconds()
	}
	
	// Log with full context
	r.logger.Error("Repository operation failed", r.layerContext.MergeWithContext(operationCtx))
	
	return r.errorHandler.WrapRepositoryError(err, operation, table)
}

// Build operation context helper (enhanced)
func (r *Repository) buildOperationContext(user account_dto.Account) map[string]interface{} {
	return r.layerContext.BuildOperationContext(core.OperationCreateUser, core.TableAccounts, core.FuncCreateUser, map[string]interface{}{
		core.FieldEmail:    maskEmail(user.Email),
		core.FieldRole:     string(user.Role),
		core.FieldBranchID: user.BranchID,
		core.FieldOwnerID:  user.OwnerID,
	})
}

// buildORMAccount converts DTO to ORM model
func (r *Repository) buildORMAccount(user account_dto.Account) (*orm.Account, error) {
	ormAccount := &orm.Account{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
		Role:     string(user.Role),
	}
	
	// Set nullable fields
	if user.BranchID != 0 {
		ormAccount.BranchID = null.Int64From(user.BranchID)
	}
	if user.OwnerID != 0 {
		ormAccount.OwnerID = null.Int64From(user.OwnerID)
	}
	if user.Avatar != "" {
		ormAccount.Avatar = null.StringFrom(user.Avatar)
	}
	if user.Title != "" {
		ormAccount.Title = null.StringFrom(user.Title)
	}
	
	if ormAccount.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	
	return ormAccount, nil
}

// Enhanced email masking function (same as before but with better structure)
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	// Find the @ symbol
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 {
		// Invalid email format, mask everything except first and last char
		if len(email) <= 2 {
			return "***"
		}
		return email[:1] + "***" + email[len(email)-1:]
	}
	
	username := email[:atIndex]
	domain := email[atIndex:]
	
	// Mask username part
	if len(username) <= 2 {
		return "**" + domain
	} else if len(username) <= 4 {
		return username[:1] + "**" + username[len(username)-1:] + domain
	} else {
		return username[:2] + "***" + username[len(username)-1:] + domain
	}
}