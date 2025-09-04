// Solution 3: Logging Decorators with Reflection-based Context
package logger

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// LogDecorator provides automatic logging for struct methods
type LogDecorator struct {
	domain    string
	component string
	layer     string
	logger    *DomainLogger
}

func NewLogDecorator(domain, component, layer string) *LogDecorator {
	return &LogDecorator{
		domain:    domain,
		component: component,
		layer:     layer,
		logger:    NewDomainLogger(domain),
	}
}

// LoggedMethod represents a method with automatic logging
type LoggedMethod struct {
	decorator *LogDecorator
	operation string
}

// WrapHandler creates a logged wrapper for handler methods
func (ld *LogDecorator) WrapHandler(handler interface{}) interface{} {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()
	
	// Create a new struct type with the same methods but with logging
	newStruct := reflect.New(handlerType.Elem())
	
	// Copy all fields
	for i := 0; i < handlerValue.Elem().NumField(); i++ {
		newStruct.Elem().Field(i).Set(handlerValue.Elem().Field(i))
	}
	
	return newStruct.Interface()
}

// LogExecution wraps any function with automatic logging
func (ld *LogDecorator) LogExecution(operation string, fn func() error) error {
	logger := ld.logger.WithOperation(operation)
	
	logger.Info(fmt.Sprintf("Starting %s operation", operation))
	start := time.Now()
	
	err := fn()
	duration := time.Since(start)
	
	if err != nil {
		logger.ErrorWithCause(fmt.Sprintf("Operation %s failed", operation), "operation_failed", map[string]interface{}{
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
	} else {
		logger.Info(fmt.Sprintf("Operation %s completed successfully", operation), map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		})
	}
	
	return err
}

// Solution 4: Code Generation Helper
// This would be used with go:generate to automatically generate logging code

//go:generate go run logging_generator.go

// HandlerConfig defines logging configuration for code generation
type HandlerConfig struct {
	Domain     string            `yaml:"domain"`
	Component  string            `yaml:"component"`
	Layer      string            `yaml:"layer"`
	Operations []OperationConfig `yaml:"operations"`
}

type OperationConfig struct {
	Name        string            `yaml:"name"`
	LogLevel    string            `yaml:"log_level"`
	CustomFields map[string]string `yaml:"custom_fields"`
}

// GenerateHandlerLogger generates a logger for a specific handler
func GenerateHandlerLogger(config HandlerConfig) string {
	template := `
// Auto-generated logger for %s
package %s

import (
	"context"
	"time"
	"english-ai-full/logger"
)

type %sLogger struct {
	logger *logger.ContextualLogger
}

func New%sLogger() *%sLogger {
	domainLogger := logger.NewDomainLogger("%s")
	return &%sLogger{
		logger: domainLogger.WithOperation(""),
	}
}

`
	
	// Generate method-specific loggers
	for _, op := range config.Operations {
		template += fmt.Sprintf(`
func (l *%sLogger) Log%s(ctx context.Context, message string, fields ...map[string]interface{}) {
	logger := l.logger.DomainLogger.WithContext(ctx).WithOperation("%s")
	logger.%s(message, fields...)
}

func (l *%sLogger) Log%sStart(ctx context.Context) {
	logger := l.logger.DomainLogger.WithContext(ctx).WithOperation("%s")
	logger.Info("%s operation started")
}

func (l *%sLogger) Log%sEnd(ctx context.Context, success bool, duration time.Duration, fields ...map[string]interface{}) {
	logger := l.logger.DomainLogger.WithContext(ctx).WithOperation("%s")
	mergedFields := make(map[string]interface{})
	mergedFields["duration_ms"] = duration.Milliseconds()
	mergedFields["success"] = success
	
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			mergedFields[k] = v
		}
	}
	
	if success {
		logger.Info("%s operation completed successfully", mergedFields)
	} else {
		logger.ErrorWithCause("%s operation failed", "operation_failed", mergedFields)
	}
}
`,
			config.Component, strings.Title(op.Name), op.Name, strings.Title(op.LogLevel),
			config.Component, strings.Title(op.Name), op.Name,
			config.Component, strings.Title(op.Name), op.Name,
			config.Component, strings.Title(op.Name), op.Name, strings.Title(op.Name),
			strings.Title(op.Name), strings.Title(op.Name))
	}
	
	return fmt.Sprintf(template, config.Domain, config.Domain, 
		strings.Title(config.Component), strings.Title(config.Component), 
		strings.Title(config.Component), config.Domain, strings.Title(config.Component))
}

// Solution 5: Aspect-Oriented Programming (AOP) Style Logging
type LoggingAspect struct {
	beforeAdvice func(context.Context, string, ...interface{})
	afterAdvice  func(context.Context, string, time.Duration, error, ...interface{})
	logger       *ContextualLogger
}

func NewLoggingAspect(domain string) *LoggingAspect {
	logger := NewDomainLogger(domain).WithOperation("")
	
	return &LoggingAspect{
		logger: logger,
		beforeAdvice: func(ctx context.Context, operation string, args ...interface{}) {
			contextLogger := logger.DomainLogger.WithContext(ctx).WithOperation(operation)
			contextLogger.Info(fmt.Sprintf("Starting %s", operation), map[string]interface{}{
				"args_count": len(args),
			})
		},
		afterAdvice: func(ctx context.Context, operation string, duration time.Duration, err error, args ...interface{}) {
			contextLogger := logger.DomainLogger.WithContext(ctx).WithOperation(operation)
			if err != nil {
				contextLogger.ErrorWithCause(fmt.Sprintf("%s failed", operation), "operation_error", map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": duration.Milliseconds(),
				})
			} else {
				contextLogger.Info(fmt.Sprintf("%s completed", operation), map[string]interface{}{
					"duration_ms": duration.Milliseconds(),
				})
			}
		},
	}
}

