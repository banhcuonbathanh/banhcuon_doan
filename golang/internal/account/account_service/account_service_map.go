package account_service

import (

	errorcustom "english-ai-full/error_custom"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/proto_qr/account"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator"
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


// handleServiceError wraps errors with service context and logs them
func (s *AccountService) handleServiceError(err error, operation string, context map[string]interface{}, startTime *time.Time) error {
	if err == nil {
		return nil
	}

	// Add timing information if provided
	if startTime != nil {
		context["duration_ms"] = time.Since(*startTime).Milliseconds()
	}

	// Add service context
	context["service"] = "account"
	context["operation"] = operation
	context["layer"] = "service"

	// Log the error
	s.logger.LogServiceCall("account", operation, false, err, context)

	// Convert to APIError and add service layer context
	apiErr := errorcustom.ConvertToAPIError(err)
	if apiErr == nil {
		apiErr = errorcustom.NewAPIError(
			"ACCOUNT_SERVICE_ERROR",
			"Account service operation failed",
			500,
		)
	}

	// Add service layer context
	apiErr.WithDomain(s.domain).
		WithLayer("service").
		WithOperation(operation).
		WithCause(err)

	// Add additional context
	for k, v := range context {
		apiErr.WithDetail(k, v)
	}

	return apiErr
}
// handleServiceSuccess logs successful operations
func (s *AccountService) handleServiceSuccess(operation string, context map[string]interface{}, startTime time.Time) {
	context["duration_ms"] = time.Since(startTime).Milliseconds()
	context["success"] = true
	s.logger.LogServiceCall("account", operation, true, nil, context)
}

func (s *AccountService) formatValidationError(err error, email string) error {
	var validationErrors []string
	
	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			validationErrors = append(validationErrors, fmt.Sprintf("%s is required", err.Field()))
		case "email":
			validationErrors = append(validationErrors, "invalid email format")
		case "min":
			validationErrors = append(validationErrors, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
		case "oneof":
			validationErrors = append(validationErrors, fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param()))
		default:
			validationErrors = append(validationErrors, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	
	return errorcustom.NewValidationError("account", "input", strings.Join(validationErrors, "; "), email)
}

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