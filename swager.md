# Safe to install - minimal impact
go install github.com/swaggo/swag/cmd/swag@latest
go get -u github.com/swaggo/http-swagger
go get -u github.com/swaggo/files

swag init -g cmd/server/main.go

 swag init -g cmd/server/main.go -o ./docs

 # 2. Add the files from artifacts
# 3. Make executable
chmod +x scripts/build-docs.sh

# 4. Test it
make docs-build
make docs-serve