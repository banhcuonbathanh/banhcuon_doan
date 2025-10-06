// golang/internal/account/account_handler/account_handler_route.go

package account_handler

import (
	"net/http"

	// utils_config "english-ai-full/utils/config"

	"github.com/go-chi/chi"
)

func RegisterRoutesAccountHandler(r *chi.Mux, accountHandler *AccountHandler) {
	// Root endpoint (public)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Server is running", "status": "ok"}`))
	})
	
	// Health check endpoint (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "account-service"}`))
	})
	
	r.Route("/accounts", func(r chi.Router) {
		// Add domain-specific middleware for all account routes
		// r.Use(errorcustom.RateLimitMiddleware("account"))
		
		// PUBLIC Authentication routes (NO JWT middleware)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", accountHandler.Register)
			r.Post("/login", accountHandler.Login)
				r.Post("/dailyCleanup", accountHandler.DailyCleanup)
			// r.Post("/logout", accountHandler.Logout)
			// r.Post("/refresh-token", accountHandler.RefreshToken)
			// r.Post("/validate-token", accountHandler.ValidateToken)
		})
		
		// ALL Password management routes (both public and protected)
		r.Route("/password", func(r chi.Router) {
			// PUBLIC password routes (NO JWT middleware)
			// r.Post("/forgot", accountHandler.ForgotPassword)
			// r.Post("/reset", accountHandler.ResetPassword)
			
			// PROTECTED password routes (WITH JWT middleware)
			r.Group(func(r chi.Router) {
				// cfg := utils_config.GetConfig()
				// if cfg != nil && cfg.JWT.SecretKey != "" {
				// 	r.Use(errorcustom.JWTValidationMiddleware(cfg.JWT.SecretKey))
				// }
				// r.Put("/change", accountHandler.ChangePassword)
			})
		})
		
		// PUBLIC Email verification routes (NO JWT middleware)
		r.Route("/email", func(r chi.Router) {
			// r.Get("/verify/{token}", accountHandler.VerifyEmail)
			// r.Post("/resend-verification", accountHandler.ResendVerification)
		})

		// PROTECTED user management routes (WITH JWT middleware)
		r.Group(func(r chi.Router) {
			// Apply JWT middleware only to this group
			// cfg := utils_config.GetConfig()
			// if cfg != nil && cfg.JWT.SecretKey != "" {
			// 	r.Use(errorcustom.JWTValidationMiddleware(cfg.JWT.SecretKey))
			// }
			// Alternative: if you have the auth middleware package working:
			// r.Use(auth.AuthMiddleware)
			
			// CRUD operations (protected)
			// r.Post("/", accountHandler.CreateAccount)
			// r.Get("/", accountHandler.FindAllUsers)
			// r.Get("/{id}", accountHandler.FindAccountByID)
			// r.Put("/{id}", accountHandler.UpdateUserByID)
			// r.Delete("/{id}", accountHandler.DeleteUser)
			
			// Profile management (protected)
			r.Route("/profile", func(r chi.Router) {
				// r.Get("/", accountHandler.GetUserProfile)      // Current user
				// r.Get("/{id}", accountHandler.GetUserProfile)  // Specific user
			})
			
			// Search and filtering (protected)
			r.Route("/search", func(r chi.Router) {
				// r.Get("/", accountHandler.SearchUsers)              // Advanced search
				// r.Get("/email/{email}", accountHandler.FindByEmail) // Find by email
				// r.Get("/role/{role}", accountHandler.FindByRole)    // Find by role
			})
			
			// Branch-related endpoints (protected)
			r.Route("/branch", func(r chi.Router) {
				// r.Get("/{branch_id}", accountHandler.FindByBranch)
				// r.Get("/{branch_id}/users", accountHandler.GetUsersByBranch)
			})
			
			// Account management (protected)
			r.Route("/manage", func(r chi.Router) {
				// r.Put("/{id}/status", accountHandler.UpdateAccountStatus)
			})
		})
	})
}

// Route documentation and usage examples
/*
API Endpoints Overview:

PUBLIC ENDPOINTS (no JWT required):
GET    /                                    - Server status check
GET    /health                              - Health check
POST   /accounts/auth/register              - User registration
POST   /accounts/auth/login                 - User login
POST   /accounts/auth/logout                - User logout
POST   /accounts/auth/refresh-token         - Refresh access token
POST   /accounts/auth/validate-token        - Validate token
POST   /accounts/password/forgot            - Request password reset (PUBLIC)
POST   /accounts/password/reset             - Reset password with token (PUBLIC)
GET    /accounts/email/verify/{token}       - Verify email address
POST   /accounts/email/resend-verification  - Resend verification email

PROTECTED ENDPOINTS (require JWT token):
POST   /accounts/                           - Create new user account
GET    /accounts/                           - Get all users (paginated)
GET    /accounts/{id}                       - Get user by ID
PUT    /accounts/{id}                       - Update user by ID
DELETE /accounts/{id}                       - Delete user by ID

GET    /accounts/profile                    - Get current user profile
GET    /accounts/profile/{id}               - Get user profile by ID
PUT    /accounts/password/change            - Change password (PROTECTED)

GET    /accounts/search/email/{email}       - Find user by email
GET    /accounts/search/role/{role}         - Find users by role
GET    /accounts/branch/{branch_id}         - Find users by branch
GET    /accounts/branch/{branch_id}/users   - Get users by branch (alternative)
GET    /accounts/search                     - Advanced user search with query params
PUT    /accounts/manage/{id}/status         - Update account status

Authentication Header Format for Protected Routes:
Authorization: Bearer <jwt_token>

Query Parameters for Search and Pagination:
- page: Page number (default: 1)
- page_size: Items per page (default: 10, max: 100)
- q: Search query
- role: Filter by role
- branch_id: Filter by branch ID
- status: Filter by status (comma-separated for multiple)
- sort_by: Sort field (default: created_at)
- sort_order: Sort order (asc/desc, default: desc)

Example Usage:
GET /accounts/search?q=john&role=admin&page=1&page_size=20&sort_by=name&sort_order=asc
GET /accounts/?page=2&page_size=50
GET /accounts/search/role/manager?page=1&page_size=10

For protected endpoints, include JWT token:
curl -H "Authorization: Bearer <your_jwt_token>" http://localhost:8080/accounts/profile
*/