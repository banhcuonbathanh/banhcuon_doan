cd golang
go run cmd/server/main.go
go run cmd/grpc-server/main.go
cd golang
docker-compose up -d mypostgres_ai
docker compose up

for testing position in golang folder command : go test ./internal/account/account_unit_test/
