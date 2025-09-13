package log_output


import (
	"fmt"
	"log"


	"english-ai-full/logger/core"
)

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level           core.Level
	Environment     string
	LogDirectory    string
	EnableConsole   bool
	EnableFileJSON  bool
	EnableFileText  bool
	MaxFileSize     int64 // in MB
	MaxFileAge      int   // in days
	EnableRotation  bool
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() LoggerConfig {
	return LoggerConfig{
		Level:          core.InfoLevel,
		Environment:    "development",
		LogDirectory:   "/Users/monghoaivu/Desktop/code/14:7 restaurant/restaurant-master/golang/logger/actuall_log",
		EnableConsole:  true,
		EnableFileJSON: true,
		EnableFileText: true,
		MaxFileSize:    100, // 100MB
		MaxFileAge:     30,  // 30 days
		EnableRotation: true,
	}
}

// SetupLogger configures and returns a logger with file output
func SetupLogger(config LoggerConfig) (*core.CoreLogger, error) {
	logger := core.NewLogger()
	
	// Set basic configuration
	logger.SetLevel(config.Level)
	logger.SetEnvironment(config.Environment)
	
	// Create output manager
	outputManager := core.NewDefaultOutputManager()
	
	// Remove default console output if disabled
	if !config.EnableConsole {
		outputManager.RemoveOutput("console")
	}
	
	// Add JSON file output if enabled
	if config.EnableFileJSON {
		jsonFileConfig := FileOutputConfig{
			BaseDir:        config.LogDirectory,
			Filename:       "app",
			MaxSize:        config.MaxFileSize * 1024 * 1024, // Convert MB to bytes
			MaxAge:         config.MaxFileAge,
			Compress:       false,
			EnableRotation: config.EnableRotation,
			Format:         JSONFormat,
		}
		
		jsonOutput, err := NewFileOutput(jsonFileConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON file output: %w", err)
		}
		
		outputManager.AddOutput("json_file", jsonOutput)
	}
	
	// Add text file output if enabled
	if config.EnableFileText {
		textFileConfig := FileOutputConfig{
			BaseDir:        config.LogDirectory,
			Filename:       "app",
			MaxSize:        config.MaxFileSize * 1024 * 1024, // Convert MB to bytes
			MaxAge:         config.MaxFileAge,
			Compress:       false,
			EnableRotation: config.EnableRotation,
			Format:         TextFormat,
		}
		
		textOutput, err := NewFileOutput(textFileConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create text file output: %w", err)
		}
		
		outputManager.AddOutput("text_file", textOutput)
	}
	
	// Set the output manager
	logger.SetOutputManager(outputManager)
	
	return logger, nil
}

// SetupProductionLogger creates a production-ready logger configuration
func SetupProductionLogger(logDir string) (*core.CoreLogger, error) {
	config := LoggerConfig{
		Level:          core.InfoLevel,
		Environment:    "production",
		LogDirectory:   logDir,
		EnableConsole:  false, // Disable console in production
		EnableFileJSON: true,  // JSON for structured logging
		EnableFileText: false, // Disable text format in production
		MaxFileSize:    500,   // 500MB
		MaxFileAge:     90,    // 90 days
		EnableRotation: true,
	}
	
	return SetupLogger(config)
}

// SetupDevelopmentLogger creates a development-friendly logger configuration
func SetupDevelopmentLogger(logDir string) (*core.CoreLogger, error) {
	config := LoggerConfig{
		Level:          core.DebugLevel,
		Environment:    "development",
		LogDirectory:   logDir,
		EnableConsole:  true, // Enable console for development
		EnableFileJSON: true, // JSON for structured logging
		EnableFileText: true, // Text for human-readable logs
		MaxFileSize:    100,  // 100MB
		MaxFileAge:     30,   // 30 days
		EnableRotation: true,
	}
	
	return SetupLogger(config)
}

