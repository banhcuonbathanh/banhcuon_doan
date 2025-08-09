// logger/types.go - Core types, constants, and structures
package logger

import (
	"log"
	"sync"

)

// Logger levels with numeric values for comparison
const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

var levelNames = map[int]string{
	DebugLevel:   "DEBUG",
	InfoLevel:    "INFO",
	WarningLevel: "WARN",
	ErrorLevel:   "ERROR",
	FatalLevel:   "FATAL",
}

// Output formats
const (
	FormatJSON   = "json"
	FormatText   = "text"
	FormatPretty = "pretty"
)

// Layer constants for better organization
const (
	LayerHandler    = "handler"
	LayerService    = "service"
	LayerRepository = "repository"
	LayerMiddleware = "middleware"
	LayerAuth       = "auth"
	LayerValidation = "validation"
	LayerCache      = "cache"
	LayerDatabase   = "database"
	LayerExternal   = "external"
	LayerSecurity   = "security"
)

// LogEntry represents a structured log entry with enhanced metadata
type LogEntry struct {
	Timestamp    string                 `json:"timestamp"`
	Level        string                 `json:"level"`
	Message      string                 `json:"message"`
	Context      map[string]interface{} `json:"context,omitempty"`
	File         string                 `json:"file,omitempty"`
	Function     string                 `json:"function,omitempty"`
	Line         int                    `json:"line,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	TraceID      string                 `json:"trace_id,omitempty"`
	Component    string                 `json:"component,omitempty"`
	Operation    string                 `json:"operation,omitempty"`
	Duration     int64                  `json:"duration_ms,omitempty"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	Environment  string                 `json:"environment,omitempty"`
	Cause        string                 `json:"cause,omitempty"`
	Layer        string                 `json:"layer,omitempty"`
}

// Logger structure with enhanced capabilities and thread safety
type Logger struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
	outputFormat  string
	enableDebug   bool
	minLevel      int
	environment   string
	component     string
	layer         string
	operation     string
	mutex         sync.RWMutex
	contextFields map[string]interface{}
}