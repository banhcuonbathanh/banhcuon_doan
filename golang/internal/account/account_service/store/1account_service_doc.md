# Account Service Documentation

## Overview
The Account Service layer provides the business logic and gRPC service implementation for user account management in the English AI application. It acts as an intermediary between the HTTP handlers and the repository layer, implementing comprehensive user management, authentication, and account operations.

## File Structure and Functions

### 1. `account_service_main.go` - Main Service Structure
**Purpose**: Core service structure, constructors, and business logic helpers

#### Functions:
- `NewAccountService(userRepo, tokenMaker, passwordHash, emailService) *ServiceStruct`
  - Creates a full-featured service instance with all dependencies
  - Includes token management, password hashing, and email capabilities

- `NewAccountServiceLegacy(userRepo) *ServiceStruct`
  - Backward compatibility constructor with minimal dependencies
  - For systems not requiring advanced features

- `modelToProto(user model.Account) *account.Account`
  - Helper method to convert internal model to protobuf format
  - Handles timestamp conversion and field mapping

#### Business Logic Helpers:
- `ValidateUserCredentials(ctx, email, password) (model.Account, error)`
  - Validates user login credentials
  - Supports both new password hasher and legacy utils.Compare

- `DeactivateUser(ctx, userID) error`
  - Deactivates user account by updating status to "inactive"

- `GetUsersByBranch(ctx, branchID) ([]model.Account, error)`
  - Retrieves all users belonging to a specific branch

#### Interface Compliance:
✅ **Implements**: `AccountServiceInterface`

---

### 2. `account_service_auth.go` - Authentication Services
**Purpose**: User authentication, registration, and token management

#### Functions:
- `Register(ctx, req *account.RegisterReq) (*account.RegisterRes, error)`
  - **Service**: User registration with comprehensive validation
  - Email uniqueness checking and password hashing
  - Automatic email verification token generation and sending
  - Comprehensive error handling for duplicate emails

- `Login(ctx, req *account.LoginReq) (*account.AccountRes, error)`
  - **Service**: User authentication with email/password
  - Password verification using configurable hasher or legacy utils
  - JWT token generation if token maker available
  - Returns user profile and access token

- `Logout(ctx, req *account.LogoutReq) (*account.LogoutRes, error)`
  - **Service**: User logout handling
  - Currently placeholder for JWT-based systems
  - Logs logout activity

- `RefreshToken(ctx, req *account.RefreshTokenReq) (*account.RefreshTokenRes, error)`
  - **Service**: JWT token refresh functionality
  - Validates refresh token and generates new access/refresh tokens
  - Comprehensive token validation

- `ValidateToken(ctx, req *account.ValidateTokenReq) (*account.ValidateTokenRes, error)`
  - **Service**: Token validation service
  - Returns token validity status and user information
  - Error handling for invalid/expired tokens

---

### 3. `account_service_user.go` - User Management Services
**Purpose**: User CRUD operations and account management

#### Functions:
- `CreateUser(ctx, req *account.AccountReq) (*account.Account, error)`
  - **Service**: Creates new user account
  - Password hashing and validation
  - Automatic welcome email sending
  - Comprehensive field validation

- `UpdateUser(ctx, req *account.UpdateUserReq) (*account.AccountRes, error)`
  - **Service**: Updates existing user information
  - Custom error handling for not found and duplicate email scenarios
  - Field validation and sanitization

- `DeleteUser(ctx, req *account.DeleteAccountReq) (*account.DeleteAccountRes, error)`
  - **Service**: User deletion with email notification
  - Retrieves user info before deletion for notification
  - Sends account deactivation email
  - Comprehensive error handling

- `UpdateAccountStatus(ctx, req *account.UpdateAccountStatusReq) (*account.UpdateAccountStatusRes, error)`
  - **Service**: Account status management
  - Validates status values and user existence
  - Admin-level operation with proper error handling

---

### 4. `account_service_search.go` - Search and Query Services
**Purpose**: User search, filtering, and retrieval operations

#### Functions:
- `FindByEmail(ctx, req *account.FindByEmailReq) (*account.AccountRes, error)`
  - **Service**: Find user by email address
  - Custom error handling for user not found scenarios
  - Email format validation

- `FindByID(ctx, req *account.FindByIDReq) (*account.FindByIDRes, error)`
  - **Service**: Find user by ID
  - Custom error handling for user not found scenarios
  - ID validation

