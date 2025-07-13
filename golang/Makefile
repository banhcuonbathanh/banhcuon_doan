# Variables
SERVER_CMD = cmd/grpc-server/main.go
TEST_USER_DIR = cmd/client
PID_FILE = .server.pid

# Phony targets
.PHONY: all run-server stop-server run-client test test-create test-get clean

# Default target
all: run-server test

# Run the server
run-server:
	@echo "Starting the gRPC server..."
	@go run $(SERVER_CMD) & echo $$! > $(PID_FILE)
# Stop the server
stop-server:
	@echo "Stopping the gRPC server..."
	@-kill `cat $(PID_FILE)` 2>/dev/null || true
	@rm -f $(PID_FILE)
# Run all tests
test:
	@echo "Running all tests..."
	@cd $(TEST_USER_DIR) && go test -v ./...

# Run specific test for CreateUser
test-create:
	@echo "Running CreateUser test..."
	@cd $(TEST_USER_DIR) && go test -v -run TestCreateUser

# Run specific test for GetUser
test-get:
	@echo "Running GetUser test..."
	@cd $(TEST_USER_DIR) && go test -v -run TestGetUser

# Clean build artifacts (if any)
clean:
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -f $(TEST_USER_DIR)/*.test
	@rm -f $(PID_FILE)

# Help target
help:
	@echo "Available targets:"
	@echo "  run-server  : Start the gRPC server"
	@echo "  stop-server : Stop the gRPC server"
	@echo "  run-client  : Run the gRPC client"
	@echo "  test        : Run all tests"
	@echo "  test-create : Run CreateUser test"
	@echo "  test-get    : Run GetUser test"
	@echo "  clean       : Clean build artifacts"
	@echo "  help        : Show this help message"

# SQLBoiler variables
DB_URL=postgres://restaurant:restaurant@localhost:5432/restaurant?sslmode=disable
MIGRATIONS_DIR=internal/db/migrations/
SQLBOILER_CONFIG=./sqlboiler.toml

.PHONY: migrate-up migrate-down migrate-force migrate-status migrate-create sqlboiler models

# Migration commands
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force

migrate-status:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

# SQLBoiler generate orm
sqlboiler:
	sqlboiler psql --config=$(SQLBOILER_CONFIG)

gen-models: sqlboiler

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           internal/proto_qr/account/account.proto \
           internal/proto_qr/branch/branch.proto

# Rebuild and restart go_app container
rebuild-app:
	docker compose down go_app
	docker compose build --no-cache go_app
	docker compose up -d go_app

# Rebuild and restart all services
rebuild-all:
	docker compose down
	docker compose build --no-cache
	docker compose up -d

# Quick restart (without rebuild)
restart-app:
	docker compose restart go_app