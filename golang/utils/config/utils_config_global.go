package utils_config

import (
	"context"
	"english-ai-full/logger/core"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	// "runtime"
	// "strconv"
	// "strings"
	"sync"
	"time"
)

var (
	globalConfigManager *ConfigManager
	globalMutex         sync.RWMutex
	globalConfig        *Config
)

// InitializeConfig initializes the global configuration manager with comprehensive logging
func InitializeConfig(configPath string) error {
	// Initialize logging context
	logger := core.NewLogger()
	logger.SetComponent(core.ConfigGlobal)
	logger.SetLayer(core.LayerConfig)
	logger.SetOperation("initialize_config")

	logger.Info("Starting global configuration initialization", map[string]interface{}{
		core.FieldOperation:   "initialize_global_config",
		core.FieldConfigPath:  configPath,
		core.FieldMessage:     fmt.Sprintf( configPath, fileExists(configPath)),
		"input_path":          configPath,
		// "input_path_absolute": getAbsolutePath(configPath),
		"input_path_exists":   fileExists(configPath),
		"working_directory":   getCurrentWorkingDir(),
	})
	startTime := time.Now()
	
	// Log mutex acquisition attempt
	logger.Debug("Attempting to acquire global configuration mutex", map[string]interface{}{
		core.FieldOperation: "acquire_mutex",
		"mutex_type":        "global_config_mutex",
	})
	
	mutexAcquireStart := time.Now()
	globalMutex.Lock()
	mutexAcquireDuration := time.Since(mutexAcquireStart)
	
	logger.Debug("Global configuration mutex acquired successfully", map[string]interface{}{
		core.FieldOperation:  "mutex_acquired",
		core.FieldDurationMS: mutexAcquireDuration.Milliseconds(),
		core.FieldSuccess:    true,
	})
	
	defer func() {
		logger.Debug("Releasing global configuration mutex", map[string]interface{}{
			core.FieldOperation: "release_mutex",
		})
		globalMutex.Unlock()
		
		releaseDuration := time.Since(startTime)
		logger.Debug("Global configuration mutex released", map[string]interface{}{
			core.FieldOperation:  "mutex_released",
			core.FieldDurationMS: releaseDuration.Milliseconds(),
		})
	}()

	// Log current global state before initialization
	logger.Debug("Current global configuration state", map[string]interface{}{
		core.FieldOperation:        "pre_init_state",
		"global_config_nil":        globalConfig == nil,
		"global_manager_nil":       globalConfigManager == nil,
		"previous_config_env":      getPreviousConfigEnvironment(),
		"previous_config_app_name": getPreviousConfigAppName(),
	})

	// Create new config manager
	logger.Debug("Creating new configuration manager", map[string]interface{}{
		core.FieldOperation: "create_config_manager",
	})
	
	managerCreateStart := time.Now()
	cm := NewConfigManager()
	managerCreateDuration := time.Since(managerCreateStart)
	
	if cm == nil {
		logger.ErrorWithCause(
			"Failed to create configuration manager",
			core.CauseServiceError,
			core.LayerService,
			"create_config_manager",
			map[string]interface{}{
				core.FieldDurationMS: managerCreateDuration.Milliseconds(),
				core.FieldOperation:  "create_config_manager",
				"manager_nil":        true,
			},
		)
		return fmt.Errorf("failed to create configuration manager")
	}
	
	logger.Debug("Configuration manager created successfully", map[string]interface{}{
		core.FieldOperation:  "config_manager_created",
		core.FieldDurationMS: managerCreateDuration.Milliseconds(),
		core.FieldSuccess:    true,
		"manager_type":       fmt.Sprintf("%T", cm),
	})

	// Create context for config loading
	logger.Debug("Creating context for configuration loading", map[string]interface{}{
		core.FieldOperation: "create_context",
	})
	ctx := context.Background()
	
	// Log context details
	logger.Debug("Context created for config loading", map[string]interface{}{
		core.FieldOperation: "context_created",
		"context_type":      fmt.Sprintf("%T", ctx),
		"context_nil":       ctx == nil,
		"has_deadline":      false, // Background context has no deadline
		"has_cancel":        false, // Background context is not cancellable
	})

	// Load configuration with enhanced error tracking

	
	configLoadStart := time.Now()
	config, err := cm.Load(ctx, configPath)
	configLoadDuration := time.Since(configLoadStart)
	
	if err != nil {
		logger.ErrorWithCause(
			"Configuration loading failed",
			core.CauseServiceError,
			core.LayerService,
			"load_config",
			map[string]interface{}{
				core.FieldError:               err.Error(),
				core.FieldConfigPath:          configPath,
				core.FieldDurationMS:          configLoadDuration.Milliseconds(),
				core.FieldOperation:           "load_config_failed",
				"input_config_path_raw":       fmt.Sprintf("%q", configPath),
				// "input_config_path_absolute":  getAbsolutePath(configPath),
				"error_type":                  fmt.Sprintf("%T", err),
				"config_nil":                  config == nil,
			},
		)
		return fmt.Errorf("failed to initialize config: %w", err)
	}
	
	logger.Info("Configuration loaded successfully", map[string]interface{}{
		core.FieldOperation:  "load_config_success",
		core.FieldSuccess:    true,
		core.FieldDurationMS: configLoadDuration.Milliseconds(),
		core.FieldConfigPath: configPath,
		core.FieldMessage:    fmt.Sprintf("Config loaded in %dms from input: %q", configLoadDuration.Milliseconds(), configPath),
		"input_path":         configPath,
		"load_duration":      configLoadDuration.Milliseconds(),
	})

	// Validate loaded config before assignment
	logger.Debug("Validating loaded configuration before global assignment", map[string]interface{}{
		core.FieldOperation: "validate_loaded_config",
	})
	
	if config == nil {
		logger.ErrorWithCause(
			"Loaded configuration is nil",
			core.CauseValidationFailed,
			core.LayerValidation,
			"validate_loaded_config",
			map[string]interface{}{
				core.FieldOperation: "config_nil_validation",
				"config_nil":        true,
			},
		)
		return fmt.Errorf("loaded configuration is nil")
	}

	// Log configuration details before assignment
	configDetails := map[string]interface{}{
		core.FieldOperation:     "pre_assignment_config_details",
		"config_environment":    config.Environment,
		"config_app_name":       config.AppName,
		"config_version":        config.Version,
		"config_debug":          config.Debug,
		"config_server_port":    config.Server.Port,
		"config_database_host":  config.Database.Host,
		"config_database_name":  config.Database.Name,
		"valid_roles_count":     len(config.ValidRoles),
		"enabled_domains_count": len(config.Domains.Enabled),
	}
	logger.Debug("Configuration details ready for global assignment", configDetails)

	// Assign to global variables
	logger.Debug("Assigning configuration to global variables", map[string]interface{}{
		core.FieldOperation: "assign_globals",
	})
	
	assignmentStart := time.Now()
	globalConfigManager = cm
	globalConfig = config
	assignmentDuration := time.Since(assignmentStart)
	
	logger.Debug("Global variables assigned successfully", map[string]interface{}{
		core.FieldOperation:  "globals_assigned",
		core.FieldDurationMS: assignmentDuration.Milliseconds(),
		core.FieldSuccess:    true,
		"manager_assigned":   globalConfigManager != nil,
		"config_assigned":    globalConfig != nil,
	})

	// Log final state verification
	logger.Debug("Verifying final global configuration state", map[string]interface{}{
		core.FieldOperation: "verify_final_state",
	})
	
	finalVerification := map[string]interface{}{
		core.FieldOperation:          "final_state_verification",
		"global_config_nil":          globalConfig == nil,
		"global_manager_nil":         globalConfigManager == nil,
		"config_environment_match":   globalConfig != nil && globalConfig.Environment == config.Environment,
		"config_app_name_match":      globalConfig != nil && globalConfig.AppName == config.AppName,
		"manager_type_match":         globalConfigManager != nil && fmt.Sprintf("%T", globalConfigManager) == fmt.Sprintf("%T", cm),
	}
	logger.Debug("Final state verification completed", finalVerification)

	// Log successful completion with comprehensive summary
	totalDuration := time.Since(startTime)
	
	// Get the actual config file that was used by the config manager
	var actualConfigFileUsed string

	
	// successSummary := map[string]interface{}{
	// 	core.FieldOperation:                "initialize_config_complete",
	// 	core.FieldDurationMS:               totalDuration.Milliseconds(),
	// 	core.FieldSuccess:                  true,
	// 	core.FieldConfigPath:               configPath,
	// 	core.FieldEnvironment:              config.Environment,
	// 	"input_config_path_raw":            fmt.Sprintf("%q", configPath),
	// 	"input_config_path_absolute":       getAbsolutePath(configPath),
	// 	"actual_config_file_used":          fmt.Sprintf("%q", actualConfigFileUsed),
	// 	"actual_config_file_used_absolute": getAbsolutePath(actualConfigFileUsed),
	// 	"input_vs_actual_paths_match":      configPath == actualConfigFileUsed,
	// 	"mutex_acquire_duration":           mutexAcquireDuration.Milliseconds(),
	// 	"manager_create_duration":          managerCreateDuration.Milliseconds(),
	// 	"config_load_duration":             configLoadDuration.Milliseconds(),
	// 	"assignment_duration":              assignmentDuration.Milliseconds(),
	// 	"total_operations":                 4, // mutex, create, load, assign
	// 	"global_state_ready":               globalConfig != nil && globalConfigManager != nil,
	// 	"config_app_name":                  config.AppName,
	// 	"config_version":                   config.Version,
	// 	"config_environment":               config.Environment,
	// 	"config_debug_mode":                config.Debug,
	// 	"initialization_success":           true,
	// 	"config_path_comparison": map[string]string{
	// 		"input_path":  configPath,
	// 		"actual_path": actualConfigFileUsed,
	// 	},
	// }
	
	logger.Info("Global configuration initialization completed successfully", map[string]interface{}{
		core.FieldOperation:   "initialize_config_complete",
		core.FieldSuccess:     true,
		core.FieldDurationMS:  totalDuration.Milliseconds(),
		core.FieldEnvironment: config.Environment,
		core.FieldConfigPath:  configPath,
		core.FieldMessage:     fmt.Sprintf("Initialization complete: Input=%q, Actual=%q, Duration=%dms", configPath, actualConfigFileUsed, totalDuration.Milliseconds()),
		"input_path":          configPath,
		"actual_path":         actualConfigFileUsed,
		"paths_match":         configPath == actualConfigFileUsed,
		"total_duration":      totalDuration.Milliseconds(),
		"config_environment":  config.Environment,
		"config_app_name":     config.AppName,
	})
	
	return nil
}

