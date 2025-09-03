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
	utils_config "english-ai-full/utils/config"

	"github.com/go-playground/validator/v10"
)

// ============================================================================
// CORE HANDLER STRUCTURE AND CONSTRUCTOR
// ============================================================================

// BaseAccountHandler provides core account handling functionality
type BaseAccountHandler struct {
	userClient   pb.AccountServiceClient
	validator    *validator.Validate
	logger       *logger.Logger
	config       *utils_config.Config
	domain       string
	errorHandler *errorcustom.UnifiedErrorHandler
	requestID    string
}

// OperationContext encapsulates common operation context data
type OperationContext struct {
	RequestID string
	Domain    string
	Operation string
	StartTime time.Time
	UserID    int64
	Context   map[string]interface{}
}

// NewBaseHandler creates a new BaseAccountHandler instance
func NewBaseHandler(userClient pb.AccountServiceClient, config *utils_config.Config) *BaseAccountHandler {
	domain := errorcustom.DomainAccount
	
	// Create logger with context
	handlerLogger := logger.NewHandlerLogger()
	handlerLogger.AddGlobalField("component", "account_handler")
	handlerLogger.AddGlobalField("layer", "handler")
	handlerLogger.AddGlobalField("domain", domain)
	
	// Initialize validator
	v := validator.New()
	
	// Register validators
	if err := registerAccountValidators(v, domain); err != nil {
		handlerLogger.Warning("Some custom validators failed to register", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	handler := &BaseAccountHandler{
		userClient:   userClient,
		validator:    v,
		logger:       handlerLogger,
		config:       config,
		domain:       domain,
		errorHandler: errorcustom.NewUnifiedErrorHandler(),
	}

	handlerLogger.Info("Base account handler initialized", map[string]interface{}{
		"domain":            domain,
		"grpc_client_ready": userClient != nil,
		"config_loaded":     config != nil,
	})

	return handler
}

// ============================================================================
// CORE UTILITY METHODS
// ============================================================================

// WithRequestID sets the request ID for the handler
func (h *BaseAccountHandler) WithRequestID(r *http.Request) *BaseAccountHandler {
	h.requestID = errorcustom.GetRequestIDFromContext(r.Context())
	return h
}

// GetDomain returns the domain this handler operates in
func (h *BaseAccountHandler) GetDomain() string {
	return h.domain
}

// GetConfig returns the configuration instance
func (h *BaseAccountHandler) GetConfig() *utils_config.Config {
	return h.config
}

// GetErrorHandler returns the unified error handler
func (h *BaseAccountHandler) GetErrorHandler() *errorcustom.UnifiedErrorHandler {
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
// OPERATION CONTEXT MANAGEMENT
// ============================================================================

// setupOperationContext creates a standardized operation context
func (h *BaseAccountHandler) setupOperationContext(r *http.Request, operation string) *OperationContext {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	opCtx := &OperationContext{
		RequestID: requestID,
		Domain:    h.domain,
		Operation: operation,
		StartTime: time.Now(),
		Context: map[string]interface{}{
			"operation":  operation,
			"component":  "account_handler",
			"domain":     h.domain,
			"request_id": requestID,
		},
	}
	
	// Try to extract user ID from context
	if userID, err := h.getUserIDFromContext(r.Context()); err == nil {
		opCtx.UserID = userID
		opCtx.Context["user_id"] = userID
	}
	
	return opCtx
}

// logOperationStart logs the beginning of an operation
func (h *BaseAccountHandler) logOperationStart(opCtx *OperationContext) {
	h.logger.Debug("Operation started", opCtx.Context)
}

// logOperationEnd logs the completion of an operation
func (h *BaseAccountHandler) logOperationEnd(opCtx *OperationContext, err error, statusCode int) {
	duration := time.Since(opCtx.StartTime)
	
	context := make(map[string]interface{})
	for k, v := range opCtx.Context {
		context[k] = v
	}
	context["duration_ms"] = duration.Milliseconds()
	context["success"] = err == nil
	context["status_code"] = statusCode
	
	if err != nil {
		context["error"] = err.Error()
		h.logger.Warning("Operation completed with error", context)
	} else {
		h.logger.Debug("Operation completed successfully", context)
	}
	
	// Log performance metrics
	logger.LogPerformance(
		opCtx.Operation,
		duration,
		context,
	)
}

// ============================================================================
// CONTEXT EXTRACTION METHODS
// ============================================================================

// getUserIDFromContext extracts user ID with domain-aware error handling
func (h *BaseAccountHandler) getUserIDFromContext(ctx context.Context) (int64, error) {
	start := time.Now()
	requestID := errorcustom.GetRequestIDFromContext(ctx)
	
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
				h.logger.Warning("User ID found as string in context, converting to int64", map[string]interface{}{
					"user_id_string": userIDStr,
					"user_id_int64":  parsedID,
					"duration_ms":    time.Since(start).Milliseconds(),
				})
				return parsedID, nil
			}
		}
		
		// Log the failed context extraction with available context keys
		contextKeys := h.getContextKeys(ctx)
		
		h.logger.Error("User ID not found in request context", map[string]interface{}{
			"available_context_keys": contextKeys,
			"expected_key":           "user_id",
			"expected_type":          "int64",
			"duration_ms":           time.Since(start).Milliseconds(),
		})
		
		// Log security event for unauthorized access attempt
		logger.LogSecurityEvent(
			"unauthorized_context_access",
			"Request made without valid user context",
			"medium",
			logContext,
		)
		
		return 0, errorcustom.NewAuthenticationError(h.domain, "user context not found")
	}
	
	h.logger.Debug("User ID successfully extracted from context", map[string]interface{}{
		"user_id":     userID,
		"duration_ms": time.Since(start).Milliseconds(),
	})
	
	return userID, nil
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

// ============================================================================
// VALIDATOR REGISTRATION
// ============================================================================

// registerAccountValidators registers custom validators for the account domain
func registerAccountValidators(v *validator.Validate, domain string) error {
	// Register password strength validator
	if err := v.RegisterValidation("strongpassword", validateStrongPassword); err != nil {
		return fmt.Errorf("failed to register strong password validator: %w", err)
	}
	
	// Register user role validator
	if err := v.RegisterValidation("userrole", validateUserRole); err != nil {
		return fmt.Errorf("failed to register user role validator: %w", err)
	}
	
	return nil
}

// validateStrongPassword validates password strength
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
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

// validateUserRole validates user role
func validateUserRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := []string{"admin", "teacher", "student"}
	
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	
	return false
}