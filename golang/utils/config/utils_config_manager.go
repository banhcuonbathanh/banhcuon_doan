package utils_config

import (
	"context"
	"english-ai-full/logger/core"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// ConfigManager manages configuration with hot-reloading and validation
type ConfigManager struct {
	config    *Config
	mu        sync.RWMutex
	validator *validator.Validate
	viper     *viper.Viper
	callbacks []ConfigChangeCallback
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		validator: validator.New(),
		viper:     viper.New(),
		callbacks: make([]ConfigChangeCallback, 0),
	}
}

// Load loads configuration from file and environment variables with comprehensive logging
func (cm *ConfigManager) Load(ctx context.Context, configPath string) (*Config, error) {
	// Initialize logging context
	logger := core.NewLogger()
	logger.SetComponent(core.ConfigManager)
logger.SetLayer(core.LayerConfig)
	logger.SetOperation(core.OperationQuery) // Using query as closest match for config loading
	
	// Add context fields for tracing with detailed input information
	fields := map[string]interface{}{
		core.FieldOperation:           "load_config",
		core.FieldConfigPath:          configPath,
		core.FieldEnvironment:         os.Getenv("ENGLISH_AI_ENVIRONMENT"),
		"input_config_path":           configPath,
		"input_config_path_empty":     configPath == "",
		"input_config_path_length":    len(configPath),
		"context_deadline_set":        ctx != nil && ctx.Err() == nil,
		"working_directory":           getCurrentWorkingDir(),
		"user_home":                   os.Getenv("HOME"),
		"config_env_vars":             getRelevantEnvVars(),
	}
	
	logger.Info("Starting configuration loading process with input analysis", fields)

	
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Set defaults first
	logger.Debug("Setting default configuration values", fields)
	cm.setDefaults()
	logger.Debug("Default configuration values set successfully", fields)

	// Configure viper
	logger.Debug("Configuring viper settings", fields)
	if configPath != "" {
		logger.Info("Using specified config file path", map[string]interface{}{
			core.FieldConfigPath: configPath,
			core.FieldOperation:  "set_config_file",
		})
		cm.viper.SetConfigFile(configPath)
	} else {
		logger.Debug("Using default config search paths", map[string]interface{}{
			core.FieldOperation: "set_default_paths",
			"search_paths": []string{".", "./config", "./configs", "/etc/english-ai/", "$HOME/.english-ai"},
			"config_name": "config",
			"config_type": "yaml",
		})
		cm.viper.SetConfigName("config")
		cm.viper.SetConfigType("yaml")
		cm.viper.AddConfigPath(".")
		cm.viper.AddConfigPath("./config")
		cm.viper.AddConfigPath("./configs")
		cm.viper.AddConfigPath("/etc/english-ai/")
		cm.viper.AddConfigPath("$HOME/.english-ai")
	}

	// Environment variable configuration
	logger.Debug("Configuring environment variable handling", map[string]interface{}{
		core.FieldOperation: "configure_env_vars",
		"env_prefix": "ENGLISH_AI",
		"auto_env": true,
	})
	cm.viper.AutomaticEnv()
	cm.viper.SetEnvPrefix("ENGLISH_AI")
	cm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	logger.Debug("Attempting to read configuration file", fields)
	configReadStartTime := time.Now()
	
	if err := cm.viper.ReadInConfig(); err != nil {
		configReadDuration := time.Since(configReadStartTime)
		
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// This is a real error, not just missing config file
			logger.ErrorWithCause(
				"Failed to read configuration file",
				core.CauseServiceError,
				core.LayerService,
				"read_config_file",
				map[string]interface{}{
					core.FieldError:          err.Error(),
					core.FieldConfigPath:     configPath,
					core.FieldDurationMS:     configReadDuration.Milliseconds(),
					core.FieldOperation:      "read_config_file",
				},
			)
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		
		// Config file not found - this might be OK if using env vars only
		logger.Warn("Configuration file not found, relying on environment variables and defaults", map[string]interface{}{
			core.FieldConfigPath:     configPath,
			core.FieldDurationMS:     configReadDuration.Milliseconds(),
			core.FieldOperation:      "config_file_not_found",
			core.FieldMessage:        "Will use environment variables and defaults",
		})
	} else {
		configReadDuration := time.Since(configReadStartTime)
		usedConfigFile := cm.viper.ConfigFileUsed()
		
		logger.Info("Configuration file read successfully", map[string]interface{}{
			core.FieldConfigPath:     usedConfigFile,
			core.FieldDurationMS:     configReadDuration.Milliseconds(),
			core.FieldOperation:      "config_file_read_success",
			core.FieldSuccess:        true,
		})
	}

	// Override with environment-specific settings
	logger.Debug("Handling environment-specific overrides", map[string]interface{}{
		core.FieldOperation: "environment_overrides",
	})
	envOverrideStartTime := time.Now()
	
	if err := cm.handleEnvironmentOverrides(); err != nil {
		envOverrideDuration := time.Since(envOverrideStartTime)
		
		logger.ErrorWithCause(
			"Failed to handle environment overrides",
			core.CauseServiceError,
			core.LayerService,
			"handle_environment_overrides",
			map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: envOverrideDuration.Milliseconds(),
				core.FieldOperation:  "environment_overrides",
			},
		)
		return nil, fmt.Errorf("error handling environment overrides: %w", err)
	}
	
	envOverrideDuration := time.Since(envOverrideStartTime)
	logger.Debug("Environment overrides processed successfully", map[string]interface{}{
		core.FieldDurationMS: envOverrideDuration.Milliseconds(),
		core.FieldOperation:  "environment_overrides_success",
		core.FieldSuccess:    true,
	})

	// Unmarshal configuration
	logger.Debug("Unmarshaling configuration into struct", map[string]interface{}{
		core.FieldOperation: "unmarshal_config",
	})
	unmarshalStartTime := time.Now()
	
	config := &Config{}
	if err := cm.viper.Unmarshal(config); err != nil {
		unmarshalDuration := time.Since(unmarshalStartTime)
		
		logger.ErrorWithCause(
			"Failed to unmarshal configuration",
			core.CauseORMConversionFailed, // Using ORM conversion as closest match
			core.LayerService,
			"unmarshal_config",
			map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: unmarshalDuration.Milliseconds(),
				core.FieldOperation:  "unmarshal_config",
			},
		)
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	unmarshalDuration := time.Since(unmarshalStartTime)
	logger.Debug("Configuration unmarshaled successfully", map[string]interface{}{
		core.FieldDurationMS: unmarshalDuration.Milliseconds(),
		core.FieldOperation:  "unmarshal_config_success",
		core.FieldSuccess:    true,
	})

	// Validate configuration
	logger.Debug("Starting configuration validation", map[string]interface{}{
		core.FieldOperation: "validate_config",
	})
	validationStartTime := time.Now()
	
	if err := cm.validateConfig(config); err != nil {
		validationDuration := time.Since(validationStartTime)
		
		logger.ErrorWithCause(
			"Configuration validation failed",
			core.CauseValidationFailed,
			core.LayerValidation,
			"validate_config",
			map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: validationDuration.Milliseconds(),
				core.FieldOperation:  "validate_config",
				core.FieldValidationStep: "basic_validation",
			},
		)
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	validationDuration := time.Since(validationStartTime)
	logger.Debug("Basic configuration validation passed", map[string]interface{}{
		core.FieldDurationMS: validationDuration.Milliseconds(),
		core.FieldOperation:  "validate_config_success",
		core.FieldSuccess:    true,
		core.FieldValidationStep: "basic_validation",
	})

	// Perform cross-field validation
	logger.Debug("Starting cross-field validation", map[string]interface{}{
		core.FieldOperation: "cross_field_validation",
	})
	crossValidationStartTime := time.Now()
	
	if err := cm.validateCrossFields(config); err != nil {
		crossValidationDuration := time.Since(crossValidationStartTime)
		
		logger.ErrorWithCause(
			"Cross-field validation failed",
			core.CauseValidationFailed,
			core.LayerValidation,
			"validate_cross_fields",
			map[string]interface{}{
				core.FieldError:      err.Error(),
				core.FieldDurationMS: crossValidationDuration.Milliseconds(),
				core.FieldOperation:  "cross_field_validation",
				core.FieldValidationStep: "cross_field_validation",
			},
		)
		return nil, fmt.Errorf("cross-field validation failed: %w", err)
	}
	
	crossValidationDuration := time.Since(crossValidationStartTime)
	logger.Debug("Cross-field validation passed", map[string]interface{}{
		core.FieldDurationMS: crossValidationDuration.Milliseconds(),
		core.FieldOperation:  "cross_field_validation_success",
		core.FieldSuccess:    true,
		core.FieldValidationStep: "cross_field_validation",
	})

	// Store the config and log its values
	logger.Debug("Storing validated configuration", map[string]interface{}{
		core.FieldOperation: "store_config",
	})
	cm.config = config
	
	// Log detailed config values (sanitized for security)
	// configSummary := cm.buildConfigSummary(config)
	// logger.Info("Configuration values loaded", configSummary)
	
	// Log successful completion with summary

	logger.Info("Configuration loading completed successfully", map[string]interface{}{
	
	
		"config.JWT.SecretKey":   config.JWT.SecretKey,

	})

	return config, nil
}


func getCurrentWorkingDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "unknown"
}

// Helper method to get relevant environment variables
func getRelevantEnvVars() map[string]string {
	envVars := make(map[string]string)
	relevantPrefixes := []string{"ENGLISH_AI_", "DATABASE_", "JWT_", "SMTP_"}
	
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			key := pair[0]
			for _, prefix := range relevantPrefixes {
				if strings.HasPrefix(key, prefix) {
					// Sanitize sensitive values
					if strings.Contains(strings.ToLower(key), "password") || 
					   strings.Contains(strings.ToLower(key), "secret") ||
					   strings.Contains(strings.ToLower(key), "key") {
						envVars[key] = "[REDACTED]"
					} else {
						envVars[key] = pair[1]
					}
					break
				}
			}
		}
	}
	return envVars
}

// Helper method to check if file exists
func fileExists(filename string) bool {
	if filename == "" {
		return false
	}
	_, err := os.Stat(filename)
	return err == nil
}

// Helper method to build sanitized config summary
func (cm *ConfigManager) buildConfigSummary(config *Config) map[string]interface{} {
	return map[string]interface{}{
		core.FieldOperation: "config_summary",
		"app_config": map[string]interface{}{
			"environment":     config.Environment,
			"app_name":        config.AppName,
			"version":         config.Version,
			"debug":           config.Debug,
		},
		"server_config": map[string]interface{}{
			"address":       config.Server.Address,
			"port":          config.Server.Port,
			"grpc_address":  config.Server.GRPCAddress,
			"grpc_port":     config.Server.GRPCPort,
			"read_timeout":  config.Server.ReadTimeout.String(),
			"write_timeout": config.Server.WriteTimeout.String(),
			"idle_timeout":  config.Server.IdleTimeout.String(),
			"tls_enabled":   config.Server.TLSEnabled,
		},
		"database_config": map[string]interface{}{
			"host":             config.Database.Host,
			"port":             config.Database.Port,
			"name":             config.Database.Name,
			"user":             config.Database.User,
			"password":         "[REDACTED]",
			"ssl_mode":         config.Database.SSLMode,
			"max_connections":  config.Database.MaxConnections,
			"max_idle_conns":   config.Database.MaxIdleConns,
			"conn_max_lifetime": config.Database.ConnMaxLifetime.String(),
			"conn_max_idle_time": config.Database.ConnMaxIdleTime.String(),
		},
		"security_config": map[string]interface{}{
			"max_login_attempts":      config.Security.MaxLoginAttempts,
			"account_lockout_minutes": config.Security.AccountLockoutMinutes,
			"session_timeout":         config.Security.SessionTimeout.String(),
			"csrf_enabled":            config.Security.CSRFEnabled,
			"cors_enabled":            config.Security.CORSEnabled,
			"require_https":           config.Security.RequireHTTPS,
			"allowed_origins_count":   len(config.Security.AllowedOrigins),
			"allowed_email_domains":   config.Security.AllowedEmailDomains,
		},
		"password_config": map[string]interface{}{
			"min_length":         config.Password.MinLength,
			"max_length":         config.Password.MaxLength,
			"require_uppercase":  config.Password.RequireUppercase,
			"require_lowercase":  config.Password.RequireLowercase,
			"require_numbers":    config.Password.RequireNumbers,
			"require_special":    config.Password.RequireSpecial,
			"special_chars":      config.Password.SpecialChars,
		},
		"pagination_config": map[string]interface{}{
			"default_size": config.Pagination.DefaultSize,
			"max_size":     config.Pagination.MaxSize,
			"limit":        config.Pagination.Limit,
		},
		"jwt_config": map[string]interface{}{
			"secret_key":                    "[REDACTED]",
			"expiration_hours":              config.JWT.ExpirationHours,
			"refresh_token_expiration_days": config.JWT.RefreshTokenExpirationDays,
			"issuer":                        config.JWT.Issuer,
			"algorithm":                     config.JWT.Algorithm,
			"refresh_threshold":             config.JWT.RefreshThreshold.String(),
		},
		"email_config": map[string]interface{}{
			"verification_enabled":      config.Email.VerificationEnabled,
			"verification_expiry_hours": config.Email.VerificationExpiryHours,
			"require_verification":      config.Email.RequireVerification,
			"smtp_host":                 config.Email.SMTPHost,
			"smtp_port":                 config.Email.SMTPPort,
			"smtp_user":                 config.Email.SMTPUser,
			"smtp_password":             "[REDACTED]",
			"from_address":              config.Email.FromAddress,
			"from_name":                 config.Email.FromName,
		},
		"rate_limit_config": map[string]interface{}{
			"enabled":     config.RateLimit.Enabled,
			"per_minute":  config.RateLimit.PerMinute,
			"per_hour":    config.RateLimit.PerHour,
			"burst_size":  config.RateLimit.BurstSize,
			"window_size": config.RateLimit.WindowSize.String(),
		},
		"logging_config": map[string]interface{}{
			"level":       config.Logging.Level,
			"format":      config.Logging.Format,
			"output":      config.Logging.Output,
			"file_path":   config.Logging.FilePath,
			"max_size":    config.Logging.MaxSize,
			"max_backups": config.Logging.MaxBackups,
			"max_age":     config.Logging.MaxAge,
			"compress":    config.Logging.Compress,
		},
		"external_apis_config": map[string]interface{}{
			"anthropic_api_url":      config.ExternalAPIs.Anthropic.APIURL,
			"anthropic_timeout":      config.ExternalAPIs.Anthropic.Timeout.String(),
			"anthropic_max_retries":  config.ExternalAPIs.Anthropic.MaxRetries,
			"anthropic_api_key":      "[REDACTED]",
			"quan_an_address":        config.ExternalAPIs.QuanAn.Address,
			"quan_an_timeout":        config.ExternalAPIs.QuanAn.Timeout.String(),
			"quan_an_max_retries":    config.ExternalAPIs.QuanAn.MaxRetries,
		},
		"business_config": map[string]interface{}{
			"valid_roles":           config.ValidRoles,
			"valid_account_statuses": config.ValidAccountStatuses,
		},
		"domains_config": map[string]interface{}{
			"enabled":                    config.Domains.Enabled,
			"default":                    config.Domains.Default,
			"error_tracking_enabled":     config.Domains.ErrorTracking.Enabled,
			"error_tracking_log_level":   config.Domains.ErrorTracking.LogLevel,
			"account_max_login_attempts": config.Domains.Account.MaxLoginAttempts,
			"account_password_complexity": config.Domains.Account.PasswordComplexity,
			"account_email_verification": config.Domains.Account.EmailVerification,
		},
		"error_handling_config": map[string]interface{}{
			"include_stack_trace":      config.ErrorHandling.IncludeStackTrace,
			"sanitize_sensitive_data":  config.ErrorHandling.SanitizeSensitiveData,
			"request_id_required":      config.ErrorHandling.RequestIDRequired,
		},
		"domain_error_policies": cm.buildDomainErrorPoliciesSummary(config.DomainErrorPolicies),
	}
}

