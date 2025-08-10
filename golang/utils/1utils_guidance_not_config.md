# Go Utils Package Documentation

This package provides essential utilities for authentication, email services, password handling, and various helper functions for a Go web application.

## Table of Contents

- [Authentication (auth.go)](#authentication)
- [Email Services (email.go)](#email-services)
- [Helper Functions (helper.go)](#helper-functions)
- [Password Utilities (password.go)](#password-utilities)
- [UUID Utilities (stringtouuid.go)](#uuid-utilities)
- [Setup and Configuration](#setup-and-configuration)
- [Usage Examples](#usage-examples)

## Authentication

The authentication module provides JWT token management with support for access tokens, refresh tokens, reset tokens, and verification tokens.

### Key Components

#### JWTTokenMaker
Main struct for handling JWT operations with configurable token durations:
- **Access Token**: 15 minutes
- **Refresh Token**: 7 days  
- **Reset Token**: 1 hour
- **Verification Token**: 24 hours

### Available Functions

```go
// Function variables for easy mocking in tests
var (
    HashPassword         = hashPassword
    GenerateJWTToken     = generateJWTToken
    GenerateRefreshToken = generateRefreshToken
    ParseToken           = parseToken
)

// Password comparison
func Compare(hashedPassword, password string) bool

// Token maker constructor
func NewJWTTokenMaker(secretKey string) *JWTTokenMaker

// Token operations
func (maker *JWTTokenMaker) CreateToken(user model.Account) (string, error)
func (maker *JWTTokenMaker) VerifyToken(tokenString string) (*model.Account, error)
func (maker *JWTTokenMaker) CreateRefreshToken(user model.Account) (string, error)
func (maker *JWTTokenMaker) ValidateRefreshToken(tokenString string) (*model.Account, error)
func (maker *JWTTokenMaker) CreateResetToken(email string) (string, error)
func (maker *JWTTokenMaker) ValidateResetToken(tokenString string) (string, error)
func (maker *JWTTokenMaker) CreateVerificationToken(email string) (string, error)
func (maker *JWTTokenMaker) ValidateVerificationToken(tokenString string) (string, error)
```

### Usage Example

```go
import "english-ai-full/utils"

// Initialize token maker
config := utils_config.GetConfig()
tokenMaker := utils.NewJWTTokenMaker(config.JWT.SecretKey)

// Create access token
user := model.Account{
    ID:       123,
    Email:    "user@example.com",
    Role:     model.RoleUser,
    BranchID: 1,
}
token, err := tokenMaker.CreateToken(user)
if err != nil {
    // Handle error
}

// Verify token
account, err := tokenMaker.VerifyToken(token)
if err != nil {
    // Handle invalid token
}

// Create password reset token
resetToken, err := tokenMaker.CreateResetToken("user@example.com")

// Validate reset token
email, err := tokenMaker.ValidateResetToken(resetToken)
```

## Email Services

Provides both SMTP and mock email services for different environments.

### SMTP Email Service

Production-ready email service using SMTP protocol.

```go
type SMTPEmailService struct {
    host     string
    port     string
    username string
    password string
    from     string
}

// Constructor
func NewSMTPEmailService(config EmailConfig) *SMTPEmailService

// Email methods
func (s *SMTPEmailService) SendWelcomeEmail(ctx context.Context, email, name string) error
func (s *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, email, resetToken string) error
func (s *SMTPEmailService) SendAccountDeactivationEmail(ctx context.Context, email, name string) error
func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, email, verificationToken string) error
func (s *SMTPEmailService) SendPasswordChangedNotification(ctx context.Context, email, name string) error
```

### Mock Email Service

Testing/development service that prints to console instead of sending emails.

```go
func NewMockEmailService() *MockEmailService
// Same interface as SMTPEmailService but prints to console
```

### Usage Example

```go
// Production setup
emailConfig := utils.EmailConfig{
    Host:     "smtp.gmail.com",
    Port:     "587",
    Username: "your-email@gmail.com",
    Password: "your-app-password",
    From:     "noreply@yourapp.com",
}
emailService := utils.NewSMTPEmailService(emailConfig)

// Development/testing setup
emailService := utils.NewMockEmailService()

// Send emails
err := emailService.SendWelcomeEmail(ctx, "user@example.com", "John Doe")
err := emailService.SendPasswordResetEmail(ctx, "user@example.com", resetToken)
err := emailService.SendVerificationEmail(ctx, "user@example.com", verificationToken)
```

## Helper Functions

Comprehensive set of utility functions for logging, context management, and request handling.

### Context Management

```go
// Merge multiple context maps
func MergeContext(base, additional map[string]interface{}) map[string]interface{}
func MergeMultipleContexts(contexts ...map[string]interface{}) map[string]interface{}

// Add single key-value to context
func AddToContext(context map[string]interface{}, key string, value interface{}) map[string]interface{}

// Create base HTTP request context
func CreateBaseContext(r *http.Request, additionalFields map[string]interface{}) map[string]interface{}
```

### Security and Logging

```go
// Mask sensitive data for logging
func MaskSensitiveValue(fieldName string, value interface{}) interface{}

// Sanitize context for safe logging
func ValidateAndSanitizeContext(context map[string]interface{}) map[string]interface{}

// Extract or generate request ID
func GetRequestID(r *http.Request) string

// Get type name for logging
func GetTypeName(obj interface{}) string

// Format duration for human-readable logs
func FormatDuration(duration time.Duration) string
```

### Usage Example

```go
// Create request context
baseCtx := utils.CreateBaseContext(r, map[string]interface{}{
    "user_id": userID,
    "operation": "create_user",
})

// Add more context
enrichedCtx := utils.AddToContext(baseCtx, "validation_errors", errors)

// Sanitize for logging (masks sensitive data)
safeCtx := utils.ValidateAndSanitizeContext(enrichedCtx)

// Log with safe context
logger.InfoContext(ctx, "User creation attempt", safeCtx)

// Mask sensitive values
maskedPassword := utils.MaskSensitiveValue("password", "secretpassword123")
// Returns: "s***3"

maskedEmail := utils.MaskSensitiveValue("email", "user@example.com")  
// Returns: "u***r@example.com"
```

## Password Utilities

Secure password hashing using bcrypt with configurable cost.

### BcryptPasswordHasher

```go
type BcryptPasswordHasher struct {
    cost int
}

// Constructors
func NewBcryptPasswordHasher() *BcryptPasswordHasher                    // Uses default cost (10)
func NewBcryptPasswordHasherWithCost(cost int) *BcryptPasswordHasher    // Custom cost

// Methods
func (h *BcryptPasswordHasher) HashPassword(password string) (string, error)
func (h *BcryptPasswordHasher) ComparePassword(hashedPassword, password string) bool

// Standalone function
func CheckPassword(password string, hashedPassword string) error
```

### Usage Example

```go
// Create hasher
hasher := utils.NewBcryptPasswordHasher()

// Hash password
hashedPassword, err := hasher.HashPassword("userpassword123")
if err != nil {
    // Handle error
}

// Compare password
isValid := hasher.ComparePassword(hashedPassword, "userpassword123")

// Or use standalone function
err := utils.CheckPassword("userpassword123", hashedPassword)
isValid := err == nil

// Custom cost for higher security (slower hashing)
secureHasher := utils.NewBcryptPasswordHasherWithCost(12)
```

## UUID Utilities

UUID parsing and formatted name generation.

```go
// Parse string to UUID
func StringToUUID(s string) (uuid.UUID, error)

// Generate timestamped unique names
func GenerateFormattedName(baseName string) string
```

### Usage Example

```go
// Parse UUID
id, err := utils.StringToUUID("550e8400-e29b-41d4-a716-446655440000")

// Generate unique filename/identifier
uniqueName := utils.GenerateFormattedName("user-avatar")
// Returns: "user-avatar-2024/01/15/14/30/45-550e8400-e29b-41d4-a716-446655440000"
```

## Setup and Configuration

### Prerequisites

Required dependencies in your `go.mod`:

```go
require (
    github.com/golang-jwt/jwt/v5 v5.0.0
    github.com/google/uuid v1.3.0
    golang.org/x/crypto v0.14.0
)
```

### Configuration Structure

Your config package should provide:

```go
type Config struct {
    JWT struct {
        SecretKey string `json:"secret_key"`
    } `json:"jwt"`
    
    Email struct {
        Host     string `json:"host"`
        Port     string `json:"port"`
        Username string `json:"username"`
        Password string `json:"password"`
        From     string `json:"from"`
    } `json:"email"`
}
```

### Model Requirements

The `model.Account` struct should include:

```go
type Account struct {
    ID       int64  `json:"id"`
    Email    string `json:"email"`
    Role     Role   `json:"role"`
    BranchID int64  `json:"branch_id"`
}

type Role string // Define your role constants
```

## Usage Examples

### Complete Authentication Flow

```go
package main

import (
    "context"
    "net/http"
    "english-ai-full/utils"
    "english-ai-full/internal/model"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    // Create base context for logging
    ctx := utils.CreateBaseContext(r, nil)
    
    // Hash password
    hasher := utils.NewBcryptPasswordHasher()
    hashedPassword, err := hasher.HashPassword(password)
    if err != nil {
        // Log with context
        logger.ErrorContext(r.Context(), "Password hashing failed", 
            utils.AddToContext(ctx, "error", err))
        return
    }
    
    // Create user account
    user := model.Account{
        ID:       123,
        Email:    "user@example.com",
        Role:     model.RoleUser,
        BranchID: 1,
    }
    
    // Generate tokens
    tokenMaker := utils.NewJWTTokenMaker(config.JWT.SecretKey)
    accessToken, err := tokenMaker.CreateToken(user)
    refreshToken, err := tokenMaker.CreateRefreshToken(user)
    verificationToken, err := tokenMaker.CreateVerificationToken(user.Email)
    
    // Send welcome email
    emailService := utils.NewSMTPEmailService(emailConfig)
    err = emailService.SendWelcomeEmail(context.Background(), user.Email, "John Doe")
    err = emailService.SendVerificationEmail(context.Background(), user.Email, verificationToken)
    
    // Generate unique identifier
    userFolder := utils.GenerateFormattedName("user-data")
    
    // Return tokens (implementation depends on your response format)
    response := map[string]string{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "user_folder":   userFolder,
    }
}

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }
        
        // Verify token
        tokenMaker := utils.NewJWTTokenMaker(config.JWT.SecretKey)
        user, err := tokenMaker.VerifyToken(strings.TrimPrefix(authHeader, "Bearer "))
        if err != nil {
            // Log with masked context
            ctx := utils.CreateBaseContext(r, map[string]interface{}{
                "error": "token_verification_failed",
                "token": utils.MaskSensitiveValue("token", authHeader),
            })
            logger.WarnContext(r.Context(), "Authentication failed", 
                utils.ValidateAndSanitizeContext(ctx))
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Add user to request context
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Testing with Mock Services

```go
func TestEmailService(t *testing.T) {
    // Use mock service in tests
    emailService := utils.NewMockEmailService()
    
    err := emailService.SendWelcomeEmail(context.Background(), 
        "test@example.com", "Test User")
    
    assert.NoError(t, err)
    // Check console output for mock email
}
```

## Security Best Practices

1. **JWT Secret Key**: Use a strong, randomly generated secret key
2. **Password Hashing**: Use appropriate bcrypt cost (10-12 for production)
3. **Token Expiration**: Keep access tokens short-lived (15 minutes)
4. **Sensitive Data**: Always use masking functions for logging
5. **Email Verification**: Implement proper email verification flow
6. **Request IDs**: Use request IDs for tracing across services

## Error Handling

The package returns standard Go errors. Always check for errors and handle them appropriately:

```go
token, err := tokenMaker.CreateToken(user)
if err != nil {
    logger.ErrorContext(ctx, "Token creation failed", 
        utils.AddToContext(baseCtx, "error", err))
    return fmt.Errorf("authentication failed: %w", err)
}
```

## Performance Considerations

- **Bcrypt Cost**: Higher cost = more secure but slower (balance security vs. performance)
- **Context Merging**: Functions create new maps to avoid mutations but use memory
- **Email Service**: Use connection pooling for high-volume email sending
- **Token Verification**: Cache public keys if using external JWT verification

This documentation provides a comprehensive guide for developers to understand and effectively use the utils package in your Go application.