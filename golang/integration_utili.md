// Package utils provides integrated management for configuration, logging, and error handling
package utils

import (
	"context"
	"fmt"
	"sync"

	"your-project/golang/utils/config"
	"your-project/golang/utils/error"
	"your-project/golang/utils/logger"
)

// UtilityManager manages all utility systems in a coordinated way
type UtilityManager struct {
	Config       *config.Config
	ConfigMgr    *config.ConfigManager
	Logger       *logger.SpecializedLogger
	ErrorHandler *error.UnifiedErrorHandler
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
	var coreLogger *logger.CoreLogger
	
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
	um.ErrorHandler = error.NewUnifiedErrorHandler()

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
func ErrorHandler() *error.UnifiedErrorHandler {
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
		return error.NewBusinessLogicError(du.Domain, "domain_disabled", 
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


example how to use 9-------------------------------------------


-----------------------------------------------------------------


// Package examples demonstrates how to use the integrated utility system
package examples

import (
	"context"
	"net/http"
	"time"

	"your-project/golang/utils"
	"your-project/golang/utils/config"
	"your-project/golang/utils/error"
	"your-project/golang/utils/logger"
)

// =============================================================================
// 1. APPLICATION INITIALIZATION
// =============================================================================

// main.go - Application startup
func main() {
	// Initialize all utilities at application start
	if err := utils.InitializeUtilities("./config.yaml"); err != nil {
		panic(err) // or handle gracefully
	}
	
	// Alternative: Use MustInitializeUtilities for cleaner code
	// utils.MustInitializeUtilities("./config.yaml")
	
	// Start your application
	startHTTPServer()
}

func startHTTPServer() {
	config := utils.Config()
	logger := utils.Logger()
	
	logger.LogBusinessEvent("system", "startup", "http_server", "starting", map[string]interface{}{
		"address": config.GetServerAddress(),
		"port":    config.Server.Port,
	})
	
	// Setup routes with middleware
	http.Handle("/api/", setupRoutes())
	
	addr := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.LogSecurityEvent("startup_failed", "critical", "", "", map[string]interface{}{
			"error": err.Error(),
			"address": addr,
		})
	}
}

// =============================================================================
// 2. HANDLER LAYER USAGE
// =============================================================================

// UserHandler demonstrates handler-level utility usage
type UserHandler struct {
	service *UserService
	utils   *utils.HandlerUtilities
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{
		service: service,
		utils:   utils.NewHandlerUtilities("account"), // domain-specific utilities
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Extract user ID using error handler
	userID, err := h.utils.ErrorHandler.ParseIDParam(r, "id", "account")
	if err != nil {
		h.utils.HandlerLogger.LogAPIRequest(r.Method, r.URL.Path, 400, time.Since(startTime), map[string]interface{}{
			"error": "invalid_user_id",
		})
		h.utils.ErrorHandler.HandleHTTPError(w, r, err)
		return
	}

	// Log the operation start
	h.utils.HandlerLogger.LogRequestStart(
		utils.GetRequestIDFromContext(r.Context()),
		r.Method,
		r.URL.Path,
		fmt.Sprintf("%d", userID),
	)

	// Call service layer
	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		duration := time.Since(startTime)
		h.utils.LogOperation("get_user", "failed", false, duration, map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		h.utils.ErrorHandler.HandleHTTPError(w, r, err)
		return
	}

	// Success response
	duration := time.Since(startTime)
	h.utils.LogOperation("get_user", "success", true, duration, map[string]interface{}{
		"user_id": userID,
		"email":   user.Email,
	})
	
	h.utils.ErrorHandler.RespondWithSuccess(w, r, user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Decode JSON request
	var req CreateUserRequest
	if err := h.utils.ErrorHandler.DecodeJSONRequest(r, &req); err != nil {
		h.utils.HandlerLogger.LogValidationError("request_body", req, "json_decode", err.Error())
		h.utils.ErrorHandler.HandleHTTPError(w, r, err)
		return
	}

	// Validate required fields (using config-driven validation)
	if err := h.validateCreateUserRequest(&req); err != nil {
		h.utils.ErrorHandler.HandleHTTPError(w, r, err)
		return
	}

	// Call service
	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		duration := time.Since(startTime)
		h.utils.LogOperation("create_user", "failed", false, duration, map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		})
		h.utils.ErrorHandler.HandleHTTPError(w, r, err)
		return
	}

	// Success
	duration := time.Since(startTime)
	h.utils.LogOperation("create_user", "success", true, duration, map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})

	w.WriteHeader(http.StatusCreated)
	h.utils.ErrorHandler.RespondWithSuccess(w, r, user)
}

func (h *UserHandler) validateCreateUserRequest(req *CreateUserRequest) error {
	config := h.utils.Config
	
	// Email validation using config
	if !config.IsEmailAllowed(req.Email) {
		return error.NewValidationError("account", "email", "Email domain not allowed", req.Email)
	}

	// Password validation using config
	minLength := config.GetMinPasswordLength()
	if len(req.Password) < minLength {
		return error.NewValidationError("account", "password", 
			fmt.Sprintf("Password must be at least %d characters", minLength), len(req.Password))
	}

	return nil
}

// =============================================================================
// 3. SERVICE LAYER USAGE
// =============================================================================

type UserService struct {
	repo  UserRepository
	utils *utils.ServiceUtilities
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo:  repo,
		utils: utils.NewServiceUtilities("account"),
	}
}

