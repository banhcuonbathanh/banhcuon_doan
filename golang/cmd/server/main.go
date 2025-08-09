package main

import (
	"english-ai-full/internal/account"
	"english-ai-full/internal/branch"

	"log"
	"net"


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
	// Load basic configuration from environment variables
	cfg, err := utils.LoadServer()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
	}

	dbConn, err := db.ConnectDataBase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	accountRepository := accountRepo.NewAccountRepository(dbConn)
	accountService := initializeAccountService(accountRepository)

	branchRepository := branch.NewBranchRepository(dbConn)
	branchService := branch.NewBranchService(branchRepository)

	// Use gRPC address from config or environment
	grpcAddress := "0.0.0.0:50051" // Default
	if cfg != nil {
		grpcAddress = cfg.GRPCAddress
	}

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