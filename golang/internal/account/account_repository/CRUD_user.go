package account_repository

import (
	"context"
	"database/sql"
	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	"english-ai-full/orm"
	"english-ai-full/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
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

func (r *Repository) Login(ctx context.Context, loginReq account_dto.LoginRequest) (account_dto.Account, error) {
    const operation = core.OperationLogin
    const table = core.TableAccounts
    const function = core.FuncLogin
    
    startTime := time.Now()
    
    // Extract request ID from context
    requestID := r.getRequestIDFromContext(ctx)
    
    // Build operation context with request ID
    operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
        "request_id": requestID,
        "email":      utils.MaskEmail(loginReq.Email),
    })
    
    // Set request ID in logger
    r.logger.SetOperation(operation)
    
    // Log the start with request ID
    r.logger.Info(core.MsgDatabaseOperationStarted, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldTable:     table,
        core.FieldFunction:  function,
        core.FieldEmail:     utils.MaskEmail(loginReq.Email),
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
                "source_method":      "Login",
                "error_location":     "context_check",
            }))
        
        return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
    }

    // Validate login request
    if err := r.validateLoginRequest(loginReq); err != nil {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)
        
        r.logger.Error("Login validation failed", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":         requestID,
            "error_message":      err.Error(),
            core.FieldDurationMS: duration.Milliseconds(),
            core.FieldTable:      table,
            core.FieldFunction:   function,
            "source_method":      "Login",
            "error_location":     "validation",
            core.FieldEmail:      utils.MaskEmail(loginReq.Email),
        }))
        
        return account_dto.Account{}, err
    }

    // Log database query attempt
    r.logger.Info("Attempting to find user by email", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldTable:     table,
        core.FieldFunction:  function,
        core.FieldEmail:     utils.MaskEmail(loginReq.Email),
        "source_method":     "Login",
        "query_step":        "find_by_email",
    }))

    // Find user by email
    ormAccount, err := orm.Accounts(
        qm.Where("email = ?", loginReq.Email),
    ).One(ctx, r.db)
    
    if err != nil {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)
        
        if errors.Is(err, sql.ErrNoRows) {
            // User not found - log as security event but don't reveal this information
            r.logger.Warn("Login attempt with non-existent email", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":         requestID,
                core.FieldDurationMS: duration.Milliseconds(),
                core.FieldTable:      table,
                core.FieldFunction:   function,
                "source_method":      "Login",
                "error_location":     "user_not_found",
                core.FieldEmail:      utils.MaskEmail(loginReq.Email),
                "security_event":     "invalid_login_attempt",
            }))
            
            // Return generic authentication error to prevent user enumeration
            return account_dto.Account{}, error_system.InvalidCredentials()
        }
        
        // Database error
        r.logger.ErrorWithCause("Database error during login", "database_error", core.LayerRepository, operation,
            r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":         requestID,
                core.FieldError:      err.Error(),
                core.FieldDurationMS: duration.Milliseconds(),
                core.FieldTable:      table,
                core.FieldFunction:   function,
                "source_method":      "Login",
                "error_location":     "database_query",
                core.FieldEmail:      utils.MaskEmail(loginReq.Email),
            }))
        
        return account_dto.Account{}, r.errorHandler.Handle(err, operation, table, operationCtx)
    }

    // Log password verification attempt
    r.logger.Debug("Verifying password", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldFunction:  function,
        core.FieldUserID:    ormAccount.ID,
        core.FieldEmail:     utils.MaskEmail(loginReq.Email),
        "source_method":     "Login",
        "verification_step": "password_check",
    }))

    // Verify password
    if !r.verifyPassword(loginReq.Password, ormAccount.Password) {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)
        
        // Log failed password attempt as security event
        r.logger.Warn("Login attempt with invalid password", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":         requestID,
            core.FieldUserID:     ormAccount.ID,
            core.FieldDurationMS: duration.Milliseconds(),
            core.FieldTable:      table,
            core.FieldFunction:   function,
            "source_method":      "Login",
            "error_location":     "password_verification",
            core.FieldEmail:      utils.MaskEmail(loginReq.Email),
            "security_event":     "invalid_password_attempt",
        }))
        
        // Return generic authentication error to prevent user enumeration
        return account_dto.Account{}, error_system.InvalidCredentials()
    }

    // Check if account is active/valid
    if err := r.validateAccountStatus(ormAccount); err != nil {
        duration := time.Since(startTime)
        r.logDatabaseOperation(operation, table, duration, false, 0)
        
        r.logger.Warn("Login attempt on inactive/suspended account", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":         requestID,
            core.FieldUserID:     ormAccount.ID,
            core.FieldDurationMS: duration.Milliseconds(),
            core.FieldTable:      table,
            core.FieldFunction:   function,
            "source_method":      "Login",
            "error_location":     "account_status_check",
            core.FieldEmail:      utils.MaskEmail(loginReq.Email),
            "account_status":     ormAccount.Status.String,
            "security_event":     "inactive_account_login_attempt",
        }))
        
        return account_dto.Account{}, err
    }

    duration := time.Since(startTime)
    r.logDatabaseOperation(operation, table, duration, true, 1)
    
    // Log successful login
    r.logger.Info("User login successful", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":           requestID,
        core.FieldUserID:       ormAccount.ID,
        core.FieldEmail:        utils.MaskEmail(loginReq.Email),
        core.FieldDurationMS:   duration.Milliseconds(),
        core.FieldOperation:    operation,
        core.FieldTable:        table,
        core.FieldFunction:     function,
        "source_method":        "Login",
        "success_step":         "login_complete",
        "user_role":           ormAccount.Role,
        "branch_id":           ormAccount.BranchID.Int64,
    }))

    // Convert to DTO and return
    userAccount := r.mapORMToDTO(ormAccount)
    
    // Clear password from response for security
    userAccount.Password = ""
    
    return userAccount, nil
}

