Step 1: Initialize Configuration
Where: In your main.go file before any application logic
Why: To load configuration from files/env vars and make it available globally

go
package main

import (
	"log"
	utils_config "english-ai-full/utils/config"
)

func main() {
	// Initialize configuration
	configPath := getEnvWithDefault("CONFIG_PATH", "./config.yaml")
	err := utils_config.InitializeConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	
	// Rest of your application code
}
Step 2: Create Configuration File
Create config.yaml in your project root:

yaml
environment: development
app_name: "English AI"
version: "1.0.0"

server:
  address: "localhost"
  port: 8888
  grpc_port: 50051

database:
  host: "localhost"
  port: 5432
  name: "english_ai_dev"
  user: "postgres"
  password: ""  # Will be set via env var

security:
  allowed_origins:
    - "http://localhost:3000"

jwt:
  secret_key: ""  # Will be set via env var
  expiration_hours: 24
Step 3: Set Environment Variables
Create .env file or set in your shell:

bash
# .env file
ENGLISH_AI_DATABASE_PASSWORD=your_db_password
ENGLISH_AI_JWT_SECRET_KEY=your_secure_secret_32chars
ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY=your_anthropic_key
Step 4: Access Configuration in Code
Example 1: In a database connection module

go
package database

import (
	utils_config "english-ai-full/utils/config"
	"database/sql"
	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	cfg := utils_config.GetConfig()
	
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	
	return sql.Open("postgres", dbURL)
}
Example 2: In an HTTP handler

go
package handlers

import (
	utils_config "english-ai-full/utils/config"
	"net/http"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := utils_config.GetConfig()
	
	if cfg.IsDevelopment() {
		w.Write([]byte("Development mode: Debug features enabled"))
	} else {
		w.Write([]byte("Production mode"))
	}
}
Step 5: Environment-Specific Configuration
Create config.production.yaml:

yaml
environment: production
debug: false

server:
  address: "0.0.0.0"
  port: 8080

database:
  host: "prod-db.example.com"
  ssl_mode: "require"
Set environment variable to load this config:

bash
export ENGLISH_AI_ENVIRONMENT=production
Step 6: Validate Configuration
Add validation rules in utils_config_type.go:

go
type DatabaseConfig struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required,min=1,max=65535"`
	Name     string `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	SSLMode  string `validate:"oneof=disable require verify-full"`
}
Step 7: Use Utility Methods
go
func main() {
	cfg := utils_config.GetConfig()
	
	// Get formatted addresses
	fmt.Println("HTTP Server:", cfg.GetServerAddress())
	fmt.Println("gRPC Server:", cfg.GetGRPCAddress())
	
	// Check environment
	if cfg.IsProduction() {
		fmt.Println("Running in PRODUCTION mode")
	}
	
	// Get database URL
	fmt.Println("Database URL:", cfg.GetDatabaseURL())
}
Step 8: Handle Configuration Changes (Advanced)
go
func setupHotReload() {
	cm := utils_config.GetConfigManager()
	
	cm.RegisterCallback(func(oldConfig, newConfig *utils_config.Config) error {
		log.Println("Configuration changed - restarting services")
		
		if oldConfig.Database.Host != newConfig.Database.Host {
			restartDatabaseConnection(newConfig.Database)
		}
		
		return nil
	})
	
	// Start watching for changes
	go func() {
		if err := cm.Watch(context.Background()); err != nil {
			log.Printf("Config watch error: %v", err)
		}
	}()
}
Step 9: Access in gRPC Services
go
package service

import (
	utils_config "english-ai-full/utils/config"
	pb "english-ai-full/internal/proto_qr/account"
)

type AccountService struct {
	pb.UnimplementedAccountServiceServer
	config *utils_config.Config
}

func NewAccountService() *AccountService {
	return &AccountService{
		config: utils_config.GetConfig(),
	}
}

func (s *AccountService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	if s.config.IsEmailVerificationRequired() {
		// Send verification email
	}
	// ...
}
Best Practices:
Secrets Management: Never commit secrets to git

yaml
# config.yaml
database:
  password: ""  # Set via ENGLISH_AI_DATABASE_PASSWORD
Environment-Specific Files:

text
config/
  base.yaml
  development.yaml
  production.yaml
Validation First: Add validation tags as you define new config sections

Use Helper Methods:

go
// Instead of:
fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)

// Use:
cfg.GetServerAddress()
Testing: Create test configurations:

go
func TestHandler(t *testing.T) {
    testConfig := &utils_config.Config{
        Environment: "testing",
        Server: utils_config.ServerConfig{Port: 8080},
    }
    utils_config.SetTestConfig(testConfig)
    
    // Run tests
}
Troubleshooting First Steps:
If config isn't loading:

go
// Add debug logging in main.go
cfg := utils_config.GetConfig()
log.Printf("Config: %+v", cfg)  // Redact secrets in production!
Common errors:

"Config validation failed": Check your YAML syntax and required fields

"Environment variable not recognized": Ensure correct prefix and naming

"Nil config": Make sure InitializeConfig() was called successfully

This configuration system provides a robust foundation for your application. Start with the basic setup (Steps 1-4), then gradually adopt more advanced features as needed.