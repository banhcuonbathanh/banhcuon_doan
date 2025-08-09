package main

import (
	"english-ai-full/internal/account"
	"english-ai-full/internal/branch"

	"log"
	"net"
	"os"

	accountRepo "english-ai-full/internal/account"
	"english-ai-full/internal/db"
	accountpb "english-ai-full/internal/proto_qr/account"
	branchpb "english-ai-full/internal/proto_qr/branch"
	"english-ai-full/utils"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func initializeAccountService(accountRepository *account.Repository) *account.ServiceStruct {
	// Initialize JWT Token Maker
	tokenMaker := utils.NewJWTTokenMaker("kIOopC3C7wA8DQH6FOF2Jfn+UZP8Q02nGxr/EgFMOmo=")
	
	// Initialize Password Hasher
	passwordHasher := utils.NewBcryptPasswordHasher()
	
	// Initialize Mock Email Service for now (can be replaced later)
	emailService := utils.NewMockEmailService()
	
	// Create account service with all dependencies
	accountService := account.NewAccountService(
		accountRepository,
		tokenMaker,
		passwordHasher,
		emailService,
	)
	
	return accountService
}

func main() {
	// Set up database connection using environment variables directly
	// This bypasses the complex config system for now
	dbConn, err := db.ConnectDataBase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	accountRepository := accountRepo.NewAccountRepository(dbConn)
	accountService := initializeAccountService(accountRepository)

	branchRepository := branch.NewBranchRepository(dbConn)
	branchService := branch.NewBranchService(branchRepository)

	// Use gRPC address from environment or default
	grpcAddress := getEnvOrDefault("GRPC_ADDRESS", "0.0.0.0:50051")

	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %s", lis.Addr())

	grpcServer := grpc.NewServer()
	accountpb.RegisterAccountServiceServer(grpcServer, accountService)
	branchpb.RegisterBranchServiceServer(grpcServer, branchService)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// Helper function to get environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

