# Error Handling System Migration Guide

## New File Structure

```
golang/internal/error_custom/
├── error_custom_core.go                 # Core types and interfaces (keep existing)
├── error_custom_code.go                 # Error codes (keep existing)
├── error_custom_constructor.go          # Basic constructors (keep existing)
├── error_custom_utilities.go            # Utilities (keep existing)
├── error_factory.go                     # NEW: Central factory
├── unified_handler.go                   # NEW: Single interface for all errors
├── middleware.go                        # NEW: Enhanced middleware
├── constants.go                         # NEW: Organized constants
├── domain/
│   ├── user_errors.go                   # NEW: User domain errors
│   ├── auth_errors.go                   # NEW: Auth domain errors
│   ├── branch_errors.go                 # NEW: Branch domain errors
│   ├── admin_errors.go                  # NEW: Admin domain errors
│   └── account_errors.go                # NEW: Account domain errors
└── layer/
    ├── handler_errors.go                # NEW: HTTP layer errors
    ├── service_errors.go                # NEW: Business logic layer errors
    └── repository_errors.go             # NEW: Data access layer errors
```

## Migration Steps

### Step 6: Update Middleware Setup

**Before:**
```go
func setupMiddleware(r chi.Router) {
    r.Use(middleware.RequestID)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
}
```

**After:**
```go
func setupMiddleware(r chi.Router) {
    errorMiddleware := errorcustom.NewErrorMiddleware()
    
    r.Use(errorMiddleware.RequestIDMiddleware)
    r.Use(errorMiddleware.AutoDomainMiddleware) // Automatic domain detection
    r.Use(errorMiddleware.LoggingMiddleware)    // Enhanced logging
    r.Use(errorMiddleware.RecoveryMiddleware)   // Domain-aware recovery
}
```

### Step 7: Update Route Setup

**Before:**
```go
func setupRoutes(r chi.Router) {
    userHandler := &UserHandler{}
    
    r.Route("/api/users", func(r chi.Router) {
        r.Get("/{user_id}", userHandler.GetUser)
        r.Post("/", userHandler.CreateUser)
    })
}
```

**After:**
```go
func setupRoutes(r chi.Router) {
    errorMiddleware := errorcustom.NewErrorMiddleware()
    userHandler := &UserHandler{}
    
    r.Route("/api/users", func(r chi.Router) {
        r.Use(errorMiddleware.DomainMiddleware(errorcustom.DomainUser))
        r.Get("/{user_id}", userHandler.GetUser)
        r.Post("/", userHandler.CreateUser)
    })
    
    r.Route("/api/auth", func(r chi.Router) {
        r.Use(errorMiddleware.DomainMiddleware(errorcustom.DomainAuth))
        // Auth routes...
    })
    
    r.Route("/api/branches", func(r chi.Router) {
        r.Use(errorMiddleware.DomainMiddleware(errorcustom.DomainBranch))
        // Branch routes...
    })
}
```

## Benefits of New System

### 1. **Consistency**
- All errors follow the same format
- Domain-aware error codes
- Standardized logging

### 2. **Maintainability**
- Centralized error definitions
- Layer-specific error handling
- Easy to add new domains

### 3. **Developer Experience**
- Single interface for all error operations
- Automatic parameter parsing with validation
- Built-in business rule validation

### 4. **Observability**
- Request ID tracking
- Domain-aware logging
- Layer information in errors

### 5. **Extensibility**
- Easy to add new error types
- Domain-specific error managers
- Layer-specific handling

## Common Migration Patterns

### Pattern 1: Convert Manual Parameter Parsing

**Before:**
```go
idStr := chi.URLParam(r, "user_id")
userID, err := strconv.ParseInt(idStr, 10, 64)
if err != nil || userID <= 0 {
    http.Error(w, "Invalid user ID", http.StatusBadRequest)
    return
}
```

**After:**
```go
userID, err := errorHandler.ParseIDParam(r, "user_id")
if err != nil {
    errorHandler.HandleHTTPError(w, r, err)
    return
}
```

### Pattern 2: Convert Validation Logic

**Before:**
```go
if user.Email == "" {
    return errorcustom.NewValidationError("user", "email", "Email is required", nil)
}

if !isValidEmail(user.Email) {
    return errorcustom.NewValidationError("user", "email", "Invalid email format", user.Email)
}

exists, _ := repo.EmailExists(user.Email)
if exists {
    return errorcustom.NewDuplicateError("user", "user", "email", user.Email)
}
```

