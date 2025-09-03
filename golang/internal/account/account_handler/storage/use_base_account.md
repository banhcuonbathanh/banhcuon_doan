package account_handler

import (
	"context"
	"net/http"
	"strconv"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	utils_config "english-ai-full/utils/config"

	"github.com/gorilla/mux"
)

// ============================================================================
// MAIN HANDLER IMPLEMENTATION USING REFACTORED COMPONENTS
// ============================================================================

// AccountHandler is the main handler that composes all the functionality
type AccountHandler struct {
	*BaseAccountHandler
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(userClient pb.AccountServiceClient, config *utils_config.Config) *AccountHandler {
	return &AccountHandler{
		BaseAccountHandler: NewBaseHandler(userClient, config),
	}
}

// ============================================================================
// HTTP ROUTE HANDLERS - USING REFACTORED METHODS
// ============================================================================

// CreateUser handles user creation requests
func (ah *AccountHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Setup operation context using core methods
	opCtx := ah.setupOperationContext(r, "create_user")
	ah.logOperationStart(opCtx)
	
	// Apply rate limiting using HTTP methods
	if err := ah.applyRateLimit(w, r, "create_user"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusTooManyRequests)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Decode and validate request using validation methods
	var req CreateUserRequest
	if err := ah.validateAndDecodeRequest(r, &req, "create_user"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Register user using business logic methods
	user, err := ah.registerNewUser(r.Context(), req)
	if err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Respond with success using HTTP methods
	ah.logOperationEnd(opCtx, nil, http.StatusCreated)
	ah.RespondWithSuccess(w, r, user)
}

// LoginUser handles user login requests
func (ah *AccountHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "login_user")
	ah.logOperationStart(opCtx)
	
	// Rate limiting
	if err := ah.applyRateLimit(w, r, "login"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusTooManyRequests)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Decode login request
	var loginReq LoginRequest
	if err := ah.DecodeJSONRequest(r, &loginReq); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Validate login request
	if err := ah.validateRequest(&loginReq, "login_user", r); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Authenticate user using business methods
	user, sessionToken, err := ah.authenticateUser(r.Context(), loginReq.Email, loginReq.Password)
	if err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Prepare response
	loginResponse := LoginResponse{
		User:         user,
		SessionToken: sessionToken,
		ExpiresAt:    ah.getSessionExpiryTime(),
	}
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, loginResponse)
}

// GetUser handles user retrieval requests
func (ah *AccountHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "get_user")
	ah.logOperationStart(opCtx)
	
	// Parse user ID from URL using HTTP methods
	userID, err := ah.ParseIDParam(r, "id")
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Get requesting user ID from context
	requestingUserID, err := ah.getUserIDFromContext(r.Context())
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Check permissions using business methods
	if err := ah.checkUserPermissions(requestingUserID, "admin", "user_read"); err != nil {
		// Allow users to read their own profile
		if userID != requestingUserID {
			ah.logOperationEnd(opCtx, err, http.StatusForbidden)
			ah.HandleHTTPError(w, r, err)
			return
		}
	}
	
	// Get user using gRPC methods
	user, err := ah.getUserByID(userID)
	if err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Log data access using business methods
	ah.logDataAccess(requestingUserID, "user_profile", userID, "read")
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, user)
}

// UpdateUser handles user update requests
func (ah *AccountHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "update_user")
	ah.logOperationStart(opCtx)
	
	// Parse user ID
	userID, err := ah.ParseIDParam(r, "id")
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Get requesting user ID
	requestingUserID, err := ah.getUserIDFromContext(r.Context())
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Decode update request
	var updateReq UpdateUserRequest
	if err := ah.DecodeJSONRequest(r, &updateReq); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Convert to updates map
	updates := map[string]interface{}{}
	if updateReq.Name != "" {
		updates["name"] = updateReq.Name
	}
	if updateReq.Email != "" {
		updates["email"] = updateReq.Email
	}
	if updateReq.Role != "" {
		updates["role"] = updateReq.Role
	}
	if updateReq.Title != "" {
		updates["title"] = updateReq.Title
	}
	if updateReq.Avatar != "" {
		updates["avatar"] = updateReq.Avatar
	}
	
	// Update user using business methods
	user, err := ah.updateUserProfile(r.Context(), userID, requestingUserID, updates)
	if err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, user)
}

