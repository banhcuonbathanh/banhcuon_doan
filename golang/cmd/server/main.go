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
	"english-ai-full/token"

	branchpb "english-ai-full/internal/proto_qr/branch"

	pb "english-ai-full/internal/proto_qr/account"

	"english-ai-full/utils"
	utils_config "english-ai-full/utils/config"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/ianschenck/envflag"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize configuration using the new system
	// configPath := getEnvWithDefault("CONFIG_PATH", "utils/config/config.yaml")
	LogerSetup()
configPath := "utils/config/config.yaml" 
	err := utils_config.InitializeConfig(configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config file: %v", err)
		log.Println("Continuing with environment variables and defaults...")
		
		// Initialize with empty path to use defaults and environment variables
		err = utils_config.InitializeConfig("")
		if err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
	}

	// Get the configuration
	cfg := utils_config.GetConfig()
	if cfg == nil {
		log.Fatalf("Configuration is nil after initialization")
	}
	// loger

	//
	// Verify JWT config specifically
	if cfg.JWT.SecretKey == "" {
		log.Fatalf("JWT secret key is not configured")
	}
	
	log.Printf("Configuration loaded successfully with JWT key length: %d", len(cfg.JWT.SecretKey))

	// 3. Initialize token maker ONLY ONCE
	log.Println("Initializing token maker...")
	if err := token.InitializeTokenMaker(); err != nil {
		log.Fatalf("Failed to initialize token maker: %v", err)
	}
	log.Println("Token maker initialized successfully")

    log.Printf("Token maker initialized successfully")
	envflag.Parse()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	r := chi.NewRouter()
	  r.Use(utils.RequestIDMiddleware)
	setupCORS(r, cfg)

	// Use environment variable with a default value
	if cfg.Environment == "development" {
		r.Use(debugMiddleware)
	}

	// Setup ONLY basic global middleware (no JWT here)
	// setupGlobalMiddleware(r, cfg)

	/**
	python server
	*/
	pythonConn, err := grpc.NewClient(":50052", opts...)
	if err != nil {
		log.Fatalf("failed to connect to Python gRPC server: %v", err)
	}
	defer pythonConn.Close()

	// Construct gRPC address properly
	grpcAddress := fmt.Sprintf("%s:%d", cfg.Server.GRPCAddress, cfg.Server.GRPCPort)
	
	conn, err := grpc.DialContext(
		context.Background(),
		grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()
	log.Println("Connection State to GRPC Server: ", conn.GetState())
	log.Println("Calling to GRPC Server: ", grpcAddress)

	branchClient := branchpb.NewBranchServiceClient(conn)
	b := branch.NewBranchHandler(branchClient)
	branch.RegisterRoutes(r, b)

	// Setup domain handlers with error handling
	setupDomainHandlers(r, conn, cfg)

	// Construct server address properly
	serverAddress := fmt.Sprintf(":%d", cfg.Server.Port)
	StartWithErrorHandling(serverAddress, r, cfg)
}

// setupCORS configures CORS middleware for the router
func setupCORS(r *chi.Mux, cfg *utils_config.Config) {
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
}

func debugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		// log.Printf("Headers: %v", r.Header)
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

// func SetupWs2(r chi.Router, orderHandler *order.OrderHandlerController, deliveryHandler *delivery.DeliveryHandlerController, cfg *utils_config.Config) {
// 	log.Println("golang/cmd/server/main.go")

// 	// Initialize the JWT token maker
// 	tokenMaker := token.NewJWTMaker(cfg.JWT.SecretKey)

// 	// Create message handlers
// 	orderMsgHandler := ws2.NewOrderMessageHandler(orderHandler)
// 	deliveryMsgHandler := ws2.NewDeliveryMessageHandler(deliveryHandler)

// 	// Create a combined message handler
// 	combinedHandler := ws2.NewCombinedMessageHandler(orderMsgHandler, deliveryMsgHandler)

// 	// Create and setup the hub
// 	hub := ws2.NewHub(combinedHandler)
// 	broadcaster := ws2.NewBroadcaster(hub)

// 	// Set broadcasters
// 	orderMsgHandler.SetBroadcaster(broadcaster)
// 	deliveryMsgHandler.SetBroadcaster(broadcaster)

// 	// Setup router with token maker
// 	wsRouter := ws2.NewWebSocketRouter(hub, tokenMaker)
// 	wsRouter.RegisterRoutes(r)

// 	go hub.Run()
// }

func Start(addr string, r *chi.Mux) error {
	log.Printf("Starting HTTP server on %s", addr)
	log.Printf("Swagger UI available at: http://localhost%s/swagger/index.html", addr)
	return http.ListenAndServe(addr, r)
}

// setupGlobalMiddleware sets up ONLY basic global middleware (NO JWT here)
// func setupGlobalMiddleware(r *chi.Mux, cfg *utils_config.Config) {
// 	// Core error handling middleware
// 	r.Use(errorcustom.RequestIDMiddleware)
// 	r.Use(errorcustom.LogHTTPMiddleware)
// 	r.Use(errorcustom.RecoveryMiddleware)

// 	// Environment-specific middleware
// 	if cfg.Environment == "development" {
// 		r.Use(errorcustom.DebugMiddleware) // Fixed function name
// 	}

// 	// ❌ REMOVED: JWT middleware is NO LONGER applied globally
// 	// if cfg.JWT.SecretKey != "" {
// 	//     r.Use(errorcustom.JWTValidationMiddleware(cfg.JWT.SecretKey))
// 	// }

// 	// Domain context middleware
// 	// r.Use(errorcustom.DomainContextMiddleware)
// }

func setupDomainHandlers(r *chi.Mux, conn *grpc.ClientConn, cfg *utils_config.Config) {
	// Account domain with error handling
	if cfg.IsDomainEnabled("account") {
		userClient := pb.NewAccountServiceClient(conn)
		accountHandler := account_handler.NewAccountHandler(userClient)
		account_handler.RegisterRoutesAccountHandler(r, accountHandler)
	}

	// Branch domain with error handling
	// if cfg.IsDomainEnabled("branch") {
	//     branchClient := branchpb.NewBranchServiceClient(conn)
	//     branchHandler := branch.NewBranchHandlerWithErrorHandling(branchClient, cfg)
	//     setupBranchRoutes(r, branchHandler)
	// }

	// Additional domain handlers...
}

func StartWithErrorHandling(addr string, r *chi.Mux, cfg *utils_config.Config) {
	log.Printf("Starting HTTP server on %s", addr)
	log.Printf("Environment: %s", cfg.Environment)

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// Enhanced error logging with context using fixed function name
		// errorcustom.LogCriticalError("server_startup_failed", map[string]interface{}{
		// 	"address": addr,
		// 	"error":   err.Error(),
		// })
		log.Fatalf("Server failed to start: %v", err)
	}
}

func LogerSetup() {

logger := core.NewLogger()
	logger.EnableOnlyLayers( core.LayerToken)

		logger.DisableLayers(core.LayerDatabase, core.LayerExternal, core.LayerConfig, core.LayerHandler, core.LayerService, core.LayerRepository,)
}