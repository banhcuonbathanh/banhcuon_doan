cd golang
go run cmd/server/main.go
go run cmd/grpc-server/main.go
docker-compose up -d mypostgres_ai
docker compose up
