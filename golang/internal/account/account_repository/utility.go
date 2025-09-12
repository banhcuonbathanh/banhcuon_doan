package account_repository

import (
	"context"
	"english-ai-full/logger/core"
	"english-ai-full/orm"
	"english-ai-full/utils"

"github.com/aarondl/sqlboiler/v4/queries/qm"
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

