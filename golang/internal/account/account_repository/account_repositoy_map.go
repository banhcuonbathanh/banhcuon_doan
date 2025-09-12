package account_repository

import (
	"english-ai-full/error_system"
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
// new 12341231231