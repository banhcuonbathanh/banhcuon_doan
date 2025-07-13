package http

import (
	"encoding/json"
	"net/http"
)

func ErrorHandler(w http.ResponseWriter, statusCode int, err error, message string) {
	response := map[string]string{
		"error":   err.Error(),
		"message": message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