// Helper method to build domain error policies summary
func (cm *ConfigManager) buildDomainErrorPoliciesSummary(policies map[string]DomainErrorPolicy) map[string]interface{} {
	summary := make(map[string]interface{})
	for domain, policy := range policies {
		summary[domain] = map[string]interface{}{
			"max_retries":      policy.MaxRetries,
			"retry_delay":      policy.RetryDelay.String(),
			"circuit_breaker":  policy.CircuitBreaker,
			"log_level":        policy.LogLevel,
			"alert_threshold":  policy.AlertThreshold,
			"enable_fallback":  policy.EnableFallback,
		}
	}
	return summary
}

// Helper method to determine config source for logging
func (cm *ConfigManager) getConfigSource() string {
	if cm.viper.ConfigFileUsed() != "" {
		return "file_and_env"
	}
	return "env_and_defaults"
}

// Reload reloads the configuration
func (cm *ConfigManager) Reload(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	oldConfig := cm.config

	if err := cm.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error rereading config: %w", err)
	}

	newConfig := &Config{}
	if err := cm.viper.Unmarshal(newConfig); err != nil {
		return fmt.Errorf("error unmarshaling reloaded config: %w", err)
	}

	if err := cm.validateConfig(newConfig); err != nil {
		return fmt.Errorf("reloaded config validation failed: %w", err)
	}

	if err := cm.validateCrossFields(newConfig); err != nil {
		return fmt.Errorf("reloaded config cross-field validation failed: %w", err)
	}

	// Notify callbacks
	for _, callback := range cm.callbacks {
		if err := callback(oldConfig, newConfig); err != nil {
			return fmt.Errorf("config change callback failed: %w", err)
		}
	}

	cm.config = newConfig
	return nil
}

