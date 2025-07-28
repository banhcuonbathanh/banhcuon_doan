# # Variables
# SERVER_CMD = cmd/grpc-server/main.go
# TEST_USER_DIR = cmd/client
# PID_FILE = .server.pid

# # Phony targets
# .PHONY: all run-server stop-server run-client test test-create test-get clean

# # Default target
# all: run-server test

# # Run the server
# run-server:
# 	@echo "Starting the gRPC server..."
# 	@go run $(SERVER_CMD) & echo $$! > $(PID_FILE)
# # Stop the server
# stop-server:
# 	@echo "Stopping the gRPC server..."
# 	@-kill `cat $(PID_FILE)` 2>/dev/null || true
# 	@rm -f $(PID_FILE)
# # Run all tests
# test:
# 	@echo "Running all tests..."
# 	@cd $(TEST_USER_DIR) && go test -v ./...

# # Run specific test for CreateUser
# test-create:
# 	@echo "Running CreateUser test..."
# 	@cd $(TEST_USER_DIR) && go test -v -run TestCreateUser

# # Run specific test for GetUser
# test-get:
# 	@echo "Running GetUser test..."
# 	@cd $(TEST_USER_DIR) && go test -v -run TestGetUser

# # Clean build artifacts (if any)
# clean:
# 	@echo "Cleaning build artifacts..."
# 	@go clean
# 	@rm -f $(TEST_USER_DIR)/*.test
# 	@rm -f $(PID_FILE)

# # Help target
# help:
# 	@echo "Available targets:"
# 	@echo "  run-server  : Start the gRPC server"
# 	@echo "  stop-server : Stop the gRPC server"
# 	@echo "  run-client  : Run the gRPC client"
# 	@echo "  test        : Run all tests"
# 	@echo "  test-create : Run CreateUser test"
# 	@echo "  test-get    : Run GetUser test"
# 	@echo "  clean       : Clean build artifacts"
# 	@echo "  help        : Show this help message"

# # SQLBoiler variables
# DB_URL=postgres://restaurant:restaurant@localhost:5432/restaurant?sslmode=disable
# MIGRATIONS_DIR=internal/db/migrations/
# SQLBOILER_CONFIG=./sqlboiler.toml

# .PHONY: migrate-up migrate-down migrate-force migrate-status migrate-create sqlboiler models

# # Migration commands
# migrate-up:
# 	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

# migrate-down:
# 	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

# migrate-force:
# 	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force

# migrate-status:
# 	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# migrate-create:
# 	@read -p "Migration name: " name; \
# 	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

# # SQLBoiler generate orm
# sqlboiler:
# 	sqlboiler psql --config=$(SQLBOILER_CONFIG)

# gen-models: sqlboiler

# proto:
# 	protoc --go_out=. --go_opt=paths=source_relative \
#            --go-grpc_out=. --go-grpc_opt=paths=source_relative \
#            internal/proto_qr/account/account.proto \
#            internal/proto_qr/branch/branch.proto

# # Rebuild and restart go_app container
# rebuild-app:
# 	docker compose down go_app
# 	docker compose build --no-cache go_app
# 	docker compose up -d go_app

# # Rebuild and restart all services
# rebuild-all:
# 	docker compose down
# 	docker compose build --no-cache
# 	docker compose up -d

# # Quick restart (without rebuild)
# restart-app:
# 	docker compose restart go_app




# Makefile for Restaurant API Project

# Variables
SERVER_CMD = cmd/grpc-server/main.go
TEST_USER_DIR = cmd/client
PID_FILE = .server.pid

# Database variables
DB_URL=postgres://restaurant:restaurant@localhost:5432/restaurant?sslmode=disable
MIGRATIONS_DIR=internal/db/migrations/
SQLBOILER_CONFIG=./sqlboiler.toml

# Documentation variables
DOCS_OUTPUT_DIR = docs/generated

.PHONY: help all run-server stop-server test test-create test-get clean
.PHONY: migrate-up migrate-down migrate-force migrate-status migrate-create sqlboiler gen-models
.PHONY: proto rebuild-app rebuild-all restart-app
.PHONY: docs docs-build docs-clean docs-serve docs-validate docs-watch docs-info dev-setup

# Default target
all: run-server test

# Help target - show all available commands
help: ## Show this help message
	@echo "🚀 Restaurant API Project - Available Commands"
	@echo ""
	@echo "📋 Main Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(all|help|run-server|stop-server|test)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🗄️  Database Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(migrate|sqlboiler|gen-models)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "📚 Documentation Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(docs)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🐳 Docker Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(rebuild|restart)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🔧 Development Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(proto|dev-setup|clean)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# SERVER COMMANDS
# =============================================================================

