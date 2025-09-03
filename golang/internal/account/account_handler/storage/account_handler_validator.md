package account_handler

import (
	"context"
	"time"

	pb "english-ai-full/internal/proto_qr/account"
	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidatePassword validates password using utils function with enhanced logging
func ValidatePassword(fl validator.FieldLevel) bool {
	// Create validation logger
	validationLog := logger.NewValidationLogger()
	validationLog.SetOperation("validate_password")

	fieldName := fl.FieldName()
	password := fl.Field().String()
	
	// Context for logging (password is masked for security)
	context := map[string]interface{}{
		"field":          fieldName,
		"password_length": len(password),
		"validation_type": "password_strength",
	}

	validationLog.Debug("Password validation started", context)

	// Perform validation using existing utility
	err := errorcustom.ValidatePassword(password)
	isValid := err == nil

	if !isValid {
		// Log validation failure with cause
		logger.LogValidationError(fieldName, "Password validation failed", "***masked***")
		
		logger.WarningWithCause(
			"Password validation failed",
			"weak_password",
			logger.LayerValidation,
			"validate_password_strength",
			utils.MergeContext(context, map[string]interface{}{
				"validation_error": err.Error(),
				"is_valid": false,
			}),
		)
	} else {
		validationLog.Debug("Password validation passed", utils.MergeContext(context, map[string]interface{}{
			"is_valid": true,
		}))
	}

	return isValid
}

// ValidateRole validates if the role is valid with enhanced logging
func ValidateRole(fl validator.FieldLevel) bool {
	// Create validation logger
	validationLog := logger.NewValidationLogger()
	validationLog.SetOperation("validate_role")

	fieldName := fl.FieldName()
	role := fl.Field().String()
	
	validRoles := map[string]bool{
		"admin":   true,
		"user":    true,
		"manager": true,
	}

	// Context for logging
	context := map[string]interface{}{
		"field":         fieldName,
		"role":          role,
		"validation_type": "role_enum",
		"valid_roles":   []string{"admin", "user", "manager"},
	}

	validationLog.Debug("Role validation started", context)

	isValid := validRoles[role]

	if !isValid {
		// Log validation failure
		logger.LogValidationError(fieldName, "Invalid role provided", role)
		
		logger.WarningWithCause(
			"Role validation failed",
			"invalid_role",
			logger.LayerValidation,
			"validate_role_enum",
			utils.MergeContext(context, map[string]interface{}{
				"is_valid": false,
				"allowed_values": []string{"admin", "user", "manager"},
			}),
		)
	} else {
		validationLog.Debug("Role validation passed", utils.MergeContext(context, map[string]interface{}{
			"is_valid": true,
		}))
	}

	return isValid
}

// ValidateEmailUnique checks if email is unique using the gRPC client with comprehensive logging
func ValidateEmailUnique(userClient pb.AccountServiceClient) validator.Func {
	return func(fl validator.FieldLevel) bool {
		start := time.Now()
		
		// Create validation logger
		validationLog := logger.NewValidationLogger()
		validationLog.SetOperation("validate_email_unique")

		fieldName := fl.FieldName()
		email := fl.Field().String()
		
		// Context for logging
		baseContext := map[string]interface{}{
			"field":           fieldName,
			"email":           email,
			"validation_type": "email_uniqueness",
		}

		validationLog.Debug("Email uniqueness validation started", baseContext)

		// Call service to check if email exists
		serviceStart := time.Now()
		_, err := userClient.FindByEmail(context.Background(), &pb.FindByEmailReq{
			Email: email,
		})
		serviceDuration := time.Since(serviceStart)

		// Log service call context
		serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
			"service":     "user-service",
			"method":      "FindByEmail",
			"duration_ms": serviceDuration.Milliseconds(),
		})

		if err == nil {
			// Email exists - validation failed
			logger.LogServiceCall("user-service", "FindByEmail", true, nil, serviceContext)
			
			// Log validation failure
			logger.LogValidationError(fieldName, "Email already exists", email)
			
			logger.WarningWithCause(
				"Email uniqueness validation failed - email already exists",
				"duplicate_email",
				logger.LayerValidation,
				"validate_email_uniqueness",
				utils.MergeContext(baseContext, map[string]interface{}{
					"is_valid": false,
					"reason": "email_already_registered",
				}),
			)

			return false // Email exists, not unique
		}

		// Check if error is "not found" (expected for unique email)
		if status.Code(err) == codes.NotFound {
			// Email is unique - validation passed
			logger.LogServiceCall("user-service", "FindByEmail", true, nil, utils.MergeContext(serviceContext, map[string]interface{}{
				"result": "email_not_found_unique",
			}))
			
			validationLog.Debug("Email uniqueness validation passed", utils.MergeContext(baseContext, map[string]interface{}{
				"is_valid": true,
				"reason": "email_not_found",
			}))

			// Log performance
			logger.LogPerformance("validate_email_unique", time.Since(start), baseContext)

			return true // Email is unique
		}

		// Unexpected error occurred
		logger.LogServiceCall("user-service", "FindByEmail", false, err, serviceContext)
		
		logger.ErrorWithCause(
			"Email uniqueness validation failed due to service error",
			"service_error",
			logger.LayerValidation,
			"validate_email_uniqueness",
			utils.MergeContext(baseContext, map[string]interface{}{
				"error": err.Error(),
				"grpc_code": status.Code(err).String(),
				"is_valid": true, // Default to true on service error to not block registration
				"fallback_behavior": "allow_on_service_error",
			}),
		)

		// Log performance even on error
		logger.LogPerformance("validate_email_unique", time.Since(start), baseContext)

		// Return true on service error to avoid blocking legitimate registrations
		// This is a business decision - could be changed based on requirements
		return true
	}
}

