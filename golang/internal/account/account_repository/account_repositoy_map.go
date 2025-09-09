package account_repository

import (
	"context"
	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"english-ai-full/orm"
	"time"

	"github.com/aarondl/null/v8"
"github.com/aarondl/sqlboiler/v4/queries/qm"

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



// Keep existing helper methods but simplify error handling in them
func (r *Repository) buildORMAccount(user account_dto.Account) (*orm.Account, error) {
	// Validation
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





// Enhanced email masking function (same as before but with better structure)

// 



func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const operation = "exists_by_email"
	
	r.logger.Debug("Checking if email exists", r.layerContext.MergeWithContext(map[string]interface{}{
		"email": utils.MaskEmail(email),
		"operation": operation,
	}))
	
	// Context timeout check
	if err := ctx.Err(); err != nil {
		return false, r.errorHandler.Handle(err, operation, "accounts", map[string]interface{}{
			"email": utils.MaskEmail(email),
		})
	}
	
	// Query database for existing email
	exists, err := orm.Accounts(qm.Where("email = ?", email)).Exists(ctx, r.db)
	if err != nil {
		r.logger.Error("Failed to check email existence", r.layerContext.MergeWithContext(map[string]interface{}{
			"email": utils.MaskEmail(email),
			"error": err.Error(),
		}))
		return false, r.errorHandler.Handle(err, operation, "accounts", map[string]interface{}{
			"email": utils.MaskEmail(email),
		})
	}
	
	r.logger.Debug("Email existence check completed", r.layerContext.MergeWithContext(map[string]interface{}{
		"email": utils.MaskEmail(email),
		"exists": exists,
	}))
	
	return exists, nil
}