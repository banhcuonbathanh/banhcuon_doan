package db

import (
	"database/sql"
	"fmt"

	utils_config "english-ai-full/utils/config"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// ConnectDataBase connects to database using the new config system
func ConnectDataBase() (*sql.DB, error) {
	// Use the new config system
	config := utils_config.GetConfig()
	if config == nil {
		// If config is not initialized, try to load it
		if err := utils_config.InitializeConfig(""); err != nil {
			return nil, fmt.Errorf("failed to initialize config: %v", err)
		}
		config = utils_config.GetConfig()
	}

	// Use the database URL from config
	databaseURL := config.Database.URL
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is not configured")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	// Configure connection pool using config values
	db.SetMaxOpenConns(config.Database.MaxConnections)
	db.SetMaxIdleConns(config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(config.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.Database.ConnMaxIdleTime)

	return db, nil
}

// ConnectDataBaseLegacy uses the legacy config system for backward compatibility
func ConnectDataBaseLegacy() (*sql.DB, error) {
	// Load legacy configuration
	cfg, err := utils_config.LoadServerLegacy()
	if err != nil {
		return nil, fmt.Errorf("failed to load legacy config: %v", err)
	}

	// Use the database URL from legacy config
	databaseURL := cfg.DatabaseURL
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is not configured")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	return db, nil
}

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
func ConnectDataBaseWithConfig(config *utils_config.Config) (*sql.DB, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	databaseURL := config.Database.URL
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is not configured")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	// Configure connection pool using config values
	db.SetMaxOpenConns(config.Database.MaxConnections)
	db.SetMaxIdleConns(config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(config.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.Database.ConnMaxIdleTime)

	return db, nil
}