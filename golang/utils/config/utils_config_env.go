package utils_config

import (
	"fmt"
	"os"
)

// handleEnvironmentOverrides handles environment-specific configuration overrides
func (cm *ConfigManager) handleEnvironmentOverrides() error {
	env := cm.viper.GetString("environment")

	switch env {
	case EnvDocker:
		return cm.handleDockerOverrides()
	case EnvProduction:
		return cm.handleProductionOverrides()
	case EnvTesting:
		return cm.handleTestingOverrides()
	}

	return nil
}

// handleDockerOverrides applies Docker-specific configuration
func (cm *ConfigManager) handleDockerOverrides() error {
	// Construct database URL from environment variables
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		dbUser := getEnvOrDefault("DB_USER", "postgres")
		dbPassword := getEnvOrDefault("DB_PASSWORD", "")
		dbName := getEnvOrDefault("DB_NAME", "english_ai")
		dbPort := getEnvOrDefault("DB_PORT", "5432")

		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName)
		cm.viper.Set("database.url", dbURL)
	}

	// Override server addresses for Docker
	cm.viper.Set("server.address", "0.0.0.0")
	cm.viper.Set("server.grpc_address", "0.0.0.0")

	return nil
}

// handleProductionOverrides applies production-specific configuration
func (cm *ConfigManager) handleProductionOverrides() error {
	// Force secure settings in production
	cm.viper.Set("debug", false)
	cm.viper.Set("security.require_https", true)
	cm.viper.Set("security.csrf_enabled", true)
	cm.viper.Set("server.tls_enabled", true)
	cm.viper.Set("logging.level", "info")

	return nil
}

// handleTestingOverrides applies testing-specific configuration
func (cm *ConfigManager) handleTestingOverrides() error {
	// Use in-memory database for testing
	cm.viper.Set("database.url", "sqlite:///:memory:")
	cm.viper.Set("security.max_login_attempts", 100) // Don't lockout during tests
	cm.viper.Set("rate_limit.enabled", false)         // Disable rate limiting in tests
	cm.viper.Set("logging.level", "error")            // Reduce log noise in tests

	return nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}