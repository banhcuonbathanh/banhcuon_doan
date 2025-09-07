package account_repository

import (

	error_custom "english-ai-full/error_custom"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger/core"
	"strings"

	"english-ai-full/orm"
	"time"

	"github.com/aarondl/null/v8"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ===== HELPER METHODS =====

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
			return "" // or some default value
		}(),
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}
}

// mapORMToProto converts ORM model to Proto format



func (r *Repository) mapORMToProto(m *orm.Account) account.Account {
	return account.Account{
		Id:       m.ID,
		BranchId: m.BranchID.Int64,
		Name:     m.Name,
		Email:    m.Email,
		Avatar:   m.Avatar.String,
		Title:    m.Title.String,
		Role:     m.Role,
		OwnerId:  m.OwnerID.Int64,
		Status:   r.mapStatusToProtoEnum(m.Status),
		CreatedAt: func() *timestamppb.Timestamp {
			if m.CreatedAt.Valid {
				return timestamppb.New(m.CreatedAt.Time)
			}
			return nil
		}(),
		UpdatedAt: func() *timestamppb.Timestamp {
			if m.UpdatedAt.Valid {
				return timestamppb.New(m.UpdatedAt.Time)
			}
			return nil
		}(),
	}
}

// Helper function to map status string to AccountStatus enum
func (r *Repository) mapStatusToProtoEnum(status null.String) account.AccountStatus {
	if !status.Valid {
		return account.AccountStatus_UNKNOWN // default for null/invalid status
	}
	
	switch status.String {
	case "active":
		return account.AccountStatus_ACTIVE
	case "inactive":
		return account.AccountStatus_INACTIVE
	case "suspended":
		return account.AccountStatus_SUSPENDED
	default:
		return account.AccountStatus_UNKNOWN
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

// Alternative helper methods with different null handling logic:

// toNullInt64FromPointer converts *int64 to null.Int64 (if you need pointer handling)
func (r *Repository) toNullInt64FromPointer(value *int64) null.Int64 {
	if value == nil {
		return null.Int64{Valid: false}
	}
	return null.Int64{Int64: *value, Valid: true}
}

// toNullStringAlways converts string to null.String (always valid, even if empty)
func (r *Repository) toNullStringAlways(value string) null.String {
	return null.String{
		String: value,
		Valid:  true,
	}
}



func (r *Repository) buildORMAccount(user account_dto.Account) (*orm.Account, error) {
	// ✅ Add validation for required fields
	if user.Email == "" {
		return nil, error_custom.NewValidationError("account", "email", "Email is required", user.Email)
	}
	if user.Name == "" {
		return nil, error_custom.NewValidationError("account", "name", "Name is required", user.Name)
	}
	if user.Password == "" {
		return nil, error_custom.NewValidationError("account", "password", "Password is required", "[REDACTED]")
	}

	now := time.Now()
	return &orm.Account{
		BranchID:  r.toNullInt64(user.BranchID),
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password, // Should already be hashed
		Avatar:    r.toNullString(user.Avatar),
		Title:     r.toNullString(user.Title),
		Role:      string(user.Role),
		OwnerID:   r.toNullInt64(user.OwnerID),
		Status:    r.toNullString(string(user.Status)),
		CreatedAt: null.TimeFrom(now),
		UpdatedAt: null.TimeFrom(now),
	}, nil
}

func (r *Repository) handleValidationError(err error, operation, table string, operationCtx map[string]interface{}) error {
	operationCtx["error_type"] = "validation"
	
	// r.logger.LogDBOperation(operation, table, false, err, operationCtx)
	
	// Return the validation error directly since it's already properly typed
	return err
}

func (r *Repository) handleInsertError(err error, operation, table string, operationCtx map[string]interface{}, startTime time.Time) error {
	duration := time.Since(startTime)
	operationCtx["duration_ms"] = duration.Milliseconds()
	operationCtx["error_type"] = "database_insert"
	
	// r.logger.LogDBOperation(operation, table, false, err, operationCtx)
	
	return r.errorHandler.HandleDatabaseError(err, "account", table, operation, operationCtx)
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



// Build operation context helper (enhanced)
func (r *Repository) buildOperationContext(user account_dto.Account) map[string]interface{} {
	return r.layerContext.BuildOperationContext(core.OperationCreateUser, core.TableAccounts, core.FuncCreateUser, map[string]interface{}{
		core.FieldEmail:    maskEmail(user.Email),
		core.FieldRole:     string(user.Role),
		core.FieldBranchID: user.BranchID,
		core.FieldOwnerID:  user.OwnerID,
	})
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

func (r *Repository) wrapError(err error, operation, table string, context map[string]interface{}, startTime *time.Time) error {
	if err == nil {
		return nil
	}

	// Add timing information if provided
	if startTime != nil {
		context["duration_ms"] = time.Since(*startTime).Milliseconds()
	}

	// Add operation context
	context["operation"] = operation
	context["table"] = table
	context["layer"] = "repository"

	// Log the error
	// r.logger.LogDBOperation(operation, table, false, err, context)

	// Delegate to error handler with full context
	return r.errorHandler.HandleDatabaseError(err, "account", table, operation, context)
}