**After:**
```go
validations := map[string]func() error{
    "email_required": func() error {
        if user.Email == "" {
            return errorcustom.NewValidationError(errorcustom.DomainUser, "email", "Email is required", nil)
        }
        return nil
    },
    "email_format": func() error {
        if !isValidEmail(user.Email) {
            return errorcustom.NewValidationError(errorcustom.DomainUser, "email", "Invalid email format", user.Email)
        }
        return nil
    },
    "email_unique": func() error {
        exists, err := repo.EmailExists(user.Email)
        if err != nil {
            return err
        }
        if exists {
            return errorHandler.NewDuplicateEmailError(user.Email)
        }
        return nil
    },
}

return errorHandler.ValidateBusinessRules(errorcustom.DomainUser, validations)
```

### Pattern 3: Convert Error Responses

**Before:**
```go
if err != nil {
    if isNotFoundError(err) {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    logger.Error("Internal error", map[string]interface{}{"error": err.Error()})
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
}
```

**After:**
```go
if err != nil {
    errorHandler.HandleHTTPError(w, r, err)
    return
}
```

## Testing the New System

### Unit Test Example

```go
func TestUserHandler_CreateUser(t *testing.T) {
    // Setup
    errorHandler := errorcustom.NewUnifiedErrorHandler()
    handler := &UserHandler{errorHandler: errorHandler}
    
    tests := []struct {
        name           string
        requestBody    string
        expectedStatus int
        expectedError  string
    }{
        {
            name:           "missing email",
            requestBody:    `{"name":"John","password":"password123"}`,
            expectedStatus: http.StatusBadRequest,
            expectedError:  "user_VALIDATION_ERROR",
        },
        {
            name:           "weak password",
            requestBody:    `{"email":"john@test.com","name":"John","password":"123"}`,
            expectedStatus: http.StatusBadRequest,
            expectedError:  "user_VALIDATION_ERROR",
        },
        {
            name:           "success",
            requestBody:    `{"email":"john@test.com","name":"John","password":"password123"}`,
            expectedStatus: http.StatusCreated,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/api/users", strings.NewReader(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            // Add domain context
            ctx := context.WithValue(req.Context(), "domain", errorcustom.DomainUser)
            req = req.WithContext(ctx)
            
            rr := httptest.NewRecorder()
            handler.CreateUser(rr, req)
            
            assert.Equal(t, tt.expectedStatus, rr.Code)
            
            if tt.expectedError != "" {
                var response errorcustom.ErrorResponse
                json.Unmarshal(rr.Body.Bytes(), &response)
                assert.Equal(t, tt.expectedError, response.Code)
            }
        })
    }
}
```

## Performance Considerations

### 1. **Error Factory Initialization**
Initialize the error factory once at application startup:

```go
var (
    errorHandler *errorcustom.UnifiedErrorHandler
    once         sync.Once
)

func getErrorHandler() *errorcustom.UnifiedErrorHandler {
    once.Do(func() {
        errorHandler = errorcustom.NewUnifiedErrorHandler()
    })
    return errorHandler
}
```

### 2. **Context Passing**
Use context efficiently for request-scoped data:

```go
// Add to context once
ctx = context.WithValue(ctx, "request_id", requestID)
ctx = context.WithValue(ctx, "domain", domain)

// Retrieve when needed
requestID := errorcustom.GetRequestIDFromContext(ctx)
```

### 3. **Logging Optimization**
Log errors at appropriate levels to avoid noise:

```go
// The system automatically determines logging levels:
// - Server errors (5xx) -> ERROR level
// - Client errors (4xx) -> INFO level (not logged by default)
// - External service errors -> WARNING level
```

## Rollback Plan

If you need to rollback:

1. **Keep old files**: Don't delete existing error files during migration
2. **Gradual migration**: Migrate one domain at a time
3. **Feature flags**: Use feature flags to toggle between old and new systems
4. **Fallback handlers**: Keep fallback error handling in critical paths

```go
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // New system
    if useNewErrorSystem {
        userID, err := errorHandler.ParseIDParam(r, "user_id")
        if err != nil {
            errorHandler.HandleHTTPError(w, r, err)
            return
        }
    } else {
        // Old system fallback
        idStr := chi.URLParam(r, "user_id")
        userID, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
            return
        }
    }
}
```

## Conclusion

This new error handling system provides:

- **Domain-aware errors** for better organization
- **Layer-specific handling** for proper separation of concerns  
- **Unified interface** for consistent usage across your application
- **Enhanced observability** with request tracking and domain context
- **Better developer experience** with automatic parameter parsing and validation

The migration can be done gradually, starting with one domain and expanding to others as you become comfortable with the new patterns.1: Update Imports

**Before:**
```go
import (
    errorcustom "english-ai-full/internal/error_custom"
)
```

