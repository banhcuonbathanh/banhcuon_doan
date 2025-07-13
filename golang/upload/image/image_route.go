package image

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

func RegisterImageRoutes(r *chi.Mux, handler *ImageHandlerController) *chi.Mux {
	r.Get("/image-atest", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server image is running"))
	})

	r.Route("/images", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/upload", handler.UploadImage)
			r.Get("/image", handler.ServeImage)
			r.Get("/thumbnail", handler.ServeThumbnail)
			r.Delete("/deleteImage", handler.DeleteImage)
		})
	})

	FileServer(r, "/uploads", http.Dir("./uploads"))

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve static files from a http.FileSystem.

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
