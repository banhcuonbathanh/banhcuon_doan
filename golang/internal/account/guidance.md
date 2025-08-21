# Go Account Service - Function Signatures Guide

## Overview
This guide provides all function signatures for the Go account service based on your codebase structure. The service handles user authentication, account management, and user operations with gRPC and HTTP interfaces.

## Core Service Interface

### Account Service (gRPC)
```go
// Account Management
func CreateUser(ctx context.Context, req *AccountReq) (*Account, error) {}
func UpdateUser(ctx context.Context, req *UpdateUserReq) (*AccountRes, error) {}
func DeleteUser(ctx context.Context, req *DeleteAccountReq) (*DeleteAccountRes, error) {}
func FindAllUsers(ctx context.Context, req *empty.Empty) (*AccountList, error) {}
func FindByID(ctx context.Context, req *FindByIDReq) (*FindByIDRes, error) {}
func FindByEmail(ctx context.Context req *FindByEmailReq) (*AccountRes, error) {}

// Authentication & Session
func Login(ctx context.Context, req *LoginReq) (*AccountRes, error) {}
func Logout(ctx context.Context, req *LogoutReq) (*LogoutRes, error) {}
func Register(ctx context.Context, req *RegisterReq) (*RegisterRes, error) {}

// Password Management
func ChangePassword(ctx context.Context, req *ChangePasswordReq) (*ChangePasswordRes, error) {}
func ResetPassword(ctx context.Context, req *ResetPasswordReq) (*ResetPasswordRes, error) {}
func ForgotPassword(ctx context.Context, req *ForgotPasswordReq) (*ForgotPasswordRes, error) {}

// Account Verification & Status
func VerifyEmail(ctx context.Context, req *VerifyEmailReq) (*VerifyEmailRes, error) {}
func ResendVerification(ctx context.Context, req *ResendVerificationReq) (*ResendVerificationRes, error) {}
func UpdateAccountStatus(ctx context.Context, req *UpdateAccountStatusReq) (*UpdateAccountStatusRes, error) {}

// Search & Filtering
func FindByRole(ctx context.Context, req *FindByRoleReq) (*AccountList, error) {}
func FindByBranch(ctx context.Context, req *FindByBranchReq) (*AccountList, error) {}
func SearchUsers(ctx context.Context, req *SearchUsersReq) (*SearchUsersRes, error) {}

// Token Management
func RefreshToken(ctx context.Context, req *RefreshTokenReq) (*RefreshTokenRes, error) {}
func ValidateToken(ctx context.Context, req *ValidateTokenReq) (*ValidateTokenRes, error) {}
```

## HTTP Handler Interface

### Authentication Handlers
```go
func Register(w http.ResponseWriter, r *http.Request) {}
func Login(w http.ResponseWriter, r *http.Request) {}
func Logout(w http.ResponseWriter, r *http.Request) {}
```

### Account Management Handlers
```go
func CreateAccount(w http.ResponseWriter, r *http.Request) {}
func FindAccountByID(w http.ResponseWriter, r *http.Request) {}
func UpdateUserByID(w http.ResponseWriter, r *http.Request) {}
func DeleteUser(w http.ResponseWriter, r *http.Request) {}
func GetUserProfile(w http.ResponseWriter, r *http.Request) {}
```

### Password Management Handlers
```go
func ChangePassword(w http.ResponseWriter, r *http.Request) {}
func ForgotPassword(w http.ResponseWriter, r *http.Request) {}
func ResetPassword(w http.ResponseWriter, r *http.Request) {}
```

### Search & Filter Handlers
```go
func FindByEmail(w http.ResponseWriter, r *http.Request) {}
func GetUsersByBranch(w http.ResponseWriter, r *http.Request) {}
func SearchUsers(w http.ResponseWriter, r *http.Request) {}
```

## Repository Interface

### Core CRUD Operations
```go
func CreateUser(ctx context.Context, user *Account) (*Account, error) {}
func FindByID(ctx context.Context, id int64) (*Account, error) {}
func FindByEmail(ctx context.Context, email string) (*Account, error) {}
func UpdateUser(ctx context.Context, user *Account) (*Account, error) {}
func DeleteUser(ctx context.Context, id int64) error {}
func FindAllUsers(ctx context.Context) ([]*Account, error) {}
```

### Advanced Search Operations
```go
func FindByRole(ctx context.Context, role string, pagination *PaginationInfo) ([]*Account, error) {}
func FindByBranch(ctx context.Context, branchID int64, pagination *PaginationInfo) ([]*Account, error) {}
func SearchUsers(ctx context.Context, req *SearchUsersRequest) ([]*Account, *PaginationInfo, error) {}
```

### Bulk Operations
```go
func BulkUpdateStatus(ctx context.Context, userIDs []int64, status AccountStatus) (*BulkOperationResult, error) {}
func BulkDeleteUsers(ctx context.Context, userIDs []int64) (*BulkOperationResult, error) {}
```

## Validation & Utility Functions

### Request Validation
```go
func ValidateLoginRequest(req *LoginRequest) error {}
func ValidateRegisterRequest(req *RegisterUserRequest) error {}
func ValidateCreateUserRequest(req *CreateUserRequest) error {}
func ValidateUpdateUserRequest(req *UpdateUserRequest) error {}
func ValidateChangePasswordRequest(req *ChangePasswordRequest) error {}
func ValidatePasswordRequirements(password string, requirements *PasswordRequirements) error {}
```