// Watch starts watching for configuration file changes
func (cm *ConfigManager) Watch(ctx context.Context) error {
	cm.viper.WatchConfig()
	cm.viper.OnConfigChange(func(e fsnotify.Event) {
		select {
		case <-ctx.Done():
			return
		default:
			if err := cm.Reload(ctx); err != nil {
				// Log error (in production, use proper logger)
				fmt.Printf("Error reloading config: %v\n", err)
			}
		}
	})

	return nil
}

// GetConfig returns the current configuration (thread-safe)
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// RegisterCallback registers a callback for configuration changes
func (cm *ConfigManager) RegisterCallback(callback ConfigChangeCallback) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.callbacks = append(cm.callbacks, callback)
}

// Validate validates the current configuration
func (cm *ConfigManager) Validate() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if err := cm.validateConfig(cm.config); err != nil {
		return err
	}

	return cm.validateCrossFields(cm.config)
}

// Stop stops the configuration manager
func (cm *ConfigManager) Stop() error {
	// Clean up resources if needed
	return nil
}

// validateConfig validates the configuration using struct tags
func (cm *ConfigManager) validateConfig(config *Config) error {
	return cm.validator.Struct(config)
}



// new 

// validateCrossFields performs cross-field validation
func (cm *ConfigManager) validateCrossFields(config *Config) error {
	// Password length validation
	if config.Password.MaxLength < config.Password.MinLength {
		return fmt.Errorf("password max_length (%d) must be greater than min_length (%d)",
			config.Password.MaxLength, config.Password.MinLength)
	}

	// Pagination validation
	if config.Pagination.DefaultSize > config.Pagination.MaxSize {
		return fmt.Errorf("pagination default_size (%d) cannot exceed max_size (%d)",
			config.Pagination.DefaultSize, config.Pagination.MaxSize)
	}

	// JWT secret validation in production only
	if config.Environment == EnvProduction && len(config.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters in production")
	}

	// HTTPS validation in production only
	if config.Environment == EnvProduction && !config.Security.RequireHTTPS {
		return fmt.Errorf("HTTPS must be enabled in production environment")
	}

	// Email verification validation (more lenient)
	if config.Email.RequireVerification && !config.Email.VerificationEnabled {
		return fmt.Errorf("email verification must be enabled if required")
	}

	// Anthropic API key validation (only warn in development, not error)
	if config.Environment != EnvProduction && config.ExternalAPIs.Anthropic.APIKey == "dummy_key_for_dev" {
		// Just log a warning instead of failing
		fmt.Println("Warning: Using dummy Anthropic API key for development")
	}

	return nil
}