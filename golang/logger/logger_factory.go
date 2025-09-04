// internal/logger/logger_factory.go - Enhanced factory with auto-configuration
package logger

import (
	"english-ai-full/logger/core"
	"os"
	"strings"
)

// Global logger instances
var (
	GlobalLogger            *core.CoreLogger
	GlobalSpecializedLogger *SpecializedLogger
)

func init() {
	GlobalLogger = NewDefaultLogger()
	GlobalSpecializedLogger = NewSpecializedLogger(GlobalLogger)
}

// AutoConfiguredLogger provides automatic configuration based on context
type AutoConfiguredLogger struct {
	*SpecializedLogger
}

// NewAutoConfiguredLogger creates a logger with automatic layer and domain detection
func NewAutoConfiguredLogger() *AutoConfiguredLogger {
	return &AutoConfiguredLogger{
		SpecializedLogger: GlobalSpecializedLogger,
	}
}

// ForHandler creates a handler-layer logger with automatic configuration
func (acl *AutoConfiguredLogger) ForHandler(component string) *SpecializedLogger {
	logger := NewSpecializedHandlerLogger()
	if component != "" {
		logger.SetComponent(component)
	}
	return logger
}

// ForService creates a service-layer logger with automatic configuration
func (acl *AutoConfiguredLogger) ForService(component string) *SpecializedLogger {
	logger := NewSpecializedServiceLogger()
	if component != "" {
		logger.SetComponent(component)
	}
	return logger
}

// ForRepository creates a repository-layer logger with automatic configuration
func (acl *AutoConfiguredLogger) ForRepository(component string) *SpecializedLogger {
	logger := NewSpecializedRepositoryLogger()
	if component != "" {
		logger.SetComponent(component)
	}
	return logger
}

// ForDatabase creates a database-layer logger with automatic configuration
func (acl *AutoConfiguredLogger) ForDatabase(component string) *SpecializedLogger {
	logger := NewSpecializedDatabaseLogger()
	if component != "" {
		logger.SetComponent(component)
	}
	return logger
}

