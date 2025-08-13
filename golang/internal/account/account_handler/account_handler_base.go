package account_handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"

	utils_config "english-ai-full/utils/config"
)

type BaseAccountHandler struct {
	userClient   pb.AccountServiceClient
	validator    *validator.Validate
	logger       *logger.Logger
	config       *utils_config.Config
	domain       string
	errorHandler *utils_config.DomainErrorHandler

	requestID    string
}

// User represents a user entity for business logic operations


// NewBaseHandler creates a new base account handler with comprehensive domain-aware setup
func NewBaseHandler(userClient pb.AccountServiceClient, config *utils_config.Config) *BaseAccountHandler {
	domain := "account" // Account handlers primarily work with user domain
	
	// Create handler-specific logger with domain context
	handlerLogger := logger.NewHandlerLogger()
	handlerLogger.AddGlobalField("component", "account_handler")
	handlerLogger.AddGlobalField("layer", "handler")
	handlerLogger.AddGlobalField("domain", domain)
	
	// Initialize validator with custom validations
	v := validator.New()
	
	// Register domain-aware custom validation functions
	logger.Debug("Registering domain-aware validation functions", map[string]interface{}{
		"validations": []string{"password", "role", "uniqueemail"},
		"component":   "account_handler",
		"domain":      domain,
	})
	
	// Register custom validators with domain context
	v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		return ValidatePasswordWithDomain(fl, domain)
	})
	v.RegisterValidation("role", ValidateRole)
	v.RegisterValidation("uniqueemail", ValidateEmailUniqueWithDomain(userClient, domain))
	
	// Register the enhanced email format validator
	v.RegisterValidation("email", errorcustom.ValidateEmailFormat)

	// Create domain-aware error handler
	errorHandler := utils_config.NewDomainAwareErrorHandler(config)

	handler := &BaseAccountHandler{
		userClient:   userClient,
		validator:    v,
		logger:       handlerLogger,
		config:       config,
		domain:       domain,
		errorHandler: errorHandler,
	}

	logger.Info("Base account handler initialized successfully", map[string]interface{}{
		"component":            "account_handler",
		"domain":               domain,
		"validator_setup":      true,
		"custom_validations":   3,
		"grpc_client_ready":    userClient != nil,
		"config_loaded":        config != nil,
		"error_handler_ready":  errorHandler != nil,
	})

	return handler
}

// ============================================================================
// ENHANCED CONTEXT AND PARAMETER EXTRACTION
// ============================================================================

// getUserIDFromContext extracts user ID with domain-aware error handling
func (h *BaseAccountHandler) getUserIDFromContext(ctx context.Context) (int64, error) {
	start := time.Now()
	requestID := errorcustom.GetRequestIDFromContext(ctx)
	
	// Create context for logging
	logContext := map[string]interface{}{
		"operation":  "get_user_id_from_context",
		"component":  "account_handler",
		"domain":     h.domain,
		"request_id": requestID,
	}
	
	h.logger.Debug("Extracting user ID from context", logContext)
	
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		// Check for alternative context keys
		if userIDStr, exists := ctx.Value("user_id").(string); exists {
			if parsedID, parseErr := strconv.ParseInt(userIDStr, 10, 64); parseErr == nil {
				logger.Warning("User ID found as string in context, converting to int64", utils.MergeContext(logContext, map[string]interface{}{
					"user_id_string": userIDStr,
					"user_id_int64":  parsedID,
					"duration_ms":    time.Since(start).Milliseconds(),
				}))
				
				return parsedID, nil
			}
		}
		
		// Log the failed context extraction with available context keys
		contextKeys := h.getContextKeys(ctx)
		
		logger.ErrorWithCause(
			"User ID not found in request context",
			"missing_user_context",
			logger.LayerHandler,
			"extract_user_id",
			utils.MergeContext(logContext, map[string]interface{}{
				"available_context_keys": contextKeys,
				"expected_key":           "user_id",
				"expected_type":          "int64",
				"duration_ms":           time.Since(start).Milliseconds(),
			}),
		)
		
		// Log security event for unauthorized access attempt
		logger.LogSecurityEvent(
			"unauthorized_context_access",
			"Request made without valid user context",
			"medium",
			utils.MergeContext(logContext, map[string]interface{}{
				"missing_context": "user_id",
				"auth_required":   true,
			}),
		)
		
		// Return domain-aware authentication error
		return 0, errorcustom.NewAuthenticationError(h.domain, "user context not found")
	}
	
	// Log successful context extraction
	h.logger.Debug("User ID successfully extracted from context", utils.MergeContext(logContext, map[string]interface{}{
		"user_id":     userID,
		"duration_ms": time.Since(start).Milliseconds(),
	}))
	
	return userID, nil
}

