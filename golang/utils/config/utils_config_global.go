package utils_config

import (
	"context"
	"fmt"
	"os"
	"sync"
)

var (
	globalConfigManager *ConfigManager
	globalMutex         sync.RWMutex
	globalConfig        *Config
)

// InitializeConfig initializes the global configuration
func InitializeConfig(configPath string) error {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	cm := NewConfigManager()
	ctx := context.Background()
	
	config, err := cm.Load(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	globalConfigManager = cm
	globalConfig = config
	
	return nil
}

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

