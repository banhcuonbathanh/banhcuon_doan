// Package utils provides integrated management for configuration, logging, and error handling
package integration

import (
	"context"
	"fmt"
	"sync"

	config "english-ai-full/utils/config"

		logger "english-ai-full/logger"

			coreLogger "english-ai-full/logger/core"

			error_custom "english-ai-full/internal/error_custom/error_custom"

)


type UtilityManager struct {
	Config       *config.Config
	ConfigMgr    *config.ConfigManager
	Logger       *logger.SpecializedLogger
	ErrorHandler *error_custom.UnifiedErrorHandler
	initialized  bool
	mu           sync.RWMutex
}

var (
	globalUtilityManager *UtilityManager
	initOnce            sync.Once
)

// InitializeUtilities sets up all utility systems with proper dependencies
func InitializeUtilities(configPath string) error {
	var initErr error
	
	initOnce.Do(func() {
		globalUtilityManager = &UtilityManager{}
		initErr = globalUtilityManager.initialize(configPath)
	})
	
	return initErr
}

// MustInitializeUtilities initializes utilities and panics on error
func MustInitializeUtilities(configPath string) {
	if err := InitializeUtilities(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize utilities: %v", err))
	}
}

// GetUtilityManager returns the global utility manager instance
func GetUtilityManager() *UtilityManager {
	if globalUtilityManager == nil {
		panic("UtilityManager not initialized. Call InitializeUtilities first.")
	}
	return globalUtilityManager
}

// initialize sets up all utility systems in the correct order
func (um *UtilityManager) initialize(configPath string) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Step 1: Initialize Configuration
	if err := um.initializeConfig(configPath); err != nil {
		return fmt.Errorf("config initialization failed: %w", err)
	}

	// Step 2: Initialize Logger with config
	if err := um.initializeLogger(); err != nil {
		return fmt.Errorf("logger initialization failed: %w", err)
	}

	// Step 3: Initialize Error Handler with logger
	if err := um.initializeErrorHandler(); err != nil {
		return fmt.Errorf("error handler initialization failed: %w", err)
	}

	// Step 4: Setup cross-system integrations
	if err := um.setupIntegrations(); err != nil {
		return fmt.Errorf("integration setup failed: %w", err)
	}

	um.initialized = true
	return nil
}

// initializeConfig sets up the configuration system
func (um *UtilityManager) initializeConfig(configPath string) error {
	// Initialize global config
	if err := config.InitializeConfig(configPath); err != nil {
		return err
	}

	// Get config and config manager references
	um.Config = config.GetConfig()
	um.ConfigMgr = config.GetConfigManager()

	return nil
}

// initializeLogger sets up the logging system based on configuration
func (um *UtilityManager) initializeLogger() error {
	// Create environment-appropriate logger
	var coreLogger *coreLogger.CoreLogger
	
	switch {
	case um.Config.IsProduction():
		coreLogger = logger.NewComponentLogger("production")
		coreLogger.SetLevel(logger.LevelInfo)
	case um.Config.IsDevelopment():
		coreLogger = logger.NewComponentLogger("development")
		coreLogger.SetLevel(logger.LevelDebug)
	case um.Config.IsStaging():
		coreLogger = logger.NewComponentLogger("staging")
		coreLogger.SetLevel(logger.LevelInfo)
	default:
		coreLogger = logger.NewDefaultLogger()
	}

	// Configure logger with app context
	coreLogger.SetComponent(um.Config.AppName)
	coreLogger.SetEnvironment(um.Config.Environment)
	coreLogger.AddContextField("version", um.Config.Version)

	// Create specialized logger
	um.Logger = logger.NewSpecializedLogger(coreLogger)

	// Set global logger for convenience functions
	logger.SetGlobalLogger(coreLogger)

	return nil
}

// initializeErrorHandler sets up the error handling system
func (um *UtilityManager) initializeErrorHandler() error {
	um.ErrorHandler = error_custom.NewUnifiedErrorHandler()

	// Configure error handler with domain settings
	if um.Config.Domains.Enabled != nil {
		// Set up domain-specific error handling based on config
		for _, domain := range um.Config.Domains.Enabled {
			// Configure domain-specific error settings if needed
			_ = domain // Use domain for specific configurations
		}
	}

	return nil
}

