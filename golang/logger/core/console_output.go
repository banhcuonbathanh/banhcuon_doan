// logger/core/console_output.go - Rich console output implementation
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Enhanced RichConsoleOutput with better error formatting
type RichConsoleOutput struct {
	useColors   bool
	showCaller  bool
	showStack   bool
	mu          sync.Mutex
}

func NewRichConsoleOutput(useColors bool) *RichConsoleOutput {
	return &RichConsoleOutput{
		useColors:  useColors,
		showCaller: true,
		showStack:  true, // Show stack trace for errors
	}
}

func (rco *RichConsoleOutput) Write(entry *LogEntry) error {
	rco.mu.Lock()
	defer rco.mu.Unlock()
	
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	
	// Build the log line with all information
	var parts []string
	
	// Timestamp with grey color for all levels
	timestampStr := fmt.Sprintf("[%s]", timestamp)
	if rco.useColors {
		timestampStr = rco.colorize("\033[90m", timestampStr) // Dark grey for all timestamps
	}
	parts = append(parts, timestampStr)
	
	// Level with color
	levelStr := entry.Level.String()
	if rco.useColors {
		levelStr = rco.colorizeLevel(entry.Level, levelStr)
	}
	parts = append(parts, levelStr)
	
	// Layer and Component
	if entry.Layer != "" {
		layerStr := fmt.Sprintf("[%s]", strings.ToUpper(entry.Layer))
		if rco.useColors {
			layerStr = rco.colorize("\033[94m", layerStr) // Light blue
		}
		parts = append(parts, layerStr)
	}
	
	if entry.Component != "" {
		componentStr := fmt.Sprintf("<%s>", entry.Component)
		if rco.useColors {
			componentStr = rco.colorize("\033[95m", componentStr) // Magenta
		}
		parts = append(parts, componentStr)
	}
	
	if entry.Operation != "" {
		operationStr := fmt.Sprintf("{%s}", entry.Operation)
		if rco.useColors {
			operationStr = rco.colorize("\033[96m", operationStr) // Cyan
		}
		parts = append(parts, operationStr)
	}
	
	// Enhanced caller information for errors
	if rco.showCaller && entry.Caller != nil && entry.Level >= ErrorLevel {
		callerStr := fmt.Sprintf("@%s.%s:%d", entry.Caller.Package, entry.Caller.Function, entry.Caller.Line)
		if rco.useColors {
			callerStr = rco.colorize("\033[93m", callerStr) // Bright yellow
		}
		parts = append(parts, callerStr)
	}
	
	// Enhanced message with domain and cause for errors/warnings
	message := rco.buildEnhancedMessage(entry)
	parts = append(parts, message)
	
	// Join main parts
	logLine := strings.Join(parts, " ")
	
	// Add essential fields as JSON on the same line if they exist
	if len(entry.Fields) > 0 {
		filteredFields := rco.filterFields(entry)
		
		if len(filteredFields) > 0 {
			fieldsJSON, err := json.Marshal(filteredFields)
			if err == nil {
				fieldStr := fmt.Sprintf(" | %s", string(fieldsJSON))
				if rco.useColors {
					fieldStr = rco.colorize("\033[90m", fieldStr) // Dark gray
				}
				logLine += fieldStr
			}
		}
	}
	
	// Write main log line to appropriate stream
	if entry.Level >= ErrorLevel {
		fmt.Fprintf(os.Stderr, "%s\n", logLine)
		
		// Add stack trace for errors if available and enabled
		if rco.showStack && len(entry.StackTrace) > 0 {
			stackStr := rco.formatStackTrace(entry.StackTrace)
			if rco.useColors {
				stackStr = rco.colorize("\033[90m", stackStr) // Dark gray
			}
			fmt.Fprintf(os.Stderr, "%s\n", stackStr)
		}
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", logLine)
	}
	
	return nil
}

func (rco *RichConsoleOutput) buildEnhancedMessage(entry *LogEntry) string {
	message := entry.Message
	
	if entry.Level >= ErrorLevel {
		// Extract domain and cause from fields for enhanced error display
		domain := ""
		cause := ""
		
		if entry.Fields != nil {
			if d, ok := entry.Fields["domain"].(string); ok {
				domain = d
			}
			if c, ok := entry.Fields["cause"].(string); ok {
				cause = c
			}
		}
		
		// If we have cause from the entry itself, use that
		if entry.Cause != "" {
			cause = entry.Cause
		}
		
		// Build enhanced error message
		if domain != "" && cause != "" {
			message = fmt.Sprintf("%s [domain=%s] [cause=%s]", message, domain, cause)
		} else if domain != "" {
			message = fmt.Sprintf("%s [domain=%s]", message, domain)
		} else if cause != "" {
			message = fmt.Sprintf("%s [cause=%s]", message, cause)
		}
	}
	
	return message
}

func (rco *RichConsoleOutput) filterFields(entry *LogEntry) map[string]interface{} {
	// Return all fields instead of filtering
	filteredFields := make(map[string]interface{})
	
	for k, v := range entry.Fields {
		// Skip domain and cause for error levels since they're in the message
		if entry.Level >= ErrorLevel && (k == "domain" || k == "cause") {
			continue
		}
		// Include ALL other fields
		filteredFields[k] = v
	}
	
	return filteredFields
}

func (rco *RichConsoleOutput) formatStackTrace(stack []CallerInfo) string {
	if len(stack) == 0 {
		return ""
	}
	
	var lines []string
	lines = append(lines, "Stack trace:")
	
	for i, frame := range stack {
		line := fmt.Sprintf("  %d. %s.%s() at %s:%d", 
			i+1, frame.Package, frame.Function, frame.File, frame.Line)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
}

func (rco *RichConsoleOutput) colorizeLevel(level Level, text string) string {
	if !rco.useColors {
		return text
	}
	
	var color string
	switch level {
	case DebugLevel:
		color = "\033[36m" // Cyan
	case InfoLevel:
		color = "\033[32m" // Green
	case WarnLevel:
		color = "\033[33m" // Yellow
	case ErrorLevel:
		color = "\033[31m" // Red
	case FatalLevel:
		color = "\033[35m\033[1m" // Bold Magenta
	default:
		return text
	}
	
	return color + text + "\033[0m"
}

func (rco *RichConsoleOutput) colorize(color, text string) string {
	if !rco.useColors {
		return text
	}
	return color + text + "\033[0m"
}

func (rco *RichConsoleOutput) Close() error {
	return nil
}