// ValidateEmailFormat validates email format with enhanced logging
func ValidateEmailFormat(fl validator.FieldLevel) bool {
	// Create validation logger
	validationLog := logger.NewValidationLogger()
	validationLog.SetOperation("validate_email_format")

	fieldName := fl.FieldName()
	email := fl.Field().String()
	
	// Context for logging
	context := map[string]interface{}{
		"field":           fieldName,
		"email":           email,
		"validation_type": "email_format",
	}

	validationLog.Debug("Email format validation started", context)

	// Use built-in email validation or custom logic
	isValid := errorcustom.IsValidEmail(email)

	if !isValid {
		// Log validation failure
		logger.LogValidationError(fieldName, "Invalid email format", email)
		
		logger.WarningWithCause(
			"Email format validation failed",
			"invalid_email_format",
			logger.LayerValidation,
			"validate_email_format",
			utils.MergeContext(context, map[string]interface{}{
				"is_valid": false,
				"expected_format": "user@domain.com",
			}),
		)
	} else {
		validationLog.Debug("Email format validation passed", utils.MergeContext(context, map[string]interface{}{
			"is_valid": true,
		}))
	}

	return isValid
}

// ValidateBranchID validates if branch ID exists with service call logging
func ValidateBranchID(userClient pb.AccountServiceClient) validator.Func {
	return func(fl validator.FieldLevel) bool {
		start := time.Now()
		
		// Create validation logger
		validationLog := logger.NewValidationLogger()
		validationLog.SetOperation("validate_branch_id")

		fieldName := fl.FieldName()
		branchID := fl.Field().Int()
		
		// Context for logging
		baseContext := map[string]interface{}{
			"field":           fieldName,
			"branch_id":       branchID,
			"validation_type": "branch_existence",
		}

		validationLog.Debug("Branch ID validation started", baseContext)

		// Skip validation for zero value (optional field)
		if branchID == 0 {
			validationLog.Debug("Branch ID validation skipped - zero value", utils.MergeContext(baseContext, map[string]interface{}{
				"is_valid": true,
				"reason": "zero_value_allowed",
			}))
			return true
		}

		// Call service to check if branch exists (assuming there's a FindBranchByID method)
		serviceStart := time.Now()
		_, err := userClient.FindByBranch(context.Background(), &pb.FindByBranchReq{
			BranchId: branchID,
		})
		serviceDuration := time.Since(serviceStart)

		// Log service call context
		serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
			"service":     "user-service",
			"method":      "FindByBranch",
			"duration_ms": serviceDuration.Milliseconds(),
		})

		if err != nil && status.Code(err) == codes.NotFound {
			// Branch not found - validation failed
			logger.LogServiceCall("user-service", "FindByBranch", true, nil, serviceContext)
			
			logger.LogValidationError(fieldName, "Branch ID not found", branchID)
			
			logger.WarningWithCause(
				"Branch ID validation failed - branch not found",
				"branch_not_found",
				logger.LayerValidation,
				"validate_branch_existence",
				utils.MergeContext(baseContext, map[string]interface{}{
					"is_valid": false,
					"reason": "branch_not_exists",
				}),
			)

			logger.LogPerformance("validate_branch_id", time.Since(start), baseContext)
			return false
		}

		if err != nil {
			// Unexpected service error
			logger.LogServiceCall("user-service", "FindByBranch", false, err, serviceContext)
			
			logger.ErrorWithCause(
				"Branch ID validation failed due to service error",
				"service_error",
				logger.LayerValidation,
				"validate_branch_existence",
				utils.MergeContext(baseContext, map[string]interface{}{
					"error": err.Error(),
					"grpc_code": status.Code(err).String(),
					"is_valid": true, // Default to true on service error
					"fallback_behavior": "allow_on_service_error",
				}),
			)

			logger.LogPerformance("validate_branch_id", time.Since(start), baseContext)
			return true // Allow on service error
		}

		// Branch exists - validation passed
		logger.LogServiceCall("user-service", "FindByBranch", true, nil, serviceContext)
		
		validationLog.Debug("Branch ID validation passed", utils.MergeContext(baseContext, map[string]interface{}{
			"is_valid": true,
			"reason": "branch_exists",
		}))

		logger.LogPerformance("validate_branch_id", time.Since(start), baseContext)
		return true
	}
}