func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	// Validate configuration
	if err := s.utils.ValidateConfig(); err != nil {
		return nil, err
	}

	// Check business rules using config
	maxLoginAttempts := s.utils.Config.GetMaxLoginAttempts()
	s.utils.ServiceLogger.LogBusinessEvent("account", "service_call", "get_user", "started", map[string]interface{}{
		"user_id": userID,
		"max_login_attempts": maxLoginAttempts,
	})

	// Call repository
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		// Wrap repository error with service context
		return nil, s.utils.ErrorHandler.WrapRepositoryError(err, "account", "get_user_by_id", map[string]interface{}{
			"user_id": userID,
		})
	}

	// Business logic validation
	if err := s.validateUserStatus(user); err != nil {
		return nil, err
	}

	s.utils.ServiceLogger.LogUserActivity(
		fmt.Sprintf("%d", user.ID),
		user.Email,
		"profile_accessed",
		"user",
		map[string]interface{}{
			"access_method": "by_id",
		},
	)

	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	// Business rule validation using config
	err := s.utils.ErrorHandler.ValidateBusinessRules("account", map[string]func() error{
		"unique_email": func() error { return s.validateUniqueEmail(ctx, req.Email) },
		"password_complexity": func() error { return s.validatePasswordComplexity(req.Password) },
		"allowed_domain": func() error { return s.validateEmailDomain(req.Email) },
	})
	if err != nil {
		return nil, err
	}

	// Create user
	user := &User{
		Email:     req.Email,
		Password:  s.hashPassword(req.Password),
		CreatedAt: time.Now(),
		Status:    "active",
	}

	// Save to repository
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, s.utils.ErrorHandler.WrapRepositoryError(err, "account", "create_user", map[string]interface{}{
			"email": req.Email,
		})
	}

	// Log user creation
	s.utils.ServiceLogger.LogUserActivity(
		fmt.Sprintf("%d", user.ID),
		user.Email,
		"account_created",
		"user",
		map[string]interface{}{
			"registration_method": "web_form",
		},
	)

	return user, nil
}

func (s *UserService) validateUniqueEmail(ctx context.Context, email string) error {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return s.utils.HandleError(err, "validate_unique_email", map[string]interface{}{
			"email": email,
		})
	}
	
	if exists {
		return error.NewDuplicateEmailError(email)
	}
	
	return nil
}

func (s *UserService) validatePasswordComplexity(password string) error {
	config := s.utils.Config
	
	if len(password) < config.GetMinPasswordLength() {
		return error.NewWeakPasswordError([]string{
			fmt.Sprintf("Minimum %d characters", config.GetMinPasswordLength()),
		})
	}
	
	if config.IsPasswordComplexityRequired() {
		// Add complexity checks
		if !s.hasSpecialChars(password) {
			return error.NewWeakPasswordError([]string{"Must contain special characters"})
		}
	}
	
	return nil
}

func (s *UserService) validateEmailDomain(email string) error {
	if !s.utils.Config.IsEmailAllowed(email) {
		allowedDomains := s.utils.Config.GetAllowedEmailDomains()
		return error.NewValidationError("account", "email", "Email domain not allowed", map[string]interface{}{
			"allowed_domains": allowedDomains,
		})
	}
	return nil
}

// =============================================================================
// 4. REPOSITORY LAYER USAGE
// =============================================================================

type UserRepository struct {
	db    Database
	utils *utils.RepositoryUtilities
}

