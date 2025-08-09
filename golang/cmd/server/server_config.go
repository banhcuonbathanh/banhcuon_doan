// cmd/server/server_config.go

package main

import (
	"context"
	"log"
	"os"

	utils_config "english-ai-full/utils/config"
	"english-ai-full/utils"
)

// ConfigSetup handles all configuration initialization
type ConfigSetup struct {
	NewConfig    *utils_config.Config
	LegacyConfig *utils.Config
	Manager      *utils_config.ConfigManager
	Context      context.Context
	Cancel       context.CancelFunc
}

// InitializeConfigs loads both new and legacy configurations
func InitializeConfigs() (*ConfigSetup, error) {
	// Load legacy config (for backward compatibility)
	legacyCfg, err := utils.LoadServer()
	if err != nil {
		return nil, err
	}

	// Initialize new configuration system
	ctx, cancel := context.WithCancel(context.Background())
	configManager := utils_config.NewConfigManager()

	// Load new configuration
	configPath := os.Getenv("CONFIG_PATH")
	config, err := configManager.Load(ctx, configPath)
	if err != nil {
		cancel()
		return nil, err
	}

	setup := &ConfigSetup{
		NewConfig:    config,
		LegacyConfig: legacyCfg,
		Manager:      configManager,
		Context:      ctx,
		Cancel:       cancel,
	}

	// Start config watching for hot-reload
	go setup.startConfigWatching()

	return setup, nil
}

// startConfigWatching enables hot-reload functionality
func (cs *ConfigSetup) startConfigWatching() {
	if err := cs.Manager.Watch(cs.Context); err != nil {
		log.Printf("Failed to start config watching: %v", err)
	}
}

// RegisterConfigCallbacks sets up configuration change handlers
func (cs *ConfigSetup) RegisterConfigCallbacks() {
	cs.Manager.RegisterCallback(func(oldConfig, newConfig *utils_config.Config) error {
		log.Printf("Configuration reloaded: %s v%s", newConfig.AppName, newConfig.Version)
		
		// Handle specific configuration changes
		if oldConfig.Logging.Level != newConfig.Logging.Level {
			log.Printf("Log level changed: %s -> %s", oldConfig.Logging.Level, newConfig.Logging.Level)
		}
		
		if oldConfig.Server.Port != newConfig.Server.Port {
			log.Printf("Server port changed: %d -> %d", oldConfig.Server.Port, newConfig.Server.Port)
		}

		return nil
	})
}

// GetServerAddress returns the server address from new config or falls back to legacy
func (cs *ConfigSetup) GetServerAddress() string {
	if cs.NewConfig != nil {
		return cs.NewConfig.GetServerAddress()
	}
	return cs.LegacyConfig.ServerAddress
}

// GetGRPCAddress returns the gRPC address from new config or falls back to legacy
func (cs *ConfigSetup) GetGRPCAddress() string {
	if cs.NewConfig != nil {
		return cs.NewConfig.GetGRPCAddress()
	}
	return cs.LegacyConfig.GRPCAddress
}

// GetJWTSecret returns JWT secret from new config or falls back to legacy
func (cs *ConfigSetup) GetJWTSecret() string {
	if cs.NewConfig != nil && cs.NewConfig.JWT.SecretKey != "" {
		return cs.NewConfig.JWT.SecretKey
	}
	return cs.LegacyConfig.JwtSecret
}

// GetDatabaseURL returns database URL from new config
func (cs *ConfigSetup) GetDatabaseURL() string {
	if cs.NewConfig != nil {
		return cs.NewConfig.GetDatabaseURL()
	}
	return ""
}

// IsProduction checks if running in production environment
func (cs *ConfigSetup) IsProduction() bool {
	if cs.NewConfig != nil {
		return cs.NewConfig.IsProduction()
	}
	return false
}

// IsDevelopment checks if running in development environment
func (cs *ConfigSetup) IsDevelopment() bool {
	if cs.NewConfig != nil {
		return cs.NewConfig.IsDevelopment()
	}
	return true // Default to development if no config
}

// Cleanup properly shuts down configuration management
func (cs *ConfigSetup) Cleanup() {
	if cs.Cancel != nil {
		cs.Cancel()
	}
	if cs.Manager != nil {
		cs.Manager.Stop()
	}
}