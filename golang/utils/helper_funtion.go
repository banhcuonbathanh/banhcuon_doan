package utils

import (
	"encoding/json"
	"net"

	"fmt"

	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	errorcustom "english-ai-full/internal/error_custom"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

// func DecodeJSON(r io.Reader, v interface{}) error {
// 	decoder := json.NewDecoder(r)
// 	decoder.DisallowUnknownFields() // Critical security feature

// 	if err := decoder.Decode(v); err != nil {
// 		return errorcustom.NewAPIError(
// 			errorcustom.ErrCodeInvalidInput,
// 			"Invalid JSON format",
// 			http.StatusBadRequest,
// 		).WithDetail("error", err.Error())
// 	}
// 	return nil
// }

// Robust JSON response handler
// func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)

// 	if err := json.NewEncoder(w).Encode(payload); err != nil {
// 		log.Printf("JSON encoding error: %v", err)
// 		// Critical fallback to prevent incomplete responses
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	}
// }

// Structured error response handler
func RespondWithAPIError(w http.ResponseWriter, apiErr *errorcustom.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)
	
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		log.Printf("JSON encoding error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// // Centralized error handler with prioritized processing
// func HandleError(w http.ResponseWriter, err error) {
// 	log.Printf("Error occurred: %v", err) // Structured logging recommended in production

// 	// Prioritize validation errors
// 	var validatorErr validator.ValidationErrors
// 	if errors.As(err, &validatorErr) {
// 		HandleValidationErrors(w, validatorErr)
// 		return
// 	}

// 	// Handle custom error types
// 	var (
// 		apiErr             *errorcustom.APIError
// 		userNotFoundErr    *errorcustom.UserNotFoundError
// 		validationErr      *errorcustom.ValidationError
// 		authErr            *errorcustom.AuthenticationError
// 		authzErr           *errorcustom.AuthorizationError
// 		duplicateEmailErr  *errorcustom.DuplicateEmailError
// 		invalidTokenErr    *errorcustom.InvalidTokenError
// 		branchNotFoundErr  *errorcustom.BranchNotFoundError
// 		passwordErr        *errorcustom.PasswordValidationError
// 	)

// 	switch {
// 	case errors.As(err, &apiErr):
// 		RespondWithAPIError(w, apiErr)
// 	case errors.As(err, &userNotFoundErr):
// 		RespondWithAPIError(w, userNotFoundErr.ToAPIError())
// 	case errors.As(err, &validationErr):
// 		RespondWithAPIError(w, validationErr.ToAPIError())
// 	case errors.As(err, &authErr):
// 		RespondWithAPIError(w, authErr.ToAPIError())
// 	case errors.As(err, &authzErr):
// 		RespondWithAPIError(w, authzErr.ToAPIError())
// 	case errors.As(err, &duplicateEmailErr):
// 		RespondWithAPIError(w, duplicateEmailErr.ToAPIError())
// 	case errors.As(err, &invalidTokenErr):
// 		RespondWithAPIError(w, invalidTokenErr.ToAPIError())
// 	case errors.As(err, &branchNotFoundErr):
// 		RespondWithAPIError(w, branchNotFoundErr.ToAPIError())
// 	case errors.As(err, &passwordErr):
// 		RespondWithAPIError(w, passwordErr.ToAPIError())
// 	default:
// 		// Secure default for unknown errors
// 		RespondWithAPIError(w, errorcustom.NewAPIError(
// 			errorcustom.ErrCodeInternalError,
// 			"An internal error occurred",
// 			http.StatusInternalServerError,
// 		))
// 	}
// }

// // Comprehensive validation error handler
// func HandleValidationErrors(w http.ResponseWriter, err validator.ValidationErrors) {
// 	validationErrors := make([]map[string]string, 0, len(err))
	
// 	for _, fieldErr := range err {
// 		validationErrors = append(validationErrors, map[string]string{
// 			"field":   strings.ToLower(fieldErr.Field()),
// 			"tag":     fieldErr.Tag(),
// 			"message": getValidationMessage(fieldErr),
// 		})
// 	}
	
// 	RespondWithAPIError(w, errorcustom.NewAPIError(
// 		errorcustom.ErrCodeValidationError,
// 		"Validation failed",
// 		http.StatusBadRequest,
// 	).WithDetail("errors", validationErrors))
// }

// Detailed validation messages
func getValidationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	case "uuid", "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "numeric", "number":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid date/time (format: %s)", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Secure ID parameter parsing
func ParseIDParam(r *http.Request, paramName string) (int64, *errorcustom.APIError) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Invalid %s: must be a positive integer", paramName),
			http.StatusBadRequest,
		).WithDetail("value", idStr)
	}

	return id, nil
}

// Secure string parameter handling
func GetStringParam(r *http.Request, paramName string, minLen int) (string, *errorcustom.APIError) {
	value := chi.URLParam(r, paramName)
	if value == "" {
		return "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	if minLen > 0 && len(value) < minLen {
		return "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("%s must be at least %d characters", paramName, minLen),
			http.StatusBadRequest,
		)
	}

	for _, r := range value {
		if r < 32 || r == 127 { // Block control characters
			return "", errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				fmt.Sprintf("Invalid characters in %s", paramName),
				http.StatusBadRequest,
			)
		}
	}

	return value, nil
}

