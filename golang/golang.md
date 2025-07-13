go get -u github.com/jackc/pgx/v4
go get -u github.com/golang-migrate/migrate/v4
go get -u google.golang.org/grpc
go get -u github.com/spf13/viper

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc/proto/user.proto

go run cmd/server/main.go

chek env path.
monghoaivu@192 ~ % echo $(go env GOPATH)/bin
/Users/monghoaivu/go/bin
monghoaivu@192 ~ %

which protoc-gen-go
which protoc-gen-go-grpc

---

syntax = "proto3";

package restaurant;

import "google/protobuf/timestamp.proto";

message Account {
int64 id = 1;
string name = 2;
string email = 3;
string password = 4;
string avatar = 5;
string role = 6;
int64 owner_id = 7;
google.protobuf.Timestamp created_at = 8;
google.protobuf.Timestamp updated_at = 9;
}

message Dish {
int64 id = 1;
string name = 2;
int32 price = 3;
string description = 4;
string image = 5;
string status = 6;
google.protobuf.Timestamp created_at = 7;
google.protobuf.Timestamp updated_at = 8;
}

message DishSnapshot {
int64 id = 1;
string name = 2;
int32 price = 3;
string description = 4;
string image = 5;
string status = 6;
int64 dish_id = 7;
google.protobuf.Timestamp created_at = 8;
google.protobuf.Timestamp updated_at = 9;
}



message Guest {
int64 id = 1;
string name = 2;
int32 table_number = 3;
string refresh_token = 4;
google.protobuf.Timestamp refresh_token_expires_at = 5;
google.protobuf.Timestamp created_at = 6;
google.protobuf.Timestamp updated_at = 7;
}

message Order {
int64 id = 1;
int64 guest_id = 2;
int32 table_number = 3;
int64 dish_snapshot_id = 4;
int32 quantity = 5;
int64 order_handler_id = 6;
string status = 7;
google.protobuf.Timestamp created_at = 8;
google.protobuf.Timestamp updated_at = 9;
}

message RefreshToken {
string token = 1;
int64 account_id = 2;
google.protobuf.Timestamp expires_at = 3;
google.protobuf.Timestamp created_at = 4;
}

message Socket {
string socket_id = 1;
int64 account_id = 2;
int64 guest_id = 3;
}

message Table {
int32 number = 1;
int32 capacity = 2;
string status = 3;
string token = 4;
google.protobuf.Timestamp created_at = 5;
google.protobuf.Timestamp updated_at = 6;
}