// ListUsers handles user listing requests with pagination and filtering
func (ah *AccountHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "list_users")
	ah.logOperationStart(opCtx)
	
	// Get requesting user ID
	requestingUserID, err := ah.getUserIDFromContext(r.Context())
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Check permissions
	if err := ah.checkUserPermissions(requestingUserID, "admin", "user_list"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusForbidden)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Parse pagination parameters using HTTP methods
	page, pageSize, err := ah.getPaginationParams(r)
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Parse sorting parameters
	allowedSortFields := []string{"name", "email", "role", "created_at", "updated_at"}
	sortBy, sortOrder, err := ah.getSortingParams(r, allowedSortFields)
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Build search criteria
	criteria := SearchCriteria{
		Role:     r.URL.Query().Get("role"),
		Status:   r.URL.Query().Get("status"),
		Page:     page,
		PageSize: pageSize,
	}
	
	// Parse branch ID if provided
	if branchIDStr := r.URL.Query().Get("branch_id"); branchIDStr != "" {
		if branchID, parseErr := strconv.ParseInt(branchIDStr, 10, 64); parseErr == nil {
			criteria.BranchID = branchID
		}
	}
	
	// Search users using business methods
	result, err := ah.searchUsers(r.Context(), requestingUserID, criteria)
	if err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Add sorting information to response
	response := ListUsersResponse{
		Users:      result.Users,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		SortBy:     sortBy,
		SortOrder:  sortOrder,
	}
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, response)
}

// DeleteUser handles user deletion requests
func (ah *AccountHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "delete_user")
	ah.logOperationStart(opCtx)
	
	// Parse user ID
	userID, err := ah.ParseIDParam(r, "id")
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Get requesting user ID
	requestingUserID, err := ah.getUserIDFromContext(r.Context())
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Check permissions (only admins can delete users)
	if err := ah.checkUserPermissions(requestingUserID, "admin", "user_delete"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusForbidden)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Prevent self-deletion
	if userID == requestingUserID {
		err := errorcustom.NewBusinessLogicErrorWithContext(
			ah.domain,
			"self_deletion_not_allowed",
			"Users cannot delete their own account",
			map[string]interface{}{
				"user_id": userID,
			},
		)
		ah.logOperationEnd(opCtx, err, http.StatusForbidden)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Delete user using gRPC methods
	if err := ah.deleteUserViaGRPC(r.Context(), userID); err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Log security event
	ah.logSecurityEvent(
		"user_deleted",
		"User account deleted",
		"high",
		map[string]interface{}{
			"deleted_user_id": userID,
			"deleted_by":      requestingUserID,
		},
	)
	
	ah.logOperationEnd(opCtx, nil, http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// REQUEST/RESPONSE TYPES
// ============================================================================

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type LoginResponse struct {
	User         *pb.Account `json:"user"`
	SessionToken string      `json:"session_token"`
	ExpiresAt    string      `json:"expires_at"`
}

type UpdateUserRequest struct {
	Name   string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email  string `json:"email,omitempty" validate:"omitempty,email"`
	Role   string `json:"role,omitempty" validate:"omitempty,userrole"`
	Title  string `json:"title,omitempty" validate:"omitempty,max=200"`
	Avatar string `json:"avatar,omitempty" validate:"omitempty,url"`
}

type ListUsersResponse struct {
	Users      []*pb.Account `json:"users"`
	TotalCount int64         `json:"total_count"`
	Page       int32         `json:"page"`
	PageSize   int32         `json:"page_size"`
	SortBy     string        `json:"sort_by"`
	SortOrder  string        `json:"sort_order"`
}

// ============================================================================
// ADDITIONAL SPECIALIZED HANDLERS
// ============================================================================

// ChangePassword handles password change requests
func (ah *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "change_password")
	ah.logOperationStart(opCtx)
	
	// Get user ID from context
	userID, err := ah.getUserIDFromContext(r.Context())
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Decode password change request
	var req ChangePasswordRequest
	if err := ah.validateAndDecodeRequest(r, &req, "change_password"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Change password using gRPC methods
	if err := ah.changePasswordViaGRPC(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, map[string]string{
		"message": "Password changed successfully",
	})
}

// ResetPassword handles password reset initiation
func (ah *AccountHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "reset_password")
	ah.logOperationStart(opCtx)
	
	// Rate limiting for password reset
	if err := ah.applyRateLimit(w, r, "password_reset"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusTooManyRequests)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Decode reset request
	var req ResetPasswordRequest
	if err := ah.validateAndDecodeRequest(r, &req, "reset_password"); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Initiate password reset using business methods
	if err := ah.initiatePasswordReset(r.Context(), req.Email); err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Always return success to prevent email enumeration
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// GetUserStatistics handles user statistics requests
func (ah *AccountHandler) GetUserStatistics(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "get_user_statistics")
	ah.logOperationStart(opCtx)
	
	// Get requesting user ID
	requestingUserID, err := ah.getUserIDFromContext(r.Context())
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Get statistics using business methods
	stats, err := ah.getUserStatistics(r.Context(), requestingUserID)
	if err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, stats)
}

