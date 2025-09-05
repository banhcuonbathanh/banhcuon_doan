package internal

import (
	"fmt"
	"net/http"
	"time"
)

// LayerContext contains standardized layer information for logging and operations
type LayerContext struct {
	Domain      string `json:"domain"`
	Layer       string `json:"layer"`
	Service     string `json:"service"`
	Version     string `json:"version,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// ToMap converts LayerContext to map[string]interface{} for logging
func (lc *LayerContext) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"domain": lc.Domain,
		"layer":  lc.Layer,
	}
	
	if lc.Service != "" {
		result["service"] = lc.Service
	}
	
	if lc.Version != "" {
		result["version"] = lc.Version
	}
	
	if lc.Environment != "" {
		result["environment"] = lc.Environment
	}
	
	return result
}

// MergeWithContext merges LayerContext with additional context data
func (lc *LayerContext) MergeWithContext(additionalCtx map[string]interface{}) map[string]interface{} {
	merged := lc.ToMap()
	
	// Add additional context
	for key, value := range additionalCtx {
		merged[key] = value
	}
	
	return merged
}


func (lc *LayerContext) BuildOperationContext(operation string, r *http.Request, requestID string) map[string]interface{} {
	operationCtx := lc.ToMap()
	
	// Add operation-specific data
	operationCtx["operation"] = operation
	operationCtx["request_id"] = requestID
	operationCtx["method"] = r.Method
	operationCtx["path"] = r.URL.Path
	operationCtx["client_ip"] = getClientIP(r)
	operationCtx["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	
	return operationCtx
}

func getClientIP(r *http.Request) string {
	// Try to get IP from X-Forwarded-For header
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	
	// Try to get IP from X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Helper methods (you might already have these)
func (lc *LayerContext) GetRequestID(r *http.Request) string {
	// Implementation to extract request ID from headers or generate one
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return requestID
}

func (lc *LayerContext) GetClientIP(r *http.Request) string {
	// Get client IP from various headers
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}