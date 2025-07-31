package account

// import (
// 	"net/http"

// 	"english-ai-full/pkg/middleware/auth"
// 	"github.com/go-chi/chi"
// )

// func RegisterRoutes(r *chi.Mux, handler Handler) {
// 	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("Server is running"))
// 	})
	
// 	r.Route("/accounts", func(r chi.Router) {
// 		// Public endpoints (no authentication required)
// 		r.Post("/register", handler.Register)
// 		r.Post("/login", handler.Login)
// 		r.Post("/logout", handler.Logout)

// 		// Protected endpoints (authentication required)
// 		r.Group(func(r chi.Router) {
// 			r.Use(auth.AuthMiddleware)
			
// 			// User management endpoints
// 			r.Post("/", handler.CreateAccount)
// 			r.Get("/{id}", handler.FindAccountByID)
// 			r.Put("/{id}", handler.UpdateUserByID)     // Changed from POST to PUT for better REST convention
// 			r.Delete("/{id}", handler.DeleteUser)
			
// 			// Additional endpoints from interface
// 			r.Get("/profile/{id}", handler.GetUserProfile)
// 			r.Get("/profile", handler.GetUserProfile)  // For current user (ID from JWT)
// 			r.Put("/{id}/password", handler.ChangePassword)
// 			r.Get("/branch/{branch_id}", handler.GetUsersByBranch)
			
// 			// Email lookup endpoint
// 			r.Get("/email/{email}", handler.FindByEmail)
// 		})
// 	})
// }