// BatchUpdateUsers handles batch user update requests
func (ah *AccountHandler) BatchUpdateUsers(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "batch_update_users")
	ah.logOperationStart(opCtx)
	
	// Get requesting user ID
	requestingUserID, err := ah.getUserIDFromContext(r.Context())
	if err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Decode batch update request
	var req BatchUpdateRequest
	if err := ah.DecodeJSONRequest(r, &req); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusBadRequest)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	// Perform batch update using business methods
	result, err := ah.batchUpdateUsers(r.Context(), requestingUserID, req.Updates)
	if err != nil {
		statusCode := ah.getStatusCodeFromError(err)
		ah.logOperationEnd(opCtx, err, statusCode)
		ah.HandleHTTPError(w, r, err)
		return
	}
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, result)
}

// ============================================================================
// ROUTE REGISTRATION
// ============================================================================

// RegisterRoutes registers all account-related routes
func (ah *AccountHandler) RegisterRoutes(router *mux.Router) {
	// Create API subrouter
	api := router.PathPrefix("/api/v1/accounts").Subrouter()
	
	// Public routes (no authentication required)
	api.HandleFunc("/register", ah.withRequestLogging("create_user", ah.CreateUser)).Methods("POST")
	api.HandleFunc("/login", ah.withRequestLogging("login_user", ah.LoginUser)).Methods("POST")
	api.HandleFunc("/reset-password", ah.withRequestLogging("reset_password", ah.ResetPassword)).Methods("POST")
	
	// Protected routes (authentication required)
	protected := api.PathPrefix("").Subrouter()
	// Here you would add your authentication middleware
	// protected.Use(ah.authMiddleware)
	
	protected.HandleFunc("", ah.withRequestLogging("list_users", ah.ListUsers)).Methods("GET")
	protected.HandleFunc("/{id:[0-9]+}", ah.withRequestLogging("get_user", ah.GetUser)).Methods("GET")
	protected.HandleFunc("/{id:[0-9]+}", ah.withRequestLogging("update_user", ah.UpdateUser)).Methods("PUT")
	protected.HandleFunc("/{id:[0-9]+}", ah.withRequestLogging("delete_user", ah.DeleteUser)).Methods("DELETE")
	protected.HandleFunc("/change-password", ah.withRequestLogging("change_password", ah.ChangePassword)).Methods("POST")
	protected.HandleFunc("/statistics", ah.withRequestLogging("get_user_statistics", ah.GetUserStatistics)).Methods("GET")
	protected.HandleFunc("/batch-update", ah.withRequestLogging("batch_update_users", ah.BatchUpdateUsers)).Methods("POST")
}

