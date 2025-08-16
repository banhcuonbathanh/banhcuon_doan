// cmd/server/main.go

package main

import (
	"context"
	_ "english-ai-full/docs"                           // Add this line at the top of imports
	"english-ai-full/internal/account/account_handler" // Add this import
	"fmt"

	"english-ai-full/internal/branch"
	error_custom "english-ai-full/internal/error_custom"
	branchpb "english-ai-full/internal/proto_qr/branch"
	"log"
	"net/http"
	"os"

	delivery "english-ai-full/internal/delivery"
	order "english-ai-full/internal/order"
	pb "english-ai-full/internal/proto_qr/account"
	ws2 "english-ai-full/internal/ws2"
	"english-ai-full/token"
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
	configPath := getEnvWithDefault("CONFIG_PATH", "utils/config/config.yaml")

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
		log.Fatalf("Configuration is nil")
	}

	envflag.Parse()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	r := chi.NewRouter()
	setupCORS(r, cfg)

	// Use environment variable with a default value
	if cfg.Environment == "development" {
		r.Use(debugMiddleware)
	}

	setupGlobalMiddleware(r, cfg)


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

	// account start
	// In your main.go or wherever you're setting up routes
	// userClient := pb.NewAccountServiceClient(conn)
	// accountHandler := account_handler.NewAccountHandler(userClient, )
	// account_handler.RegisterRoutesAccountHandler(r, accountHandler)

	// account end
	branchClient := branchpb.NewBranchServiceClient(conn)
	b := branch.NewBranchHandler(branchClient)
	branch.RegisterRoutes(r, b)

	// websocket
	//set_client := pb_set.NewSetServiceClient(conn)
	//set_hdl := set.NewSetHandler(set_client, cfg.JWT.SecretKey)
	//set.RegisterSetRoutes(r, set_hdl)

	// dish
	//dish_client := pb_dish.NewDishServiceClient(conn)
	//dish_hdl := dish.NewDishHandler(dish_client, cfg.JWT.SecretKey)
	//dish.RegisterDishRoutes(r, dish_hdl)

	// table
	//table_client := pb_tables.NewTableServiceClient(conn)
	//table_hdl := tables.NewTableHandler(table_client)
	//tables.RegisterTablesRoutes(r, table_hdl)

	// guest
	//guests_client := pb_guests.NewGuestServiceClient(conn)
	//guests_hdl := guests.NewGuestHandler(guests_client, cfg.JWT.SecretKey)
	//guests.RegisterGuestRoutes(r, guests_hdl)

	// order
	//order_client := pb_order.NewOrderServiceClient(conn)
	//order_hdl := order.NewOrderHandler(order_client, cfg.JWT.SecretKey)
	//order.RegisterOrderRoutes(r, order_hdl)

	// delivery
	//delivery_client := pb_delivery.NewDeliveryServiceClient(conn)
	//delivery_hdl := delivery.NewDeliveryHandler(delivery_client, cfg.JWT.SecretKey)
	//delivery.RegisterDeliveryRoutes(r, delivery_hdl)

	//SetupWs2(r, order_hdl, delivery_hdl, cfg)
	//
	//r.Get("/image", func(w http.ResponseWriter, r *http.Request) {
	//
	//	file, err := os.Open("upload/quananqr/public/pexels-ella-olsson-572949-1640777.jpg")
	//	if err != nil {
	//		http.Error(w, "Image not found.", http.StatusNotFound)
	//		return
	//	}
	//	defer file.Close()
	//
	//	img, _, err := image.Decode(file)
	//	if err != nil {
	//		http.Error(w, "Error decoding image.", http.StatusInternalServerError)
	//		return
	//	}
	//
	//	w.Header().Set("Content-Type", "image/jpeg")
	//	jpeg.Encode(w, img, nil)
	//})
	//
	//hdl_image := image_upload.NewImageHandler(cfg.JWT.SecretKey)
	//
	//image_upload.RegisterImageRoutes(r, hdl_image)

	// Construct server address properly
	serverAddress := fmt.Sprintf(":%d", cfg.Server.Port)
	Start(serverAddress, r)
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

func SetupWs2(r chi.Router, orderHandler *order.OrderHandlerController, deliveryHandler *delivery.DeliveryHandlerController, cfg *utils_config.Config) {
	log.Println("golang/cmd/server/main.go")

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
	// wsRouter := ws2.NewWebSocketRouter(hub)

	wsRouter := ws2.NewWebSocketRouter(hub, tokenMaker)
	wsRouter.RegisterRoutes(r)

	go hub.Run()
}

func Start(addr string, r *chi.Mux) error {
	log.Printf("Starting HTTP server on %s", addr)
	log.Printf("Swagger UI available at: http://localhost%s/swagger/index.html", addr)
	return http.ListenAndServe(addr, r)
}

// new 12121212
func setupGlobalMiddleware(r *chi.Mux, cfg *utils_config.Config) {
    // Core error handling middleware
    r.Use(error_custom.RequestIDMiddleware)
    r.Use(error_custom.LogHTTPMiddleware)
    r.Use(error_custom.RecoveryMiddleware)

    // Environment-specific middleware
    if cfg.Environment == "development" {
        r.Use(error_custom.DebugMiddleware)
    }

    // JWT validation middleware for protected routes
    if cfg.JWT.SecretKey != "" {
        r.Use(error_custom.JWTValidationMiddleware(cfg.JWT.SecretKey))
    }

    // Domain context middleware
    r.Use(error_custom.DomainContextMiddleware)
}

func setupDomainHandlers(r *chi.Mux, conn *grpc.ClientConn, cfg *utils_config.Config) {
    // Account domain with error handling
    if cfg.IsDomainEnabled("account") {
        userClient := pb.NewAccountServiceClient(conn)
        accountHandler := account_handler.NewAccountHandler(userClient, cfg)
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
        // Enhanced error logging with context
        error_custom.LogCriticalError("server_startup_failed", map[string]interface{}{
            "address": addr,
            "error":   err.Error(),
        })
        log.Fatalf("Server failed to start: %v", err)
    }
}
// new 12121212