// Factory functions for specialized loggers with enhanced auto-configuration
func NewDefaultSpecializedLogger() *SpecializedLogger {
	coreLogger := NewDefaultLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedComponentLogger(component string) *SpecializedLogger {
	coreLogger := NewComponentLogger(component)
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedLayerLogger(layer string) *SpecializedLogger {
	coreLogger := NewLayerLogger(layer)
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedHandlerLogger() *SpecializedLogger {
	coreLogger := NewHandlerLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedServiceLogger() *SpecializedLogger {
	coreLogger := NewServiceLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedRepositoryLogger() *SpecializedLogger {
	coreLogger := NewRepositoryLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedMiddlewareLogger() *SpecializedLogger {
	coreLogger := NewMiddlewareLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedAuthLogger() *SpecializedLogger {
	coreLogger := NewAuthLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedValidationLogger() *SpecializedLogger {
	coreLogger := NewValidationLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedCacheLogger() *SpecializedLogger {
	coreLogger := NewCacheLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedDatabaseLogger() *SpecializedLogger {
	coreLogger := NewDatabaseLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedExternalLogger() *SpecializedLogger {
	coreLogger := NewExternalLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedSecurityLogger() *SpecializedLogger {
	coreLogger := NewSecurityLogger()
	return NewSpecializedLogger(coreLogger)
}

// Enhanced factory functions for new layers
func NewSpecializedGatewayLogger() *SpecializedLogger {
	coreLogger := NewGatewayLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedQueueLogger() *SpecializedLogger {
	coreLogger := NewQueueLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedSchedulerLogger() *SpecializedLogger {
	coreLogger := NewSchedulerLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedWebsocketLogger() *SpecializedLogger {
	coreLogger := NewWebsocketLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedEmailLogger() *SpecializedLogger {
	coreLogger := NewEmailLogger()
	return NewSpecializedLogger(coreLogger)
}

// Core logger factories for new layers
func NewGatewayLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("gateway")
	logger.SetLayer(core.LayerGateway)
	return logger
}

func NewQueueLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("queue")
	logger.SetLayer(core.LayerQueue)
	return logger
}

func NewSchedulerLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("scheduler")
	logger.SetLayer(core.LayerScheduler)
	return logger
}

func NewWebsocketLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("websocket")
	logger.SetLayer(core.LayerWebsocket)
	return logger
}

func NewEmailLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("email")
	logger.SetLayer(core.LayerEmail)
	return logger
}

// Enhanced factory functions with configuration
func NewLoggerWithConfig(config LoggerConfig) *SpecializedLogger {
	coreLogger := NewDefaultLogger()
	
	if config.Component != "" {
		coreLogger.SetComponent(config.Component)
	}
	if config.Layer != "" {
		coreLogger.SetLayer(config.Layer)
	}
	if config.Operation != "" {
		coreLogger.SetOperation(config.Operation)
	}
	if config.Environment != "" {
		coreLogger.SetEnvironment(config.Environment)
	}
	if config.Level != nil {
		coreLogger.SetLevel(*config.Level)
	}
	
	for key, value := range config.ContextFields {
		coreLogger.AddContextField(key, value)
	}
	
	return NewSpecializedLogger(coreLogger)
}

// LoggerConfig holds configuration for logger creation
type LoggerConfig struct {
	Component     string
	Layer         string
	Operation     string
	Environment   string
	Level         *core.Level
	ContextFields map[string]interface{}
}

// Convenience constructor functions
func WithContext(fields map[string]interface{}) *SpecializedLogger {
	logger := NewDefaultSpecializedLogger()
	for key, value := range fields {
		logger.AddContextField(key, value)
	}
	return logger
}

func WithConfig(component, layer, operation, environment string) *SpecializedLogger {
	config := LoggerConfig{
		Component:     component,
		Layer:         layer,
		Operation:     operation,
		Environment:   environment,
		ContextFields: make(map[string]interface{}),
	}
	return NewLoggerWithConfig(config)
}

func WithLayer(layer string) *SpecializedLogger {
	return NewSpecializedLayerLogger(layer)
}

func WithComponent(component string) *SpecializedLogger {
	return NewSpecializedComponentLogger(component)
}

// Smart factory that auto-detects layer from component name
func NewSmartLogger(component string) *SpecializedLogger {
	layer := detectLayerFromComponent(component)
	
	config := LoggerConfig{
		Component:     component,
		Layer:         layer,
		ContextFields: make(map[string]interface{}),
	}
	
	return NewLoggerWithConfig(config)
}

// detectLayerFromComponent automatically detects layer from component name patterns
func detectLayerFromComponent(component string) string {
	component = strings.ToLower(component)
	
	// Handler patterns
	if strings.Contains(component, "handler") || strings.Contains(component, "controller") || 
	   strings.Contains(component, "endpoint") || strings.Contains(component, "router") {
		return core.LayerHandler
	}
	
	// Service patterns
	if strings.Contains(component, "service") || strings.Contains(component, "business") ||
	   strings.Contains(component, "logic") || strings.Contains(component, "usecase") {
		return core.LayerService
	}
	
	// Repository patterns
	if strings.Contains(component, "repository") || strings.Contains(component, "repo") ||
	   strings.Contains(component, "dao") || strings.Contains(component, "storage") {
		return core.LayerRepository
	}
	
	// Database patterns
	if strings.Contains(component, "database") || strings.Contains(component, "db") ||
	   strings.Contains(component, "sql") || strings.Contains(component, "query") {
		return core.LayerDatabase
	}
	
	// Middleware patterns
	if strings.Contains(component, "middleware") || strings.Contains(component, "interceptor") ||
	   strings.Contains(component, "filter") {
		return core.LayerMiddleware
	}
	
	// Auth patterns
	if strings.Contains(component, "auth") || strings.Contains(component, "jwt") ||
	   strings.Contains(component, "token") || strings.Contains(component, "session") {
		return core.LayerAuth
	}
	
	// Cache patterns
	if strings.Contains(component, "cache") || strings.Contains(component, "redis") ||
	   strings.Contains(component, "memory") {
		return core.LayerCache
	}
	
	// External patterns
	if strings.Contains(component, "client") || strings.Contains(component, "api") ||
	   strings.Contains(component, "external") || strings.Contains(component, "http") {
		return core.LayerExternal
	}
	
	// Gateway patterns
	if strings.Contains(component, "gateway") || strings.Contains(component, "proxy") {
		return core.LayerGateway
	}
	
	// Queue patterns
	if strings.Contains(component, "queue") || strings.Contains(component, "worker") ||
	   strings.Contains(component, "job") || strings.Contains(component, "task") {
		return core.LayerQueue
	}
	
	// Scheduler patterns
	if strings.Contains(component, "scheduler") || strings.Contains(component, "cron") ||
	   strings.Contains(component, "timer") {
		return core.LayerScheduler
	}
	
	// Websocket patterns
	if strings.Contains(component, "websocket") || strings.Contains(component, "ws") ||
	   strings.Contains(component, "socket") {
		return core.LayerWebsocket
	}
	
	// Email patterns
	if strings.Contains(component, "email") || strings.Contains(component, "mail") ||
	   strings.Contains(component, "smtp") {
		return core.LayerEmail
	}
	
	// Security patterns
	if strings.Contains(component, "security") || strings.Contains(component, "encrypt") ||
	   strings.Contains(component, "decrypt") || strings.Contains(component, "hash") {
		return core.LayerSecurity
	}
	
	// Validation patterns
	if strings.Contains(component, "validation") || strings.Contains(component, "validator") ||
	   strings.Contains(component, "validate") {
		return core.LayerValidation
	}
	
	// Default to service layer
	return core.LayerService
}

// Core logger factories (existing ones from global.go)
func NewDefaultLogger() *core.CoreLogger {
	logger := core.NewLogger()
	
	environment := getEnvironment()
	logger.SetEnvironment(environment)
	logger.SetLevel(getMinLogLevel(environment))
	
	logger.AddContextField("environment", environment)
	logger.AddContextField("service", getServiceName())
	
	return logger
}

func NewComponentLogger(component string) *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent(component)
	return logger
}

func NewLayerLogger(layer string) *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetLayer(layer)
	return logger
}

func NewHandlerLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("handler")
	logger.SetLayer(core.LayerHandler)
	return logger
}

func NewServiceLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("service")
	logger.SetLayer(core.LayerService)
	return logger
}

func NewRepositoryLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("repository")
	logger.SetLayer(core.LayerRepository)
	return logger
}

func NewMiddlewareLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("middleware")
	logger.SetLayer(core.LayerMiddleware)
	return logger
}

func NewAuthLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("auth")
	logger.SetLayer(core.LayerAuth)
	return logger
}

func NewValidationLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("validation")
	logger.SetLayer(core.LayerValidation)
	return logger
}

func NewCacheLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("cache")
	logger.SetLayer(core.LayerCache)
	return logger
}

func NewDatabaseLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("database")
	logger.SetLayer(core.LayerDatabase)
	return logger
}

func NewExternalLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("external")
	logger.SetLayer(core.LayerExternal)
	return logger
}

func NewSecurityLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("security")
	logger.SetLayer(core.LayerSecurity)
	return logger
}

// Helper functions
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func getServiceName() string {
	service := os.Getenv("SERVICE_NAME")
	if service == "" {
		service = "unknown-service"
	}
	return service
}

func getMinLogLevel(environment string) core.Level {
	switch strings.ToLower(environment) {
	case "production", "prod":
		return core.InfoLevel
	case "staging", "stage":
		return core.InfoLevel
	case "testing", "test":
		return core.DebugLevel
	default: // development
		return core.DebugLevel
	}
}