// validateLoginRequest validates the login request
func (r *Repository) validateLoginRequest(req account_dto.LoginRequest) error {
    if req.Email == "" {
        return error_system.ValidationError("email", "Email is required")
    }
    if req.Password == "" {
        return error_system.ValidationError("password", "Password is required")
    }
    
    // Basic email format validation
    if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
        return error_system.ValidationError("email", "Invalid email format")
    }
    
    return nil
}

// verifyPassword compares the provided password with the stored hashed password
func (r *Repository) verifyPassword(providedPassword, hashedPassword string) bool {
    // If you're using bcrypt:
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
    return err == nil
}

// validateAccountStatus checks if the account is in a valid state for login
func (r *Repository) validateAccountStatus(account *orm.Account) error {
    if !account.Status.Valid {
        return error_system.InvalidCredentials()
    }
    
    switch strings.ToLower(account.Status.String) {
    case "active":
        return nil // Account is valid for login
    case "inactive":
        return error_system.InvalidCredentials()
    case "suspended":
        return error_system.AccountSuspended()
    case "pending":
        return error_system.InvalidCredentials()
    default:
        return error_system.InvalidCredentials()
    }
}

// FIXED: FindByEmail with proper nullable field handling
func (r *Repository) FindByEmail(ctx context.Context, email string) (account_dto.Account, error) {
    const operation = "find_by_email"
    const table = "accounts"
    const function = "find_by_email"
    startTime := time.Now()
    
    // Extract request ID from context
    requestID := r.getRequestIDFromContext(ctx)
    
    // Build operation context with request ID
    operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
        "request_id": requestID,
        "email":      utils.MaskEmail(email),
    })

    // Set operation in logger
    r.logger.SetOperation(operation)

    // Log operation start
    r.logger.Info(core.MsgOperationStarted, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id": requestID,
        "operation":  operation,
        "email":      utils.MaskEmail(email),
        "layer":      core.LayerRepository,
    }))

    // Context cancellation check
    if err := ctx.Err(); err != nil {
        r.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled,
            core.LayerRepository, operation, r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id": requestID,
            }))
        appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
        return account_dto.Account{}, appErr
    }

    // Input validation
    if email == "" {
        err := fmt.Errorf("email cannot be empty")
        r.logger.Error("Invalid email parameter", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id": requestID,
            "error":      err.Error(),
        }))
        appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
        return account_dto.Account{}, appErr
    }

    // Use SQLBoiler ORM instead of raw SQL to handle nullable fields properly
    r.logger.Debug("Executing database query", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id": requestID,
        "operation":  operation,
        "email":      utils.MaskEmail(email),
    }))

    dbStartTime := time.Now()
    
    // Use SQLBoiler query with proper nullable field handling
    ormAccount, err := orm.Accounts(
        qm.Where("email = ?", email),
        qm.And("deleted_at IS NULL"),
    ).One(ctx, r.db)
    
    dbDuration := time.Since(dbStartTime)

    if err != nil {
        operationCtx["database_error"] = "failed to query user by email"
        operationCtx["database_duration_ms"] = dbDuration.Milliseconds()

        if errors.Is(err, sql.ErrNoRows) {
            r.logger.Info("User not found by email", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":           requestID,
                "email":                utils.MaskEmail(email),
                "database_duration_ms": dbDuration.Milliseconds(),
                "result":               "not_found",
            }))

            operationCtx["error_type"] = "not_found"
            appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
            return account_dto.Account{}, appErr
        }

        r.logger.ErrorWithDomainAndCause("Database query failed", r.layerContext.Domain,
            core.CauseDatabaseError, core.LayerRepository, operation,
            r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":           requestID,
                "email":                utils.MaskEmail(email),
                "database_duration_ms": dbDuration.Milliseconds(),
                "error":                err.Error(),
            }))
        appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
        return account_dto.Account{}, appErr
    }

    // Convert ORM to DTO using the existing mapping function
    user := r.mapORMToDTO(ormAccount)

    // Log successful operation
    totalDuration := time.Since(startTime)
    r.logger.Info(core.MsgOperationCompleted, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":           requestID,
        "operation":            operation,
        "user_id":              user.ID,
        "email":                utils.MaskEmail(user.Email),
        "role":                 user.Role,
        "branch_id":            user.BranchID,
        "status":               user.Status,
        "success":              true,
        "duration_ms":          totalDuration.Milliseconds(),
        "database_duration_ms": dbDuration.Milliseconds(),
    }))

    return user, nil
}

