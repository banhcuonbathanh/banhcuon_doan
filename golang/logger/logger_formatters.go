// logger/formatters.go - Different log formatting implementations
package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
	 "english-ai-full/logger/core"

)

// Formatter interface for different output formats
type Formatter interface {
	Format(entry *core.LogEntry) (string, error)
}

// JSONFormatter formats log entries as JSON
type JSONFormatter struct {
	timestampFormat string
	prettyPrint     bool
}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		timestampFormat: time.RFC3339Nano,
		prettyPrint:     false,
	}
}

func NewPrettyJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		timestampFormat: time.RFC3339Nano,
		prettyPrint:     true,
	}
}

func (jf *JSONFormatter) Format(entry *core.LogEntry) (string, error) {
	data := make(map[string]interface{})
	
	// Core fields
	data["timestamp"] = entry.Timestamp.Format(jf.timestampFormat)
	data["level"] = entry.Level.String()
	data["message"] = entry.Message
	
	// Add caller info if available
	if entry.Caller != "" {
		data["caller"] = entry.Caller
	}
	
	// Add metadata fields
	if entry.Component != "" {
		data["component"] = entry.Component
	}
	if entry.Layer != "" {
		data["layer"] = entry.Layer
	}
	if entry.Operation != "" {
		data["operation"] = entry.Operation
	}
	if entry.Environment != "" {
		data["environment"] = entry.Environment
	}
	if entry.Cause != "" {
		data["cause"] = entry.Cause
	}
	if entry.RequestID != "" {
		data["request_id"] = entry.RequestID
	}
	if entry.UserID != "" {
		data["user_id"] = entry.UserID
	}
	if entry.TraceID != "" {
		data["trace_id"] = entry.TraceID
	}
	if entry.Duration > 0 {
		data["duration_ms"] = float64(entry.Duration.Nanoseconds()) / 1000000.0
	}
	if entry.ErrorCode != "" {
		data["error_code"] = entry.ErrorCode
	}
	
	// Add custom fields
	for key, value := range entry.Fields {
		// Avoid overwriting core fields
		if _, exists := data[key]; !exists {
			data[key] = value
		}
	}
	
	var jsonData []byte
	var err error
	
	if jf.prettyPrint {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}
	
	if err != nil {
		return "", err
	}
	
	return string(jsonData), nil
}

// TextFormatter formats log entries as simple text
type TextFormatter struct {
	timestampFormat string
	showCaller      bool
}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		timestampFormat: "2006-01-02 15:04:05.000",
		showCaller:      true,
	}
}

func (tf *TextFormatter) Format(entry *core.LogEntry) (string, error) {
	var parts []string
	
	// Timestamp
	parts = append(parts, entry.Timestamp.Format(tf.timestampFormat))
	
	// Level
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.Level.String())))
	
	// Layer and Component info
	if entry.Layer != "" {
		parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.Layer)))
	}
	if entry.Component != "" {
		parts = append(parts, fmt.Sprintf("<%s>", entry.Component))
	}
	if entry.Operation != "" {
		parts = append(parts, fmt.Sprintf("{%s}", entry.Operation))
	}
	
	// Caller
	if tf.showCaller && entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("(%s)", entry.Caller))
	}
	
	// Message
	parts = append(parts, entry.Message)
	
	result := strings.Join(parts, " ")
	
	// Fields - show only essential ones in text format
	if len(entry.Fields) > 0 {
		var fieldParts []string
		
		// Prioritize important fields
		essentialFields := []string{"user_id", "email", "ip", "status_code", "duration_ms", "error", "cause"}
		
		for _, field := range essentialFields {
			if value, exists := entry.Fields[field]; exists {
				fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", field, value))
			}
		}
		
		if len(fieldParts) > 0 {
			result += " {" + strings.Join(fieldParts, " ") + "}"
		}
	}
	
	return result, nil
}

// PrettyFormatter formats log entries with colors and enhanced readability
type PrettyFormatter struct {
	colors          bool
	showTime        bool
	showLevel       bool
	showCaller      bool
	showMetadata    bool
	timestampFormat string
}

func NewPrettyFormatter(colors bool) *PrettyFormatter {
	return &PrettyFormatter{
		colors:          colors,
		showTime:        true,
		showLevel:       true,
		showCaller:      true,
		showMetadata:    true,
		timestampFormat: "15:04:05.000",
	}
}

