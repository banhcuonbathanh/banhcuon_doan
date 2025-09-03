# Account Service Layer Structure

This document outlines the organized structure of the account service layer, split into logical components following the pattern from your handler layer.

## File Structure

```
internal/account/
├── account_service_main.go          # Main service structure and constructors
├── account_service_auth.go          # Authentication services (Login, Register, Logout, Token management)
├── account_service_user.go          # User management services (Create, Update, Delete, Status)
├── account_service_search.go        # Search and query services (FindBy*, SearchUsers)
├── account_service_password.go      # Password management services (Change, Forgot, Reset)
├── account_service_email.go         # Email verification services (Verify, Resend)
└── account_service_errors.go        # Common error definitions
```

## Service Methods Distribution

### account_service_main.go
- `ServiceStruct` definition
- `NewAccountService()` - Main constructor
- `NewAccountServiceLegacy()` - Backward compatibility constructor
- `modelToProto()` - Helper conversion method
- `ValidateUserCredentials()` - Business logic helper
- `DeactivateUser()` - Business logic helper
- `GetUsersByBranch()` - Business logic helper
- Interface compliance check

### account_service_auth.go
- `Register()` - User registration
- `Login()` - User authentication
- `Logout()` - User logout
- `RefreshToken()` - Token refresh
- `ValidateToken()` - Token validation

### account_service_user.go
- `CreateUser()` - User creation
- `UpdateUser()` - User updates
- `DeleteUser()` - User deletion
- `UpdateAccountStatus()` - Account status management

### account_service_search.go
- `FindByEmail()` - Find user by email
- `FindByID()` - Find user by ID
- `FindAllUsers()` - Get all users
- `FindByRole()` - Find users by role
- `FindByBranch()` - Find users by branch
- `SearchUsers()` - Advanced search with pagination and filtering

### account_service_password.go
- `ChangePassword()` - Password change
- `ForgotPassword()` - Forgot password request
- `ResetPassword()` - Password reset

### account_service_email.go
- `VerifyEmail()` - Email verification
- `ResendVerification()` - Resend verification email

### account_service_errors.go
- Common error definitions used across all service methods

## Benefits of This Structure

1. **Separation of Concerns**: Each file handles a specific domain of functionality
2. **Maintainability**: Easy to locate and modify specific features
3. **Scalability**: New features can be added to appropriate files
4. **Testing**: Individual components can be tested in isolation
5. **Code Organization**: Similar to your handler layer structure
6. **Team Development**: Multiple developers can work on different files simultaneously

## Usage Example

```go
// All methods are available on the same ServiceStruct instance
service := NewAccountService(userRepo, tokenMaker, passwordHash, emailService)

// Authentication methods
loginRes, err := service.Login(ctx, loginReq)
registerRes, err := service.Register(ctx, registerReq)

// User management methods
user, err := service.CreateUser(ctx, createReq)
updateRes, err := service.UpdateUser(ctx, updateReq)

// Search methods
users, err := service.FindAllUsers(ctx, &emptypb.Empty{})
searchRes, err := service.SearchUsers(ctx, searchReq)

// Password management methods
changeRes, err := service.ChangePassword(ctx, changeReq)
resetRes, err := service.ResetPassword(ctx, resetReq)
```

## Dependencies

Make sure to import the following packages across the service files:

```go
import (
    "context"
    "errors"
    "time"

    "english-ai-full/internal/model"
    "english-ai-full/internal/proto_qr/account"
    logg "english-ai-full/logger"
    "english-ai-full/utils"

    pkgerrors "github.com/pkg/errors"
    "google.golang.org/protobuf/types/known/emptypb"
    "google.golang.org/protobuf/types/known/timestamppb"
)
```

This structure maintains all the functionality from your original single service file while providing better organization and maintainability.