// getPaginationParams extracts pagination with domain-aware validation
func (h *BaseAccountHandler) getPaginationParams(r *http.Request) (page, pageSize int32, err error) {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	// Use the enhanced pagination utility with domain context
	limit, offset, err := errorcustom.GetPaginationParamsWithDomain(r, h.domain)
	if err != nil {
		return 0, 0, err
	}
	
	// Convert offset-based to page-based pagination
	page = int32((offset / limit) + 1)
	pageSize = int32(limit)
	
	// Log pagination parameters
	h.logger.Debug("Pagination parameters processed", map[string]interface{}{
		"page":       page,
		"page_size":  pageSize,
		"limit":      limit,
		"offset":     offset,
		"domain":     h.domain,
		"request_id": requestID,
	})
	
	return page, pageSize, nil
}

// getSortingParams extracts sorting parameters with domain-aware validation
func (h *BaseAccountHandler) getSortingParams(r *http.Request, allowedFields []string) (sortBy, sortOrder string, err error) {
	return errorcustom.GetSortParamsWithDomain(r, allowedFields, h.domain)
}

// ============================================================================
// ENHANCED REQUEST VALIDATION
// ============================================================================

// validateRequest performs comprehensive domain-aware request validation
func (h *BaseAccountHandler) validateRequest(request interface{}, operation string, r *http.Request) error {
	start := time.Now()
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	context := map[string]interface{}{
		"operation":        "validate_request",
		"target_operation": operation,
		"request_type":     utils.GetTypeName(request),
		"domain":           h.domain,
		"request_id":       requestID,
	}
	
	h.logger.Debug("Starting domain-aware request validation", context)
	
	if err := h.validator.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Create error collection for multiple validation errors
			errorCollection := errorcustom.NewErrorCollection(h.domain)
			
			for _, ve := range validationErrors {
				validationErr := errorcustom.NewValidationError(
					h.domain,
					ve.Field(),
					h.getValidationMessage(ve),
					utils.MaskSensitiveValue(ve.Field(), ve.Value()),
				)
				errorCollection.Add(validationErr)
				
				// Log individual validation error with domain context
				logger.LogValidationError(
					ve.Field(),
					ve.Tag(),
					ve.Value(),
				)
			}
			
			logger.WarningWithCause(
				"Domain-aware request validation failed",
				"validation_failed",
				logger.LayerHandler,
				operation,
				utils.MergeContext(context, map[string]interface{}{
					"error_count": len(validationErrors),
					"duration_ms": time.Since(start).Milliseconds(),
				}),
			)
			
			return errorCollection.ToAPIError()
		} else {
			// Handle unexpected validation error with domain context
			logger.ErrorWithCause(
				"Unexpected validation system error in domain",
				"validation_system_error",
				logger.LayerHandler,
				operation,
				utils.MergeContext(context, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(start).Milliseconds(),
				}),
			)
			
			// Return domain-aware system error
			systemErr := errorcustom.NewSystemError(
				h.domain,
				"validator",
				"struct_validation",
				"Validation system error",
				err,
			)
			return systemErr
		}
	}
	
	// Log successful validation
	h.logger.Debug("Domain-aware request validation completed successfully", utils.MergeContext(context, map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}))
	
	// Log performance metrics
	logger.LogPerformance("request_validation", time.Since(start), context)
	
	return nil
}

// ============================================================================
// GRPC ERROR HANDLING INTEGRATION
// ============================================================================

// handleGRPCError converts gRPC errors to domain-aware API errors
func (h *BaseAccountHandler) handleGRPCError(err error, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	
	// Use the enhanced gRPC error parser with domain context
	domainErr := errorcustom.ParseGRPCError(err, h.domain, operation, context)
	
	// Apply domain-specific error handling through configuration
	processedErr := h.errorHandler.HandleError(h.domain, domainErr)
	
	return processedErr
}

// ============================================================================
// ENHANCED CUSTOM VALIDATORS WITH DOMAIN SUPPORT
// ============================================================================

