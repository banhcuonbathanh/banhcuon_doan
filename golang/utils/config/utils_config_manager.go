package utils_config

import (
	"context"
	"fmt"
	"strings"
	"sync"

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

// Load loads configuration from file and environment variables
func (cm *ConfigManager) Load(ctx context.Context, configPath string) (*Config, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Set defaults first
	cm.setDefaults()

	// Configure viper
	if configPath != "" {
		cm.viper.SetConfigFile(configPath)
	} else {
		cm.viper.SetConfigName("config")
		cm.viper.SetConfigType("yaml")
		cm.viper.AddConfigPath(".")
		cm.viper.AddConfigPath("./config")
		cm.viper.AddConfigPath("./configs")
		cm.viper.AddConfigPath("/etc/english-ai/")
		cm.viper.AddConfigPath("$HOME/.english-ai")
	}

	// Environment variable configuration
	cm.viper.AutomaticEnv()
	cm.viper.SetEnvPrefix("ENGLISH_AI")
	cm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment-specific settings
	if err := cm.handleEnvironmentOverrides(); err != nil {
		return nil, fmt.Errorf("error handling environment overrides: %w", err)
	}

	config := &Config{}
	if err := cm.viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := cm.validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Perform cross-field validation
	if err := cm.validateCrossFields(config); err != nil {
		return nil, fmt.Errorf("cross-field validation failed: %w", err)
	}

	cm.config = config
	return config, nil
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