- `FindAllUsers(ctx, req *emptypb.Empty) (*account.AccountList, error)`
  - **Service**: Retrieve all users
  - Returns complete user list with count
  - Service error handling for repository failures

- `FindByRole(ctx, req *account.FindByRoleReq) (*account.AccountList, error)`
  - **Service**: Find users by role
  - Role validation and filtering
  - Returns filtered user list

- `FindByBranch(ctx, req *account.FindByBranchReq) (*account.AccountList, error)`
  - **Service**: Find users by branch ID
  - Branch validation and filtering
  - Returns branch-specific user list

- `SearchUsers(ctx, req *account.SearchUsersReq) (*account.SearchUsersRes, error)`
  - **Service**: Advanced user search with filtering and pagination
  - Supports query, role, branch, and status filtering
  - Pagination and sorting capabilities
  - Comprehensive pagination metadata

---

### 5. `account_service_password.go` - Password Management Services
**Purpose**: Password operations, reset, and change functionality

#### Functions:
- `ChangePassword(ctx, req *account.ChangePasswordReq) (*account.ChangePasswordRes, error)`
  - **Service**: Password change with current password verification
  - Current password validation before change
  - New password hashing and storage
  - Password change notification email

- `ForgotPassword(ctx, req *account.ForgotPasswordReq) (*account.ForgotPasswordRes, error)`
  - **Service**: Forgot password request handling
  - Security-conscious response (doesn't reveal email existence)
  - Reset token generation and storage
  - Password reset email sending

- `ResetPassword(ctx, req *account.ResetPasswordReq) (*account.ResetPasswordRes, error)`
  - **Service**: Password reset using reset token
  - Token validation with expiration checking
  - New password hashing and storage
  - Comprehensive token error handling

---

### 6. `account_service_email.go` - Email Services
**Purpose**: Email verification and email-based operations

❌ **MISSING FILE** - Need to implement:
- `VerifyEmail(ctx, req *account.VerifyEmailReq) (*account.VerifyEmailRes, error)`
  - Verify email address using verification token
  - Handle invalid/expired tokens
  - Update user email verification status

- `ResendVerification(ctx, req *account.ResendVerificationReq) (*account.ResendVerificationRes, error)`
  - Resend email verification
  - Generate new verification token
  - Security-conscious response for non-existent emails

---

### 7. `account_service_errors.go` - Error Definitions
**Purpose**: Common error definitions and error handling utilities

❌ **MISSING FILE** - Need to implement:
- Service-specific error definitions
- Error mapping utilities
- Common error response patterns
- Error logging and monitoring helpers

---

## Service Methods Implementation Status

### ✅ **Implemented Methods:**
- **Authentication**: Register, Login, Logout, RefreshToken, ValidateToken
- **User Management**: CreateUser, UpdateUser, DeleteUser, UpdateAccountStatus
- **Search Operations**: FindByEmail, FindByID, FindAllUsers, FindByRole, FindByBranch, SearchUsers

### ❌ **Missing Methods (need implementation):**
- **Email Operations**: VerifyEmail, ResendVerification
- **Branch Operations**: GetUsersByBranch (handler exists but service method needs proper implementation)

---

## Comparison with Account Handler Requirements

### Handler Requirements vs Service Implementation:

| Handler Method | Service Implementation | Status |
|----------------|----------------------|---------|
| Register | ✅ Register | Complete |
| Login | ✅ Login | Complete |
| Logout | ✅ Logout | Complete |
| RefreshToken | ✅ RefreshToken | Complete |
| ValidateToken | ✅ ValidateToken | Complete |
| CreateAccount | ✅ CreateUser | Complete |
| FindAccountByID | ✅ FindByID | Complete |
| UpdateUserByID | ✅ UpdateUser | Complete |
| DeleteUser | ✅ DeleteUser | Complete |
| GetUserProfile | ⚠️ FindByID (can be used) | Needs alias/wrapper |
| VerifyEmail | ❌ Missing | **Need to implement** |
| ResendVerification | ❌ Missing | **Need to implement** |
| UpdateAccountStatus | ✅ UpdateAccountStatus | Complete |
| FindByEmail | ✅ FindByEmail | Complete |
| FindAllUsers | ✅ FindAllUsers | Complete |
| ChangePassword | ✅ ChangePassword | Complete |
| ForgotPassword | ✅ ForgotPassword | Complete |
| ResetPassword | ✅ ResetPassword | Complete |
| FindByRole | ✅ FindByRole | Complete |
| FindByBranch | ✅ FindByBranch | Complete |
| SearchUsers | ✅ SearchUsers | Complete |
| GetUsersByBranch | ⚠️ FindByBranch (can be used) | Needs alias/wrapper |

---

## Dependencies and Configuration

### Service Dependencies:
- **Repository**: `AccountRepositoryInterface` - Data access layer
- **Token Maker**: `TokenMakerInterface` - JWT token operations (optional)
- **Password Hasher**: `PasswordHasherInterface` - Password hashing (optional)
- **Email Service**: `EmailServiceInterface` - Email operations (optional)
- **Logger**: Custom logger for structured logging

### Optional Dependencies:
The service supports graceful degradation when optional dependencies are not available:
- Without TokenMaker: No JWT token generation/validation
- Without PasswordHasher: Falls back to utils.Compare for passwords
- Without EmailService: No email notifications sent

---

## Error Handling Strategy

### Current Error Handling:
- **Custom Errors**: Uses `errorcustom` package for structured error responses
- **Error Wrapping**: Uses `pkgerrors.WithStack()` for error context
- **String Matching**: Detects specific error types using string matching
- **Service Errors**: Creates service-specific errors with context

### Error Types Handled:
- User not found errors (by ID and email)
- Duplicate email errors
- Invalid token errors
- Password mismatch errors
- Repository errors
- Service errors with retry capability

---

## Security Features

### Authentication Security:
- Password hashing with configurable hasher
- JWT token-based authentication
- Refresh token validation
- Token expiration handling

### Password Security:
- Current password verification for changes
- Reset token generation and validation
- Password strength validation (via utils)
- Security-conscious responses for sensitive operations

### Email Security:
- Generic responses for non-existent emails
- Verification token generation
- Automated email notifications

---

## Performance Considerations

### Async Operations:
- Email sending in goroutines to avoid blocking
- Background email notifications
- Non-blocking email failures with logging

### Pagination Support:
- Advanced search with pagination
- Configurable page sizes
- Efficient result counting

---

## Missing Implementation Tasks

### 1. Email Service Methods (High Priority)
```go
// Need to implement in account_service_email.go
func (s *ServiceStruct) VerifyEmail(ctx context.Context, req *account.VerifyEmailReq) (*account.VerifyEmailRes, error)
func (s *ServiceStruct) ResendVerification(ctx context.Context, req *account.ResendVerificationReq) (*account.ResendVerificationRes, error)
```

### 2. Service Method Aliases (Medium Priority)
```go
// Add convenience methods or aliases
func (s *ServiceStruct) GetUserProfile(ctx context.Context, req *account.GetUserProfileReq) (*account.AccountRes, error)
func (s *ServiceStruct) GetUsersByBranch(ctx context.Context, req *account.GetUsersByBranchReq) (*account.AccountList, error)
```

### 3. Error Definitions File (Medium Priority)
- Create `account_service_errors.go`
- Centralize error definitions
- Add error mapping utilities

### 4. Service Configuration (Low Priority)
- Add service configuration struct
- Configurable timeouts and limits
- Feature flags for optional functionality

---

## Usage Examples

### Basic Service Usage:
```go
// Create service with full dependencies
service := NewAccountService(userRepo, tokenMaker, passwordHash, emailService)

// Authentication
loginRes, err := service.Login(ctx, &account.LoginReq{
    Email: "user@example.com",
    Password: "password123",
})

// User management
user, err := service.CreateUser(ctx, &account.AccountReq{
    Name: "John Doe",
    Email: "john@example.com",
    Role: "user",
})

// Search operations
users, err := service.SearchUsers(ctx, &account.SearchUsersReq{
    Query: "john",
    Pagination: &account.PaginationReq{Page: 1, PageSize: 10},
})
```

### Legacy Service Usage:
```go
// Create service with minimal dependencies
service := NewAccountServiceLegacy(userRepo)

// Still supports basic operations
users, err := service.FindAllUsers(ctx, &emptypb.Empty{})
```

This service layer provides a robust foundation for account management with comprehensive error handling, security features, and optional dependency support for graceful degradation.