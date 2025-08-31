// cmd/server/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "english-ai-full/docs"
	"english-ai-full/internal/account/account_handler"
	"english-ai-full/internal/branch"
	delivery "english-ai-full/internal/delivery"
	error_custom "english-ai-full/internal/error_custom"
	order "english-ai-full/internal/order"
	pb "english-ai-full/internal/proto_qr/account"
	branchpb "english-ai-full/internal/proto_qr/branch"
	ws2 "english-ai-full/internal/ws2"
	"english-ai-full/logger"
	"english-ai-full/token"

		"english-ai-full/utils/config"

	// Import the integration package
	"english-ai-full/integration"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/ianschenck/envflag"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Step 1: Initialize all utilities using the integration package
	configPath := getEnvWithDefault("CONFIG_PATH", "utils/config/config.yaml")
	
	if err := integration.InitializeUtilities(configPath); err != nil {
		log.Fatalf("Failed to initialize utilities: %v", err)
	}

	// Step 2: Get utility manager and components
	utils := integration.GetUtilityManager()
	cfg := utils.Config
	logger := utils.Logger
	errorHandler := utils.ErrorHandler

	// Log application startup
	logger.LogBusinessEvent("system", "startup", "main", "starting", map[string]interface{}{
		"config_path": configPath,
		"environment": cfg.Environment,
		"version":     cfg.Version,
	})

	// Step 3: Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.LogBusinessEvent("system", "shutdown", "main", "signal_received", nil)
		cancel()
	}()

	envflag.Parse()

	// Step 4: Setup router with integrated middleware
	r := chi.NewRouter()
	setupCORSWithIntegration(r, cfg, logger)
	setupGlobalMiddlewareWithIntegration(r, cfg, errorHandler)

	// Step 5: Setup gRPC connections with error handling
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Python server connection
	pythonConn, err := grpc.NewClient(":50052", opts...)
	if err != nil {
		logger.LogSecurityEvent("grpc_connection_failed", "critical", "python", "", map[string]interface{}{
			"error":   err.Error(),
			"address": ":50052",
		})
		log.Fatalf("failed to connect to Python gRPC server: %v", err)
	}
	defer pythonConn.Close()

	// Main gRPC server connection
	grpcAddress := fmt.Sprintf("%s:%d", cfg.Server.GRPCAddress, cfg.Server.GRPCPort)
	conn, err := grpc.DialContext(ctx, grpcAddress, opts...)
	if err != nil {
		logger.LogSecurityEvent("grpc_connection_failed", "critical", "main", "", map[string]interface{}{
			"error":   err.Error(),
			"address": grpcAddress,
		})
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	logger.LogBusinessEvent("system", "grpc", "connection", "established", map[string]interface{}{
		"state":   conn.GetState().String(),
		"address": grpcAddress,
	})

	// Step 6: Setup domain handlers using integration utilities
	setupDomainHandlersWithIntegration(r, conn, cfg, utils)

	// Step 7: Setup WebSocket if needed
	// if cfg.IsDomainEnabled("websocket") {
	// 	orderClient := pb.NewOrderServiceClient(conn) // This should be the correct import
	// 	orderHandler := order.NewOrderHandler(orderClient, cfg.JWT.SecretKey)
		
	// 	deliveryClient := pb.NewDeliveryServiceClient(conn) // This should be the correct import  
	// 	deliveryHandler := delivery.NewDeliveryHandler(deliveryClient, cfg.JWT.SecretKey)
		
	// 	SetupWs2WithIntegration(r, orderHandler, deliveryHandler, cfg, logger)
	// }

	// Step 8: Start server with graceful shutdown
	startServerWithIntegration(ctx, r, cfg, logger, utils)
}

