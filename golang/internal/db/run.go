// Add this to your db package or create a new file: golang/internal/db/connection.go
package db

import (
	"database/sql"
	"fmt"
	"os"
)

// ConnectDataBase connects to the database using environment variables
func ConnectDataBase() (*sql.DB, error) {
	// Get database connection details from environment variables
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "")
	dbname := getEnvOrDefault("DB_NAME", "restaurant")
	
	// Construct connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	// Alternative: Use DATABASE_URL if provided
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		connStr = dbURL
	}
	
	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	fmt.Printf("Successfully connected to database at %s:%s\n", host, port)
	return db, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
// ConnectDataBaseLegacy uses the legacy config system for backward compatibility
// func ConnectDataBaseLegacy() (*sql.DB, error) {
// 	// Load legacy configuration
// 	cfg, err := utils_config.LoadServerLegacy()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load legacy config: %v", err)
// 	}

// 	// Use the database URL from legacy config
// 	databaseURL := cfg.DatabaseURL
// 	if databaseURL == "" {
// 		return nil, fmt.Errorf("database URL is not configured")
// 	}

// 	db, err := sql.Open("postgres", databaseURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to open database connection: %v", err)
// 	}

// 	// Test the connection
// 	if err := db.Ping(); err != nil {
// 		db.Close()
// 		return nil, fmt.Errorf("unable to ping database: %v", err)
// 	}

// 	return db, nil
// }

// ConnectDataBaseWithURL alternative function if you want to keep the parameter-based approach
func ConnectDataBaseWithURL(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		// Fall back to config if no URL provided
		return ConnectDataBase()
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	return db, nil
}

// ConnectDataBaseWithConfig connects using a provided config (useful for testing)
// func ConnectDataBaseWithConfig(config *utils_config.Config) (*sql.DB, error) {
// 	if config == nil {
// 		return nil, fmt.Errorf("config cannot be nil")
// 	}

// 	databaseURL := config.Database.URL
// 	if databaseURL == "" {
// 		return nil, fmt.Errorf("database URL is not configured")
// 	}

// 	db, err := sql.Open("postgres", databaseURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to open database connection: %v", err)
// 	}

// 	if err := db.Ping(); err != nil {
// 		db.Close()
// 		return nil, fmt.Errorf("unable to ping database: %v", err)
// 	}

// 	// Configure connection pool using config values
// 	db.SetMaxOpenConns(config.Database.MaxConnections)
// 	db.SetMaxIdleConns(config.Database.MaxIdleConns)
// 	db.SetConnMaxLifetime(config.Database.ConnMaxLifetime)
// 	db.SetConnMaxIdleTime(config.Database.ConnMaxIdleTime)

// 	return db, nil
// }