// ValidatePasswordWithDomain validates password with domain-specific requirements
func ValidatePasswordWithDomain(fl validator.FieldLevel, domain string) bool {
	password := fl.Field().String()
	
	// Use the enhanced password validation with domain context
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

// ValidateRole validates user roles (keeping existing logic)

// ============================================================================
// ENHANCED HELPER METHODS
// ============================================================================

// handleDomainError handles errors with full domain context and configuration
func (h *BaseAccountHandler) handleDomainError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	// Process error through domain-aware error handler
	processedErr := h.errorHandler.HandleError(h.domain, err)
	
	// Handle the processed error with domain context
	errorcustom.HandleDomainError(w, processedErr, h.domain, requestID)
}

// validateAndDecodeRequest combines JSON decoding with domain-aware validation
func (h *BaseAccountHandler) validateAndDecodeRequest(r *http.Request, target interface{}, operation string) error {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	// Decode JSON with domain context
	if err := errorcustom.DecodeJSONWithDomain(r.Body, target, h.domain, requestID); err != nil {
		return err
	}
	
	// Validate with domain context
	if err := h.validator.Struct(target); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Create error collection for multiple validation errors
			errorCollection := errorcustom.NewErrorCollection(h.domain)
			
			for _, ve := range validationErrors {
				validationErr := errorcustom.NewValidationError(
					h.domain,
					ve.Field(),
					h.getValidationMessage(ve),
					utils.MaskSensitiveValue(ve.Field(), ve.Value()),
				)
				errorCollection.Add(validationErr)
			}
			
			return errorCollection.ToAPIError()
		}
		
		// System validation error
		return errorcustom.NewSystemError(
			h.domain,
			"validator",
			"struct_validation",
			"Validation system error",
			err,
		)
	}
	
	return nil
}

// getValidationMessage returns domain-aware validation messages
func (h *BaseAccountHandler) getValidationMessage(fe validator.FieldError) string {

	
	// Domain-specific validation messages
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
		
	case "password":
		return "Password must meet security requirements"
		
	case "uniqueemail":
		return "An account with this email address already exists"
		
	case "role":
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
// DOMAIN-AWARE REQUEST PARSING
// ============================================================================

// parseIDParam safely parses ID parameters with domain validation
func (h *BaseAccountHandler) parseIDParam(r *http.Request, paramName string) (int64, error) {
	return errorcustom.ParseIDParamWithDomain(r, paramName, h.domain)
}

// parseStringParam safely parses string parameters with domain validation
func (h *BaseAccountHandler) parseStringParam(r *http.Request, paramName string, minLen int) (string, error) {
	return errorcustom.GetStringParamWithDomain(r, paramName, h.domain, minLen)
}

// ============================================================================
// ENHANCED LOGGING METHODS
// ============================================================================

// logHandlerStart logs the beginning of a handler operation with domain context
func (h *BaseAccountHandler) logHandlerStart(r *http.Request, operation string, additionalContext map[string]interface{}) map[string]interface{} {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	clientIP := errorcustom.GetClientIP(r)
	
	baseContext := utils.CreateBaseContext(r, utils.MergeContext(additionalContext, map[string]interface{}{
		"operation":          operation,
		"component":          "account_handler",
		"domain":             h.domain,
		"handler_start_time": time.Now(),
		"request_id":         requestID,
		"client_ip":          clientIP,
	}))
	
	logger.InfoWithOperation(
		"Domain handler operation started",
		logger.LayerHandler,
		operation,
		baseContext,
	)
	
	return baseContext
}

// logHandlerEnd logs the completion with enhanced domain context
func (h *BaseAccountHandler) logHandlerEnd(r *http.Request, operation string, statusCode int, startTime time.Time, additionalContext map[string]interface{}) {
	duration := time.Since(startTime)
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	context := utils.CreateBaseContext(r, utils.MergeContext(additionalContext, map[string]interface{}{
		"operation":   operation,
		"component":   "account_handler",
		"domain":      h.domain,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"request_id":  requestID,
	}))
	
	// Log API request completion with domain context
	logger.LogAPIRequest(r.Method, r.URL.Path, statusCode, duration, context)
	
	// Log performance metrics with domain tagging
	logger.LogPerformance(fmt.Sprintf("%s_%s", h.domain, operation), duration, context)
	
	// Enhanced status-based logging
	if statusCode >= 200 && statusCode < 300 {
		logger.InfoWithOperation(
			"Domain handler operation completed successfully",
			logger.LayerHandler,
			operation,
			context,
		)
	} else if statusCode >= 400 && statusCode < 500 {
		logger.WarningWithCause(
			"Domain handler operation completed with client error",
			"client_error",
			logger.LayerHandler,
			operation,
			context,
		)
	} else if statusCode >= 500 {
		logger.ErrorWithCause(
			"Domain handler operation completed with server error",
			"server_error",
			logger.LayerHandler,
			operation,
			context,
		)
		
		// Alert on server errors in production
		if h.config.IsProduction() {
			h.alertOpsTeam(operation, statusCode, context)
		}
	}
}

// ============================================================================
// ENHANCED RESPONSE METHODS
// ============================================================================

// respondWithSuccess sends successful response with domain context
func (h *BaseAccountHandler) respondWithSuccess(w http.ResponseWriter, r *http.Request, data interface{}, operation string) {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	errorcustom.RespondWithDomainSuccess(w, data, h.domain, requestID)
}

// respondWithError sends error response with domain context
func (h *BaseAccountHandler) respondWithError(w http.ResponseWriter, r *http.Request, err error, operation string) {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	// Process error through domain handler
	processedErr := h.errorHandler.HandleError(h.domain, err)
	
	// Handle with domain context
	errorcustom.HandleDomainError(w, processedErr, h.domain, requestID)
}

// ============================================================================
// BUSINESS LOGIC HELPERS WITH DOMAIN AWARENESS
// ============================================================================

// checkUserPermissions validates user permissions with domain context
func (h *BaseAccountHandler) checkUserPermissions(userID int64, requiredRole string, resource string) error {
	// Get user from context or service
	user, err := h.getUserByID(userID)
	if err != nil {
		return err
	}
	
	// Check role-based access
	if !h.hasRole(user.Role, requiredRole) {
		return errorcustom.NewAuthorizationErrorWithContext(
			h.domain,
			"access",
			resource,
			map[string]interface{}{
				"user_id":       userID,
				"required_role": requiredRole,
				"current_role":  user.Role,
				"resource":      resource,
			},
		)
	}
	
	return nil
}

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
	
	// Check if email verification is required
	if h.config.IsEmailVerificationRequired() {
		// Add email verification requirement to context
		context["email_verification_required"] = true
	}
	
	// Check password complexity requirements
	if h.config.IsPasswordComplexityRequired() {
		if password, ok := context["password"].(string); ok {
			if err := errorcustom.ValidatePasswordWithDomain(password, h.domain, ""); err != nil {
				errorCollection.Add(err)
			}
		}
	}
	
	// Additional domain-specific business rules
	if email, ok := context["email"].(string); ok {
		// Check for business email domains if required
		if h.isBusinessEmailRequired() && !h.isBusinessEmail(email) {
			businessErr := errorcustom.NewBusinessLogicErrorWithContext(
				h.domain,
				"business_email_required",
				"Registration requires a business email address",
				map[string]interface{}{
					"email":            email,
					"business_domains": h.getAllowedBusinessDomains(),
				},
			)
			errorCollection.Add(businessErr)
		}
	}
	
	if errorCollection.HasErrors() {
		return errorCollection.ToAPIError()
	}
	
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
// UTILITY METHODS
// ============================================================================

// getUserByID fetches user with domain-aware error handling
func (h *BaseAccountHandler) getUserByID(userID int64) (*pb.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req := &pb.FindByIDReq{Id: userID}
	resp, err := h.userClient.FindByID(ctx, req)
	
	if err != nil {
		return nil, h.handleGRPCError(err, "get_user", map[string]interface{}{
			"user_id": userID,
		})
	}
	
	// If you need to create a new Account (e.g., to exclude password)
	user := &pb.Account{
		Id:        resp.Account.Id,
		BranchId:  resp.Account.BranchId,
		Name:      resp.Account.Name,
		Email:     resp.Account.Email,
		// Password:  "", // Exclude password for security
		Avatar:    resp.Account.Avatar,
		Title:     resp.Account.Title,
		Role:      resp.Account.Role,
		OwnerId:   resp.Account.OwnerId,
		CreatedAt: resp.Account.CreatedAt,
		UpdatedAt: resp.Account.UpdatedAt,
	}
	
	return user, nil
}

// hasRole checks if user has required role
func (h *BaseAccountHandler) hasRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"student": 1,
		"teacher": 2,
		"admin":   3,
	}
	
	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]
	
	if !userExists || !requiredExists {
		return false
	}
	
	return userLevel >= requiredLevel
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

