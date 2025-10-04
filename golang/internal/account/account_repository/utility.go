package account_repository

import (
	"context"
	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"
	"english-ai-full/orm"
	"english-ai-full/utils"
	"fmt"
	"strings"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

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
func (r *Repository) validateForeignKeys(ctx context.Context, user account_dto.Account) error {
    const operation = core.OperationCreateUser
    const table = core.TableAccounts
    const function = core.FuncCreateUser
    
    // Extract request ID from context
    requestID := r.getRequestIDFromContext(ctx)
    
    // Build operation context for error handling
    operationCtx := r.layerContext.BuildOperationContext(operation, table, function, map[string]interface{}{
        "request_id": requestID,
        "email":      utils.MaskEmail(user.Email),
        "branch_id":  user.BranchID,
        "owner_id":   user.OwnerID,
    })

    // Log foreign key validation start
    r.logger.Debug("Starting foreign key validation", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldFunction:  function,
        "source_method":     "CreateUser",
        "validation_step":   "foreign_key_validation",
        "branch_id":         user.BranchID,
        "owner_id":          user.OwnerID,
    }))

    // Validate branch exists
    var branchExists bool
    err := r.db.QueryRowContext(ctx, 
        "SELECT EXISTS(SELECT 1 FROM branches WHERE id = $1)", 
        user.BranchID).Scan(&branchExists)
    if err != nil {
        r.logger.Error("Failed to validate branch existence", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":        requestID,
            core.FieldError:     err.Error(),
            core.FieldOperation: operation,
            core.FieldFunction:  function,
            "source_method":     "CreateUser",
            "error_location":    "branch_validation_query",
            "branch_id":         user.BranchID,
        }))
        return r.errorHandler.Handle(err, operation, table, operationCtx)
    }
    if !branchExists {
        r.logger.Warn("Branch validation failed - branch does not exist", r.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":        requestID,
            core.FieldOperation: operation,
            core.FieldFunction:  function,
            "source_method":     "CreateUser",
            "validation_step":   "branch_existence_check",
            "branch_id":         user.BranchID,
        }))
        return fmt.Errorf("branch with ID %d does not exist", user.BranchID)
    }

    // Validate owner exists (if provided and not zero)
    if user.OwnerID != 0 {
        var ownerExists bool
        err := r.db.QueryRowContext(ctx,
            "SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)",
            user.OwnerID).Scan(&ownerExists)
        if err != nil {
            r.logger.Error("Failed to validate owner existence", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":        requestID,
                core.FieldError:     err.Error(),
                core.FieldOperation: operation,
                core.FieldFunction:  function,
                "source_method":     "CreateUser",
                "error_location":    "owner_validation_query",
                "owner_id":          user.OwnerID,
            }))
            return r.errorHandler.Handle(err, operation, table, operationCtx)
        }
        if !ownerExists {
            r.logger.Warn("Owner validation failed - owner does not exist", r.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":        requestID,
                core.FieldOperation: operation,
                core.FieldFunction:  function,
                "source_method":     "CreateUser",
                "validation_step":   "owner_existence_check",
                "owner_id":          user.OwnerID,
            }))
            return fmt.Errorf("owner with ID %d does not exist", user.OwnerID)
        }
    }

    // Log successful validation
    r.logger.Debug("Foreign key validation completed successfully", r.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":        requestID,
        core.FieldOperation: operation,
        core.FieldFunction:  function,
        "source_method":     "CreateUser",
        "validation_step":   "foreign_key_validation_complete",
        "branch_id":         user.BranchID,
        "owner_id":          user.OwnerID,
    }))

    return nil
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




// verifyPassword compares the provided password with the stored hashed password
func (r *Repository) verifyPassword(providedPassword, hashedPassword string) bool {
    // If you're using bcrypt:
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
    return err == nil
}
