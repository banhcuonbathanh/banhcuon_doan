package account_handler

import (
	"net/http"

	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/pkg/middleware/auth"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	user      pb.AccountServiceClient
	validator *validator.Validate
}
// Alternative route structure with more organized grouping
func RegisterRoutesAccountHandler(r *chi.Mux, handler Handler) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is running"))
	})
	
	r.Route("/accounts", func(r chi.Router) {
		// Authentication routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handler.Register)
			r.Post("/login", handler.Login)
			r.Post("/logout", handler.Logout)
			r.Post("/refresh-token", handler.RefreshToken)
			r.Post("/validate-token", handler.ValidateToken)
		})
		
		// Password management routes
		r.Route("/password", func(r chi.Router) {
			r.Post("/forgot", handler.ForgotPassword)
			r.Post("/reset", handler.ResetPassword)
			
			// Protected password change
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Put("/change", handler.ChangePassword)
			})
		})
		
		// Email verification routes
		r.Route("/email", func(r chi.Router) {
			r.Get("/verify/{token}", handler.VerifyEmail)
			r.Post("/resend-verification", handler.ResendVerification)
		})

		// Protected user management routes
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			
			// CRUD operations
			r.Post("/", handler.CreateAccount)
			r.Get("/", handler.FindAllUsers)
			r.Get("/{id}", handler.FindAccountByID)
			r.Put("/{id}", handler.UpdateUserByID)
			r.Delete("/{id}", handler.DeleteUser)
			
			// Profile management
			r.Route("/profile", func(r chi.Router) {
				r.Get("/", handler.GetUserProfile)      // Current user
				r.Get("/{id}", handler.GetUserProfile)  // Specific user
			})
			
			// Search and filtering
			r.Route("/search", func(r chi.Router) {
				r.Get("/", handler.SearchUsers)              // Advanced search
				r.Get("/email/{email}", handler.FindByEmail) // Find by email
				r.Get("/role/{role}", handler.FindByRole)    // Find by role
			})
			
			// Branch-related endpoints
			r.Route("/branch", func(r chi.Router) {
				r.Get("/{branch_id}", handler.FindByBranch)
				r.Get("/{branch_id}/users", handler.GetUsersByBranch)
			})
			
			// Account management
			r.Route("/manage", func(r chi.Router) {
				r.Put("/{id}/status", handler.UpdateAccountStatus)
			})
		})
	})
}

// Route documentation and usage examples
/*
API Endpoints Overview:

PUBLIC ENDPOINTS:
POST   /accounts/register              - User registration
POST   /accounts/login                 - User login
POST   /accounts/logout                - User logout
POST   /accounts/forgot-password       - Request password reset
POST   /accounts/reset-password        - Reset password with token
GET    /accounts/verify-email/{token}  - Verify email address
POST   /accounts/resend-verification   - Resend verification email
POST   /accounts/refresh-token         - Refresh access token
POST   /accounts/validate-token        - Validate token

PROTECTED ENDPOINTS (require authentication):
POST   /accounts/                      - Create new user account
GET    /accounts/                      - Get all users (paginated)
GET    /accounts/{id}                  - Get user by ID
PUT    /accounts/{id}                  - Update user by ID
DELETE /accounts/{id}                  - Delete user by ID

GET    /accounts/profile               - Get current user profile
GET    /accounts/profile/{id}          - Get user profile by ID
PUT    /accounts/password/change       - Change password

GET    /accounts/email/{email}         - Find user by email
GET    /accounts/role/{role}           - Find users by role
GET    /accounts/branch/{branch_id}    - Find users by branch
GET    /accounts/branch/{branch_id}/users - Get users by branch (alternative)
GET    /accounts/search                - Advanced user search with query params
PUT    /accounts/{id}/status           - Update account status

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
GET /accounts/role/manager?page=1&page_size=10
*/