func NewUserRepository(db Database) *UserRepository {
	return &UserRepository{
		db:    db,
		utils: utils.NewRepositoryUtilities("account"),
	}
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	startTime := time.Now()
	
	query := "SELECT id, email, password, status, created_at FROM users WHERE id = ?"
	user := &User{}
	
	err := r.db.GetContext(ctx, user, query, id)
	duration := time.Since(startTime)
	
	// Log database operation
	r.utils.RepoLogger.LogDatabaseOperation("select", "users", duration, err == nil, 1)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, error.NewUserNotFoundByID(id)
		}
		
		// Handle database error with context
		return nil, r.utils.ErrorHandler.HandleDatabaseError(err, "account", "users", "select", map[string]interface{}{
			"user_id": id,
			"query":   query,
		})
	}

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	startTime := time.Now()
	
	query := `INSERT INTO users (email, password, status, created_at) 
			  VALUES (?, ?, ?, ?) RETURNING id`
	
	err := r.db.GetContext(ctx, &user.ID, query, user.Email, user.Password, user.Status, user.CreatedAt)
	duration := time.Since(startTime)
	
	r.utils.RepoLogger.LogDatabaseOperation("insert", "users", duration, err == nil, 1)
	
	if err != nil {
		// Check for duplicate email constraint
		if r.isDuplicateEmailError(err) {
			return error.NewDuplicateEmailError(user.Email)
		}
		
		return r.utils.ErrorHandler.HandleDatabaseError(err, "account", "users", "insert", map[string]interface{}{
			"email":      user.Email,
			"operation":  "create_user",
		})
	}

	return nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	startTime := time.Now()
	
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)"
	var exists bool
	
	err := r.db.GetContext(ctx, &exists, query, email)
	duration := time.Since(startTime)
	
	r.utils.RepoLogger.LogDatabaseOperation("select", "users", duration, err == nil, 1)
	
	if err != nil {
		return false, r.utils.ErrorHandler.HandleDatabaseError(err, "account", "users", "exists_check", map[string]interface{}{
			"email": email,
		})
	}

	return exists, nil
}

// =============================================================================
// 5. MIDDLEWARE USAGE
// =============================================================================

func setupRoutes() http.Handler {
	mux := http.NewServeMux()
	
	// Setup error middleware
	errorMiddleware := error.NewErrorMiddleware()
	
	// Chain middleware
	handler := errorMiddleware.LoggingMiddleware(
		errorMiddleware.RequestIDMiddleware(
			errorMiddleware.DomainMiddleware("account")(
				errorMiddleware.RecoveryMiddleware(mux),
			),
		),
	)
	
	// Add routes
	userHandler := NewUserHandler(NewUserService(NewUserRepository(db)))
	mux.HandleFunc("/api/users/", userHandler.GetUser)
	mux.HandleFunc("/api/users", userHandler.CreateUser)
	
	return handler
}

// =============================================================================
// 6. HEALTH CHECK AND MONITORING
// =============================================================================

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	um := utils.GetUtilityManager()
	health := um.HealthCheck(r.Context())
	
	// Log health check
	utils.Logger().LogHealthCheck("utility_system", "healthy", 0, health)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	um := utils.GetUtilityManager()
	metrics := um.GetMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// =============================================================================
// 7. GRACEFUL SHUTDOWN
// =============================================================================