// Robust password validation
func ValidatePassword(password string) *errorcustom.APIError {
	const (
		minLength = 8
		maxLength = 128
	)

	if len(password) < minLength {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			fmt.Sprintf("Password must be at least %d characters", minLength),
			http.StatusBadRequest,
		)
	}

	if len(password) > maxLength {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			fmt.Sprintf("Password cannot exceed %d characters", maxLength),
			http.StatusBadRequest,
		)
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?/", c):
			hasSpecial = true
		}
	}

	var requirements []string
	if !hasUpper {
		requirements = append(requirements, "uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "lowercase letter")
	}
	if !hasDigit {
		requirements = append(requirements, "digit")
	}
	if !hasSpecial {
		requirements = append(requirements, "special character")
	}

	if len(requirements) > 0 {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			"Password does not meet security requirements",
			http.StatusBadRequest,
		).WithDetail("requirements", requirements)
	}

	return nil
}

// Safe pagination parameters
func GetPaginationParams(r *http.Request) (limit, offset int64, apiErr *errorcustom.APIError) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit = 10
	offset = 0

	if limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 64)
		switch {
		case err != nil:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid limit parameter",
				http.StatusBadRequest,
			)
		case l < 1:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Limit must be at least 1",
				http.StatusBadRequest,
			)
		case l > 100:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Limit cannot exceed 100",
				http.StatusBadRequest,
			)
		default:
			limit = l
		}
	}

	if offsetStr != "" {
		o, err := strconv.ParseInt(offsetStr, 10, 64)
		switch {
		case err != nil:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid offset parameter",
				http.StatusBadRequest,
			)
		case o < 0:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Offset cannot be negative",
				http.StatusBadRequest,
			)
		default:
			offset = o
		}
	}

	return limit, offset, nil
}

// Safe sorting parameters
func GetSortParams(r *http.Request, allowedFields []string) (sortBy, sortOrder string, apiErr *errorcustom.APIError) {
	sortBy = strings.ToLower(r.URL.Query().Get("sort_by"))
	sortOrder = strings.ToLower(r.URL.Query().Get("sort_order"))

	// Set safe defaults
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Invalid sort order. Use 'asc' or 'desc'",
			http.StatusBadRequest,
		)
	}

	// Validate field against allowlist
	if len(allowedFields) > 0 {
		valid := false
		for _, field := range allowedFields {
			if sortBy == field {
				valid = true
				break
			}
		}

		if !valid {
			return "", "", errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				fmt.Sprintf("Invalid sort field. Allowed: %v", allowedFields),
				http.StatusBadRequest,
			)
		}
	}

	return sortBy, sortOrder, nil
}

// Utility function for pagination metadata
func CalculatePagination(total, limit, offset int) (currentPage, totalPages int) {
	currentPage = (offset / limit) + 1
	totalPages = total / limit
	if total%limit > 0 {
		totalPages++
	}
	return currentPage, totalPages
}

// func ValidatePasswordWithDetails(password string) error {
// 	var requirements []string
	
// 	if len(password) < 8 {
// 		requirements = append(requirements, "at least 8 characters")
// 	}
	
// 	var (
// 		hasUpper   = false
// 		hasLower   = false
// 		hasDigit   = false
// 		hasSpecial = false
// 	)
	
// 	for _, c := range password {
// 		switch {
// 		case unicode.IsUpper(c):
// 			hasUpper = true
// 		case unicode.IsLower(c):
// 			hasLower = true
// 		case unicode.IsDigit(c):
// 			hasDigit = true
// 		case strings.ContainsRune("!@#$%^&*", c):
// 			hasSpecial = true
// 		}
// 	}
	
// 	if !hasUpper {
// 		requirements = append(requirements, "at least one uppercase letter")
// 	}
// 	if !hasLower {
// 		requirements = append(requirements, "at least one lowercase letter")
// 	}
// 	if !hasDigit {
// 		requirements = append(requirements, "at least one digit")
// 	}
// 	if !hasSpecial {
// 		requirements = append(requirements, "at least one special character (!@#$%^&*)")
// 	}
	
// 	if len(requirements) > 0 {
// 		return &errorcustom.PasswordValidationError{
// 			Requirements: requirements,
// 		}
// 	}
	
// 	return nil
// }

func CalculatePaginationBounds(start, end, total int) (int, int) {
	if start >= total {
		start = total
	}
	if end >= total {
		end = total
	}
	return start, end
}

func RespondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, map[string]string{"error": message})
}

// new 121312312


func GetClientIP(r *http.Request) string {
    // Get IP from X-Forwarded-For header
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        // Take the first IP (client IP)
        return strings.Split(forwarded, ",")[0]
    }
    
    // Get IP from X-Real-IP header
    realIP := r.Header.Get("X-Real-IP")
    if realIP != "" {
        return realIP
    }
    
    // Get IP from request RemoteAddr
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return ip
}