// setupCORSWithIntegration configures CORS middleware using integrated logging
// setupCORSWithIntegration configures CORS middleware using integrated logging
func setupCORSWithIntegration(r *chi.Mux, cfg *utils_config.Config, logger *logger.SpecializedLogger) {
	logger.LogBusinessEvent("system", "middleware", "cors", "configuring", map[string]interface{}{
		"allowed_origins": cfg.ExternalAPIs.QuanAn.Address,
		"environment":     cfg.Environment,
	})

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			cfg.ExternalAPIs.QuanAn.Address,
			"http://localhost:*",
			"http://localhost:8888",
			"http://localhost:8080",
			"*", // Remove in production
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type", "X-CSRF-Token",
			"X-Table-Token", "X-Requested-With",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

// setupGlobalMiddlewareWithIntegration sets up middleware using the integration package
func setupGlobalMiddlewareWithIntegration(r *chi.Mux, cfg *utils_config.Config, errorHandler *error_custom.UnifiedErrorHandler) {
	// Core error handling middleware from integration
	r.Use(error_custom.RequestIDMiddleware)
	r.Use(error_custom.LogHTTPMiddleware)
	r.Use(error_custom.RecoveryMiddleware)

	// Environment-specific middleware
	if cfg.Environment == "development" {
		// r.Use(error_custom.DebugMiddleware)
		r.Use(debugMiddlewareWithIntegration)
	}

	// JWT validation middleware for protected routes
	if cfg.JWT.SecretKey != "" {
		r.Use(error_custom.JWTValidationMiddleware(cfg.JWT.SecretKey))
	}

	// Domain context middleware
	r.Use(error_custom.DomainContextMiddleware)
}

// setupDomainHandlersWithIntegration sets up domain handlers using integration utilities
func setupDomainHandlersWithIntegration(r *chi.Mux, conn *grpc.ClientConn, cfg *utils_config.Config, utils *integration.UtilityManager) {
	// Account domain
	if cfg.IsDomainEnabled("account") {
		accountUtils := integration.NewDomainUtilities("account")
		if err := accountUtils.ValidateConfig(); err != nil {
			accountUtils.Logger.LogSecurityEvent("domain_config_invalid", "error", "account", "", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			userClient := pb.NewAccountServiceClient(conn)
			accountHandler := account_handler.NewAccountHandler(userClient, cfg)
			account_handler.RegisterRoutesAccountHandler(r, accountHandler)
			
			accountUtils.LogOperation("handler_registration", "success", true, nil, map[string]interface{}{
				"routes_registered": true,
			})
		}
	}

	// Branch domain
	if cfg.IsDomainEnabled("branch") {
		branchUtils := integration.NewDomainUtilities("branch")
		if err := branchUtils.ValidateConfig(); err != nil {
			branchUtils.Logger.LogSecurityEvent("domain_config_invalid", "error", "branch", "", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			branchClient := branchpb.NewBranchServiceClient(conn)
			branchHandler := branch.NewBranchHandler(branchClient)
			branch.RegisterRoutes(r, branchHandler)
			
			branchUtils.LogOperation("handler_registration", "success", true, nil, map[string]interface{}{
				"routes_registered": true,
			})
		}
	}

	// Add other domain handlers as needed...
}

// SetupWs2WithIntegration sets up WebSocket with integration logging
func SetupWs2WithIntegration(r chi.Router, orderHandler *order.OrderHandlerController, deliveryHandler *delivery.DeliveryHandlerController, cfg *utils_config.Config,logger *logger.SpecializedLogger) {
	wsUtils := integration.NewDomainUtilities("websocket")
	wsUtils.Logger.LogBusinessEvent("websocket", "setup", "initialization", "starting", nil)

	// Initialize the JWT token maker
	tokenMaker := token.NewJWTMaker(cfg.JWT.SecretKey)

	// Create message handlers
	orderMsgHandler := ws2.NewOrderMessageHandler(orderHandler)
	deliveryMsgHandler := ws2.NewDeliveryMessageHandler(deliveryHandler)

	// Create a combined message handler
	combinedHandler := ws2.NewCombinedMessageHandler(orderMsgHandler, deliveryMsgHandler)

	// Create and setup the hub
	hub := ws2.NewHub(combinedHandler)
	broadcaster := ws2.NewBroadcaster(hub)

	// Set broadcasters
	orderMsgHandler.SetBroadcaster(broadcaster)
	deliveryMsgHandler.SetBroadcaster(broadcaster)

	// Setup router with token maker
	wsRouter := ws2.NewWebSocketRouter(hub, tokenMaker)
	wsRouter.RegisterRoutes(r)

	go hub.Run()

	wsUtils.LogOperation("websocket_setup", "completed", true, nil, map[string]interface{}{
		"hub_started":        true,
		"routes_registered": true,
	})
}

// startServerWithIntegration starts the server with integrated error handling and graceful shutdown
func startServerWithIntegration(ctx context.Context, r *chi.Mux, cfg *utils_config.Config, logger *logger.SpecializedLogger, utils *integration.UtilityManager) {
	serverAddress := fmt.Sprintf(":%d", cfg.Server.Port)
	
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	logger.LogBusinessEvent("system", "server", "startup", "starting", map[string]interface{}{
		"address":       serverAddress,
		"environment":   cfg.Environment,
		"swagger_url":   fmt.Sprintf("http://localhost%s/swagger/index.html", serverAddress),
		"read_timeout":  cfg.Server.ReadTimeout,
		"write_timeout": cfg.Server.WriteTimeout,
		"idle_timeout":  cfg.Server.IdleTimeout,
	})

	// Start server in goroutine for graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogSecurityEvent("server_startup_failed", "critical", "main", "", map[string]interface{}{
				"address": serverAddress,
				"error":   err.Error(),
			})
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Starting HTTP server on %s", serverAddress)
	log.Printf("Swagger UI available at: http://localhost%s/swagger/index.html", serverAddress)

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	logger.LogBusinessEvent("system", "server", "shutdown", "starting", nil)
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.LogSecurityEvent("server_shutdown_failed", "error", "main", "", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.LogBusinessEvent("system", "server", "shutdown", "completed", nil)
	}

	// Shutdown utility manager
	if err := utils.Shutdown(shutdownCtx); err != nil {
		logger.LogSecurityEvent("utility_shutdown_failed", "error", "main", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// debugMiddlewareWithIntegration provides debug middleware with integration logging
func debugMiddlewareWithIntegration(next http.Handler) http.Handler {
	debugUtils := integration.NewHandlerUtilities("debug")
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		debugUtils.HandlerLogger.LogBusinessEvent("debug", "request", "incoming", "received", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		})
		next.ServeHTTP(w, r)
	})
}

// getEnvWithDefault returns environment variable or default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Health check endpoint setup (optional)
func setupHealthCheck(r *chi.Mux, utils *integration.UtilityManager) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		health := utils.HealthCheck(r.Context())
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// You'll need to marshal this to JSON
		// json.NewEncoder(w).Encode(health)
		fmt.Fprintf(w, "%+v", health)
	})

	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := utils.GetMetrics()
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// You'll need to marshal this to JSON
		// json.NewEncoder(w).Encode(metrics)
		fmt.Fprintf(w, "%+v", metrics)
	})
}