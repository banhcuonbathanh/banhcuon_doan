package account_handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

// ============================================================================
// REQUEST VALIDATION
// ============================================================================

// validateRequest performs comprehensive domain-aware request validation
func (h *BaseAccountHandler) validateRequest(request interface{}, operation string, r *http.Request) error {
	start := time.Now()
	opCtx := h.setupOperationContext(r, "validate_request")
	opCtx.Context["target_operation"] = operation
	opCtx.Context["request_type"] = utils.GetTypeName(request)
	
	h.logger.Debug("Starting domain-aware request validation", opCtx.Context)
	
	if err := h.validator.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorCollection := errorcustom.NewErrorCollection(h.domain)
			
			for _, ve := range validationErrors {
				validationErr := errorcustom.NewValidationError(
					h.domain,
					ve.Field(),
					h.getValidationMessage(ve),
					utils.MaskSensitiveValue(ve.Field(), ve.Value()),
				)
				errorCollection.Add(validationErr)
				
				logger.LogValidationError(ve.Field(), ve.Tag(), ve.Value())
			}
			
			h.logger.Warning("Domain-aware request validation failed", utils.MergeContext(opCtx.Context, map[string]interface{}{
				"error_count": len(validationErrors),
				"duration_ms": time.Since(start).Milliseconds(),
			}))
			
			return errorCollection.ToAPIError()
		} else {
			h.logger.Error("Unexpected validation system error in domain", utils.MergeContext(opCtx.Context, map[string]interface{}{
				"error":       err.Error(),
				"duration_ms": time.Since(start).Milliseconds(),
			}))
			
			return errorcustom.NewSystemError(
				h.domain,
				"validator",
				"struct_validation",
				"Validation system error",
				err,
			)
		}
	}
	
	h.logger.Debug("Domain-aware request validation completed successfully", utils.MergeContext(opCtx.Context, map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}))
	
	logger.LogPerformance("request_validation", time.Since(start), opCtx.Context)
	
	return nil
}

// validateAndDecodeRequest combines JSON decoding with domain-aware validation
func (h *BaseAccountHandler) validateAndDecodeRequest(r *http.Request, target interface{}, operation string) error {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	// Decode JSON with domain context
	if err := errorcustom.DecodeJSONWithDomain(r.Body, target, h.domain, requestID); err != nil {
		return err
	}
	
	// Validate with domain context
	return h.validateRequest(target, operation, r)
}

// ============================================================================
// BUSINESS RULE VALIDATION
// ============================================================================

// validateBusinessRules validates domain-specific business rules
func (h *BaseAccountHandler) validateBusinessRules(operation string, context map[string]interface{}) error {
	switch operation {
	case "user_registration":
		return h.validateUserRegistrationRules(context)
	case "user_login":
		return h.validateUserLoginRules(context)
	case "user_update":
		return h.validateUserUpdateRules(context)
	default:
		return nil
	}
}

// validateUserRegistrationRules applies user registration business rules
func (h *BaseAccountHandler) validateUserRegistrationRules(context map[string]interface{}) error {
	errorCollection := errorcustom.NewErrorCollection(h.domain)
	
	// Extract data from context
	email, _ := context["email"].(string)
	password, _ := context["password"].(string)
	
	// Validate email requirements
	if err := h.validateEmailRequirements(email); err != nil {
		errorCollection.Add(err)
	}
	
	// Validate password requirements
	if err := h.validatePasswordRequirements(password); err != nil {
		errorCollection.Add(err)
	}
	
	// Validate business domain rules
	if err := h.validateBusinessDomainRules(context); err != nil {
		errorCollection.Add(err)
	}
	
	if errorCollection.HasErrors() {
		return errorCollection.ToAPIError()
	}
	
	return nil
}