### Data Transformation
```go
func ToAccountResponse(account *Account) *AccountLoginResponse {}
func ToUserProfile(account *Account) *UserProfile {}
func ToLoginResponse(account *Account, tokens *TokenPair) *LoginResponse {}
func ToRegisterResponse(account *Account) *RegisterResponse {}
func ToCreateUserResponse(account *Account) *CreateUserResponse {}
```

### Token Management
```go
func GenerateTokenPair(userID int64, email string, role string, branchID int64) (*TokenPair, error) {}
func ValidateAccessToken(token string) (*TokenClaims, error) {}
func ValidateRefreshToken(token string) (*RefreshTokenClaims, error) {}
func RefreshAccessToken(refreshToken string) (*TokenPair, error) {}
func RevokeToken(token string) error {}
```

### Password Utilities
```go
func HashPassword(password string) (string, error) {}
func ComparePassword(hashedPassword, password string) error {}
func GenerateResetToken(email string) (string, error) {}
func ValidateResetToken(token string) (*TokenClaims, error) {}
func GenerateVerificationToken(email string) (string, error) {}
```

## Error Handling Functions

### Custom Error Constructors
```go
func NewValidationError(field, message, tag, value string) *ValidationError {}
func NewNotFoundError(domain, resourceType string, identifiers, context map[string]interface{}) *NotFoundError {}
func NewUnauthorizedError(message string, context map[string]interface{}) *UnauthorizedError {}
func NewForbiddenError(message string, context map[string]interface{}) *ForbiddenError {}
func NewConflictError(resource, field, value string, context map[string]interface{}) *ConflictError {}
func NewInternalServerError(message string, err error, context map[string]interface{}) *InternalServerError {}
```

### Error Response Builders
```go
func BuildErrorResponse(err error) *ErrorResponse {}
func BuildValidationErrorResponse(errors []ValidationError) *ValidationErrorResponse {}
func BuildAPIResponse(success bool, message string, data interface{}, err error) *APIResponse {}
```

## Middleware Functions

### Authentication Middleware
```go
func AuthMiddleware(next http.Handler) http.Handler {}
func RequireRole(roles ...Role) func(http.Handler) http.Handler {}
func RequireOwnership() func(http.Handler) http.Handler {}
```

### Logging & Monitoring
```go
func LoggingMiddleware(next http.Handler) http.Handler {}
func RequestIDMiddleware(next http.Handler) http.Handler {}
func RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {}
```

## Configuration & Setup Functions

### Service Setup
```go
func NewAccountService(repo AccountRepository, tokenManager TokenManager, logger Logger) *AccountService {}
func NewAccountHandler(service AccountService, validator Validator, logger Logger) *AccountHandler {}
func NewAccountRepository(db *sql.DB, logger Logger) *AccountRepository {}
```

### Database Setup
```go
func InitializeDatabase(connectionString string) (*sql.DB, error) {}
func RunMigrations(db *sql.DB, migrationPath string) error {}
func CreateTables(db *sql.DB) error {}
func SeedInitialData(db *sql.DB) error {}
```

## Health Check & Monitoring

### Health Check Functions
```go
func HealthCheck() *HealthCheckResponse {}
func DatabaseHealthCheck(db *sql.DB) error {}
func ServiceHealthCheck() *HealthCheck {}
```

### Statistics Functions
```go
func GetUserStatistics() (*UserStatisticsResponse, error) {}
func GetBranchStatistics(branchID int64) (*BranchStats, error) {}
func GetActivityLogs(userID int64, limit int) ([]*ActivityLog, error) {}
```

## Import/Export Functions

### Data Export
```go
func ExportUsers(ctx context.Context, req *ExportRequest) ([]byte, error) {}
func ExportToCSV(users []*Account, fields []string) ([]byte, error) {}
func ExportToJSON(users []*Account, fields []string) ([]byte, error) {}
```

### Data Import
```go
func ImportUsers(ctx context.Context, req *ImportRequest) (*ImportResult, error) {}
func ImportFromCSV(data []byte, options map[string]interface{}) (*ImportResult, error) {}
func ImportFromJSON(data []byte, options map[string]interface{}) (*ImportResult, error) {}
```

## Usage Guidelines

### Basic Usage Flow
1. **Setup**: Initialize service with dependencies
2. **Authentication**: Use Register/Login endpoints
3. **Account Management**: CRUD operations on user accounts
4. **Search**: Use search functions with pagination
5. **Security**: Implement proper token validation and role-based access

### Best Practices
- Always validate input using provided validation functions
- Use proper error handling with custom error types
- Implement pagination for list operations
- Use middleware for cross-cutting concerns
- Log important operations for audit trails
- Implement proper token lifecycle management

### Error Handling
- Use structured error responses
- Provide meaningful error messages
- Include validation details for bad requests
- Log errors appropriately based on severity

### Security Considerations
- Always hash passwords before storage
- Validate tokens on protected endpoints
- Implement proper role-based access control
- Use secure token generation and validation
- Implement rate limiting where appropriate