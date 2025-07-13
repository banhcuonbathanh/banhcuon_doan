package main

import (
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

func main() {
	cfg, err := utils.LoadServer()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		defer dbConn.Close()
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	accountRepository := accountRepo.NewAccountRepository(dbConn)
	accountService := accountRepo.NewAccountService(accountRepository)

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
