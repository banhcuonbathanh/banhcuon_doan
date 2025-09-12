package account_repository

import (
	"context"
	"database/sql"

	"strings"
	"time"

	"english-ai-full/error_system"
	common "english-ai-full/internal"
	"english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	"english-ai-full/orm"
	"english-ai-full/utils"
	utils_config "english-ai-full/utils/config"

	"github.com/aarondl/null/v8"
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

// ===== HELPER METHODS =====

// buildORMAccount creates an ORM Account from DTO with validation
func (r *Repository) buildORMAccount(user account_dto.Account) (*orm.Account, error) {
	// Validation - return AppErrors directly for validation failures
	if user.Email == "" {
		return nil, error_system.ValidationError("email", "Email is required")
	}
	if user.Name == "" {
		return nil, error_system.ValidationError("name", "Name is required") 
	}
	if user.Password == "" {
		return nil, error_system.ValidationError("password", "Password is required")
	}

	now := time.Now()
	return &orm.Account{
		BranchID:  r.toNullInt64(user.BranchID),
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		Avatar:    r.toNullString(user.Avatar),
		Title:     r.toNullString(user.Title),
		Role:      string(user.Role),
		OwnerID:   r.toNullInt64(user.OwnerID),
		Status:    r.toNullString(user.Status),
		CreatedAt: null.TimeFrom(now),
		UpdatedAt: null.TimeFrom(now),
	}, nil
}

// mapORMToDTO converts ORM model to DTO
func (r *Repository) mapORMToDTO(m *orm.Account) account_dto.Account {
	return account_dto.Account{
		ID:        m.ID,
		BranchID:  m.BranchID.Int64,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		Avatar:    m.Avatar.String,
		Title:     m.Title.String,
		Role:      account_dto.Role(m.Role),
		OwnerID:   m.OwnerID.Int64,
		Status: func() string {
			if m.Status.Valid {
				return m.Status.String
			}
			return ""
		}(),
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}
}

// mapDTOToORM converts DTO to ORM model (helper for complex operations)
func (r *Repository) mapDTOToORM(dto account_dto.Account) *orm.Account {
	return &orm.Account{
		ID:        dto.ID,
		BranchID:  null.Int64{Int64: dto.BranchID, Valid: dto.BranchID > 0},
		Name:      dto.Name,
		Email:     dto.Email,
		Password:  dto.Password,
		Avatar:    null.String{String: dto.Avatar, Valid: dto.Avatar != ""},
		Title:     null.String{String: dto.Title, Valid: dto.Title != ""},
		Role:      string(dto.Role),
		OwnerID:   null.Int64{Int64: dto.OwnerID, Valid: dto.OwnerID > 0},
		Status:    null.String{String: dto.Status, Valid: dto.Status != ""},
		CreatedAt: null.Time{Time: dto.CreatedAt, Valid: !dto.CreatedAt.IsZero()},
		UpdatedAt: null.Time{Time: dto.UpdatedAt, Valid: !dto.UpdatedAt.IsZero()},
	}
}

// Helper methods for converting to null types

// toNullInt64 converts an int64 to null.Int64
// Treats 0 as null/invalid, positive values as valid
func (r *Repository) toNullInt64(value int64) null.Int64 {
	return null.Int64{
		Int64: value,
		Valid: value > 0,
	}
}

// toNullString converts a string to null.String
// Treats empty string as null/invalid, non-empty strings as valid
func (r *Repository) toNullString(value string) null.String {
	return null.String{
		String: value,
		Valid:  value != "",
	}
}

// logDatabaseOperation logs database operations with performance metrics
func (r *Repository) logDatabaseOperation(operation, table string, duration time.Duration, success bool, rowsAffected int64) {
	// Build database context with performance metrics
	dbCtx := r.layerContext.BuildDatabaseContext(operation, table, core.FuncCreateUser, time.Now().Add(-duration), success, rowsAffected)
	
	if success {
		r.logger.Info(core.MsgDatabaseOperationSuccess, dbCtx)
	} else {
		r.logger.Error(core.MsgDatabaseOperationFailed, dbCtx)
	}
}

// determineCause determines error cause from database errors
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