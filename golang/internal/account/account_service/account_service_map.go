package account_service

import (
	"context"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/proto_qr/account"
	"fmt"

	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// mapDTOToProto converts DTO to Proto format
func (s *AccountService) mapDTOToProto(dto account_dto.Account) *account.Account {
	return &account.Account{
		Id:       dto.ID,
		BranchId: dto.BranchID,
		Name:     dto.Name,
		Email:    dto.Email,
		Avatar:   dto.Avatar,
		Title:    dto.Title,
		Role:     string(dto.Role),
		OwnerId:  dto.OwnerID,
		Status:   s.mapStatusToProtoEnum(dto.Status),
		CreatedAt: func() *timestamppb.Timestamp {
			if !dto.CreatedAt.IsZero() {
				return timestamppb.New(dto.CreatedAt)
			}
			return nil
		}(),
		UpdatedAt: func() *timestamppb.Timestamp {
			if !dto.UpdatedAt.IsZero() {
				return timestamppb.New(dto.UpdatedAt)
			}
			return nil
		}(),
	}
}

// mapStatusToProtoEnum converts status string to AccountStatus enum
func (s *AccountService) mapStatusToProtoEnum(status string) account.AccountStatus {
	switch strings.ToLower(status) {
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

// mapProtoToDTO converts Proto to DTO format
func (s *AccountService) mapProtoToDTO(proto *account.Account) account_dto.Account {
	return account_dto.Account{
		ID:       proto.Id,
		BranchID: proto.BranchId,
		Name:     proto.Name,
		Email:    proto.Email,
		Avatar:   proto.Avatar,
		Title:    proto.Title,
		Role:     account_dto.Role(proto.Role),
		OwnerID:  proto.OwnerId,
		Status:   s.mapProtoEnumToStatus(proto.Status),
		CreatedAt: func() time.Time {
			if proto.CreatedAt != nil {
				return proto.CreatedAt.AsTime()
			}
			return time.Time{}
		}(),
		UpdatedAt: func() time.Time {
			if proto.UpdatedAt != nil {
				return proto.UpdatedAt.AsTime()
			}
			return time.Time{}
		}(),
	}
}

// mapProtoEnumToStatus converts AccountStatus enum to status string
func (s *AccountService) mapProtoEnumToStatus(status account.AccountStatus) string {
	switch status {
	case account.AccountStatus_ACTIVE:
		return "active"
	case account.AccountStatus_INACTIVE:
		return "inactive"
	case account.AccountStatus_SUSPENDED:
		return "suspended"
	default:
		return "unknown"
	}
}



// ===== HELPER METHODS =====

// buildOperationContext creates standardized context for logging and error handling
func (s *AccountService) buildOperationContext(operation string, userInfo map[string]interface{}) map[string]interface{} {
	ctx := map[string]interface{}{
		"service":   "account",
		"operation": operation,
		"domain":    s.domain,
	}
	
	// Merge user-specific context
	for key, value := range userInfo {
		ctx[key] = value
	}
	
	return ctx
}

// validateUserInput performs comprehensive input validation


func (s *AccountService)  convertDTOToProto(dto *account_dto.Account) *account.Account {
	return &account.Account{
		Id:        dto.ID,
		BranchId:  dto.BranchID,
		Name:      dto.Name,
		Email:     dto.Email,
		Password:  dto.Password,
		Avatar:    dto.Avatar,
		Title:     dto.Title,
		Role:      string(dto.Role), 
		OwnerId:   dto.OwnerID,
		Status:    convertStatusToProto(dto.Status),
		CreatedAt: timestamppb.New(dto.CreatedAt),
		UpdatedAt: timestamppb.New(dto.UpdatedAt),
	}
}

func convertStatusToProto(status string) account.AccountStatus {
	switch strings.ToUpper(status) {
	case "ACTIVE":
		return account.AccountStatus_ACTIVE
	case "INACTIVE":
		return account.AccountStatus_INACTIVE
	case "SUSPENDED":
		return account.AccountStatus_SUSPENDED
	default:
		return account.AccountStatus_UNKNOWN
	}
}

func (s *AccountService) getRequestIDFromContext(ctx context.Context, req *account.AccountReq) string {
    // First try to get from context
    if requestID, ok := ctx.Value("request_id").(string); ok {
        return requestID
    }
    

    // Fallback: generate a new request ID
    return fmt.Sprintf("service_req_%d", time.Now().UnixNano())
}