// setupIntegrations configures cross-system integrations
func (um *UtilityManager) setupIntegrations() error {
	// Register config change callback for logger reconfiguration
	um.ConfigMgr.RegisterCallback(func(oldConfig, newConfig *config.Config) error {
		return um.reconfigureOnConfigChange(oldConfig, newConfig)
	})

	// Log successful initialization
	um.Logger.LogBusinessEvent("system", "initialization", "utility_manager", "initialized", map[string]interface{}{
		"config_path":  "loaded",
		"environment":  um.Config.Environment,
		"app_name":     um.Config.AppName,
		"version":      um.Config.Version,
		"domains":      um.Config.Domains.Enabled,
	})

	return nil
}

// reconfigureOnConfigChange handles configuration changes
func (um *UtilityManager) reconfigureOnConfigChange(oldConfig, newConfig *config.Config) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Update config reference
	um.Config = newConfig

	// Reconfigure logger if needed
	if oldConfig.Environment != newConfig.Environment || 
	   oldConfig.Debug != newConfig.Debug {
		return um.reconfigureLogger()
	}

	return nil
}

// reconfigureLogger updates logger configuration
func (um *UtilityManager) reconfigureLogger() error {
	// Update logger level based on new config
	if um.Config.Debug {
		um.Logger.SetLevel(logger.LevelDebug)
	} else if um.Config.IsProduction() {
		um.Logger.SetLevel(logger.LevelInfo)
	}

	um.Logger.LogBusinessEvent("system", "configuration", "logger", "reconfigured", map[string]interface{}{
		"environment": um.Config.Environment,
		"debug_mode": um.Config.Debug,
	})

	return nil
}

// IsInitialized returns whether the utility manager is initialized
func (um *UtilityManager) IsInitialized() bool {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.initialized
}

// Shutdown gracefully shuts down all utility systems
func (um *UtilityManager) Shutdown(ctx context.Context) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if !um.initialized {
		return nil
	}

	var shutdownErrors []error

	// Log shutdown start
	um.Logger.LogBusinessEvent("system", "shutdown", "utility_manager", "started", nil)

	// Stop config watcher
	if err := um.ConfigMgr.Stop(); err != nil {
		shutdownErrors = append(shutdownErrors, fmt.Errorf("config manager shutdown: %w", err))
	}

	// Close logger outputs (if applicable)
	// Note: Depends on your logger implementation
	// if closer, ok := um.Logger.(io.Closer); ok {
	//     if err := closer.Close(); err != nil {
	//         shutdownErrors = append(shutdownErrors, fmt.Errorf("logger shutdown: %w", err))
	//     }
	// }

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}

	um.Logger.LogBusinessEvent("system", "shutdown", "utility_manager", "completed", nil)
	um.initialized = false
	return nil
}

// Convenience Functions for Global Access

// Config returns the global configuration
func Config() *config.Config {
	return GetUtilityManager().Config
}

// Logger returns the global logger
func Logger() *logger.SpecializedLogger {
	return GetUtilityManager().Logger
}

// ErrorHandler returns the global error handler
func ErrorHandler() *error_custom.UnifiedErrorHandler {
	return GetUtilityManager().ErrorHandler
}

// Domain-Specific Utility Factories

// NewDomainUtilities creates domain-specific utility instances
func NewDomainUtilities(domain string) *DomainUtilities {
	um := GetUtilityManager()
	
	// Create domain-specific logger
	domainLogger := logger.NewSpecializedComponentLogger(domain)
	domainLogger.SetComponent(domain)
	domainLogger.AddContextField("domain", domain)

	return &DomainUtilities{
		Domain:       domain,
		Config:       um.Config,
		Logger:       domainLogger,
		ErrorHandler: um.ErrorHandler,
	}
}

// DomainUtilities provides domain-specific utility access
type DomainUtilities struct {
	Domain       string
	Config       *config.Config
	Logger       *logger.SpecializedLogger
	ErrorHandler *error.UnifiedErrorHandler
}

// LogOperation logs a domain operation with consistent formatting
func (du *DomainUtilities) LogOperation(operation, result string, success bool, duration interface{}, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	
	metadata["domain"] = du.Domain
	metadata["operation"] = operation
	metadata["success"] = success
	
	if duration != nil {
		metadata["duration"] = duration
	}

	if success {
		du.Logger.LogBusinessEvent(du.Domain, "operation", operation, result, metadata)
	} else {
		du.Logger.LogSecurityEvent("operation_failed", "warning", "", "", metadata)
	}
}