run-server: ## Start the gRPC server
	@echo "🚀 Starting the gRPC server..."
	@go run $(SERVER_CMD) & echo $$! > $(PID_FILE)

stop-server: ## Stop the gRPC server
	@echo "🛑 Stopping the gRPC server..."
	@-kill `cat $(PID_FILE)` 2>/dev/null || true
	@rm -f $(PID_FILE)

restart-server: stop-server run-server ## Restart the gRPC server

# =============================================================================
# TESTING COMMANDS
# =============================================================================

test: ## Run all tests
	@echo "🧪 Running all tests..."
	@cd $(TEST_USER_DIR) && go test -v ./...

test-create: ## Run CreateUser test
	@echo "🧪 Running CreateUser test..."
	@cd $(TEST_USER_DIR) && go test -v -run TestCreateUser

test-get: ## Run GetUser test
	@echo "🧪 Running GetUser test..."
	@cd $(TEST_USER_DIR) && go test -v -run TestGetUser

test-with-docs: docs-build test ## Build docs and run tests

# =============================================================================
# DATABASE COMMANDS
# =============================================================================

migrate-up: ## Run database migrations up
	@echo "📊 Running database migrations up..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down: ## Run database migrations down (1 step)
	@echo "📊 Running database migrations down..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-force: ## Force database migration version
	@echo "📊 Forcing database migration..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force

migrate-status: ## Check database migration status
	@echo "📊 Checking database migration status..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-create: ## Create new database migration
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

sqlboiler: ## Generate ORM models using SQLBoiler
	@echo "🔧 Generating ORM models..."
	@sqlboiler psql --config=$(SQLBOILER_CONFIG)

gen-models: sqlboiler ## Alias for sqlboiler

# =============================================================================
# DOCUMENTATION COMMANDS
# =============================================================================

docs: docs-build ## Build complete API documentation

docs-build: ## Build Swagger documentation from YAML files
	@echo "📚 Building API documentation..."
	@chmod +x scripts/build-docs.sh
	@./scripts/build-docs.sh

docs-clean: ## Clean generated documentation files
	@echo "🧹 Cleaning documentation build..."
	@rm -rf $(DOCS_OUTPUT_DIR)/
	@echo "✅ Cleaned $(DOCS_OUTPUT_DIR)/"

docs-serve: docs-build ## Build and serve documentation locally
	@echo "🌐 Starting documentation server..."
	@if command -v swagger-ui-serve >/dev/null 2>&1; then \
		echo "📖 Serving Swagger UI at http://localhost:3000"; \
		swagger-ui-serve $(DOCS_OUTPUT_DIR)/swagger.yaml; \
	elif command -v http-server >/dev/null 2>&1 && [ -d "$(DOCS_OUTPUT_DIR)/html" ]; then \
		echo "📖 Serving HTML docs at http://localhost:8080"; \
		cd $(DOCS_OUTPUT_DIR)/html && http-server -p 8080; \
	else \
		echo "⚠️  No local server available. Options:"; \
		echo "   1. Install swagger-ui-serve: npm install -g swagger-ui-serve"; \
		echo "   2. Install http-server: npm install -g http-server"; \
		echo "   3. Upload $(DOCS_OUTPUT_DIR)/swagger.yaml to https://editor.swagger.io/"; \
		echo ""; \
		echo "📄 Generated file: $(DOCS_OUTPUT_DIR)/swagger.yaml"; \
	fi

docs-validate: docs-build ## Validate generated Swagger documentation
	@echo "🔍 Validating Swagger documentation..."
	@if command -v swagger-codegen >/dev/null 2>&1; then \
		swagger-codegen validate -i $(DOCS_OUTPUT_DIR)/swagger.yaml; \
		echo "✅ Swagger validation completed"; \
	else \
		echo "⚠️  swagger-codegen not found. Install it for validation:"; \
		echo "   brew install swagger-codegen"; \
	fi

docs-watch: ## Watch for changes and rebuild documentation
	@echo "👀 Watching for documentation changes..."
	@if command -v fswatch >/dev/null 2>&1; then \
		echo "🔄 Watching docs/swagger/ for changes..."; \
		fswatch -o docs/swagger/ | xargs -n1 -I{} make docs-build; \
	elif command -v inotifywait >/dev/null 2>&1; then \
		echo "🔄 Watching docs/swagger/ for changes..."; \
		while inotifywait -r -e modify docs/swagger/; do make docs-build; done; \
	else \
		echo "⚠️  No file watcher available. Install fswatch or inotify-tools"; \
		echo "   macOS: brew install fswatch"; \
		echo "   Linux: apt-get install inotify-tools"; \
	fi

