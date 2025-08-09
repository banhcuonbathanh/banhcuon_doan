package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectDataBase connects to the database using environment variables with fallbacks
func ConnectDataBase() (*sql.DB, error) {
	// Check if DATABASE_URL is provided (takes priority)
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return ConnectDataBaseWithURL(dbURL)
	}

	// Get database connection details from environment variables
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "restaurant")
	password := getEnvOrDefault("DB_PASSWORD", "restaurant")
	dbname := getEnvOrDefault("DB_NAME", "restaurant")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	// Handle Docker environment - if running outside Docker but connecting to Docker containers
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "docker" && host != "localhost" && host != "127.0.0.1" {
		fmt.Printf("Warning: Connecting to Docker service '%s' from outside Docker. Using localhost instead.\n", host)
		host = "localhost"
	}

	// Construct connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := getEnvAsDuration("DB_CONN_MAX_LIFETIME", "1h")
	connMaxIdleTime := getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", "10m")

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// Test the connection with retry logic
	if err := pingWithRetry(db, 3, time.Second*2); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database after retries: %w", err)
	}

	fmt.Printf("Successfully connected to database at %s:%s (database: %s)\n", host, port, dbname)
	return db, nil
}

// ConnectDataBaseWithURL connects using a provided database URL
func ConnectDataBaseWithURL(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL cannot be empty")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection with URL: %w", err)
	}

	// Configure connection pool with defaults
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 10)

	// Test the connection
	if err := pingWithRetry(db, 3, time.Second*2); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database with URL: %w", err)
	}

	fmt.Println("Successfully connected to database using DATABASE_URL")
	return db, nil
}

// ConnectDataBaseWithConfig connects using a config struct (for future config integration)
func ConnectDataBaseWithConfig(config DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	if err := pingWithRetry(db, 3, time.Second*2); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Printf("Successfully connected to database at %s:%d (database: %s)\n", 
		config.Host, config.Port, config.Name)
	return db, nil
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxConnections  int           `yaml:"max_connections"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// Helper functions

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer or returns default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvAsDuration gets environment variable as duration or returns default
func getEnvAsDuration(key string, defaultValue string) time.Duration {
	value := getEnvOrDefault(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// If parsing fails, parse the default
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	// Final fallback
	return time.Hour
}

// pingWithRetry attempts to ping the database with retry logic
func pingWithRetry(db *sql.DB, maxRetries int, delay time.Duration) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxRetries-1 {
				fmt.Printf("Database ping attempt %d failed: %v. Retrying in %v...\n", 
					i+1, err, delay)
				time.Sleep(delay)
			}
		}
	}
	
	return fmt.Errorf("database ping failed after %d attempts: %w", maxRetries, lastErr)
}

// IsConnectionValid checks if the database connection is still valid
func IsConnectionValid(db *sql.DB) bool {
	if db == nil {
		return false
	}
	return db.Ping() == nil
}

// CloseConnection safely closes the database connection
func CloseConnection(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}