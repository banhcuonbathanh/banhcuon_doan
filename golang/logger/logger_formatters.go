// logger/formatters.go - Different log formatting implementations
package logger

import (
	"fmt"
	"strings"
	"time"
)

// Format log entry as pretty text with enhanced error info
func (l *Logger) formatPretty(entry LogEntry) string {
	timestamp := time.Now().Format("15:04:05.000")
	
	// Get level emoji and color
	levelDisplay := formatLevel(entry.Level)
	
	// Build main message
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, levelDisplay)
	
	// Add layer if present
	if entry.Layer != "" {
		parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.Layer)))
	}
	
	// Add component if present
	if entry.Component != "" {
		parts = append(parts, fmt.Sprintf("<%s>", entry.Component))
	}
	
	// Add operation if present
	if entry.Operation != "" {
		parts = append(parts, fmt.Sprintf("{%s}", entry.Operation))
	}
	
	parts = append(parts, entry.Message)
	
	mainLine := strings.Join(parts, " ")
	
	// Add important context on the same line
	var contextParts []string
	
	// Add key identifiers
	if email, ok := entry.Context["email"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("user=%v", email))
	}
	if ip, ok := entry.Context["ip"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("ip=%v", ip))
	}
	if entry.Duration > 0 {
		contextParts = append(contextParts, fmt.Sprintf("took=%dms", entry.Duration))
	}
	if statusCode, ok := entry.Context["status_code"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("status=%v", statusCode))
	}
	if reason, ok := entry.Context["failure_reason"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("reason=%v", reason))
	}
	if errorMsg, ok := entry.Context["error"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("error=%v", errorMsg))
	}
	
	// Add cause if present (for errors)
	if entry.Cause != "" {
		contextParts = append(contextParts, fmt.Sprintf("cause=%s", entry.Cause))
	}
	
	if len(contextParts) > 0 {
		mainLine += " | " + strings.Join(contextParts, " ")
	}
	
	// Add file/line info for debug/error levels
	if entry.Level == "DEBUG" || entry.Level == "ERROR" {
		if entry.File != "" {
			mainLine += fmt.Sprintf(" (%s:%d)", entry.File, entry.Line)
		}
	}
	
	return mainLine
}

// Format log entry as simple text with enhanced error info
func (l *Logger) formatText(entry LogEntry) string {
	timestamp := time.Now().Format("15:04:05")
	
	// Build message with layer and operation info
	var msgParts []string
	if entry.Layer != "" {
		msgParts = append(msgParts, fmt.Sprintf("[%s]", entry.Layer))
	}
	if entry.Operation != "" {
		msgParts = append(msgParts, fmt.Sprintf("{%s}", entry.Operation))
	}
	msgParts = append(msgParts, entry.Message)
	
	msg := fmt.Sprintf("[%s] %s: %s", timestamp, entry.Level, strings.Join(msgParts, " "))
	
	// Add minimal essential context
	if entry.Context != nil {
		var essentials []string
		
		// Only show really important stuff
		if email, ok := entry.Context["email"]; ok {
			essentials = append(essentials, fmt.Sprintf("user=%v", email))
		}
		if entry.Duration > 0 {
			essentials = append(essentials, fmt.Sprintf("%dms", entry.Duration))
		}
		if errorMsg, ok := entry.Context["error"]; ok {
			essentials = append(essentials, fmt.Sprintf("error=%v", errorMsg))
		}
		if entry.Cause != "" {
			essentials = append(essentials, fmt.Sprintf("cause=%s", entry.Cause))
		}
		
		if len(essentials) > 0 {
			msg += " (" + strings.Join(essentials, " ") + ")"
		}
	}
	
	return msg
}

// Format level with emoji and colors
func formatLevel(level string) string {
	switch level {
	case "DEBUG":
		return "🔍 DEBUG"
	case "INFO":
		return "ℹ️  INFO"
	case "WARN":
		return "⚠️  WARN"
	case "ERROR":
		return "❌ ERROR"
	case "FATAL":
		return "💀 FATAL"
	default:
		return level
	}
}