// HandleError wraps error handling with domain context
func (du *DomainUtilities) HandleError(err error, operation string, context map[string]interface{}) error {
	if context == nil {
		context = make(map[string]interface{})
	}
	
	context["domain"] = du.Domain
	context["operation"] = operation
	
	return du.ErrorHandler.HandleError(du.Domain, err)
}

// ValidateConfig checks domain-specific configuration
func (du *DomainUtilities) ValidateConfig() error {
	if !du.Config.IsDomainEnabled(du.Domain) {
		return error_custom.NewBusinessLogicError(du.Domain, "domain_disabled", 
			fmt.Sprintf("Domain %s is not enabled in configuration", du.Domain))
	}
	return nil
}

// Layer-Specific Utility Factories

// NewHandlerUtilities creates utilities optimized for HTTP handlers
func NewHandlerUtilities(domain string) *HandlerUtilities {
	domainUtils := NewDomainUtilities(domain)
	handlerLogger := logger.NewSpecializedHandlerLogger()
	handlerLogger.AddContextField("layer", "handler")
	handlerLogger.AddContextField("domain", domain)

	return &HandlerUtilities{
		DomainUtilities: domainUtils,
		HandlerLogger:   handlerLogger,
	}
}

// HandlerUtilities provides handler-specific utilities
type HandlerUtilities struct {
	*DomainUtilities
	HandlerLogger *logger.SpecializedLogger
}

// NewServiceUtilities creates utilities optimized for service layer
func NewServiceUtilities(domain string) *ServiceUtilities {
	domainUtils := NewDomainUtilities(domain)
	serviceLogger := logger.NewSpecializedServiceLogger()
	serviceLogger.AddContextField("layer", "service")
	serviceLogger.AddContextField("domain", domain)

	return &ServiceUtilities{
		DomainUtilities: domainUtils,
		ServiceLogger:   serviceLogger,
	}
}

// ServiceUtilities provides service-specific utilities
type ServiceUtilities struct {
	*DomainUtilities
	ServiceLogger *logger.SpecializedLogger
}

// NewRepositoryUtilities creates utilities optimized for repository layer
func NewRepositoryUtilities(domain string) *RepositoryUtilities {
	domainUtils := NewDomainUtilities(domain)
	repoLogger := logger.NewSpecializedRepositoryLogger()
	repoLogger.AddContextField("layer", "repository")
	repoLogger.AddContextField("domain", domain)

	return &RepositoryUtilities{
		DomainUtilities: domainUtils,
		RepoLogger:      repoLogger,
	}
}

// RepositoryUtilities provides repository-specific utilities
type RepositoryUtilities struct {
	*DomainUtilities
	RepoLogger *logger.SpecializedLogger
}

// Health Check and Monitoring

// HealthCheck performs health checks on all utility systems
func (um *UtilityManager) HealthCheck(ctx context.Context) map[string]interface{} {
	um.mu.RLock()
	defer um.mu.RUnlock()

	health := map[string]interface{}{
		"utility_manager": map[string]interface{}{
			"initialized": um.initialized,
			"status":      "healthy",
		},
		"config": map[string]interface{}{
			"loaded":      um.Config != nil,
			"environment": "",
			"domains":     []string{},
		},
		"logger": map[string]interface{}{
			"available": um.Logger != nil,
			"level":     "",
		},
		"error_handler": map[string]interface{}{
			"available": um.ErrorHandler != nil,
		},
	}

	if um.Config != nil {
		health["config"].(map[string]interface{})["environment"] = um.Config.Environment
		health["config"].(map[string]interface{})["domains"] = um.Config.Domains.Enabled
	}

	return health
}

// GetMetrics returns utility system metrics
func (um *UtilityManager) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"initialized":       um.IsInitialized(),
		"config_loaded":     um.Config != nil,
		"logger_available":  um.Logger != nil,
		"error_handler_available": um.ErrorHandler != nil,
		"environment":       um.Config.Environment,
		"app_name":         um.Config.AppName,
		"version":          um.Config.Version,
	}
}