// validateEmailRequirements validates email-specific business rules
func (h *BaseAccountHandler) validateEmailRequirements(email string) error {
	// Basic email format validation
	if err := errorcustom.ValidateEmailWithDomain(email, h.domain, ""); err != nil {
		return err
	}
	
	// Check for business email domains if required
	if h.isBusinessEmailRequired() && !h.isBusinessEmail(email) {
		return errorcustom.NewBusinessLogicErrorWithContext(
			h.domain,
			"business_email_required",
			"Registration requires a business email address",
			map[string]interface{}{
				"email":            email,
				"business_domains": h.getAllowedBusinessDomains(),
			},
		)
	}
	
	return nil
}

// validatePasswordRequirements validates password-specific business rules
func (h *BaseAccountHandler) validatePasswordRequirements(password string) error {
	// Use enhanced password validation with domain context
	if err := errorcustom.ValidatePasswordWithDomain(password, h.domain, ""); err != nil {
		return err
	}
	
	// Additional password complexity checks if required
	if h.config.IsPasswordComplexityRequired() {
		if !h.isPasswordComplex(password) {
			return errorcustom.NewValidationError(
				h.domain,
				"password",
				"Password does not meet complexity requirements",
				"[MASKED]",
			)
		}
	}
	
	return nil
}

// validateBusinessDomainRules validates domain-specific business rules
func (h *BaseAccountHandler) validateBusinessDomainRules(context map[string]interface{}) error {
	// Check if email verification is required
	if h.config.IsEmailVerificationRequired() {
		context["email_verification_required"] = true
	}
	
	// Add any other business domain rules here
	// For example: organization limits, subscription checks, etc.
	
	return nil
}

// validateUserLoginRules applies user login business rules
func (h *BaseAccountHandler) validateUserLoginRules(context map[string]interface{}) error {
	email, _ := context["email"].(string)
	
	// Check max login attempts from configuration
	maxAttempts := h.config.GetMaxLoginAttempts()
	currentAttempts, _ := context["failed_attempts"].(int)
	
	if currentAttempts >= maxAttempts {
		return errorcustom.NewAccountLockedError(email, "max_login_attempts_exceeded")
	}
	
	return nil
}

// validateUserUpdateRules applies user update business rules
func (h *BaseAccountHandler) validateUserUpdateRules(context map[string]interface{}) error {
	// Add any user update specific business rules here
	return nil
}

// ============================================================================
// FIELD-SPECIFIC VALIDATION
// ============================================================================

// validateEmailFormat validates email format with domain context
func (h *BaseAccountHandler) validateEmailFormat(email string) error {
	return errorcustom.ValidateEmailWithDomain(email, h.domain, "")
}

// validatePasswordStrength validates password strength with domain context
func (h *BaseAccountHandler) validatePasswordStrength(password string) error {
	return errorcustom.ValidatePasswordWithDomain(password, h.domain, "")
}

// validateUserRole validates if the role is allowed for this domain
func (h *BaseAccountHandler) validateUserRole(role string) error {
	validRoles := []string{"admin", "teacher", "student"}
	
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}
	
	return errorcustom.NewValidationError(
		h.domain,
		"role",
		"Invalid user role. Must be: admin, teacher, or student",
		role,
	)
}

// ============================================================================
// CUSTOM VALIDATOR FUNCTIONS
// ============================================================================

// ValidatePasswordWithDomain validates password with domain-specific requirements
func ValidatePasswordWithDomain(fl validator.FieldLevel, domain string) bool {
	password := fl.Field().String()
	
	if err := errorcustom.ValidatePasswordWithDomain(password, domain, ""); err != nil {
		return false
	}
	
	return true
}