// Around wraps a function with logging aspects
func (la *LoggingAspect) Around(operation string, fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		start := time.Now()
		
		// Before advice
		la.beforeAdvice(ctx, operation)
		
		// Execute function
		err := fn(ctx)
		
		// After advice
		duration := time.Since(start)
		la.afterAdvice(ctx, operation, duration, err)
		
		return err
	}
}

// Solution 6: Simplified Usage Pattern for Your Current Code
type SimpleAccountHandler struct {
	log *ContextualLogger
	// ... other dependencies
}

func NewSimpleAccountHandler() *SimpleAccountHandler {
	return &SimpleAccountHandler{
		log: NewAccountLogger().WithOperation(""), // Operation will be auto-detected or set per method
	}
}

// Your Register method simplified
func (h *SimpleAccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	// One line to get fully contextualized logger
	log := h.log.DomainLogger.WithContext(r.Context()).WithOperation("register")
	
	log.Info("Register function started", map[string]interface{}{
		"endpoint": r.URL.Path,
		"method":   r.Method,
	})
	
	startTime := time.Now()
	requestID := getRequestID(r)
	
	log.LogRequestStart(requestID, r.Method, r.URL.Path, "")
	
	// Context timeout check
	if err := r.Context().Err(); err != nil {
		log.ErrorWithCause("Request context cancelled", "context_timeout")
		log.LogRequestEnd(requestID, http.StatusRequestTimeout, time.Since(startTime))
		return
	}
	
	// Parse request
	var registerRequest account_dto.CreateUserRequest
	if err := decodeJSONRequest(r, &registerRequest); err != nil {
		log.ErrorWithCause("Failed to parse request body", "json_decode_error")
		log.LogRequestEnd(requestID, http.StatusBadRequest, time.Since(startTime))
		return
	}
	
	// Validation
	if err := validateStruct(registerRequest); err != nil {
		log.LogStructValidationError("CreateUserRequest", registerRequest, "Request validation failed")
		log.LogRequestEnd(requestID, http.StatusBadRequest, time.Since(startTime))
		return
	}
	
	// Service call
	log.Info("Calling CreateUser service", map[string]interface{}{
		"service": "UserService",
		"method":  "CreateUser",
		"email":   registerRequest.Email,
	})
	
	// ... service call logic ...
	
	// Success
	log.Info("User registration completed successfully", map[string]interface{}{
		"user_id":     "created_user_id",
		"email":       registerRequest.Email,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})
	log.LogRequestEnd(requestID, http.StatusCreated, time.Since(startTime))
}

// Solution 7: Configuration-based Logger Factory
type LoggerFactory struct {
	configs map[string]DomainConfig
}

type DomainConfig struct {
	Domain    string `yaml:"domain"`
	Layer     string `yaml:"layer"`  
	Component string `yaml:"component"`
	Level     string `yaml:"level"`
}

func NewLoggerFactory(configFile string) *LoggerFactory {
	// Load configuration from YAML file
	configs := make(map[string]DomainConfig)
	
	// Example configs (would normally load from file)
	configs["account"] = DomainConfig{Domain: "account", Layer: "handler", Component: "account", Level: "info"}
	configs["product"] = DomainConfig{Domain: "product", Layer: "handler", Component: "product", Level: "info"}
	configs["order"] = DomainConfig{Domain: "order", Layer: "handler", Component: "order", Level: "info"}
	
	return &LoggerFactory{configs: configs}
}

func (lf *LoggerFactory) GetLogger(domain string) *ContextualLogger {
	config, exists := lf.configs[domain]
	if !exists {
		// Fallback to default
		config = DomainConfig{Domain: domain, Layer: "handler", Component: domain, Level: "info"}
	}
	
	return NewDomainLogger(config.Domain).WithOperation("")
}

// Usage example:
var loggerFactory = NewLoggerFactory("config/logging.yaml")

func init() {
	// Pre-create loggers for all domains
	AccountLog = loggerFactory.GetLogger("account")
	ProductLog = loggerFactory.GetLogger("product")
	OrderLog = loggerFactory.GetLogger("order")
	// ... etc
}

var (
	AccountLog *ContextualLogger
	ProductLog *ContextualLogger
	OrderLog   *ContextualLogger
	// ... add all your domains
)