// ValidateUserID validates if user ID exists with service call logging
func ValidateUserID(userClient pb.AccountServiceClient) validator.Func {
	return func(fl validator.FieldLevel) bool {
		start := time.Now()
		
		// Create validation logger
		validationLog := logger.NewValidationLogger()
		validationLog.SetOperation("validate_user_id")

		fieldName := fl.FieldName()
		userID := fl.Field().Int()
		
		// Context for logging
		baseContext := map[string]interface{}{
			"field":           fieldName,
			"user_id":         userID,
			"validation_type": "user_existence",
		}

		validationLog.Debug("User ID validation started", baseContext)

		// Skip validation for zero value (optional field)
		if userID == 0 {
			validationLog.Debug("User ID validation skipped - zero value", utils.MergeContext(baseContext, map[string]interface{}{
				"is_valid": true,
				"reason": "zero_value_allowed",
			}))
			return true
		}

		// Call service to check if user exists
		serviceStart := time.Now()
		_, err := userClient.FindByID(context.Background(), &pb.FindByIDReq{
			Id: userID,
		})
		serviceDuration := time.Since(serviceStart)

		// Log service call context
		serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
			"service":     "user-service",
			"method":      "FindByID",
			"duration_ms": serviceDuration.Milliseconds(),
		})

		if err != nil && status.Code(err) == codes.NotFound {
			// User not found - validation failed
			logger.LogServiceCall("user-service", "FindByID", true, nil, serviceContext)
			
			logger.LogValidationError(fieldName, "User ID not found", userID)
			
			logger.WarningWithCause(
				"User ID validation failed - user not found",
				"user_not_found",
				logger.LayerValidation,
				"validate_user_existence",
				utils.MergeContext(baseContext, map[string]interface{}{
					"is_valid": false,
					"reason": "user_not_exists",
				}),
			)

			logger.LogPerformance("validate_user_id", time.Since(start), baseContext)
			return false
		}

		if err != nil {
			// Unexpected service error
			logger.LogServiceCall("user-service", "FindByID", false, err, serviceContext)
			
			logger.ErrorWithCause(
				"User ID validation failed due to service error",
				"service_error",
				logger.LayerValidation,
				"validate_user_existence",
				utils.MergeContext(baseContext, map[string]interface{}{
					"error": err.Error(),
					"grpc_code": status.Code(err).String(),
					"is_valid": true, // Default to true on service error
					"fallback_behavior": "allow_on_service_error",
				}),
			)

			logger.LogPerformance("validate_user_id", time.Since(start), baseContext)
			return true // Allow on service error
		}

		// User exists - validation passed
		logger.LogServiceCall("user-service", "FindByID", true, nil, serviceContext)
		
		validationLog.Debug("User ID validation passed", utils.MergeContext(baseContext, map[string]interface{}{
			"is_valid": true,
			"reason": "user_exists",
		}))

		logger.LogPerformance("validate_user_id", time.Since(start), baseContext)
		return true
	}
}