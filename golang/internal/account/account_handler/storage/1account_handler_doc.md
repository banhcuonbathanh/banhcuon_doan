# Account Handler Documentation

## Overview
The Account Handler module provides comprehensive user account management functionality for the English AI application. It includes authentication, user management, email verification, password management, and search capabilities.

## File Structure and Functions

### 1. `account_handler_main.go` - Main Entry Point
**Purpose**: Main handler struct and interface compliance

#### Functions:
- `NewAccountHandler(userClient pb.AccountServiceClient) *AccountHandler`
  - Creates a new account handler instance
  - Initializes base handler with gRPC client

#### Interface Compliance:
✅ **Implemented Methods:**
- Register, Login, Logout
- RefreshToken, ValidateToken
- CreateAccount, FindAccountByID, UpdateUserByID, DeleteUser
- GetUserProfile, VerifyEmail, ResendVerification
- UpdateAccountStatus

❌ **Missing Methods (need implementation):**
- FindByEmail, FindAllUsers, ChangePassword
- ForgotPassword, ResetPassword, FindByRole
- FindByBranch, SearchUsers, GetUsersByBranch

---

### 2. `account_handler_base.go` - Base Handler
**Purpose**: Base functionality and common utilities

#### Functions:
- `NewBaseHandler(userClient pb.AccountServiceClient) *BaseAccountHandler`
  - Creates base handler with validator setup
  - Registers custom validation rules

- `getUserIDFromContext(ctx context.Context) (int64, error)`
  - Extracts user ID from request context
  - Used for authentication-required endpoints

- `getPaginationParams(r *http.Request) (page, pageSize int32, apiErr *errorcustom.APIError)`
  - Parses pagination parameters from query string
  - Validates page and page_size parameters
  - Default: page=1, page_size=10, max=100

---

### 3. `account_handler_auth.go` - Authentication
**Purpose**: User authentication and registration

#### Functions:
- `Login(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/auth/login`
  - Authenticates user with email/password
  - Generates JWT access and refresh tokens
  - Comprehensive error handling and logging
  - Security features: IP tracking, user agent logging

- `Register(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/auth/register`
  - Creates new user account
  - Password validation and hashing
  - Duplicate email detection
  - Enhanced error handling

- `Logout(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/auth/logout`
  - Logs user logout activity
  - Token invalidation (if implemented)

---

### 4. `account_handler_password.go` - Password Management
**Purpose**: Token management and password operations

#### Functions:
- `RefreshToken(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/auth/refresh-token`
  - Refreshes expired access tokens
  - Validates refresh token

- `ValidateToken(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/auth/validate-token`
  - Validates JWT token from Authorization header
  - Returns token validity and expiration