// FIXED: FindByEmailWithoutPassword with proper nullable field handling
func (r *Repository) FindByEmailWithoutPassword(ctx context.Context, email string) (account_dto.Account, error) {
    const operation = "find_by_email_no_password"
    const table = "accounts"
    const function = "find_by_email"
    startTime := time.Now()
    
    // Extract request ID from context
    requestID := r.getRequestIDFromContext(ctx)
    
    // Build operation context
    operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
        "request_id": requestID,
        "email":      utils.MaskEmail(email),
    })

    // Set operation in logger
    r.logger.SetOperation(operation)

    // Log operation start
    r.logger.Info(core.MsgOperationStarted, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id": requestID,
        "operation":  operation,
        "email":      utils.MaskEmail(email),
        "layer":      core.LayerRepository,
    }))

    // Context cancellation check
    if err := ctx.Err(); err != nil {
        r.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled,
            core.LayerRepository, operation, r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id": requestID,
            }))

        appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
        return account_dto.Account{}, appErr
    }

    // Input validation
    if email == "" {
        err := fmt.Errorf("email cannot be empty")
        r.logger.Error("Invalid email parameter", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id": requestID,
            "error":      err.Error(),
        }))
        appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
        return account_dto.Account{}, appErr
    }

    // Log database query execution
    dbStartTime := time.Now()
    r.logger.Debug("Executing database query (without password)", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id": requestID,
        "operation":  operation,
        "email":      utils.MaskEmail(email),
    }))

    // Use SQLBoiler query with proper nullable field handling
    ormAccount, err := orm.Accounts(
        qm.Where("email = ?", email),
        qm.And("deleted_at IS NULL"),
    ).One(ctx, r.db)

    dbDuration := time.Since(dbStartTime)

    if err != nil {
        operationCtx["database_error"] = "failed to query user by email"
        operationCtx["database_duration_ms"] = dbDuration.Milliseconds()

        if errors.Is(err, sql.ErrNoRows) {
            r.logger.Info("User not found by email", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":           requestID,
                "email":                utils.MaskEmail(email),
                "database_duration_ms": dbDuration.Milliseconds(),
                "result":               "not_found",
            }))

            operationCtx["error_type"] = "not_found"
            appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
            return account_dto.Account{}, appErr
        }

        r.logger.ErrorWithDomainAndCause("Database query failed", r.layerContext.Domain,
            core.CauseDatabaseError, core.LayerRepository, operation,
            r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":           requestID,
                "email":                utils.MaskEmail(email),
                "database_duration_ms": dbDuration.Milliseconds(),
                "error":                err.Error(),
            }))

        appErr := r.errorHandler.Handle(err, operation, r.layerContext.Domain, operationCtx)
        return account_dto.Account{}, appErr
    }

    // Convert ORM to DTO using the existing mapping function
    user := r.mapORMToDTO(ormAccount)
    
    // Clear password from response for security
    user.Password = ""

    // Log successful operation
    totalDuration := time.Since(startTime)
    r.logger.Info(core.MsgOperationCompleted, r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":           requestID,
        "operation":            operation,
        "user_id":              user.ID,
        "email":                utils.MaskEmail(user.Email),
        "role":                 user.Role,
        "branch_id":            user.BranchID,
        "status":               user.Status,
        "success":              true,
        "duration_ms":          totalDuration.Milliseconds(),
        "database_duration_ms": dbDuration.Milliseconds(),
        "password_excluded":    true,
    }))

    return user, nil
}