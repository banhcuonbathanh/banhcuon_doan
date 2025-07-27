# Configuration Files Examples

## config.yaml (Development)
```yaml
# Development configuration
environment: development
app_name: "English AI"
version: "1.0.0"
debug: true

server:
  address: "localhost"
  port: 8080
  grpc_address: "localhost"
  grpc_port: 50051
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  tls_enabled: false

database:
  host: "localhost"
  port: 5432
  name: "english_ai_dev"
  user: "postgres"
  # password set via environment variable: ENGLISH_AI_DATABASE_PASSWORD
  ssl_mode: "disable"
  max_connections: 25
  max_idle_conns: 10
  conn_max_lifetime: "1h"
  conn_max_idle_time: "10m"

security:
  max_login_attempts: 5
  account_lockout_minutes: 15
  session_timeout: "24h"
  csrf_enabled: true
  cors_enabled: true
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:3001"
  require_https: false

password:
  min_length: 8
  max_length: 128
  require_uppercase: true
  require_lowercase: true
  require_numbers: true
  require_special: true
  special_chars: "!@#$%^&*()_+-=[]{}|;:,.<>?/"

pagination:
  default_size: 10
  max_size: 100
  limit: 1000

jwt:
  # secret_key set via environment variable: ENGLISH_AI_JWT_SECRET_KEY
  expiration_hours: 24
  refresh_token_expiration_days: 30
  issuer: "english-ai-dev"
  algorithm: "HS256"
  refresh_threshold: "2h"

email:
  verification_enabled: true
  verification_expiry_hours: 24
  require_verification: false
  smtp_host: "localhost"
  smtp_port: 1025  # MailHog for development
  smtp_user: ""
  smtp_password: ""
  from_address: "noreply@english-ai.dev"
  from_name: "English AI Development"
  templates:
    verification_template: "./templates/email/verification.html"
    welcome_template: "./templates/email/welcome.html"
    reset_password_template: "./templates/email/reset_password.html"

rate_limit:
  enabled: true
  per_minute: 60
  per_hour: 3600
  burst_size: 10
  window_size: "1m"

external_apis:
  anthropic:
    # api_key set via environment variable: ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY
    api_url: "https://api.anthropic.com"
    timeout: "30s"
    max_retries: 3
  quan_an:
    address: "localhost:8081"
    timeout: "10s"
    max_retries: 3

logging:
  level: "debug"
  format: "text"
  output: "stdout"
  max_size: 100
  max_backups: 3
  max_age: 28
  compress: false

valid_roles:
  - "admin"
  - "user"
  - "manager"
  - "teacher"
  - "student"
```

## config-production.yaml
```yaml
# Production configuration
environment: production
app_name: "English AI"
version: "1.0.0"
debug: false

server:
  address: "0.0.0.0"
  port: 8080
  grpc_address: "0.0.0.0"
  grpc_port: 50051
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  tls_enabled: true
  cert_file: "/etc/ssl/certs/english-ai.crt"
  key_file: "/etc/ssl/private/english-ai.key"

database:
  host: "postgres-primary.internal"
  port: 5432
  name: "english_ai_prod"
  user: "english_ai_user"
  # password set via environment variable: ENGLISH_AI_DATABASE_PASSWORD
  ssl_mode: "require"
  max_connections: 50
  max_idle_conns: 20
  conn_max_lifetime: "2h"
  conn_max_idle_time: "15m"

security:
  max_login_attempts: 3
  account_lockout_minutes: 30
  session_timeout: "8h"
  csrf_enabled: true
  cors_enabled: true
  allowed_origins:
    - "https://english-ai.com"
    - "https://app.english-ai.com"
  require_https: true

password:
  min_length: 12
  max_length: 256
  require_uppercase: true
  require_lowercase: true
  require_numbers: true
  require_special: true
  special_chars: "!@#$%^&*()_+-=[]{}|;:,.<>?/"

pagination:
  default_size: 20
  max_size: 100
  limit: 1000

jwt:
  # secret_key set via environment variable: ENGLISH_AI_JWT_SECRET_KEY (must be 64+ chars)
  expiration_hours: 8
  refresh_token_expiration_days: 7
  issuer: "english-ai.com"
  algorithm: "HS256"
  refresh_threshold: "1h"

email:
  verification_enabled: true
  verification_expiry_hours: 2
  require_verification: true
  smtp_host: "smtp.sendgrid.net"
  smtp_port: 587
  smtp_user: "apikey"
  # smtp_password set via environment variable: ENGLISH_AI_EMAIL_SMTP_PASSWORD
  from_address: "noreply@english-ai.com"
  from_name: "English AI"
  templates:
    verification_template: "/app/templates/email/verification.html"
    welcome_template: "/app/templates/email/welcome.html"
    reset_password_template: "/app/templates/email/reset_password.html"

rate_limit:
  enabled: true
  per_minute: 30
  per_hour: 1800
  burst_size: 5
  window_size: "1m"

external_apis:
  anthropic:
    # api_key set via environment variable: ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY
    api_url: "https://api.anthropic.com"
    timeout: "60s"
    max_retries: 5
  quan_an:
    address: "quan-an-service.internal:8081"
    timeout: "15s"
    max_retries: 3

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/english-ai/app.log"
  max_size: 200
  max_backups: 10
  max_age: 90
  compress: true

valid_roles:
  - "admin"
  - "user"
  - "manager"
  - "teacher"
  - "student"
```

