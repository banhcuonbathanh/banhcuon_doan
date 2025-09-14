// cmd/server/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"english-ai-full/internal/account/account_handler" // Add this import
	"english-ai-full/internal/branch"
	"english-ai-full/logger/core"
	log_output "english-ai-full/logger/output"

	branchpb "english-ai-full/internal/proto_qr/branch"
	pb "english-ai-full/internal/proto_qr/account"

	utils_config "english-ai-full/utils/config"
	"english-ai-full/utils"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/ianschenck/envflag"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Global logger instance
var appLogger *core.CoreLogger

func main() {
	// Initialize logging first
	initializeLogging()

	appLogger.Info("=== SERVER STARTUP INITIATED ===", map[string]interface{}{
		"component": "main",
		"operation": "server_startup",
	})

	// Initialize configuration using the new system
	configPath := getEnvWithDefault("CONFIG_PATH", "utils/config/config.yaml")

	err := utils_config.InitializeConfig(configPath)
	if err != nil {
		appLogger.WarnWithCause(
			"Failed to load config file, continuing with environment variables",
			core.CauseResourceNotFound,
			core.LayerHandler,
			"config_initialization",
			map[string]interface{}{
				"config_path": configPath,
				"error":      err.Error(),
			},
		)
		
		// Initialize with empty path to use defaults and environment variables
		err = utils_config.InitializeConfig("")
		if err != nil {
			appLogger.ErrorWithCause(
				"Failed to initialize config",
				"config_initialization_failed",
				core.LayerHandler,
				"config_initialization",
				map[string]interface{}{
					"error": err.Error(),
				},
			)
			log.Fatalf("Failed to initialize config: %v", err)
		}
	}

	// Get the configuration
	cfg := utils_config.GetConfig()
	if cfg == nil {
		appLogger.Fatal("Configuration is nil", map[string]interface{}{
			"component": "main",
			"operation": "config_validation",
		})
	}

	appLogger.Info("Configuration loaded successfully", map[string]interface{}{
		"environment": cfg.Environment,
		"server_port": cfg.Server.Port,
		"grpc_port":   cfg.Server.GRPCPort,
	})

	envflag.Parse()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	r := chi.NewRouter()
	r.Use(utils.RequestIDMiddleware)
	r.Use(loggingMiddleware) // Add logging middleware
	setupCORS(r, cfg)

	// Use environment variable with a default value
	if cfg.Environment == "development" {
		r.Use(debugMiddleware)
		appLogger.Debug("Debug middleware enabled", map[string]interface{}{
			"environment": cfg.Environment,
		})
	}

	/**
	python server
	*/
	appLogger.Info("Connecting to Python gRPC server", map[string]interface{}{
		"address": ":50052",
	})

	pythonConn, err := grpc.NewClient(":50052", opts...)
	if err != nil {
		appLogger.ErrorWithCause(
			"Failed to connect to Python gRPC server",
			core.CauseConnectionRefused,
			core.LayerExternal,
			"grpc_connection",
			map[string]interface{}{
				"address": ":50052",
				"error":   err.Error(),
			},
		)
		log.Fatalf("failed to connect to Python gRPC server: %v", err)
	}
	defer pythonConn.Close()

	appLogger.Info("Python gRPC connection established", map[string]interface{}{
		"address": ":50052",
		"status":  "connected",
	})

	// Construct gRPC address properly
	grpcAddress := fmt.Sprintf("%s:%d", cfg.Server.GRPCAddress, cfg.Server.GRPCPort)
	
	appLogger.Info("Connecting to main gRPC server", map[string]interface{}{
		"address": grpcAddress,
	})

	conn, err := grpc.DialContext(
		context.Background(),
		grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		appLogger.ErrorWithCause(
			"Failed to connect to gRPC server",
			core.CauseConnectionRefused,
			core.LayerExternal,
			"grpc_connection",
			map[string]interface{}{
				"address": grpcAddress,
				"error":   err.Error(),
			},
		)
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	appLogger.Info("Main gRPC connection established", map[string]interface{}{
		"address":          grpcAddress,
		"connection_state": conn.GetState().String(),
	})

	branchClient := branchpb.NewBranchServiceClient(conn)
	b := branch.NewBranchHandler(branchClient)
	branch.RegisterRoutes(r, b)

	appLogger.Info("Branch routes registered", map[string]interface{}{
		"component": "branch",
		"status":    "registered",
	})

	// Setup domain handlers with error handling
	setupDomainHandlers(r, conn, cfg)

	// Construct server address properly
	serverAddress := fmt.Sprintf(":%d", cfg.Server.Port)
	
	appLogger.Info("=== SERVER STARTUP COMPLETED ===", map[string]interface{}{
		"server_address": serverAddress,
		"environment":    cfg.Environment,
		"swagger_ui":     fmt.Sprintf("http://localhost%s/swagger/index.html", serverAddress),
	})

	StartWithErrorHandling(serverAddress, r, cfg)
}

// initializeLogging sets up the logging system
func initializeLogging() {
	// Create logger with file output
	config := log_output.LoggerConfig{
		Level:          core.InfoLevel,
		Environment:    "development",
		LogDirectory:   getEnvWithDefault("LOG_DIRECTORY", "./logs"), // Make configurable
		EnableConsole:  true,
		EnableFileJSON: true,
		EnableFileText: true,
		MaxFileSize:    100, // 100MB
		MaxFileAge:     30,  // 30 days
		EnableRotation: true,
	}

	var err error
	appLogger, err = log_output.SetupLogger(config)
	if err != nil {
		log.Printf("Warning: Failed to setup file logging: %v", err)
		// Fallback to console-only logging
		appLogger = core.NewLogger()
		appLogger.SetLevel(core.InfoLevel)
		appLogger.SetEnvironment("development")
	}

	// Configure logger context
	appLogger.SetComponent("restaurant-backend")
	appLogger.AddContextField("service_name", "restaurant-backend")
	appLogger.AddContextField("version", "1.0.0")
	appLogger.AddContextField("pid", os.Getpid())

	log.Println("Logging system initialized successfully")
}

// loggingMiddleware adds request logging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := core.GetTimestamp()
		
		// Get request ID from context if available - handle the case where utils.GetRequestID might not exist
		var requestID string
		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			requestID = reqID
		} else if reqID := r.Context().Value("request_id"); reqID != nil {
			if id, ok := reqID.(string); ok {
				requestID = id
			}
		}
		
		appLogger.InfoWithOperation(
			"Request started",
			core.LayerHandler,
			"http_request",
			map[string]interface{}{
				core.FieldRequestID: requestID,
				core.FieldMethod:    r.Method,
				core.FieldPath:      r.URL.Path,
				core.FieldClientIP:  r.RemoteAddr,
				core.FieldUserAgent: r.Header.Get("User-Agent"),
			},
		)

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)

		duration := core.GetTimestamp() - start
		
		appLogger.InfoWithOperation(
			"Request completed",
			core.LayerHandler,
			"http_request",
			map[string]interface{}{
				core.FieldRequestID:  requestID,
				core.FieldMethod:     r.Method,
				core.FieldPath:       r.URL.Path,
				core.FieldStatusCode: wrapped.statusCode,
				core.FieldDurationMS: duration,
			},
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// setupCORS configures CORS middleware for the router
func setupCORS(r *chi.Mux, cfg *utils_config.Config) {
	appLogger.Debug("Setting up CORS", map[string]interface{}{
		"allowed_origins": []string{
			cfg.ExternalAPIs.QuanAn.Address,
			"http://localhost:*",
			"http://localhost:8888",
			"http://localhost:8080",
		},
	})

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			cfg.ExternalAPIs.QuanAn.Address, 
			"http://localhost:*",
			"http://localhost:8888",
			"http://localhost:8080",
			"*", // Allow all origins for development (remove in production)
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Table-Token",
			"X-Requested-With",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	appLogger.Debug("CORS configuration completed")
}

func debugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appLogger.Debug("Debug middleware - incoming request", map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"headers": fmt.Sprintf("%v", r.Header),
		})
		next.ServeHTTP(w, r)
	})
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func Start(addr string, r *chi.Mux) error {
	appLogger.Info("Starting HTTP server", map[string]interface{}{
		"address":    addr,
		"swagger_ui": fmt.Sprintf("http://localhost%s/swagger/index.html", addr),
	})
	return http.ListenAndServe(addr, r)
}

func setupDomainHandlers(r *chi.Mux, conn *grpc.ClientConn, cfg *utils_config.Config) {
	appLogger.Info("Setting up domain handlers", map[string]interface{}{
		"operation": "domain_setup",
	})

	// Account domain with error handling
	if cfg.IsDomainEnabled("account") {
		appLogger.Info("Setting up account domain", map[string]interface{}{
			"domain": "account",
			"status": "enabled",
		})

		userClient := pb.NewAccountServiceClient(conn)
		accountHandler := account_handler.NewAccountHandler(userClient)
		account_handler.RegisterRoutesAccountHandler(r, accountHandler)

		appLogger.Info("Account domain setup completed", map[string]interface{}{
			"domain": "account",
			"status": "configured",
		})
	} else {
		appLogger.Debug("Account domain disabled", map[string]interface{}{
			"domain": "account",
			"status": "disabled",
		})
	}

	appLogger.Info("Domain handlers setup completed")
}

func StartWithErrorHandling(addr string, r *chi.Mux, cfg *utils_config.Config) {
	appLogger.Info("=== STARTING SERVER WITH ERROR HANDLING ===", map[string]interface{}{
		"address":     addr,
		"environment": cfg.Environment,
		"timeouts": map[string]interface{}{
			"read":  cfg.Server.ReadTimeout.String(),
			"write": cfg.Server.WriteTimeout.String(),
			"idle":  cfg.Server.IdleTimeout.String(),
		},
	})
	log.Printf("Starting HTTP server on %s", addr)
	log.Printf("Environment: %s", cfg.Environment)

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	appLogger.Info("HTTP server configuration completed", map[string]interface{}{
		"server_config": map[string]interface{}{
			"address":       addr,
			"read_timeout":  cfg.Server.ReadTimeout.String(),
			"write_timeout": cfg.Server.WriteTimeout.String(),
			"idle_timeout":  cfg.Server.IdleTimeout.String(),
		},
	})

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.ErrorWithCause(
			"Server failed to start",
			core.CauseNetworkError,
			core.LayerHandler,
			"server_startup",
			map[string]interface{}{
				"address": addr,
				"error":   err.Error(),
			},
		)
		// Enhanced error logging with context using fixed function name
		// errorcustom.LogCriticalError("server_startup_failed", map[string]interface{}{
		// 	"address": addr,
		// 	"error":   err.Error(),
		// })
		log.Fatalf("Server failed to start: %v", err)
	}

	appLogger.Info("Server shutdown completed gracefully")
}