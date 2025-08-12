// golang/internal/account/account_handler/account_handler_route.go

package account_handler

import (
	"net/http"
	error_custom "english-ai-full/internal/error_custom"
	// pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/pkg/middleware/auth"

	"github.com/go-chi/chi"
)
// Update your existing RegisterRoutesAccountHandler function

func RegisterRoutesAccountHandler(r *chi.Mux, accountHandler *AccountHandler) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is running"))
	})
	
	r.Route("/accounts", func(r chi.Router) {
		// Add domain-specific middleware
		r.Use(error_custom.DomainMiddleware("account")) // or "account" 
		r.Use(error_custom.RateLimitMiddleware("account"))
		
		// Authentication routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", accountHandler.Register)
			r.Post("/login", accountHandler.Login)
			r.Post("/logout", accountHandler.Logout)
			r.Post("/refresh-token", accountHandler.RefreshToken)
			r.Post("/validate-token", accountHandler.ValidateToken)
		})
		
		// Password management routes
		r.Route("/password", func(r chi.Router) {
			r.Post("/forgot", accountHandler.ForgotPassword)
			r.Post("/reset", accountHandler.ResetPassword)
			
			// Protected password change
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Put("/change", accountHandler.ChangePassword)
			})
		})
		
		// Email verification routes
		r.Route("/email", func(r chi.Router) {
			r.Get("/verify/{token}", accountHandler.VerifyEmail)
			r.Post("/resend-verification", accountHandler.ResendVerification)
		})

		// Protected user management routes
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			
			// CRUD operations
			r.Post("/", accountHandler.CreateAccount)
			r.Get("/", accountHandler.FindAllUsers)
			r.Get("/{id}", accountHandler.FindAccountByID)
			r.Put("/{id}", accountHandler.UpdateUserByID)
			r.Delete("/{id}", accountHandler.DeleteUser)
			
			// Profile management
			r.Route("/profile", func(r chi.Router) {
				r.Get("/", accountHandler.GetUserProfile)      // Current user
				r.Get("/{id}", accountHandler.GetUserProfile)  // Specific user
			})
			
			// Search and filtering
			r.Route("/search", func(r chi.Router) {
				r.Get("/", accountHandler.SearchUsers)              // Advanced search
				r.Get("/email/{email}", accountHandler.FindByEmail) // Find by email
				r.Get("/role/{role}", accountHandler.FindByRole)    // Find by role
			})
			
			// Branch-related endpoints
			r.Route("/branch", func(r chi.Router) {
				r.Get("/{branch_id}", accountHandler.FindByBranch)
				r.Get("/{branch_id}/users", accountHandler.GetUsersByBranch)
			})
			
			// Account management
			r.Route("/manage", func(r chi.Router) {
				r.Put("/{id}/status", accountHandler.UpdateAccountStatus)
			})
		})
	})
}

// Alternative factory function if you prefer to keep the Handler struct approach
// func NewHandler(userClient pb.AccountServiceClient) *AccountHandler {
// 	return NewAccountHandler(userClient)
// }

// Route documentation and usage examples
/*
API Endpoints Overview:

PUBLIC ENDPOINTS:
POST   /accounts/auth/register              - User registration
POST   /accounts/auth/login                 - User login
POST   /accounts/auth/logout                - User logout
POST   /accounts/auth/refresh-token         - Refresh access token
POST   /accounts/auth/validate-token        - Validate token
POST   /accounts/password/forgot            - Request password reset
POST   /accounts/password/reset             - Reset password with token
GET    /accounts/email/verify/{token}       - Verify email address
POST   /accounts/email/resend-verification  - Resend verification email

PROTECTED ENDPOINTS (require authentication):
POST   /accounts/                           - Create new user account
GET    /accounts/                           - Get all users (paginated)
GET    /accounts/{id}                       - Get user by ID
PUT    /accounts/{id}                       - Update user by ID
DELETE /accounts/{id}                       - Delete user by ID

GET    /accounts/profile                    - Get current user profile
GET    /accounts/profile/{id}               - Get user profile by ID
PUT    /accounts/password/change            - Change password

GET    /accounts/search/email/{email}       - Find user by email
GET    /accounts/search/role/{role}         - Find users by role
GET    /accounts/branch/{branch_id}         - Find users by branch
GET    /accounts/branch/{branch_id}/users   - Get users by branch (alternative)
GET    /accounts/search                     - Advanced user search with query params
PUT    /accounts/manage/{id}/status         - Update account status

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
*/