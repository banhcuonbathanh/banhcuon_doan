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

// new 121212121212

func initializeAccountService(accountRepository *account.Repository) *account.ServiceStruct {
	// Initialize JWT Token Maker
	tokenMaker := utils.NewJWTTokenMaker("your-super-secret-jwt-key-here")
	
	// Initialize Password Hasher
	passwordHasher := utils.NewBcryptPasswordHasher()
	
	// Initialize Email Service
	// Option 1: Real SMTP Email Service
	emailConfig := utils.EmailConfig{
		Host:     "smtp.gmail.com",  // or your SMTP server
		Port:     "587",
		Username: "your-email@gmail.com",
		Password: "your-app-password",
		From:     "your-email@gmail.com",
	}
	emailService := utils.NewSMTPEmailService(emailConfig)
	
	// Option 2: Mock Email Service (for development/testing)
	// emailService := email.NewMockEmailService()
	
	// Create account service with all dependencies
	accountService := account.NewAccountService(
		accountRepository,
		tokenMaker,
		passwordHasher,
		emailService,
	)
	
	return accountService
}

// Alternative: Gradual implementation
func initializeAccountServiceGradual(accountRepository *account.Repository) *account.ServiceStruct {
	// Start with just token maker and password hasher
	tokenMaker := utils.NewJWTTokenMaker("your-super-secret-jwt-key-here")
	passwordHasher := utils.NewBcryptPasswordHasher()
	
	// Use mock email service for now
	emailService := utils.NewMockEmailService()
	
	accountService := account.NewAccountService(
		accountRepository,
		tokenMaker,
		passwordHasher,
		emailService, // or nil if you don't want any email functionality
	)
	
	return accountService
}

// new 1212121212
func main() {
	cfg, err := utils.LoadServer()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConn, err := db.ConnectDataBase(cfg.DatabaseURL)
	if err != nil {
		defer dbConn.Close()
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
// // new 121212121
// tokenMaker := utils.NewJWTTokenMaker(secretKey) // secretKey should be from config

// // 2. Create the password hasher
// passwordHasher := utils.NewBcryptPasswordHasher() // or whatever implementation you have

// // 3. Create the email service
// emailService := email.NewEmailService(emailConfig) 
// // new 12121212
	accountRepository := accountRepo.NewAccountRepository(dbConn)
	accountService := initializeAccountService(accountRepository)

	branchRepository := branch.NewBranchRepository(dbConn)
	branchService := branch.NewBranchService(branchRepository)

	//dishrepo := dish.NewDishRepository(dbConn)
	//dishService := dish.NewDishService(dishrepo)

	//guestsrepo := guests.NewGuestRepository(dbConn)
	//guestsService := guests.NewGuestService(guestsrepo)
	//
	//orderrepo := order.NewOrderRepository(dbConn)
	//orderService := order.NewOrderService(orderrepo)
	//tablerepo := tables.NewTableRepository(dbConn, "asdfEWQR1234%#$@")
	//tableService := tables.NewTableService(tablerepo)
	//
	//setrepo := set.NewSetRepository(dbConn)
	//setService := set.NewSetService(setrepo)
	//
	//deliveryrepo := delivery.NewDeliveryRepository(dbConn)
	//deliveryService := delivery.NewDeliveryService(deliveryrepo)

	lis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %s", lis.Addr())

	grpcServer := grpc.NewServer()
	accountpb.RegisterAccountServiceServer(grpcServer, accountService)
	branchpb.RegisterBranchServiceServer(grpcServer, branchService)
	//pb_delivery.RegisterDeliveryServiceServer(grpcServer, deliveryService)
	//pb_set.RegisterSetServiceServer(grpcServer, setService)
	//pb_table.RegisterTableServiceServer(grpcServer, tableService)
	//dishPb.RegisterDishServiceServer(grpcServer, dishService)
	//pb_guests.RegisterGuestServiceServer(grpcServer, guestsService)
	//pb_order.RegisterOrderServiceServer(grpcServer, orderService)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}

}
