# Uses ./configs/config.yaml or falls back to default search paths
go run main.go
# Uses ./configs/config-production.yaml
GO_ENV=production go run main.go

# Uses ./configs/config-testing.yaml  
GO_ENV=testing go run main.go

# Uses specific config file
CONFIG_PATH=./my-custom-config.yaml go run main.go

# Override any config value
ENGLISH_AI_SERVER_PORT=9000 go run main.go