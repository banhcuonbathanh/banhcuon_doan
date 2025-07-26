cd golang
go run cmd/server/main.go
go run cmd/grpc-server/main.go
cd golang
docker-compose up -d mypostgres_ai
docker compose up

for testing position in golang folder command : go test ./internal/account/account_unit_test/

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative golang/internal/proto_qr/account/account.proto

monghoaivu@mongs-MacBook-Pro golang % go test ./internal/account/account_unit_test/account_authentication.go