docs-info: ## Show documentation build information
	@echo "📊 Documentation Build Information"
	@echo ""
	@echo "📁 Documentation structure:"
	@if command -v tree >/dev/null 2>&1; then \
		tree docs/ 2>/dev/null || true; \
	else \
		find docs/ -type f 2>/dev/null | sort || echo "   docs/ directory not found"; \
	fi
	@echo ""
	@echo "📄 Generated files:"
	@if [ -d "$(DOCS_OUTPUT_DIR)" ]; then \
		ls -la $(DOCS_OUTPUT_DIR)/; \
	else \
		echo "   No generated files. Run 'make docs-build'"; \
	fi
	@echo ""
	@echo "🔧 Available tools:"
	@echo -n "   swagger-codegen: "; command -v swagger-codegen >/dev/null && echo "✅" || echo "❌"
	@echo -n "   swagger-ui-serve: "; command -v swagger-ui-serve >/dev/null && echo "✅" || echo "❌"
	@echo -n "   yq: "; command -v yq >/dev/null && echo "✅" || echo "❌"
	@echo -n "   yamllint: "; command -v yamllint >/dev/null && echo "✅" || echo "❌"

# =============================================================================
# PROTOCOL BUFFERS
# =============================================================================

proto: ## Generate Go code from protocol buffers
	@echo "🔧 Generating Go code from protocol buffers..."
	@protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           internal/proto_qr/account/account.proto \
           internal/proto_qr/branch/branch.proto

# =============================================================================
# DOCKER COMMANDS
# =============================================================================

rebuild-app: ## Rebuild and restart go_app container
	@echo "🐳 Rebuilding go_app container..."
	@docker compose down go_app
	@docker compose build --no-cache go_app
	@docker compose up -d go_app

rebuild-all: ## Rebuild and restart all services
	@echo "🐳 Rebuilding all containers..."
	@docker compose down
	@docker compose build --no-cache
	@docker compose up -d

restart-app: ## Quick restart go_app (without rebuild)
	@echo "🐳 Restarting go_app container..."
	@docker compose restart go_app

# =============================================================================
# DEVELOPMENT SETUP
# =============================================================================

dev-setup: ## Install development dependencies
	@echo "🔧 Installing development dependencies..."
	@echo "📦 Installing Node.js packages..."
	@if command -v npm >/dev/null 2>&1; then \
		npm install -g swagger-ui-serve http-server; \
		echo "✅ Node.js packages installed"; \
	else \
		echo "⚠️  npm not found. Install Node.js first"; \
	fi
	@echo "📦 Installing Go tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@if command -v brew >/dev/null 2>&1; then \
		echo "📦 Installing additional tools with Homebrew..."; \
		brew install swagger-codegen yq yamllint 2>/dev/null || echo "⚠️  Some tools may already be installed"; \
	fi
	@echo "✅ Development setup completed"

# =============================================================================
# UTILITY COMMANDS
# =============================================================================

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@go clean
	@rm -f $(TEST_USER_DIR)/*.test
	@rm -f $(PID_FILE)
	@rm -rf $(DOCS_OUTPUT_DIR)
	@echo "✅ Cleaned build artifacts"

# Combined workflows
full-setup: dev-setup migrate-up docs-build ## Complete development setup

full-test: docs-build test docs-validate ## Run comprehensive testing

quick-docs: docs-serve ## Quick build and serve documentation

# Development workflow shortcuts
dev: run-server docs-serve ## Start server and documentation (development mode)

# Production ready build
build-prod: clean proto gen-models docs-build ## Production build sequence

# Status check
status: ## Show project status
	@echo "📊 Project Status"
	@echo ""
	@echo "🔄 Server status:"
	@if [ -f $(PID_FILE) ]; then \
		echo "   Server PID: `cat $(PID_FILE)`"; \
		if ps -p `cat $(PID_FILE)` > /dev/null; then \
			echo "   Status: ✅ Running"; \
		else \
			echo "   Status: ❌ Not running (stale PID file)"; \
		fi \
	else \
		echo "   Status: ❌ Not running"; \
	fi
	@echo ""
	@echo "📚 Documentation:"
	@if [ -f "$(DOCS_OUTPUT_DIR)/swagger.yaml" ]; then \
		echo "   Status: ✅ Built ($(shell stat -f%Sm $(DOCS_OUTPUT_DIR)/swagger.yaml 2>/dev/null || stat -c%y $(DOCS_OUTPUT_DIR)/swagger.yaml 2>/dev/null || echo 'unknown date'))"; \
	else \
		echo "   Status: ❌ Not built"; \
	fi
	@echo ""
	@echo "🗄️  Database:"
	@echo "   Migration status: $(shell migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version 2>/dev/null || echo 'Unable to check')"

# Default target when just running 'make'
.DEFAULT_GOAL := help