// internal/common/layer_context.go
package common


import (
	"fmt"
	"net/http"
	"time"

	"english-ai-full/logger/core"
)

// BaseLayerContext contains common layer information shared across all layers
type BaseLayerContext struct {
	Domain      string `json:"domain"`
	Layer       string `json:"layer"`
	Service     string `json:"service"`
	Version     string `json:"version,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// ToMap converts BaseLayerContext to map[string]interface{} for logging
func (lc *BaseLayerContext) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"domain": lc.Domain,
		"layer":  lc.Layer,
	}
	
	if lc.Service != "" {
		result["service"] = lc.Service
	}
	
	if lc.Version != "" {
		result["version"] = lc.Version
	}
	
	if lc.Environment != "" {
		result["environment"] = lc.Environment
	}
	
	return result
}

// MergeWithContext merges BaseLayerContext with additional context data
func (lc *BaseLayerContext) MergeWithContext(additionalCtx map[string]interface{}) map[string]interface{} {
	merged := lc.ToMap()
	
	// Add additional context
	for key, value := range additionalCtx {
		merged[key] = value
	}
	
	return merged
}

// ConfigureLogger configures a core logger with this layer context
func (lc *BaseLayerContext) ConfigureLogger(logger *core.CoreLogger) {
	logger.SetLayer(lc.Layer)
	logger.SetEnvironment(lc.Environment)
	
	// Add context fields
	logger.AddContextField("domain", lc.Domain)
	logger.AddContextField("service", lc.Service)
	if lc.Version != "" {
		logger.AddContextField("version", lc.Version)
	}
}

// =============================================================================
// HANDLER LAYER CONTEXT
// =============================================================================

// HandlerLayerContext extends BaseLayerContext with handler-specific functionality
type HandlerLayerContext struct {
	*BaseLayerContext
}

// NewHandlerLayerContext creates a new handler layer context
func NewHandlerLayerContext(domain, service string) *HandlerLayerContext {
	return &HandlerLayerContext{
		BaseLayerContext: &BaseLayerContext{
			Domain:      domain,
			Layer:       core.LayerHandler,
			Service:     service,
			Version:     "1.0.0",
			Environment: "production",
		},
	}
}

// NewHandlerLogger creates a pre-configured core logger for this handler context
func (hlc *HandlerLayerContext) NewHandlerLogger() *core.CoreLogger {
	logger := core.NewLogger()
	hlc.ConfigureLogger(logger)
	logger.SetComponent("handler")
	return logger
}

// BuildOperationContext creates operation-specific context with request details
func (hlc *HandlerLayerContext) BuildOperationContext(operation string, r *http.Request, requestID string) map[string]interface{} {
	operationCtx := hlc.ToMap()
	
	// Add operation-specific data
	operationCtx["operation"] = operation
	operationCtx["request_id"] = requestID
	operationCtx["method"] = r.Method
	operationCtx["path"] = r.URL.Path
	operationCtx["client_ip"] = hlc.GetClientIP(r)
	operationCtx["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	
	return operationCtx
}

// GetRequestID extracts or generates request ID
func (hlc *HandlerLayerContext) GetRequestID(r *http.Request) string {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return requestID
}

// GetClientIP extracts client IP from various headers
func (hlc *HandlerLayerContext) GetClientIP(r *http.Request) string {
	// Try to get IP from X-Forwarded-For header
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	
	// Try to get IP from X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// =============================================================================
// SERVICE LAYER CONTEXT
// =============================================================================

// ServiceLayerContext extends BaseLayerContext with service-specific functionality
type ServiceLayerContext struct {
	*BaseLayerContext
}

// NewServiceLayerContext creates a new service layer context
func NewServiceLayerContext(domain, service string) *ServiceLayerContext {
	return &ServiceLayerContext{
		BaseLayerContext: &BaseLayerContext{
			Domain:      domain,
			Layer:       core.LayerService,
			Service:     service,
			Version:     "1.0.0",
			Environment: "production",
		},
	}
}

// NewServiceLogger creates a pre-configured core logger for this service context
func (slc *ServiceLayerContext) NewServiceLogger() *core.CoreLogger {
	logger := core.NewLogger()
	slc.ConfigureLogger(logger)
	logger.SetComponent("service")
	return logger
}

// BuildOperationContext creates operation-specific context for service operations
func (slc *ServiceLayerContext) BuildOperationContext(operation string, additionalCtx map[string]interface{}) map[string]interface{} {
	operationCtx := slc.ToMap()
	
	// Add operation-specific data
	operationCtx["operation"] = operation
	operationCtx["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	
	// Merge additional context
	for key, value := range additionalCtx {
		operationCtx[key] = value
	}
	
	return operationCtx
}

// BuildServiceCallContext creates context for external service calls
func (slc *ServiceLayerContext) BuildServiceCallContext(operation, targetService, method string, additionalCtx map[string]interface{}) map[string]interface{} {
	operationCtx := slc.BuildOperationContext(operation, additionalCtx)
	operationCtx["target_service"] = targetService
	operationCtx["service_method"] = method
	operationCtx["call_type"] = "outbound"
	
	return operationCtx
}

// =============================================================================
// REPOSITORY LAYER CONTEXT
// =============================================================================

// RepositoryLayerContext extends BaseLayerContext with repository-specific functionality
type RepositoryLayerContext struct {
	*BaseLayerContext
}

// NewRepositoryLayerContext creates a new repository layer context
func NewRepositoryLayerContext(domain, service string) *RepositoryLayerContext {
	return &RepositoryLayerContext{
		BaseLayerContext: &BaseLayerContext{
			Domain:      domain,
			Layer:       core.LayerRepository,
			Service:     service,
			Version:     "1.0.0",
			Environment: "production",
		},
	}
}

// NewRepositoryLogger creates a pre-configured core logger for this repository context
func (rlc *RepositoryLayerContext) NewRepositoryLogger() *core.CoreLogger {
	logger := core.NewLogger()
	rlc.ConfigureLogger(logger)
	logger.SetComponent("repository")
	return logger
}

// BuildOperationContext creates operation-specific context for database operations
func (rlc *RepositoryLayerContext) BuildOperationContext(operation, table, function string, additionalCtx map[string]interface{}) map[string]interface{} {
	operationCtx := rlc.ToMap()
	
	// Add operation-specific data
	operationCtx["operation"] = operation
	operationCtx["table"] = table
	operationCtx["function"] = function
	operationCtx["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	
	// Merge additional context
	for key, value := range additionalCtx {
		operationCtx[key] = value
	}
	
	return operationCtx
}

// BuildDatabaseContext creates context for database operations with performance metrics
func (rlc *RepositoryLayerContext) BuildDatabaseContext(operation, table, function string, startTime time.Time, success bool, rowsAffected int64) map[string]interface{} {
	operationCtx := rlc.ToMap()
	duration := time.Since(startTime)
	
	operationCtx["operation"] = operation
	operationCtx["table"] = table
	operationCtx["function"] = function
	operationCtx["success"] = success
	operationCtx["duration_ms"] = duration.Milliseconds()
	operationCtx["rows_affected"] = rowsAffected
	operationCtx["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	
	return operationCtx
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// CreateLoggerForLayer creates a properly configured logger for any layer
func CreateLoggerForLayer(layerType, domain, service, component string) *core.CoreLogger {
	logger := core.NewLogger()
	logger.SetLayer(layerType)
	logger.SetComponent(component)
	logger.SetEnvironment("production")
	
	// Add context fields
	logger.AddContextField("domain", domain)
	logger.AddContextField("service", service)
	logger.AddContextField("version", "1.0.0")
	
	return logger
}