http://localhost:3000/table/1?token=MTo0OkF2YWlsYWJsZTo0ODg3NDI2MzQz.0v1DNiYriLs
fmt.Printf("golang/quanqr/order/order_handler.go ordersResponse %v\n", ordersResponse)
docker-compose up mypostgres_ai

cd quananqr1
npm run dev

cd english-app-fe-nextjs

cd golang

go get -u github.com/go-chi/chi/v5
cd golang
go run cmd/server/main.go
cd golang
go run cmd/grcp-server/main.go

cd golang && cd cmd && cd python && source env/bin/activate
python server/python_server.py
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. python_proto/claude/claude.proto
go run cmd/client/main.go
======================================= postgres ======================
psql -U myuser -d mydatabase

# psql -U myuser -d mydatabase

DROP DATABASE mydatabase;
TRUNCATE TABLE schema\*migrations, users; delete all data
\dt : list all table
\d orders
\d users
\d comments
\d sessions
\d reading_test_models;

\d orders
SELECT * FROM dish_order_items;
SELECT * FROM tables;
SELECT * FROM users;
SELECT _ FROM users;
SELECT \_ FROM dishes;
SELECT * FROM orders;
SELECT \_ FROM users;
SELECT \* FROM sessions;
DELETE FROM sessions;
\d order_items
mydatabase=# \d users
SELECT \* FROM reading_tests;
DROP TABLE schema_migrations;
DELETE FROM schema_migrations;
DELETE FROM reading_tests;
\l
\c testdb
testdb=# \dT+ paragraph_content
UPDATE users
SET is_admin = true
WHERE id = 1;
migrate -database postgres://myuser:mypassword@localhost:5432/mydatabase?sslmode=disable force 7

-- List all tables in the public schema
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_type = 'BASE TABLE';

-- List all custom types (including ENUMs)
SELECT t.typname AS enum_name,
e.enumlabel AS enum_value
FROM pg_type t
JOIN pg_enum e ON t.oid = e.enumtypid
JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
WHERE n.nspname = 'public';

-- Drop all tables in the public schema
DO $$
DECLARE
r RECORD;
BEGIN
FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
END LOOP;
END $$;
DROP TYPE IF EXISTS question_type CASCADE;
-- Drop the question_type ENUM
DROP TYPE IF EXISTS question_type CASCADE;

-- Verify that all tables are dropped
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_type = 'BASE TABLE';

-- Verify that the question_type ENUM is dropped
SELECT t.typname AS enum_name,
e.enumlabel AS enum_value
FROM pg_type t
JOIN pg_enum e ON t.oid = e.enumtypid
JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
WHERE n.nspname = 'public';
=================================================== docker =======================
branch delivery2 and makeorder
docker-compose up -d
docker-compose up
docker compose build go_app_ai
docker compose down
docker-compose up mypostgres_ai

---

docker-compose stop nextjs_app

docker-compose build nextjs_app

docker-compose up -d nextjs_app

docker-compose up -d --build nextjs_app (all above command in 1 shot)

project-root/
├── docker-compose.yml
├── quananqr1/
│ ├── Dockerfile
│ ├── .dockerignore <-- Place it here
│ ├── src/
│ ├── public/
│ ├── package.json
│ └── package-lock.json
└── golang/
├── Dockerfile
└── .dockerignore <-- You can have another one here for Golang
//

curl http://localhost:8888/qr/guest/test

curl http://go_app_ai:8888/qr/guest/test
========================================= golang ==============================

go run cmd/server/main.go

Run the desired commands using make <target>. For example:

To run the server: make run-server
To run the client: make run-client
To run all tests: make test
To run only the CreateUser test: make test-create
To run only the GetUser test: make test-getf
To clean build artifacts: make clean
To see available commands: make help

make stop-server

go test -v test/test-api/test-api.go
golang/
============================================== git hub ================================
git branch make_order1
git checkout make_order1
git branch -d web-sokcert-new-strtuchture delete branch
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc/proto/python_proto/claude/claude.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc/proto/python_proto/helloworld.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc-python/ielts/proto/ielts.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc/proto/claude/claude.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc/proto/comment/comment.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc/proto/user.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ecomm-grpc/proto/reading/reading.proto
git checkout nextjs-fe-readiding-add-more-clean-architextture
git merge golang-new-server-for-grpc
git commit
git push origin dev

golang/ecomm-grpc/proto/reading/reading.proto

git checkout -b golang: create new branch

reading_test_models
section_models
passage_models
schema_migrations
paragraph_content_models
question_models
users
sessions

Jump back to the golang branch:
git checkout test_isadmin

Merge the golang branch with the python branch:
Jump back to the golang branch:
git checkout test_isadmin

Merge the golang branch with the python branch:
git merge guest
git merge --no-ff guest

Update the changes to the remote repository:
git push origin test_isadmin

Jump back to the python branch:
git checkout guest

git branch
========================================= golang ==============================

====================================== project proto ============================

cd project_protos

go mod init project_proto

source env/bin/activate

cd python
python server/greeter_server.py

python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. python_proto/helloworld.proto

python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. python_proto/claude/claude.proto

------------------------------------- quan an qr ------------
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/delivery/delivery.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/set/set.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/account/account.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/dish/dish.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/dishsnapshot/dishsnapshot.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/guest/guest.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/order/order.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative quanqr/proto_qr/table/table.proto

http://localhost:8888/images/image?filename=Screenshot%202024-02-20%20at%2014.37.22.png&path=folder1/folder2