**After:**
```go
import (
    errorcustom "english-ai-full/internal/error_custom"
    "english-ai-full/internal/error_custom/domain"
    "english-ai-full/internal/error_custom/layer"
)
```

### Step 2: Initialize Error Handlers

**Before:**
```go
// Scattered error handling throughout code
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "user_id")
    if idStr == "" {
        http.Error(w, "Missing user ID", http.StatusBadRequest)
        return
    }
    
    userID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    // ... rest of handler
}
```

**After:**
```go
// Initialize once at application startup
var errorHandler = errorcustom.NewUnifiedErrorHandler()

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID, err := errorHandler.ParseIDParam(r, "user_id")
    if err != nil {
        errorHandler.HandleHTTPError(w, r, err)
        return
    }
    // ... rest of handler
}
```

### Step 3: Update Handler Layer

**Before:**
```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Manual validation
    if user.Email == "" {
        http.Error(w, "Email is required", http.StatusBadRequest)
        return
    }
    
    if len(user.Password) < 8 {
        http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
        return
    }
    
    createdUser, err := h.userService.CreateUser(r.Context(), &user)
    if err != nil {
        // Generic error handling
        logger.Error("Failed to create user", map[string]interface{}{"error": err.Error()})
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(createdUser)
}
```

**After:**
```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        // Automatic JSON decode error handling with detailed context
        errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    // Validation is now handled in service layer with business rules
    createdUser, err := h.userService.CreateUser(r.Context(), &user)
    if err != nil {
        // Domain-aware error handling with automatic logging
        errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(createdUser)
}
```

### Step 4: Update Service Layer

**Before:**
```go
func (s *UserService) CreateUser(ctx context.Context, user *User) (*User, error) {
    // Manual validation
    if user.Email == "" {
        return nil, errors.New("email is required")
    }
    
    // Check if email exists
    exists, err := s.userRepo.EmailExists(user.Email)
    if err != nil {
        logger.Error("Failed to check email existence", map[string]interface{}{"error": err.Error()})
        return nil, errors.New("internal error")
    }
    if exists {
        return nil, errors.New("email already exists")
    }
    
    // Create user
    createdUser, err := s.userRepo.Create(user)
    if err != nil {
        logger.Error("Failed to create user", map[string]interface{}{"error": err.Error()})
        return nil, errors.New("failed to create user")
    }
    
    return createdUser, nil
}
```

**After:**
```go
func (s *UserService) CreateUser(ctx context.Context, user *User) (*User, error) {
    // Business rule validation with automatic error collection
    validations := map[string]func() error{
        "email_required": func() error {
            if user.Email == "" {
                return errorcustom.NewValidationError(errorcustom.DomainUser, "email", "Email is required", nil)
            }
            return nil
        },
        "email_unique": func() error {
            exists, err := s.userRepo.EmailExists(user.Email)
            if err != nil {
                return err // Will be wrapped by service error handler
            }
            if exists {
                return errorHandler.NewDuplicateEmailError(user.Email)
            }
            return nil
        },
        "password_strength": func() error {
            if len(user.Password) < 8 {
                return errorHandler.NewWeakPasswordError([]string{"at least 8 characters"})
            }
            return nil
        },
    }
    
    if err := errorHandler.ValidateBusinessRules(errorcustom.DomainUser, validations); err != nil {
        return nil, err
    }
    
    // Create user with automatic error wrapping
    createdUser, err := s.userRepo.Create(user)
    if err != nil {
        return nil, errorHandler.WrapRepositoryError(err, errorcustom.DomainUser, "create_user", map[string]interface{}{
            "email": user.Email,
        })
    }
    
    return createdUser, nil
}
```

### Step 5: Update Repository Layer

**Before:**
```go
func (r *UserRepository) GetByID(userID int64) (*User, error) {
    user := &User{}
    err := r.db.QueryRow("SELECT id, email, name FROM users WHERE id = ?", userID).
        Scan(&user.ID, &user.Email, &user.Name)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("user not found")
        }
        logger.Error("Database error", map[string]interface{}{"error": err.Error()})
        return nil, errors.New("database error")
    }
    
    return user, nil
}
```

**After:**
```go
func (r *UserRepository) GetByID(userID int64) (*User, error) {
    user := &User{}
    err := r.db.QueryRow("SELECT id, email, name FROM users WHERE id = ?", userID).
        Scan(&user.ID, &user.Email, &user.Name)
    
    if err != nil {
        // Automatic database error handling with context
        return nil, errorHandler.HandleDatabaseError(err, errorcustom.DomainUser, "users", "get_by_id", map[string]interface{}{
            "user_id": userID,
        })
    }
    
    return user, nil
}
```

### Step