func (pf *PrettyFormatter) Format(entry *core.LogEntry) (string, error) {
	var parts []string
	
	// Timestamp
	if pf.showTime {
		timeStr := entry.Timestamp.Format(pf.timestampFormat)
		if pf.colors {
			timeStr = pf.colorize("\033[90m", timeStr) // Dark gray
		}
		parts = append(parts, fmt.Sprintf("[%s]", timeStr))
	}
	
	// Level with emoji and colors
	if pf.showLevel {
		levelStr := pf.formatLevel(entry.Level)
		parts = append(parts, levelStr)
	}
	
	// Layer information
	if pf.showMetadata {
		if entry.Layer != "" {
			layerStr := fmt.Sprintf("[%s]", strings.ToUpper(entry.Layer))
			if pf.colors {
				layerStr = pf.colorize("\033[94m", layerStr) // Light blue
			}
			parts = append(parts, layerStr)
		}
		
		if entry.Component != "" {
			componentStr := fmt.Sprintf("<%s>", entry.Component)
			if pf.colors {
				componentStr = pf.colorize("\033[95m", componentStr) // Magenta
			}
			parts = append(parts, componentStr)
		}
		
		if entry.Operation != "" {
			operationStr := fmt.Sprintf("{%s}", entry.Operation)
			if pf.colors {
				operationStr = pf.colorize("\033[96m", operationStr) // Cyan
			}
			parts = append(parts, operationStr)
		}
	}
	
	// Caller information
	if pf.showCaller && entry.Caller != "" {
		caller := fmt.Sprintf("(%s)", entry.Caller)
		if pf.colors {
			caller = pf.colorize("\033[36m", caller) // Cyan
		}
		parts = append(parts, caller)
	}
	
	// Message
	message := entry.Message
	if pf.colors && entry.Level >= core.ErrorLevel {
		message = pf.colorize("\033[1m", message) // Bold for errors
	}
	
	result := strings.Join(parts, " ") + " " + message
	
	// Important context on the same line
	if len(entry.Fields) > 0 {
		contextParts := pf.formatImportantContext(entry)
		if len(contextParts) > 0 {
			contextStr := " | " + strings.Join(contextParts, " ")
			if pf.colors {
				contextStr = pf.colorize("\033[90m", contextStr) // Dark gray
			}
			result += contextStr
		}
	}
	
	return result, nil
}

func (pf *PrettyFormatter) formatImportantContext(entry *core.LogEntry) []string {
	var contextParts []string
	
	// Prioritize important fields for inline display
	importantFields := map[string]string{
		"email":        "user",
		"user_id":      "uid",
		"ip":           "ip",
		"status_code":  "status",
		"duration_ms":  "took",
		"error":        "error",
		"cause":        "cause",
		"failure_reason": "reason",
	}
	
	for field, label := range importantFields {
		if value, exists := entry.Fields[field]; exists {
			var displayValue string
			switch field {
			case "duration_ms":
				displayValue = fmt.Sprintf("%vms", value)
			case "email":
				displayValue = fmt.Sprintf("%v", value) // Assuming it's already masked
			default:
				displayValue = fmt.Sprintf("%v", value)
			}
			
			fieldStr := fmt.Sprintf("%s=%s", label, displayValue)
			if pf.colors {
				labelStr := pf.colorize("\033[33m", label) // Yellow
				fieldStr = fmt.Sprintf("%s=%s", labelStr, displayValue)
			}
			contextParts = append(contextParts, fieldStr)
		}
	}
	
	// Add cause from entry if present
	if entry.Cause != "" {
		causeStr := fmt.Sprintf("cause=%s", entry.Cause)
		if pf.colors {
			causeStr = pf.colorize("\033[31m", causeStr) // Red
		}
		contextParts = append(contextParts, causeStr)
	}
	
	return contextParts
}

func (pf *PrettyFormatter) formatLevel(level core.Level) string {
	var levelStr string
	var color string
	
	switch level {
	case core.DebugLevel:
		levelStr = "🔍 DEBUG"
		color = "\033[36m" // Cyan
	case core.InfoLevel:
		levelStr = "ℹ️  INFO"
		color = "\033[32m" // Green
	case core.WarnLevel:
		levelStr = "⚠️  WARN"
		color = "\033[33m" // Yellow
	case core.ErrorLevel:
		levelStr = "❌ ERROR"
		color = "\033[31m" // Red
	case core.FatalLevel:
		levelStr = "💀 FATAL"
		color = "\033[35m\033[1m" // Bold Magenta
	default:
		levelStr = level.String()
		color = "\033[37m" // White
	}
	
	if pf.colors {
		return pf.colorize(color, levelStr)
	}
	
	return levelStr
}

func (pf *PrettyFormatter) colorize(color, text string) string {
	if !pf.colors {
		return text
	}
	return color + text + "\033[0m"
}

// Helper function to get caller information
func getCaller(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	
	function := runtime.FuncForPC(pc).Name()
	
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}
	
	if lastDot := strings.LastIndex(function, "."); lastDot >= 0 {
		function = function[lastDot+1:]
	}
	
	return fmt.Sprintf("%s:%d", file, line)
}