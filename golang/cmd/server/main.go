package main

import (

		"english-ai-full/internal/account/account_handler" // Add this import
	"context"
	"english-ai-full/internal/account"
	"english-ai-full/internal/branch"
	branchpb "english-ai-full/internal/proto_qr/branch"
	"log"
	"net/http"
	"os"

	delivery "english-ai-full/internal/delivery"
	order "english-ai-full/internal/order"
	pb "english-ai-full/internal/proto_qr/account"
	ws2 "english-ai-full/internal/ws2"
	"english-ai-full/token"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/ianschenck/envflag"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := utils.LoadServer()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	envflag.Parse()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{cfg.QuanAnAddress, "http://localhost:*"},
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

	// Use environment variable with a default value
	if getEnvWithDefault("GO_ENV", "development") == "development" {
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

	conn, err := grpc.DialContext(
		context.Background(),
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()
	log.Println("Connection State to GRPC Server: ", conn.GetState())
	log.Println("Calling to GRPC Server: ", cfg.GRPCAddress)
// account start


	accountClient := pb.NewAccountServiceClient(conn)

		accountHandler := account_handler.NewAccountHandler(accountClient)
	account.RegisterRoutes(r, accountHandler)
	// h := account.New(accountClient)
	// account.RegisterRoutes(r, h)
// account end
	branchClient := branchpb.NewBranchServiceClient(conn)
	b := branch.NewBranchHandler(branchClient)
	branch.RegisterRoutes(r, b)

	// websocket
	//set_client := pb_set.NewSetServiceClient(conn)
	//set_hdl := set.NewSetHandler(set_client, cfg.JwtSecret)
	//set.RegisterSetRoutes(r, set_hdl)

	// dish
	//dish_client := pb_dish.NewDishServiceClient(conn)
	//dish_hdl := dish.NewDishHandler(dish_client, cfg.JwtSecret)
	//dish.RegisterDishRoutes(r, dish_hdl)

	// table
	//table_client := pb_tables.NewTableServiceClient(conn)
	//table_hdl := tables.NewTableHandler(table_client)
	//tables.RegisterTablesRoutes(r, table_hdl)

	// guest
	//guests_client := pb_guests.NewGuestServiceClient(conn)
	//guests_hdl := guests.NewGuestHandler(guests_client, cfg.JwtSecret)
	//guests.RegisterGuestRoutes(r, guests_hdl)

	// order
	//order_client := pb_order.NewOrderServiceClient(conn)
	//order_hdl := order.NewOrderHandler(order_client, cfg.JwtSecret)
	//order.RegisterOrderRoutes(r, order_hdl)

	// delivery
	//delivery_client := pb_delivery.NewDeliveryServiceClient(conn)
	//delivery_hdl := delivery.NewDeliveryHandler(delivery_client, cfg.JwtSecret)
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
	//hdl_image := image_upload.NewImageHandler(cfg.JwtSecret)
	//
	//image_upload.RegisterImageRoutes(r, hdl_image)

	Start(cfg.ServerAddress, r)

}

func setupGlobalMiddleware(r *chi.Mux, cfg *utils.Config) {
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers for every response
			w.Header().Set("Access-Control-Allow-Origin", cfg.QuanAnAddress)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Table-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
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

func SetupWs2(r chi.Router, orderHandler *order.OrderHandlerController, deliveryHandler *delivery.DeliveryHandlerController, cfg *utils.Config) {
	log.Println("golang/cmd/server/main.go")

	// Initialize the JWT token maker

	tokenMaker := token.NewJWTMaker(cfg.JwtSecret)

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
	return http.ListenAndServe(addr, r)
}