func gracefulShutdown() {
	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Log shutdown start
	utils.Logger().LogBusinessEvent("system", "shutdown", "application", "started", nil)
	
	// Shutdown utility systems
	if err := utils.GetUtilityManager().Shutdown(ctx); err != nil {
		utils.Logger().LogSecurityEvent("shutdown_error", "error", "", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	utils.Logger().LogBusinessEvent("system", "shutdown", "application", "completed", nil)
}

// =============================================================================
// 8. TESTING UTILITIES
// =============================================================================

// TestUtilities provides utilities specifically for testing
type TestUtilities struct {
	Config       *config.Config
	Logger       *logger.SpecializedLogger
	ErrorHandler *error.UnifiedErrorHandler
}

// NewTestUtilities creates utilities optimized for testing
func NewTestUtilities() *TestUtilities {
	// Create test configuration
	testConfig := &config.Config{
		Environment: config.EnvTesting,
		AppName:     "test-app",
		Version:     "test",
		Debug:       true,
		// Add minimal required config for tests
	}
	
	// Create test logger (typically with higher verbosity)
	testLogger := logger.NewSpecializedLogger(logger.NewComponentLogger("test"))
	testLogger.SetLevel(logger.LevelDebug)
	testLogger.AddContextField("testing", true)
	
	return &TestUtilities{
		Config:       testConfig,
		Logger:       testLogger,
		ErrorHandler: error.NewUnifiedErrorHandler(),
	}
}

// MockUtilityManager for testing
type MockUtilityManager struct {
	*TestUtilities
	MockResponses map[string]interface{}
}

func NewMockUtilityManager() *MockUtilityManager {
	return &MockUtilityManager{
		TestUtilities: NewTestUtilities(),
		MockResponses: make(map[string]interface{}),
	}
}

// =============================================================================
// 9. CONFIGURATION PATTERNS
// =============================================================================

// ConfigBasedFeatureToggle demonstrates config-driven feature flags
func (s *UserService) isFeatureEnabled(feature string) bool {
	// Use configuration to toggle features
	switch feature {
	case "email_verification":
		return s.utils.Config.IsEmailVerificationRequired()
	case "password_complexity":
		return s.utils.Config.IsPasswordComplexityRequired()
	case "rate_limiting":
		return s.utils.Config.IsRateLimitEnabled()
	default:
		return false
	}
}

// ConfigBasedValidation shows how to use config for dynamic validation
func (h *UserHandler) validateWithConfig(data interface{}) error {
	config := h.utils.Config
	
	// Different validation rules based on environment
	if config.IsProduction() {
		// Stricter validation in production
		return h.strictValidation(data)
	} else if config.IsDevelopment() {
		// Relaxed validation in development
		return h.relaxedValidation(data)
	}
	
	return h.standardValidation(data)
}

// =============================================================================
// 10. PERFORMANCE MONITORING
// =============================================================================

// PerformanceMonitor wraps operations with performance tracking
type PerformanceMonitor struct {
	logger *logger.SpecializedLogger
	config *config.Config
}

func NewPerformanceMonitor(domain string) *PerformanceMonitor {
	return &PerformanceMonitor{
		logger: utils.NewDomainUtilities(domain).Logger,
		config: utils.Config(),
	}
}

func (pm *PerformanceMonitor) TrackOperation(operationName string, fn func() error) error {
	startTime := time.Now()
	
	err := fn()
	duration := time.Since(startTime)
	
	// Log performance metrics
	pm.logger.LogPerformance(operationName, duration, err == nil, map[string]interface{}{
		"operation": operationName,
		"success":   err == nil,
	})
	
	// Alert on slow operations (configurable threshold)
	if duration > time.Duration(pm.config.GetPerformanceThreshold()) {
		pm.logger.LogSecurityEvent("slow_operation", "warning", "", "", map[string]interface{}{
			"operation": operationName,
			"duration":  duration,
			"threshold": pm.config.GetPerformanceThreshold(),
		})
	}
	
	return err
}

// =============================================================================
// 11. ASYNC PROCESSING WITH UTILITIES
// =============================================================================

// AsyncProcessor demonstrates async operations with proper utility integration
type AsyncProcessor struct {
	utils       *utils.DomainUtilities
	workerCount int
	jobQueue    chan Job
}

type Job struct {
	ID     string
	Type   string
	Data   interface{}
	Domain string
}

func NewAsyncProcessor(domain string, workerCount int) *AsyncProcessor {
	return &AsyncProcessor{
		utils:       utils.NewDomainUtilities(domain),
		workerCount: workerCount,
		jobQueue:    make(chan Job, workerCount*2),
	}
}

func (ap *AsyncProcessor) Start(ctx context.Context) {
	for i := 0; i < ap.workerCount; i++ {
		go ap.worker(ctx, i)
	}
	
	ap.utils.Logger.LogBusinessEvent(ap.utils.Domain, "async_processor", "workers", "started", map[string]interface{}{
		"worker_count": ap.workerCount,
	})
}

func (ap *AsyncProcessor) worker(ctx context.Context, workerID int) {
	workerLogger := logger.NewSpecializedLogger(logger.NewComponentLogger("async_worker"))
	workerLogger.AddContextField("worker_id", workerID)
	workerLogger.AddContextField("domain", ap.utils.Domain)
	
	for {
		select {
		case <-ctx.Done():
			workerLogger.LogBusinessEvent(ap.utils.Domain, "worker", "shutdown", "graceful", nil)
			return
		case job := <-ap.jobQueue:
			ap.processJob(ctx, job, workerLogger)
		}
	}
}

func (ap *AsyncProcessor) processJob(ctx context.Context, job Job, workerLogger *logger.SpecializedLogger) {
	startTime := time.Now()
	
	defer func() {
		if r := recover(); r != nil {
			workerLogger.LogSecurityEvent("job_panic", "critical", "", "", map[string]interface{}{
				"job_id":   job.ID,
				"job_type": job.Type,
				"panic":    r,
			})
		}
	}()
	
	workerLogger.LogBusinessEvent(job.Domain, "job", job.Type, "started", map[string]interface{}{
		"job_id": job.ID,
	})
	
	// Process job with error handling
	err := ap.executeJob(ctx, job)
	duration := time.Since(startTime)
	
	if err != nil {
		// Handle job error
		handledErr := ap.utils.HandleError(err, "process_job", map[string]interface{}{
			"job_id":   job.ID,
			"job_type": job.Type,
			"duration": duration,
		})
		
		workerLogger.LogBusinessEvent(job.Domain, "job", job.Type, "failed", map[string]interface{}{
			"job_id":   job.ID,
			"error":    handledErr.Error(),
			"duration": duration,
		})
		return
	}
	
	workerLogger.LogPerformance(fmt.Sprintf("job_%s", job.Type), duration, true, map[string]interface{}{
		"job_id": job.ID,
	})
}

func (ap *AsyncProcessor) executeJob(ctx context.Context, job Job) error {
	// Job execution logic here
	// This would contain your actual business logic
	return nil
}

// =============================================================================
// 12. CACHING WITH UTILITIES
// =============================================================================

// CacheManager integrates caching with logging and error handling
type CacheManager struct {
	cache map[string]interface{}
	utils *utils.DomainUtilities
	mu    sync.RWMutex
}

func NewCacheManager(domain string) *CacheManager {
	return &CacheManager{
		cache: make(map[string]interface{}),
		utils: utils.NewDomainUtilities(domain),
	}
}

func (cm *CacheManager) Get(key string) (interface{}, bool) {
	startTime := time.Now()
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	value, exists := cm.cache[key]
	duration := time.Since(startTime)
	
	// Log cache operation
	cm.utils.Logger.LogCacheOperation("get", key, exists, duration)
	
	return value, exists
}

func (cm *CacheManager) Set(key string, value interface{}) {
	startTime := time.Now()
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.cache[key] = value
	duration := time.Since(startTime)
	
	cm.utils.Logger.LogCacheOperation("set", key, true, duration)
	
	// Log cache size if it's getting large
	if len(cm.cache) > cm.utils.Config.GetCacheSizeThreshold() {
		cm.utils.Logger.LogMetric("cache_size", float64(len(cm.cache)), "entries", nil)
	}
}

// =============================================================================
// 13. BEST PRACTICES SUMMARY
// =============================================================================

/*
BEST PRACTICES FOR USING THE INTEGRATED UTILITY SYSTEM:

1. INITIALIZATION:
   - Always initialize utilities at application start
   - Use MustInitializeUtilities for critical applications
   - Handle initialization errors gracefully

2. LAYER SEPARATION:
   - Use domain-specific utilities (NewDomainUtilities)
   - Use layer-specific utilities (Handler/Service/Repository)
   - Don't cross layer boundaries with utility instances

3. ERROR HANDLING:
   - Always use domain context in errors
   - Wrap repository errors at service layer
   - Use specific error types (NotFound, Validation, etc.)
   - Log errors with appropriate context

4. LOGGING:
   - Use specialized logging methods for different types of events
   - Include relevant context in all log entries
   - Log performance metrics for critical operations
   - Use appropriate log levels based on environment

5. CONFIGURATION:
   - Use config for feature toggles and business rules
   - Environment-specific behavior through config
   - Validate configuration at service initialization
   - React to config changes when needed

6. PERFORMANCE:
   - Monitor operation duration
   - Log slow operations
   - Use metrics for system health
   - Cache frequently accessed config values

7. TESTING:
   - Use TestUtilities for unit tests
   - Mock utility components when needed
   - Test error handling paths
   - Validate logging output in tests

8. ASYNC OPERATIONS:
   - Propagate context properly
   - Handle panics gracefully
   - Log async operation lifecycle
   - Use appropriate error handling

9. GRACEFUL SHUTDOWN:
   - Stop utilities in correct order
   - Log shutdown process
   - Handle shutdown timeouts
   - Clean up resources properly

10. MONITORING:
    - Implement health checks
    - Expose metrics endpoints
    - Monitor utility system health
    - Alert on critical failures
*/



# Complete Go Utility Integration Project Structure

## Recommended Directory Structure

```
your-project/
├── golang/
│   ├── main.go                          # Application entry point
│   ├── go.mod                          # Go module file
│   ├── go.sum                          # Dependencies checksum
│   │
│   ├── config/
│   │   ├── config.yaml                 # Main configuration
│   │   ├── config.development.yaml     # Development overrides
│   │   ├── config.production.yaml      # Production overrides
│   │   └── config.staging.yaml         # Staging overrides
│   │
│   ├── utils/                          # Integrated utility system
│   │   ├── manager.go                  # Main utility manager
│   │   ├── domain.go                   # Domain-specific utilities
│   │   ├── testing.go                  # Testing utilities
│   │   │
│   │   ├── config/                     # Configuration system
│   │   │   ├── utils_config_type.go
│   │   │   ├── utils_config_manager.go
│   │   │   ├── utils_config_default.go
│   │   │   ├── utils_config_env.go
│   │   │   ├── utils_config_global.go
│   │   │   ├── utils_config_interface.go
│   │   │   └── utils_config_utility.go
│   │   │
│   │   ├── logger/                     # Logging system
│   │   │   ├── core/
│   │   │   │   ├── logger_core.go
│   │   │   │   └── logger_type.go
│   │   │   ├── logger_specialized.go
│   │   │   ├── logger_output.go
│   │   │   ├── logger_formatters.go
│   │   │   ├── logger_factory.go
│   │   │   └── logger_global.go
│   │   │
│   │   └── error/                      # Error handling system
│   │       ├── error_types.go
│   │       ├── error_codes.go
│   │       ├── error_constructors.go
│   │       ├── error_handlers.go
│   │       ├── error_middleware.go
│   │       └── error_http.go
│   │
│   ├── internal/                       # Internal application packages
│   │   ├── domain/                     # Domain models and interfaces
│   │   │   ├── account/               # Account domain
│   │   │   │   ├── model.go
│   │   │   │   ├── repository.go
│   │   │   │   ├── service.go
│   │   │   │   └── handler.go
│   │   │   ├── auth/                   # Authentication domain
│   │   │   └── admin/                  # Admin domain
│   │   │
│   │   ├── infrastructure/             # Infrastructure layer
│   │   │   ├── database/
│   │   │   │   ├── connection.go
│   │   │   │   └── migrations/
│   │   │   ├── http/
│   │   │   │   ├── server.go
│   │   │   │   ├── middleware.go
│   │   │   │   └── routes.go
│   │   │   └── cache/
│   │   │       └── manager.go
│   │   │
│   │   └── application/                # Application services
│   │       ├── services/
│   │       └── dto/
│   │
│   ├── api/                           # API definitions
│   │   ├── v1/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── routes.go
│   │   └── proto/                      # gRPC definitions (if using)
│   │
│   ├── pkg/                           # Public packages
│   │   ├── validator/
│   │   ├── jwt/
│   │   └── encryption/
│   │
│   ├── scripts/                       # Build and deployment scripts
│   │   ├── build.sh
│   │   ├── deploy.sh
│   │   └── migrate.sh
│   │
│   ├── tests/                         # Test files
│   │   ├── integration/
│   │   ├── unit/
│   │   └── fixtures/
│   │
│   └── docs/                          # Documentation
│       ├── api/
│       ├── deployment/
│       └── architecture.md
```

## Implementation Files

### 1. Main Application Entry Point

```go
// golang/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"your-project/golang/utils"
	"your-project/golang/internal/infrastructure/http"
	"your-project/golang/internal/infrastructure/database"
)

func main() {
	// Initialize utilities first
	configPath := getConfigPath()
	if err := utils.InitializeUtilities(configPath); err != nil {
		log.Fatal("Failed to initialize utilities:", err)
	}

	// Get utility manager for cleanup
	um := utils.GetUtilityManager()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := um.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Initialize database
	db, err := database.Initialize(utils.Config())
	if err != nil {
		utils.Logger().LogSecurityEvent("startup_failed", "critical", "", "", map[string]interface{}{
			"component": "database",
			"error": err.Error(),
		})
		log.Fatal("Database initialization failed:", err)
	}
	defer db.Close()

	// Initialize HTTP server
	server := http.NewServer(utils.Config(), db)
	
	// Start server in goroutine
	go func() {
		utils.Logger().LogBusinessEvent("system", "startup", "http_server", "starting", map[string]interface{}{
			"address": server.Addr,
		})
		
		if err := server.Start(); err != nil {
			utils.Logger().LogSecurityEvent("server_error", "critical", "", "", map[string]interface{}{
				"error": err.Error(),
			})
			log.Fatal("Server start failed:", err)
		}
	}()

	// Wait for shutdown signal
	waitForShutdown()
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		utils.Logger().LogSecurityEvent("shutdown_error", "error", "", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	utils.Logger().LogBusinessEvent("system", "shutdown", "application", "completed", nil)
}

func getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "./config/config.yaml"
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.Logger().LogBusinessEvent("system", "shutdown", "signal", "received", nil)
}
```

### 2. Utility Manager Domain Extensions

```go
// golang/utils/domain.go
package utils

import (
	"fmt"
	"time"
)

// DomainManager handles domain-specific utility coordination
type DomainManager struct {
	domain    string
	utilities map[string]*DomainUtilities
}

func NewDomainManager(domain string) *DomainManager {
	return &DomainManager{
		domain:    domain,
		utilities: make(map[string]*DomainUtilities),
	}
}

// GetOrCreateUtilities gets or creates utilities for a specific component
func (dm *DomainManager) GetOrCreateUtilities(component string) *DomainUtilities {
	key := fmt.Sprintf("%s.%s", dm.domain, component)
	
	if utils, exists := dm.utilities[key]; exists {
		return utils
	}
	
	utils := NewDomainUtilities(dm.domain)
	utils.Logger.SetComponent(component)
	dm.utilities[key] = utils
	
	return utils
}

// OperationTracker provides standardized operation tracking
type OperationTracker struct {
	domain    string
	operation string
	startTime time.Time
	logger    *logger.SpecializedLogger
	context   map[string]interface{}
}

func NewOperationTracker(domain, operation string) *OperationTracker {
	return &OperationTracker{
		domain:    domain,
		operation: operation,
		startTime: time.Now(),
		logger:    NewDomainUtilities(domain).Logger,
		context:   make(map[string]interface{}),
	}
}

func (ot *OperationTracker) AddContext(key string, value interface{}) *OperationTracker {
	ot.context[key] = value
	return ot
}

func (ot *OperationTracker) Success(result interface{}) {
	duration := time.Since(ot.startTime)
	ot.context["duration"] = duration
	ot.context["result"] = result
	
	ot.logger.LogBusinessEvent(ot.domain, "operation", ot.operation, "success", ot.context)
}

func (ot *OperationTracker) Error(err error) {
	duration := time.Since(ot.startTime)
	ot.context["duration"] = duration
	ot.context["error"] = err.Error()
	
	ot.logger.LogBusinessEvent(ot.domain, "operation", ot.operation, "failed", ot.context)
}
```

### 3. HTTP Server Implementation

```go
// golang/internal/infrastructure/http/server.go
package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"your-project/golang/utils"
	"your-project/golang/utils/config"
	"your-project/golang/api/v1"
)

type Server struct {
	server *http.Server
	config *config.Config
	utils  *utils.DomainUtilities
}

func NewServer(cfg *config.Config, db Database) *Server {
	utils := utils.NewDomainUtilities("system")
	
	mux := http.NewServeMux()
	
	// Setup routes
	v1.RegisterRoutes(mux, db)
	
	// Setup middleware
	handler := setupMiddleware(mux)
	
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	
	return &Server{
		server: server,
		config: cfg,
		utils:  utils,
	}
}

func (s *Server) Start() error {
	s.utils.LogOperation("start_server", "starting", true, nil, map[string]interface{}{
		"address": s.server.Addr,
	})
	
	if s.config.Server.TLSEnabled {
		return s.server.ListenAndServeTLS(s.config.Server.CertFile, s.config.Server.KeyFile)
	}
	
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.utils.LogOperation("shutdown_server", "starting", true, nil, nil)
	return s.server.Shutdown(ctx)
}

func setupMiddleware(handler http.Handler) http.Handler {
	// Create error middleware
	errorMW := error.NewErrorMiddleware()
	
	// Chain middleware in order
	return errorMW.LoggingMiddleware(
		errorMW.RequestIDMiddleware(
			errorMW.AutoDomainMiddleware(
				errorMW.RecoveryMiddleware(
					RateLimitMiddleware(handler),
				),
			),
		),
	)
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implement rate limiting logic using config
		config := utils.Config()
		
		if config.IsRateLimitEnabled() {
			// Check rate limit
			if isRateLimited(r) {
				utils.ErrorHandler().HandleHTTPError(w, r, 
					error.NewRateLimitExceededError("system", "api_request", 
						config.RateLimit.PerMinute, "per minute"))
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}
```

### 4. Testing Framework

```go
// golang/utils/testing.go
package utils

import (
	"testing"
	"context"
	"time"

	"your-project/golang/utils/config"
	"your-project/golang/utils/logger"
	"your-project/golang/utils/error"
)

// TestSuite provides comprehensive testing utilities
type TestSuite struct {
	t        *testing.T
	config   *config.Config
	logger   *logger.SpecializedLogger
	errorHandler *error.UnifiedErrorHandler
	cleanup  []func()
}

func NewTestSuite(t *testing.T) *TestSuite {
	// Create test configuration
	cfg := &config.Config{
		Environment: config.EnvTesting,
		AppName:     "test-suite",
		Version:     "test",
		Debug:       true,
		Server: config.ServerConfig{
			Address: "localhost",
			Port:    0, // Random port for testing
		},
		Database: config.DatabaseConfig{
			Name: "test_db",
			User: "test_user",
		},
	}
	
	// Create test logger with memory output
	testLogger := logger.NewSpecializedLogger(logger.NewComponentLogger("test"))
	testLogger.SetLevel(logger.LevelDebug)
	
	return &TestSuite{
		t:            t,
		config:       cfg,
		logger:       testLogger,
		errorHandler: error.NewUnifiedErrorHandler(),
		cleanup:      make([]func(), 0),
	}
}

func (ts *TestSuite) Config() *config.Config {
	return ts.config
}

func (ts *TestSuite) Logger() *logger.SpecializedLogger {
	return ts.logger
}

func (ts *TestSuite) ErrorHandler() *error.UnifiedErrorHandler {
	return ts.errorHandler
}

func (ts *TestSuite) AddCleanup(fn func()) {
	ts.cleanup = append(ts.cleanup, fn)
}

func (ts *TestSuite) Teardown() {
	for i := len(ts.cleanup) - 1; i >= 0; i-- {
		ts.cleanup[i]()
	}
}

// AssertNoError fails the test if err is not nil
func (ts *TestSuite) AssertNoError(err error) {
	ts.t.Helper()
	if err != nil {
		ts.t.Errorf("Expected no error, got: %v", err)
	}
}

// AssertError fails the test if err is nil
func (ts *TestSuite) AssertError(err error) {
	ts.t.Helper()
	if err == nil {
		ts.t.Error("Expected error, got nil")
	}
}

// AssertErrorType checks if error is of specific type
func (ts *TestSuite) AssertErrorType(err error, expectedType string) {
	ts.t.Helper()
	if err == nil {
		ts.t.Error("Expected error, got nil")
		return
	}
	
	if domainErr, ok := err.(error.DomainError); ok {
		if domainErr.GetErrorType() != expectedType {
			ts.t.Errorf("Expected error type %s, got %s", expectedType, domainErr.GetErrorType())
		}
	} else {
		ts.t.Errorf("Expected domain error, got %T", err)
	}
}

// CreateTestContext creates a context with test values
func (ts *TestSuite) CreateTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "test-request-123")
	ctx = context.WithValue(ctx, "domain", "test")
	ctx = context.WithValue(ctx, "user_id", int64(123))
	return ctx
}

// Integration test helpers
func (ts *TestSuite) WithDatabase(testFn func(db Database)) {
	// Setup test database
	db := setupTestDB(ts.config)
	ts.AddCleanup(func() { db.Close() })
	
	testFn(db)
}

func (ts *TestSuite) WithHTTPServer(testFn func(baseURL string)) {
	// Setup test HTTP server
	server := setupTestServer(ts.config)
	
	go server.Start()
	ts.AddCleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	})
	
	baseURL := fmt.Sprintf("http://localhost:%d", ts.config.Server.Port)
	testFn(baseURL)
}
```

## Key Implementation Strategies

### 1. **Centralized Initialization**
- Single entry point for all utilities
- Proper dependency order (Config → Logger → Error Handler)
- Graceful failure handling

### 2. **Domain-Driven Organization**
- Each domain gets its own utility instance
- Domain-specific logging and error handling
- Configuration validation per domain

### 3. **Layer Separation**
- Handler, Service, Repository utilities
- Appropriate logging levels per layer
- Error propagation with context

### 4. **Testing Integration**
- Dedicated test utilities
- Mock-friendly design
- Integration test helpers

### 5. **Performance Monitoring**
- Operation tracking
- Performance metrics
- Configurable thresholds

### 6. **Graceful Degradation**
- Fallback configurations
- Error recovery
- Health checking

This structure provides a robust, scalable foundation for managing your three utility systems efficiently while maintaining clean architecture principles and enabling comprehensive testing and monitoring.