// Helper function to get goroutine ID (for debugging purposes)
func getGoroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.Atoi(idField)
	return id
}


// Helper function to get previous config environment (if exists)
func getPreviousConfigEnvironment() string {
	if globalConfig != nil {
		return globalConfig.Environment
	}
	return "none"
}

// Helper function to get previous config app name (if exists)
func getPreviousConfigAppName() string {
	if globalConfig != nil {
		return globalConfig.AppName
	}
	return "none"
}

// // Helper function to get goroutine ID (for debugging purposes)
// func getGoroutineID() int {
// 	var buf [64]byte
// 	n := runtime.Stack(buf[:], false)
// 	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
// 	id, _ := strconv.Atoi(idField)
// 	return id
// }

// // Helper function to get current working directory safely
// func getCurrentWorkingDir() string {
// 	if wd, err := os.Getwd(); err == nil {
// 		return wd
// 	}
// 	return "unknown"
// }

// // Helper function to get previous config environment (if exists)
// func getPreviousConfigEnvironment() string {
// 	if globalConfig != nil {
// 		return globalConfig.Environment
// 	}
// 	return "none"
// }

// // Helper function to get previous config app name (if exists)
// func getPreviousConfigAppName() string {
// 	if globalConfig != nil {
// 		return globalConfig.AppName
// 	}
// 	return "none"
// }