// Example usage functions
func ExampleUsage() {
	// Method 1: Using your specific directory with default config
	config := DefaultConfig()
	logger, err := SetupLogger(config)
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}
	
	// Method 2: Using development setup
	// logger, err := SetupDevelopmentLogger("/Users/monghoaivu/Desktop/code/14:7 restaurant/restaurant-master/golang/logger/actuall_log")
	// if err != nil {
	//     log.Fatalf("Failed to setup development logger: %v", err)
	// }
	
	// Method 3: Using production setup
	// logger, err := SetupProductionLogger("/Users/monghoaivu/Desktop/code/14:7 restaurant/restaurant-master/golang/logger/actuall_log")
	// if err != nil {
	//     log.Fatalf("Failed to setup production logger: %v", err)
	// }
	
	// Configure logger with your application context
	logger.SetComponent("user-service")
	logger.SetLayer(core.LayerService)
	logger.AddContextField("service_name", "restaurant-backend")
	logger.AddContextField("version", "1.0.0")
	
	// Test different log levels
	logger.Info("Application started", map[string]interface{}{
		"port": 8080,
		"env":  "development",
	})
	
	logger.Debug("Database connection established", map[string]interface{}{
		"database": "postgresql",
		"host":     "localhost:5432",
	})
	
	logger.Warn("High memory usage detected", map[string]interface{}{
		"memory_usage": "85%",
		"threshold":    "80%",
	})
	
	// Example error with enhanced logging
	logger.ErrorWithCause(
		"Failed to create user account",
		core.CauseDuplicateEmail,
		core.LayerRepository,
		core.OperationCreateUser,
		map[string]interface{}{
			"attempted_email": "user@example.com",
			"table":          core.TableUsers,
			"function":       core.FuncCreateUser,
		},
	)
	
	// Example with domain and cause
	logger.ErrorWithDomainAndCause(
		"Authentication failed",
		"auth",
		core.CauseInvalidCredentials,
		core.LayerAuth,
		core.OperationLogin,
		map[string]interface{}{
			"attempted_email": "invalid@example.com",
			"method":         "POST",
			"endpoint":       "/api/auth/login",
		},
	)
}

// QuickSetup provides a simple way to setup the logger with your directory
func QuickSetup() *core.CoreLogger {
	logDir := "/Users/monghoaivu/Desktop/code/14:7 restaurant/restaurant-master/golang/logger/actuall_log"
	
	logger, err := SetupDevelopmentLogger(logDir)
	if err != nil {
		log.Printf("Warning: Failed to setup file logging: %v", err)
		// Fallback to console-only logging
		logger = core.NewLogger()
		logger.SetLevel(core.DebugLevel)
		logger.SetEnvironment("development")
	}
	
	return logger
}

// Advanced setup with custom file naming
func SetupLoggerWithCustomNames(baseDir, appName string) (*core.CoreLogger, error) {
	logger := core.NewLogger()
	logger.SetLevel(core.InfoLevel)
	logger.SetEnvironment("development")
	
	outputManager := core.NewDefaultOutputManager()
	
	// Create separate log files for different log levels
	configs := []struct {
		name     string
		filename string
		level    core.Level
	}{
		{"app_json", appName, core.InfoLevel},
		{"error_json", fmt.Sprintf("%s-errors", appName), core.ErrorLevel},
		{"debug_text", fmt.Sprintf("%s-debug", appName), core.DebugLevel},
	}
	
	for _, config := range configs {
		fileConfig := FileOutputConfig{
			BaseDir:        baseDir,
			Filename:       config.filename,
			MaxSize:        100 * 1024 * 1024, // 100MB
			MaxAge:         30,
			EnableRotation: true,
			Format:         JSONFormat,
		}
		
		if config.name == "debug_text" {
			fileConfig.Format = TextFormat
		}
		
		output, err := NewFileOutput(fileConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s output: %w", config.name, err)
		}
		
		outputManager.AddOutput(config.name, output)
	}
	
	logger.SetOutputManager(outputManager)
	return logger, nil
}