- `ChangePassword(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `PUT /accounts/password/change` (Protected)
  - Changes user password with current password verification
  - Password strength validation

- `ForgotPassword(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/password/forgot`
  - Initiates password reset process
  - Security: Always returns success message

- `ResetPassword(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/password/reset`
  - Resets password using reset token
  - Token validation and expiration check

---

### 5. `account_handler_email.go` - Email Management
**Purpose**: Email verification and email-based operations

#### Functions:
- `VerifyEmail(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/email/verify/{token}`
  - Verifies email address using verification token
  - Handles invalid/expired tokens

- `ResendVerification(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/email/resend-verification`
  - Resends email verification
  - Security: Generic response for non-existent emails

- `FindByEmail(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/search/email/{email}` (Protected)
  - Finds user by email address
  - Email format validation

---

### 6. `account_handler_user.go` - User CRUD Operations
**Purpose**: User creation, updating, and deletion

#### Functions:
- `CreateAccount(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `POST /accounts/` (Protected)
  - Creates new user account (admin function)
  - Full validation and password hashing
  - Duplicate email handling

- `UpdateUserByID(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `PUT /accounts/{id}` (Protected)
  - Updates user information by ID
  - Field validation and not-found handling

- `DeleteUser(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `DELETE /accounts/{id}` (Protected)
  - Soft/hard delete user account
  - Admin-level operation

---

### 7. `account_handler_search.go` - Search and Retrieval
**Purpose**: User search, filtering, and retrieval operations

#### Functions:
- `FindAccountByID(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/{id}` (Protected)
  - Retrieves user by ID
  - Not-found error handling

- `GetUserProfile(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/profile` or `GET /accounts/profile/{id}` (Protected)
  - Gets current user or specific user profile
  - Context-based user ID extraction

- `FindAllUsers(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/` (Protected)
  - Retrieves all users with pagination
  - Manual pagination implementation

- `FindByRole(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/search/role/{role}` (Protected)
  - Finds users by role (admin, user, manager)
  - Role validation

- `FindByBranch(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/branch/{branch_id}` (Protected)
  - Finds users by branch ID
  - Branch ID validation

- `SearchUsers(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/search` (Protected)
  - Advanced search with multiple filters
  - Query parameters: q, role, branch_id, status, sort_by, sort_order
  - Pagination and sorting support

- `GetUsersByBranch(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `GET /accounts/branch/{branch_id}/users` (Protected)
  - Alternative endpoint for branch-based user retrieval
  - Pagination support

---

### 8. `account_handler_account_management.go` - Account Management
**Purpose**: Administrative account management functions

#### Functions:
- `UpdateAccountStatus(w http.ResponseWriter, r *http.Request)`
  - **Endpoint**: `PUT /accounts/manage/{id}/status` (Protected)
  - Updates account status (active, inactive, suspended, pending)
  - Admin-level operation with validation

---

### 9. `account_handler_validator.go` - Custom Validators
**Purpose**: Custom validation functions for form data

#### Functions:
- `ValidatePassword(fl validator.FieldLevel) bool`
  - Custom password validation using utils
  - Integrates with go-playground/validator

- `ValidateRole(fl validator.FieldLevel) bool`
  - Validates role values (admin, user, manager)
  - Whitelist-based validation

- `ValidateEmailUnique(userClient pb.AccountServiceClient) validator.Func`
  - Checks email uniqueness via gRPC
  - Returns validator function for registration

---

### 10. `account_handler_route.go` - Route Registration
**Purpose**: HTTP route configuration and API documentation

#### Functions:
- `RegisterRoutesAccountHandler(r *chi.Mux, accountHandler *AccountHandler)`
  - Registers all account-related routes
  - Groups routes by functionality
  - Applies middleware (auth) where needed

- `NewHandler(userClient pb.AccountServiceClient) *AccountHandler`
  - Alternative factory function
  - Returns configured AccountHandler

#### Route Groups:
1. **Authentication** (`/accounts/auth/`)
   - register, login, logout, refresh-token, validate-token

2. **Password Management** (`/accounts/password/`)
   - forgot, reset, change (protected)

3. **Email Operations** (`/accounts/email/`)
   - verify/{token}, resend-verification

4. **User CRUD** (`/accounts/`)
   - POST, GET, PUT, DELETE (all protected)

5. **Profile Management** (`/accounts/profile/`)
   - GET current/specific user (protected)

6. **Search Operations** (`/accounts/search/`)
   - Advanced search, email lookup, role filtering (protected)

7. **Branch Operations** (`/accounts/branch/`)
   - Branch-based user retrieval (protected)

8. **Account Management** (`/accounts/manage/`)
   - Status updates (protected)

---

### 11. `account_handler_token.go` - Legacy Token Handler
**Purpose**: Legacy/commented token handling functions
**Status**: Currently commented out, functionality moved to `account_handler_password.go`

---

## API Endpoints Summary

### Public Endpoints (No Authentication Required)
| Method | Endpoint | Function | Description |
|--------|----------|----------|-------------|
| POST | `/accounts/auth/register` | Register | User registration |
| POST | `/accounts/auth/login` | Login | User authentication |
| POST | `/accounts/auth/logout` | Logout | User logout |
| POST | `/accounts/auth/refresh-token` | RefreshToken | Token refresh |
| POST | `/accounts/auth/validate-token` | ValidateToken | Token validation |
| POST | `/accounts/password/forgot` | ForgotPassword | Password reset request |
| POST | `/accounts/password/reset` | ResetPassword | Password reset |
| GET | `/accounts/email/verify/{token}` | VerifyEmail | Email verification |
| POST | `/accounts/email/resend-verification` | ResendVerification | Resend verification |

### Protected Endpoints (Authentication Required)
| Method | Endpoint | Function | Description |
|--------|----------|----------|-------------|
| POST | `/accounts/` | CreateAccount | Create user |
| GET | `/accounts/` | FindAllUsers | List all users |
| GET | `/accounts/{id}` | FindAccountByID | Get user by ID |
| PUT | `/accounts/{id}` | UpdateUserByID | Update user |
| DELETE | `/accounts/{id}` | DeleteUser | Delete user |
| GET | `/accounts/profile` | GetUserProfile | Current user profile |
| GET | `/accounts/profile/{id}` | GetUserProfile | User profile by ID |
| PUT | `/accounts/password/change` | ChangePassword | Change password |
| GET | `/accounts/search/email/{email}` | FindByEmail | Find by email |
| GET | `/accounts/search/role/{role}` | FindByRole | Find by role |
| GET | `/accounts/branch/{branch_id}` | FindByBranch | Find by branch |
| GET | `/accounts/branch/{branch_id}/users` | GetUsersByBranch | Users by branch |
| GET | `/accounts/search` | SearchUsers | Advanced search |
| PUT | `/accounts/manage/{id}/status` | UpdateAccountStatus | Update status |

## Query Parameters

### Search and Pagination
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10, max: 100)
- `q`: Search query
- `role`: Filter by role
- `branch_id`: Filter by branch ID
- `status`: Filter by status (comma-separated)
- `sort_by`: Sort field (default: created_at)
- `sort_order`: Sort order (asc/desc, default: desc)

## Error Handling
All handlers implement comprehensive error handling with:
- Custom error types from `errorcustom` package
- Detailed logging with context
- Security-conscious error messages
- HTTP status code mapping
- Validation error details

## Security Features
- JWT token-based authentication
- Password hashing and validation
- IP address and User-Agent logging
- Rate limiting considerations
- Input validation and sanitization
- Generic responses for security-sensitive operations

## Dependencies
- **gRPC Client**: `pb.AccountServiceClient`
- **Validation**: `github.com/go-playground/validator/v10`
- **Routing**: `github.com/go-chi/chi`
- **Utils**: Custom utility functions for JWT, hashing, validation
- **Logging**: Custom logger with structured logging
- **Error Handling**: Custom error types and handlers