// ValidateEmailUniqueWithDomain validates email uniqueness with domain context
func ValidateEmailUniqueWithDomain(client pb.AccountServiceClient, domain string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		
		// First validate email format with domain context
		if err := errorcustom.ValidateEmailWithDomain(email, domain, ""); err != nil {
			return false
		}
		
		// Check uniqueness via gRPC using FindByEmail
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		req := &pb.FindByEmailReq{Email: email}
		resp, err := client.FindByEmail(ctx, req)
		
		if err != nil {
			// If error is "not found", email is unique (good)
			if strings.Contains(err.Error(), "not found") || 
			   strings.Contains(err.Error(), "no rows") {
				return true
			}
			
			// Log other gRPC errors but don't fail validation here
			logger.Warning("Email uniqueness check failed", map[string]interface{}{
				"email":     email,
				"error":     err.Error(),
				"domain":    domain,
				"operation": "validate_email_unique",
			})
			return true // Allow validation to pass, will be caught later
		}
		
		// If we got a response with an account, email exists (not unique)
		if resp != nil && resp.Account != nil {
			return false
		}
		
		return true // Email is unique
	}
}

// ============================================================================
// VALIDATION MESSAGE GENERATION
// ============================================================================

// getValidationMessage returns domain-aware validation messages
func (h *BaseAccountHandler) getValidationMessage(fe validator.FieldError) string {
	switch h.domain {
	case "user":
		return h.getUserValidationMessage(fe)
	default:
		return h.getGenericValidationMessage(fe)
	}
}

// getUserValidationMessage returns user domain-specific validation messages
func (h *BaseAccountHandler) getUserValidationMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()
	
	switch tag {
	case "required":
		if field == "Email" {
			return "Email address is required for user registration"
		}
		if field == "Password" {
			return "Password is required for user account"
		}
		return fmt.Sprintf("%s is required", field)
		
	case "email":
		return "Please provide a valid email address"
		
	case "strongpassword":
		return "Password must contain at least 8 characters with uppercase, lowercase, numbers, and special characters"
		
	case "uniqueemail":
		return "An account with this email address already exists"
		
	case "userrole":
		return "Invalid user role. Must be: admin, teacher, or student"
		
	case "min":
		if field == "Password" {
			return fmt.Sprintf("Password must be at least %s characters long", param)
		}
		return fmt.Sprintf("%s must be at least %s characters", field, param)
		
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
		
	default:
		return h.getGenericValidationMessage(fe)
	}
}

// getGenericValidationMessage returns generic validation messages
func (h *BaseAccountHandler) getGenericValidationMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ============================================================================
// VALIDATION HELPER METHODS
// ============================================================================

// isPasswordComplex checks if password meets complexity requirements
func (h *BaseAccountHandler) isPasswordComplex(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// isBusinessEmailRequired checks configuration for business email requirement
func (h *BaseAccountHandler) isBusinessEmailRequired() bool {
	// This would be configured in your domain configuration
	// For now, return false unless specifically configured
	return false
}

// isBusinessEmail validates if email is from business domain
func (h *BaseAccountHandler) isBusinessEmail(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	
	businessDomains := h.getAllowedBusinessDomains()
	emailDomain := strings.Split(email, "@")[1]
	
	for _, domain := range businessDomains {
		if emailDomain == domain {
			return true
		}
	}
	
	return false
}

// getAllowedBusinessDomains gets business domains from configuration
func (h *BaseAccountHandler) getAllowedBusinessDomains() []string {
	// This would come from your configuration
	return []string{"company.com", "enterprise.com", "business.org"}
}

// checkEmailUniqueness validates email uniqueness via gRPC
func (h *BaseAccountHandler) checkEmailUniqueness(ctx context.Context, email string) error {
	req := &pb.FindByEmailReq{Email: email}
	resp, err := h.userClient.FindByEmail(ctx, req)
	
	if err != nil {
		// If user not found, email is unique (good)
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		
		// Handle other gRPC errors
		return h.errorHandler.HandleExternalServiceError(
			err, h.domain, "user_service", "check_email_uniqueness", true,
		)
	}
	
	// If we found a user, email is not unique
	if resp != nil && resp.Account != nil {
		return h.errorHandler.HandleBusinessRuleViolation(
			h.domain,
			"email_uniqueness",
			"An account with this email already exists",
			map[string]interface{}{
				"email":              email,
				"existing_account_id": resp.Account.Id,
			},
		)
	}
	
	return nil
}