=============================== ws ========================

ws://localhost:8888/ws?userId=8&userName=vy1_2024_11_07_14_53_51_24e6a4af-9052-4dfd-8cc9-45fce2cc264d&isGuest=true

{
"connection": {
"url": "ws://localhost:8888/ws?userId=8&userName=vy1_2024_11_07_14_53_51_24e6a4af-9052-4dfd-8cc9-45fce2cc264d&isGuest=true"
},
"messages": {
"directMessage": {
"type": "DIRECT_MESSAGE",
"content": "Hello, this is a test message",
"toUser": "recipient_user_id",
"tableId": "table_123",
"orderId": "order_456"
},
"newOrder": {
"type": "NEW_ORDER",
"content": {
"orderId": 123,
"orderData": {
"guest_id": null,
"user_id": 8,
"is_guest": true,
"table_number": 1,
"order_handler_id": 456,
"status": "PENDING",
"created_at": "2024-11-07T14:53:51Z",
"updated_at": "2024-11-07T14:53:51Z",
"total_price": 25.50,
"dish_items": [
{
"dish_id": 1,
"quantity": 2
}
],
"set_items": [
{
"set_id": 1,
"quantity": 1
}
],
"bow_chili": 1,
"bow_no_chili": 0,
"takeAway": false,
"chiliNumber": 2,
"table_token": "table_123_token",
"order_name": "Order #123"
}
},
"toUser": "recipient_user_id",
"tableId": "table_123",
"orderId": "order_456"
},
"orderStatusUpdate": {
"type": "ORDER_STATUS_UPDATE",
"content": {
"orderId": 123,
"status": "COMPLETED",
"timestamp": "2024-11-07T14:53:51Z"
},
"toUser": "recipient_user_id",
"tableId": "table_123",
"orderId": "order_456"
}
}
}

ws://localhost:8888/ws?userId=8&userName=vy1_2024_11_07_14_53_51_24e6a4af-9052-4dfd-8cc9-45fce2cc264d&isGuest=true

ws://localhost:8888/ws?userId=9&userName=dung_2024_11_08_12_43_15_0ed49e95-07c3-489f-a6f3-f6a8dcef835a&isGuest=true

ws-----------

wsService := NewWebSocketService(messageRepo, orderHandler)
wsHandler := NewWebSocketHandler(wsService)

r := chi.NewRouter()
r = RegisterWebSocketRoutes(r, wsHandler, orderHandler)

Users: ws://localhost:8888/ws/user/1?username=John
Guests: ws://localhost:8888/ws/guest/1?guestname=Guest1

ws://localhost:8888/ws?userId=8&userName=vy1_2024_11_07_14_53_51_24e6a4af-9052-4dfd-8cc9-45fce2cc264d&isGuest=true

ws://localhost:8888/ws?userId=1&userName=John&isGuest=false

ws://localhost:8888/ws?userId=9&userName=Johnguest&isGuest=true

ws://localhost:8888/ws/user/1?token=abc123&tableToken=table456

ws://localhost:8888/ws/user/2?token=abc124&tableToken=table455
ws://localhost:8888/ws/employee/1?token=abc123&tableToken=table455
ws://localhost:8888/ws/admin/1?token=abc123&tableToken=table455

ws smessage

// 1. Create Delivery Message Structure
{
"type": "delivery",
"action": "create",
"payload": {
"fromUserId": "user_123",
"toUserId": "staff_456",
"type": "delivery",
"action": "create",
"payload": {
"guest_id": null,
"user_id": 1,
"is_guest": false,
"table_number": 1,
"order_handler_id": 1,
"status": "Pending",
"total_price": 2500,
"dish_items": [
{
"dish_id": 1,
"quantity": 2
},
{
"dish_id": 2,
"quantity": 1
}
],
"bow_chili": 1,
"bow_no_chili": 1,
"take_away": false,
"chili_number": 3,
"table_token": "MTp0YWJsZTo0ODg0Mjk0Mjk0.666YJoUIKKI",
"client_name": "John Doe",
"delivery_address": "123 Main St, Springfield",
"delivery_contact": "555-1234",
"delivery_notes": "Ring the doorbell upon arrival",
"scheduled_time": "2024-11-04T14:00:00Z",
"order_id": 456789,
"delivery_fee": 200,
"delivery_status": "pending"
}
},
"role": "User",
"roomId": ""
}

// 2. Update Delivery Status Message Structure
{
"type": "delivery",
"action": "update_status",
"payload": {
"delivery_id": "delivery_789",
"status": "in_progress"
},
"role": "Employee",
"roomId": ""
}

// 3. Assign Delivery Message Structure
{
"type": "delivery",
"action": "assign",
"payload": {
"delivery_id": "delivery_789",
"driver_id": "driver_123"
},
"role": "Employee",
"roomId": ""
}

int64 bow_chili = 13;

import Image from 'next/image'

------------------------ linux -------------------
find . -name "\*.tsx" | grep -E "login|dialog"

grep -r "from.*LoginDialog" .grep -r "from.*login-dialog" .
grep -r "from.\*LoginDialog" .

find . -type f \( -name "_.tsx" -o -name "_.jsx" \) -exec grep -l "useAuthStore" {} \;

grep -r "from.\*login-dialog" . ---> ok

-------------------------- zustand optimization -----------------------------
how to optimize Zustand store subscriptions to prevent unnecessary rerenders.
