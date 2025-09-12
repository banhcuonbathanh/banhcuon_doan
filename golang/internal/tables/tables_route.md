package tables_test


import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func RegisterTablesRoutes(r *chi.Mux, handler *TablesHandlerController) *chi.Mux {
	r.Get("/table/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("table test is running"))
	})

	r.Route("/table", func(r chi.Router) {
		// Public routes (no authentication required)
		r.Get("/list", handler.GetTableList)
		r.Get("/{tableNumber}", handler.GetTableDetail)

		// Protected routes (authentication might be required)
		r.Group(func(r chi.Router) {
			log.Print("table/table_route.go")
			// Add authentication middleware here if needed
			// r.Use(middleware.SomeAuthMiddleware)

			r.Post("/", handler.CreateTable)
			r.Put("/{tableNumber}", handler.UpdateTable)
			r.Delete("/{tableNumber}", handler.DeleteTable)
		})
	})

	return r
}