package utils

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// golang/cmd/server/main.go (add this middleware function)

func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if request ID already exists in headers
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            // Generate new request ID if not provided
            requestID = generateRequestID()
        }
        
        // Add to request context
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        r = r.WithContext(ctx)
        
        // Add to response header
        w.Header().Set("X-Request-ID", requestID)
        
        next.ServeHTTP(w, r)
    })
}

func generateRequestID() string {
    // Generate UUID v4 or use any ID generation method
    return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), randomString(8))
}

func randomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}