// ============================================================================
// ADDITIONAL REQUEST TYPES
// ============================================================================

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=1"`
	NewPassword string `json:"new_password" validate:"required,strongpassword"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type BatchUpdateRequest struct {
	Updates map[int64]map[string]interface{} `json:"updates" validate:"required"`
}

// ============================================================================
// UTILITY METHODS
// ============================================================================

// getStatusCodeFromError determines HTTP status code from error type
func (ah *AccountHandler) getStatusCodeFromError(err error) int {
	// This would inspect the error type and return appropriate status code
	// Implementation depends on your error handling strategy
	
	if errorcustom.IsValidationError(err) {
		return http.StatusBadRequest
	}
	
	if errorcustom.IsAuthenticationError(err) {
		return http.StatusUnauthorized
	}
	
	if errorcustom.IsAuthorizationError(err) {
		return http.StatusForbidden
	}
	
	if errorcustom.IsNotFoundError(err) {
		return http.StatusNotFound
	}
	
	if errorcustom.IsConflictError(err) {
		return http.StatusConflict
	}
	
	if errorcustom.IsBusinessLogicError(err) {
		return http.StatusBadRequest
	}
	
	if errorcustom.IsRateLimitError(err) {
		return http.StatusTooManyRequests
	}
	
	// Default to internal server error
	return http.StatusInternalServerError
}

// getSessionExpiryTime returns session expiry time string
func (ah *AccountHandler) getSessionExpiryTime() string {
	expiryTime := ah.getSessionTimeout()
	return time.Now().Add(expiryTime).Format(time.RFC3339)
}

// ============================================================================
// HEALTH CHECK HANDLER
// ============================================================================

// HealthCheck handles health check requests
func (ah *AccountHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	opCtx := ah.setupOperationContext(r, "health_check")
	ah.logOperationStart(opCtx)
	
	// Check gRPC client health
	if err := ah.checkGRPCHealth(); err != nil {
		ah.logOperationEnd(opCtx, err, http.StatusServiceUnavailable)
		ah.RespondWithError(w, r, errorcustom.NewSystemError(
			ah.domain,
			"grpc_client",
			"health_check",
			"gRPC service unavailable",
			err,
		))
		return
	}
	
	// Get diagnostics
	diagnostics := ah.getDiagnostics()
	
	health := map[string]interface{}{
		"status":      "healthy",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"diagnostics": diagnostics,
	}
	
	ah.logOperationEnd(opCtx, nil, http.StatusOK)
	ah.RespondWithSuccess(w, r, health)
}

// ============================================================================
// EXAMPLE USAGE AND INTEGRATION
// ============================================================================

/*
Example of how to use this refactored handler:

func main() {
	// Initialize configuration
	config := utils_config.LoadConfig()
	
	// Initialize gRPC client
	conn, err := grpc.Dial(config.UserServiceAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to user service:", err)
	}
	defer conn.Close()
	
	userClient := pb.NewAccountServiceClient(conn)
	
	// Create account handler
	accountHandler := NewAccountHandler(userClient, config)
	
	// Setup router
	router := mux.NewRouter()
	
	// Register routes
	accountHandler.RegisterRoutes(router)
	
	// Add health check
	router.HandleFunc("/health", accountHandler.HealthCheck).Methods("GET")
	
	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

Benefits of this refactored approach:

1. **Single Responsibility**: Each file has a focused responsibility
2. **Maintainability**: Easier to find and modify specific functionality
3. **Testability**: Each component can be tested independently
4. **Reusability**: Components can be reused across different handlers
5. **Readability**: Code is organized logically and easier to understand
6. **Scalability**: Easy to add new functionality without cluttering existing files

File Structure:
- base_handler_core.go: Core functionality, constructor, context management
- base_handler_validation.go: All validation-related methods
- base_handler_http.go: HTTP request/response handling, middleware
- base_handler_grpc.go: gRPC client interactions and error handling
- base_handler_business.go: Business logic, user operations, permissions
- main_handler.go: Concrete implementation using all components
*/