// // Helper function to get goroutine ID (for debugging purposes)
// func getGoroutineID() int {
// 	var buf [64]byte
// 	n := runtime.Stack(buf[:], false)
// 	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
// 	id, _ := strconv.Atoi(idField)
// 	return id
// }

// // Helper function to get current working directory safely
// // func getCurrentWorkingDir() string {
// // 	if wd, err := os.Getwd(); err == nil {
// // 		return wd
// // 	}
// // 	return "unknown"
// // }

// // Helper function to get previous config environment (if exists)
// func getPreviousConfigEnvironment() string {
// 	if globalConfig != nil {
// 		return globalConfig.Environment
// 	}
// 	return "none"
// }

// // Helper function to get previous config app name (if exists)
// func getPreviousConfigAppName() string {
// 	if globalConfig != nil {
// 		return globalConfig.AppName
// 	}
// 	return "none"
// }

// GetConfig returns the global configuration
func GetConfig() *Config {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalConfig
}

// GetConfigManager returns the global config manager
func GetConfigManager() *ConfigManager {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalConfigManager
}

// ReloadGlobalConfig reloads the global configuration
func ReloadGlobalConfig() error {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if globalConfigManager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	ctx := context.Background()
	err := globalConfigManager.Reload(ctx)
	if err != nil {
		return err
	}

	globalConfig = globalConfigManager.GetConfig()
	return nil
}

// Legacy configuration structure for backward compatibility
type LegacyServerConfig struct {
	DatabaseURL string
	Port        int
	Address     string
}

// LoadServerLegacy loads configuration using legacy format for backward compatibility
func LoadServerLegacy() (*LegacyServerConfig, error) {
	// Try to load from global config first
	config := GetConfig()
	if config != nil {
		return &LegacyServerConfig{
			DatabaseURL: config.Database.URL,
			Port:        config.Server.Port,
			Address:     config.Server.Address,
		}, nil
	}

	// Fallback to environment variables
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Try to construct from individual env vars
		dbHost := getEnvOrDefault("DB_HOST", "localhost")
		dbPort := getEnvOrDefault("DB_PORT", "5432")
		dbUser := getEnvOrDefault("DB_USER", "postgres")
		dbPassword := getEnvOrDefault("DB_PASSWORD", "")
		dbName := getEnvOrDefault("DB_NAME", "english_ai")
		
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := fmt.Sscanf(portStr, "%d", &port); err != nil || p != 1 {
			port = 8080
		}
	}

	address := getEnvOrDefault("ADDRESS", "localhost")

	return &LegacyServerConfig{
		DatabaseURL: databaseURL,
		Port:        port,
		Address:     address,
	}, nil
}

// MustInitializeConfig initializes config and panics on error
func MustInitializeConfig(configPath string) {
	if err := InitializeConfig(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}
}

// IsConfigInitialized returns true if global config is initialized
func IsConfigInitialized() bool {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalConfig != nil
}

