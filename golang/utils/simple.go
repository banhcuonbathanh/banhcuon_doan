package utils

import (
	"fmt"
	"os"
	"strconv"
	"english-ai-full/utils/config"
)

// Simple configuration struct for backward compatibility
// type SimpleServerConfig struct {
// 	DatabaseURL string
// 	Port        int
// 	Address     string
// 	GRPCAddress string
// }

// LoadServer loads basic server configuration from environment variables
// func LoadServer() (*SimpleServerConfig, error) {
// 	config := &SimpleServerConfig{
// 		DatabaseURL: constructDatabaseURL(),
// 		Port:        8888, // Default port
// 		Address:     getEnvOrDefault("ADDRESS", "0.0.0.0"),
// 		GRPCAddress: getEnvOrDefault("GRPC_ADDRESS", "0.0.0.0:50051"),
// 	}
	
// 	// Override with DATABASE_URL if provided
// 	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
// 		config.DatabaseURL = dbURL
// 	}
	
// 	return config, nil
// }

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

// new start 


// SimpleServerConfig provides backward compatibility
type SimpleServerConfig struct {
	DatabaseURL    string
	ServerAddress  string
	GRPCAddress    string
	JwtSecret      string
	QuanAnAddress  string // Added this field for backward compatibility
}

// Config type alias for backward compatibility
type Config = utils_config.Config

// LoadServer provides backward compatibility with the old LoadServer function
func LoadServer() (*SimpleServerConfig, error) {
	// Try to load from new config system first
	cfg := utils_config.GetConfig()
	if cfg != nil {
		return &SimpleServerConfig{
			DatabaseURL:   cfg.Database.URL,
			ServerAddress: fmt.Sprintf(":%d", cfg.Server.Port),
			GRPCAddress:   fmt.Sprintf("%s:%d", cfg.Server.GRPCAddress, cfg.Server.GRPCPort),
			JwtSecret:     cfg.JWT.SecretKey,
			QuanAnAddress: fmt.Sprintf("http://%s", cfg.ExternalAPIs.QuanAn.Address),
		}, nil
	}

	// Fallback to environment variables for backward compatibility
	return loadFromEnvironment(), nil
}

// loadFromEnvironment loads configuration from environment variables
func loadFromEnvironment() *SimpleServerConfig {
	// Database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Construct from individual environment variables
		dbHost := getEnvWithDefault("DB_HOST", "localhost")
		dbPort := getEnvWithDefault("DB_PORT", "5432")
		dbUser := getEnvWithDefault("DB_USER", "postgres")
		dbPassword := getEnvWithDefault("DB_PASSWORD", "")
		dbName := getEnvWithDefault("DB_NAME", "english_ai")
		sslMode := getEnvWithDefault("DB_SSLMODE", "disable")
		
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
	}

	// Server address
	serverAddress := fmt.Sprintf(":%s", getEnvWithDefault("SERVER_PORT", "8888"))
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		port := getEnvWithDefault("SERVER_PORT", "8888")
		serverAddress = fmt.Sprintf("%s:%s", addr, port)
	}

	// gRPC address
	grpcHost := getEnvWithDefault("GRPC_ADDRESS", "localhost")
	grpcPort := getEnvWithDefault("GRPC_PORT", "50051")
	grpcAddress := fmt.Sprintf("%s:%s", grpcHost, grpcPort)

	// JWT secret
	jwtSecret := getEnvWithDefault("JWT_SECRET", "kIOopC3C7wA8DQH6FOF2Jfn+UZP8Q02nGxr/EgFMOmo=")

	// QuanAn address
	quanAnAddress := getEnvWithDefault("QUAN_AN_ADDRESS", "localhost:8081")
	if !contains(quanAnAddress, "http") {
		quanAnAddress = "http://" + quanAnAddress
	}

	return &SimpleServerConfig{
		DatabaseURL:   databaseURL,
		ServerAddress: serverAddress,
		GRPCAddress:   grpcAddress,
		JwtSecret:     jwtSecret,
		QuanAnAddress: quanAnAddress,
	}
}

// getEnvWithDefault gets environment variable or returns default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer or returns default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    s[:len(substr)] == substr || 
		    s[len(s)-len(substr):] == substr ||
		    containsAt(s, substr))
}

// containsAt checks if string contains substring at any position
func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
// new end