## config-testing.yaml
```yaml
# Testing configuration
environment: testing
app_name: "English AI Test"
version: "test"
debug: true

server:
  address: "localhost"
  port: 0  # Random available port
  grpc_address: "localhost"
  grpc_port: 0  # Random available port
  read_timeout: "5s"
  write_timeout: "5s"
  idle_timeout: "10s"
  tls_enabled: false

database:
  url: "sqlite:///:memory:"
  host: "localhost"
  port: 5432
  name: "test_db"
  user: "test"
  password: "test"
  ssl_mode: "disable"
  max_connections: 5
  max_idle_conns: 2
  conn_max_lifetime: "1h"
  conn_max_idle_time: "10m"

security:
  max_login_attempts: 100
  account_lockout_minutes: 1
  session_timeout: "1h"
  csrf_enabled: false
  cors_enabled: true
  allowed_origins:
    - "*"
  require_https: false

password:
  min_length: 4
  max_length: 128
  require_uppercase: false
  require_lowercase: false
  require_numbers: false
  require_special: false
  special_chars: "!@#$%^&*"

pagination:
  default_size: 5
  max_size: 50
  limit: 100

jwt:
  secret_key: "test-secret-key-32-characters-long"
  expiration_hours: 1
  refresh_token_expiration_days: 1
  issuer: "english-ai-test"
  algorithm: "HS256"
  refresh_threshold: "10m"

email:
  verification_enabled: false
  verification_expiry_hours: 1
  require_verification: false
  smtp_host: "localhost"
  smtp_port: 1025
  from_address: "test@example.com"
  from_name: "Test"

rate_limit:
  enabled: false
  per_minute: 1000
  per_hour: 10000
  burst_size: 100
  window_size: "1m"

external_apis:
  anthropic:
    api_key: "test-anthropic-key"
    api_url: "http://localhost:8082"  # Mock server
    timeout: "5s"
    max_retries: 0
  quan_an:
    address: "localhost:8083"  # Mock server
    timeout: "5s"
    max_retries: 0

logging:
  level: "error"
  format: "text"
  output: "stdout"
  max_size: 10
  max_backups: 1
  max_age: 1
  compress: false

valid_roles:
  - "admin"
  - "user"
```

## Docker Environment Variables (.env)
```bash
# Application
ENGLISH_AI_ENVIRONMENT=docker
ENGLISH_AI_APP_NAME="English AI Docker"
ENGLISH_AI_DEBUG=false

# Database (Docker Compose)
DB_HOST=postgres
DB_PORT=5432
DB_NAME=english_ai
DB_USER=english_ai_user
DB_PASSWORD=your_secure_database_password

# JWT
ENGLISH_AI_JWT_SECRET_KEY=your-super-secure-jwt-secret-key-64-characters-long-for-production

# Anthropic API
ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY=your_anthropic_api_key

# Email (optional)
ENGLISH_AI_EMAIL_SMTP_PASSWORD=your_smtp_password

# Override any other settings
ENGLISH_AI_SERVER_PORT=8080
ENGLISH_AI_LOGGING_LEVEL=info
```

## Docker Compose Override
```yaml
# docker-compose.override.yml
version: '3.8'
services:
  app:
    environment:
      - ENGLISH_AI_ENVIRONMENT=docker
      - ENGLISH_AI_DATABASE_URL=postgres://english_ai_user:${DB_PASSWORD}@postgres:5432/english_ai?sslmode=disable
      - ENGLISH_AI_JWT_SECRET_KEY=${JWT_SECRET_KEY}
      - ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    ports:
      - "8080:8080"
      - "50051:50051"
```

## Environment-specific Usage Examples

### Development Setup
```bash
# Set required environment variables
export ENGLISH_AI_JWT_SECRET_KEY="dev-secret-key-32-characters-long"
export ENGLISH_AI_DATABASE_PASSWORD="devpassword"
export ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY="your-dev-api-key"

# Run with default config
go run main.go

# Or specify config file
go run main.go -config=config-dev.yaml
```

### Production Deployment
```bash
# Set secure environment variables
export ENGLISH_AI_JWT_SECRET_KEY="production-jwt-secret-key-that-is-exactly-64-characters-long"
export ENGLISH_AI_DATABASE_PASSWORD="super-secure-production-password"
export ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY