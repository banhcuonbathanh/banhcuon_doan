
package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

func DecodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func RespondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, map[string]string{"error": message})
}

func HandleValidationErrors(w http.ResponseWriter, err error) {
	validationErrors := make([]string, 0)
	for _, fieldErr := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, fieldErr.Error())
	}
	RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":   "validation failed",
		"details": validationErrors,
	})
}

func ParseIDParam(r *http.Request, paramName string) (int64, error) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}
	return strconv.ParseInt(idStr, 10, 64)
}

func ValidatePassword(password string) bool {
	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)
	
	if len(password) < 8 {
		return false
	}
	
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*", c):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}


// Utils package extensions
func GetPaginationParams(r *http.Request) (limit, offset int64) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit = 10
	offset = 0

	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil {
			offset = o
		}
	}
	return
}

func GetSortParams(r *http.Request) (sortBy, sortOrder string) {
	sortBy = r.URL.Query().Get("sort_by")
	sortOrder = r.URL.Query().Get("sort_order")
	
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}
	return
}

func CalculatePaginationBounds(start, end, total int) (int, int) {
	if start >= total {
		start = total
	}
	if end >= total {
		end = total
	}
	return start, end
}