// alertOpsTeam sends alerts for critical errors in production
func (h *BaseAccountHandler) alertOpsTeam(operation string, statusCode int, context map[string]interface{}) {
	if !h.config.IsProduction() {
		return
	}
	
	alertContext := utils.MergeContext(context, map[string]interface{}{
		"alert_type":   "critical_error",
		"domain":       h.domain,
		"operation":    operation,
		"status_code":  statusCode,
		"timestamp":    time.Now().UTC(),
		"environment":  h.config.Environment,
	})
	
	logger.LogSecurityEvent(
		"critical_system_error",
		"Critical error in account handler",
		"high",
		alertContext,
	)
	
	// Here you would integrate with your alerting system
	// e.g., PagerDuty, Slack, email notifications
}

// getContextKeys extracts available context keys for debugging
func (h *BaseAccountHandler) getContextKeys(ctx context.Context) []string {
	keys := []string{}
	
	// Check for common context keys
	commonKeys := []string{
		"user_id", "user_email", "request_id", "trace_id", 
		"span_id", "tenant_id", "domain", "client_ip",
	}
	
	for _, key := range commonKeys {
		if value := ctx.Value(key); value != nil {
			keys = append(keys, key)
		}
	}
	
	return keys
}

// ============================================================================
// DOMAIN-AWARE ACCOUNT HANDLER
// ============================================================================




// ============================================================================
// INTERFACE COMPLIANCE AND ADDITIONAL HELPERS
// ============================================================================

// GetDomain returns the domain this handler operates in
func (h *BaseAccountHandler) GetDomain() string {
	return h.domain
}

// GetConfig returns the configuration instance
func (h *BaseAccountHandler) GetConfig() *utils_config.Config {
	return h.config
}

// GetErrorHandler returns the domain error handler
func (h *BaseAccountHandler) GetErrorHandler() *utils_config.DomainErrorHandler {
	return h.errorHandler
}

// GetLogger returns the logger instance
func (h *BaseAccountHandler) GetLogger() *logger.Logger {
	return h.logger
}

// SetLoggerOperation sets operation context for enhanced logging
func (h *BaseAccountHandler) SetLoggerOperation(operation string) {
	h.logger.SetOperation(operation)
}

// ============================================================================
// ENHANCED MIDDLEWARE AND REQUEST PROCESSING
// ============================================================================

// withRequestLogging wraps handler functions with comprehensive request logging
func (h *BaseAccountHandler) withRequestLogging(operation string, handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		
		// Start operation logging
		context := h.logHandlerStart(r, operation, map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		})
		
		// Create response writer wrapper to capture status code
		responseWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Execute handler
		handler(responseWriter, r)
		
		// End operation logging
		h.logHandlerEnd(r, operation, responseWriter.statusCode, startTime, context)
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture status codes
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// ============================================================================
// ADVANCED VALIDATION HELPERS
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
// CONTEXT MANAGEMENT HELPERS
// ============================================================================

// enrichContext adds common context information to request context
func (h *BaseAccountHandler) enrichContext(ctx context.Context, operation string, additionalData map[string]interface{}) context.Context {
	// Add operation context
	ctx = context.WithValue(ctx, "operation", operation)
	ctx = context.WithValue(ctx, "domain", h.domain)
	ctx = context.WithValue(ctx, "component", "account_handler")
	
	// Add additional data to context
	for key, value := range additionalData {
		ctx = context.WithValue(ctx, key, value)
	}
	
	return ctx
}

// extractUserContext extracts user-related information from context
func (h *BaseAccountHandler) extractUserContext(ctx context.Context) map[string]interface{} {
	userContext := make(map[string]interface{})
	
	if userID := ctx.Value("user_id"); userID != nil {
		userContext["user_id"] = userID
	}
	
	if userEmail := ctx.Value("user_email"); userEmail != nil {
		userContext["user_email"] = userEmail
	}
	
	if userRole := ctx.Value("user_role"); userRole != nil {
		userContext["user_role"] = userRole
	}
	
	return userContext
}

// ============================================================================
// PERFORMANCE AND MONITORING HELPERS
// ============================================================================

