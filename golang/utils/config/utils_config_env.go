package utils_config

import (
	"fmt"
	"os"
	"strconv"
)









// handleEnvironmentOverrides handles environment-specific configuration overrides
func (cm *ConfigManager) handleEnvironmentOverrides() error {
	env := cm.viper.GetString("environment")

	// Check for APP_ENV environment variable to override environment
	if appEnv := os.Getenv("APP_ENV"); appEnv != "" {
		env = appEnv
		cm.viper.Set("environment", env)
	}

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
		
		// Also set individual database fields
		cm.viper.Set("database.host", dbHost)
		cm.viper.Set("database.user", dbUser)
		cm.viper.Set("database.password", dbPassword)
		cm.viper.Set("database.name", dbName)
		if port, err := strconv.Atoi(dbPort); err == nil {
			cm.viper.Set("database.port", port)
		}
	}

	// Override server addresses for Docker
	cm.viper.Set("server.address", "0.0.0.0")
	cm.viper.Set("server.grpc_address", "0.0.0.0")

	// Handle other environment variables for Docker
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cm.viper.Set("jwt.secret_key", jwtSecret)
	}

	if anthropicKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicKey != "" {
		cm.viper.Set("external_apis.anthropic.api_key", anthropicKey)
	}

	if anthropicURL := os.Getenv("ANTHROPIC_API_URL"); anthropicURL != "" {
		cm.viper.Set("external_apis.anthropic.api_url", anthropicURL)
	}

	if quanAnAddr := os.Getenv("QUAN_AN_ADDRESS"); quanAnAddr != "" {
		cm.viper.Set("external_apis.quan_an.address", quanAnAddr)
	}

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

	// Ensure JWT secret is set from environment in production
	if jwtSecret := os.Getenv("ENGLISH_AI_JWT_SECRET_KEY"); jwtSecret != "" {
		cm.viper.Set("jwt.secret_key", jwtSecret)
	}

	// Ensure database password is set from environment
	if dbPassword := os.Getenv("ENGLISH_AI_DATABASE_PASSWORD"); dbPassword != "" {
		cm.viper.Set("database.password", dbPassword)
	}

	// Ensure Anthropic API key is set from environment
	if anthropicKey := os.Getenv("ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY"); anthropicKey != "" {
		cm.viper.Set("external_apis.anthropic.api_key", anthropicKey)
	}

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