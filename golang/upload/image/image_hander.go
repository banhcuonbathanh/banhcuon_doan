package image

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"io"
	"net/http"
	"os"
	"path/filepath"

	"english-ai-full/token"
)

type ImageHandlerController struct {
	ctx        context.Context

	TokenMaker *token.JWTMaker
}

func NewImageHandler( secretKey string) *ImageHandlerController {
	return &ImageHandlerController{
		ctx:        context.Background(),

		TokenMaker: token.NewJWTMaker(secretKey),
	}
}

func (h *ImageHandlerController) UploadImage(w http.ResponseWriter, r *http.Request) {
	log.Println("UploadImage function called")

	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retrieving file from form: %v", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadPath := r.FormValue("path")
	log.Printf("Received upload path: %s", uploadPath)
	if uploadPath == "" {
		log.Println("Upload path is empty")
		http.Error(w, "Upload path is required", http.StatusBadRequest)
		return
	}

	uploadPath = filepath.Clean(uploadPath)
	if strings.HasPrefix(uploadPath, "..") {
		log.Printf("Invalid upload path detected: %s", uploadPath)
		http.Error(w, "Invalid upload path", http.StatusBadRequest)
		return
	}

	fullUploadPath := filepath.Join("uploads", uploadPath)
	log.Printf("Full upload path: %s", fullUploadPath)

	err = os.MkdirAll(fullUploadPath, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory structure: %v", err)
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}
	log.Println("Directory structure created successfully")

	dstPath := filepath.Join(fullUploadPath, handler.Filename)
	log.Printf("Destination file path: %s", dstPath)
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Error creating destination file: %v", err)
		http.Error(w, "Error creating the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file contents: %v", err)
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
		return
	}
	log.Println("File copied successfully")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "File uploaded successfully",
		"filename": handler.Filename,
		"path":     filepath.Join(uploadPath, handler.Filename),
	})
	log.Println("Upload completed successfully")
}


func (h *ImageHandlerController) ServeImage(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	path = filepath.Clean(path)
	if strings.HasPrefix(path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	filepath := filepath.Join("uploads", path, filename)
	http.ServeFile(w, r, filepath)
}


func (h *ImageHandlerController) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder. You'd typically generate or serve a thumbnail here.
	// For now, we'll just serve the original image.
	h.ServeImage(w, r)
}

func (h *ImageHandlerController) DeleteImage(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	err := os.Remove(filepath.Join("uploads", filename))
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting file", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File deleted successfully"})
}