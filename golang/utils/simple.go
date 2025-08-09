// Add this to your utils package: golang/utils/simple.go
package utils

import (
	"fmt"
	"os"
)

// Simple configuration struct for backward compatibility
type SimpleServerConfig struct {
	DatabaseURL string
	Port        int
	Address     string
	GRPCAddress string
}

// LoadServer loads basic server configuration from environment variables
func LoadServer() (*SimpleServerConfig, error) {
	config := &SimpleServerConfig{
		DatabaseURL: constructDatabaseURL(),
		Port:        8888, // Default port
		Address:     getEnvOrDefault("ADDRESS", "0.0.0.0"),
		GRPCAddress: getEnvOrDefault("GRPC_ADDRESS", "0.0.0.0:50051"),
	}
	
	// Override with DATABASE_URL if provided
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}
	
	return config, nil
}

// constructDatabaseURL constructs database URL from environment variables
func constructDatabaseURL() string {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "")
	dbname := getEnvOrDefault("DB_NAME", "restaurant")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}