// measureOperation measures operation duration and logs performance metrics
func (h *BaseAccountHandler) measureOperation(operation string, fn func() error) error {
	start := time.Now()
	
	h.logger.Debug("Starting measured operation", map[string]interface{}{
		"operation": operation,
		"domain":    h.domain,
	})
	
	err := fn()
	duration := time.Since(start)
	
	// Log performance metrics
	logger.LogPerformance(fmt.Sprintf("%s_%s", h.domain, operation), duration, map[string]interface{}{
		"operation":   operation,
		"domain":      h.domain,
		"success":     err == nil,
		"duration_ms": duration.Milliseconds(),
	})
	
	if err != nil {
		h.logger.Warning("Measured operation completed with error", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
	} else {
		h.logger.Debug("Measured operation completed successfully", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"duration_ms": duration.Milliseconds(),
		})
	}
	
	return err
}

// ============================================================================
// SECURITY HELPERS
// ============================================================================

// validateRequestOrigin validates the origin of the request for security


// logSecurityEvent logs security-related events with domain context
func (h *BaseAccountHandler) logSecurityEvent(eventType string, description string, severity string, context map[string]interface{}) {
	securityContext := utils.MergeContext(context, map[string]interface{}{
		"domain":    h.domain,
		"component": "account_handler",
		"timestamp": time.Now().UTC(),
	})
	
	logger.LogSecurityEvent(eventType, description, severity, securityContext)
}

// ============================================================================
// CONFIGURATION ACCESS HELPERS
// ============================================================================

// getMaxLoginAttempts returns the maximum login attempts from configuration
func (h *BaseAccountHandler) getMaxLoginAttempts() int {
	if h.config != nil {
		return h.config.GetMaxLoginAttempts()
	}
	return 5 // Default fallback
}

// getSessionTimeout returns the session timeout from configuration


// isRateLimitEnabled checks if rate limiting is enabled
func (h *BaseAccountHandler) isRateLimitEnabled() bool {
	if h.config != nil {
		return h.config.IsRateLimitEnabled()
	}
	return false
}

// ============================================================================
// ERROR RECOVERY AND RESILIENCE
// ============================================================================

// withGRPCRetry executes gRPC operations with retry logic
func (h *BaseAccountHandler) withGRPCRetry(operation string, fn func() error) error {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		
		if err == nil {
			if attempt > 0 {
				h.logger.Info("gRPC operation succeeded after retry", map[string]interface{}{
					"operation": operation,
					"attempt":   attempt + 1,
					"domain":    h.domain,
				})
			}
			return nil
		}
		
		// Check if error is retryable
		if !errorcustom.IsRetryableError(err) {
			h.logger.Warning("gRPC operation failed with non-retryable error", map[string]interface{}{
				"operation": operation,
				"attempt":   attempt + 1,
				"error":     err.Error(),
				"domain":    h.domain,
			})
			return err
		}
		
		if attempt < maxRetries-1 {
			delay := time.Duration(attempt+1) * baseDelay
			h.logger.Warning("gRPC operation failed, retrying", map[string]interface{}{
				"operation":    operation,
				"attempt":      attempt + 1,
				"max_retries":  maxRetries,
				"retry_delay":  delay.String(),
				"error":        err.Error(),
				"domain":       h.domain,
			})
			
			time.Sleep(delay)
		} else {
			h.logger.Error("gRPC operation failed after all retries", map[string]interface{}{
				"operation":   operation,
				"attempts":    maxRetries,
				"final_error": err.Error(),
				"domain":      h.domain,
			})
		}
	}
	
	return fmt.Errorf("operation %s failed after %d attempts", operation, maxRetries)
}

// ============================================================================
// HEALTH CHECK AND DIAGNOSTICS
// ============================================================================


// getDiagnostics returns diagnostic information about the handler
func (h *BaseAccountHandler) getDiagnostics() map[string]interface{} {
	return map[string]interface{}{
		"domain":                h.domain,
		"component":            "account_handler",
		"grpc_client_ready":    h.userClient != nil,
		"validator_ready":      h.validator != nil,
		"logger_ready":         h.logger != nil,
		"config_ready":         h.config != nil,
		"error_handler_ready":  h.errorHandler != nil,
		"environment":          h.config.Environment,
		"initialized_at":       time.Now().UTC(),
	}
}

func (h *BaseAccountHandler) WithRequestID(r *http.Request) *BaseAccountHandler {
	h.requestID = errorcustom.GetRequestIDFromContext(r.Context())
	return h
}