// logger/core/file_output.go - File output implementation
package log_output

import (
	"encoding/json"
	"english-ai-full/logger/core"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileOutput writes logs to files with rotation support
type FileOutput struct {
	baseDir        string
	filename       string
	maxSize        int64  // Maximum file size in bytes
	maxAge         int    // Maximum file age in days
	compress       bool   // Compress old log files
	currentFile    *os.File
	currentSize    int64
	mu             sync.Mutex
	enableRotation bool
	format         OutputFormat
}

type OutputFormat int

const (
	JSONFormat OutputFormat = iota
	TextFormat
)

// FileOutputConfig holds configuration for file output
type FileOutputConfig struct {
	BaseDir        string       // Base directory for log files
	Filename       string       // Base filename (without extension)
	MaxSize        int64        // Maximum file size in bytes (0 = no limit)
	MaxAge         int          // Maximum file age in days (0 = no limit)
	Compress       bool         // Compress old log files
	EnableRotation bool         // Enable file rotation
	Format         OutputFormat // Output format (JSON or Text)
}

// NewFileOutput creates a new file output with the given configuration
func NewFileOutput(config FileOutputConfig) (*FileOutput, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", config.BaseDir, err)
	}

	fo := &FileOutput{
		baseDir:        config.BaseDir,
		filename:       config.Filename,
		maxSize:        config.MaxSize,
		maxAge:         config.MaxAge,
		compress:       config.Compress,
		enableRotation: config.EnableRotation,
		format:         config.Format,
	}

	// Open initial log file
	if err := fo.openLogFile(); err != nil {
		return nil, err
	}

	return fo, nil
}

// Write implements the Output interface
func (fo *FileOutput) Write(entry *core.LogEntry) error {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	// Check if rotation is needed
	if fo.enableRotation && fo.needsRotation() {
		if err := fo.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	// Format the log entry
	var logLine string
	var err error

	switch fo.format {
	case JSONFormat:
		logLine, err = fo.formatJSON(entry)
	case TextFormat:
		logLine, err = fo.formatText(entry)
	default:
		logLine, err = fo.formatText(entry)
	}

	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}

	// Write to file
	n, err := fo.currentFile.WriteString(logLine + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	fo.currentSize += int64(n)

	// Force sync for error and fatal levels
	if entry.Level >= core.ErrorLevel {
		fo.currentFile.Sync()
	}

	return nil
}

// formatJSON formats the log entry as JSON
func (fo *FileOutput) formatJSON(entry *core.LogEntry) (string, error) {
	// Create a structured log entry for JSON output
	jsonEntry := map[string]interface{}{
		"timestamp":   entry.Timestamp.Format(time.RFC3339Nano),
		"level":       entry.Level.String(),
		"message":     entry.Message,
		"environment": entry.Environment,
	}

	// Add optional fields
	if entry.Component != "" {
		jsonEntry["component"] = entry.Component
	}
	if entry.Layer != "" {
		jsonEntry["layer"] = entry.Layer
	}
	if entry.Operation != "" {
		jsonEntry["operation"] = entry.Operation
	}
	if entry.Cause != "" {
		jsonEntry["cause"] = entry.Cause
	}
	if entry.RequestID != "" {
		jsonEntry["request_id"] = entry.RequestID
	}
	if entry.UserID != "" {
		jsonEntry["user_id"] = entry.UserID
	}

	// Add caller information
	if entry.Caller != nil {
		jsonEntry["caller"] = map[string]interface{}{
			"function": entry.Caller.Function,
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
			"package":  entry.Caller.Package,
		}
	}

	// Add stack trace for errors
	if len(entry.StackTrace) > 0 {
		stackTrace := make([]map[string]interface{}, len(entry.StackTrace))
		for i, frame := range entry.StackTrace {
			stackTrace[i] = map[string]interface{}{
				"function": frame.Function,
				"file":     frame.File,
				"line":     frame.Line,
				"package":  frame.Package,
			}
		}
		jsonEntry["stack_trace"] = stackTrace
	}

	// Add custom fields
	if len(entry.Fields) > 0 {
		for k, v := range entry.Fields {
			jsonEntry[k] = v
		}
	}

	jsonBytes, err := json.Marshal(jsonEntry)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// formatText formats the log entry as human-readable text
func (fo *FileOutput) formatText(entry *core.LogEntry) (string, error) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	
	// Build the log line
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, entry.Level.String())

	if entry.Layer != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Layer))
	}
	if entry.Component != "" {
		parts = append(parts, fmt.Sprintf("<%s>", entry.Component))
	}
	if entry.Operation != "" {
		parts = append(parts, fmt.Sprintf("{%s}", entry.Operation))
	}

	// Add caller information for errors
	if entry.Caller != nil && entry.Level >= core.ErrorLevel {
		parts = append(parts, fmt.Sprintf("@%s.%s:%d", 
			entry.Caller.Package, entry.Caller.Function, entry.Caller.Line))
	}

	// Enhanced message with domain and cause for errors/warnings
	message := entry.Message
	if entry.Level >= core.ErrorLevel {
		if entry.Cause != "" {
			message = fmt.Sprintf("%s [cause=%s]", message, entry.Cause)
		}
		if domain, ok := entry.Fields["domain"].(string); ok && domain != "" {
			message = fmt.Sprintf("%s [domain=%s]", message, domain)
		}
	}
	parts = append(parts, message)

	logLine := fmt.Sprintf("%s", joinParts(parts))

	// Add fields as JSON on the same line if they exist
	if len(entry.Fields) > 0 {
		filteredFields := fo.filterFields(entry)
		if len(filteredFields) > 0 {
			fieldsJSON, err := json.Marshal(filteredFields)
			if err == nil {
				logLine += fmt.Sprintf(" | %s", string(fieldsJSON))
			}
		}
	}

	// Add stack trace for errors
	if len(entry.StackTrace) > 0 && entry.Level >= core.ErrorLevel {
		logLine += "\nStack trace:"
		for i, frame := range entry.StackTrace {
			logLine += fmt.Sprintf("\n  %d. %s.%s() at %s:%d", 
				i+1, frame.Package, frame.Function, frame.File, frame.Line)
		}
	}

	return logLine, nil
}

// filterFields filters out fields that are already displayed in the main log line
func (fo *FileOutput) filterFields(entry *core.LogEntry) map[string]interface{} {
	filteredFields := make(map[string]interface{})
	for k, v := range entry.Fields {
		// Skip domain and cause for error levels since they're in the message
		if entry.Level >= core.ErrorLevel && (k == "domain" || k == "cause") {
			continue
		}
		// Include important fields
		if k == "request_id" || k == "user_id" || k == "email" || k == "method" || 
		   k == "endpoint" || k == "status_code" || k == "duration_ms" || 
		   k == "error_code" || k == "service" || k == "table" || k == "function" ||
		   k == "attempted_email" || k == "attempted_role" || k == "error" {
			filteredFields[k] = v
		}
	}
	return filteredFields
}

// joinParts joins log parts with spaces
func joinParts(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}
	return result
}

// needsRotation checks if the current log file needs rotation
func (fo *FileOutput) needsRotation() bool {
	if fo.maxSize <= 0 {
		return false
	}
	return fo.currentSize >= fo.maxSize
}

// rotateFile rotates the current log file
func (fo *FileOutput) rotateFile() error {
	// Close current file
	if fo.currentFile != nil {
		fo.currentFile.Close()
	}

	// Rename current file with timestamp
	currentPath := fo.getCurrentFilePath()
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	
	var rotatedPath string
	switch fo.format {
	case JSONFormat:
		rotatedPath = filepath.Join(fo.baseDir, fmt.Sprintf("%s-%s.json", fo.filename, timestamp))
	default:
		rotatedPath = filepath.Join(fo.baseDir, fmt.Sprintf("%s-%s.log", fo.filename, timestamp))
	}

	if err := os.Rename(currentPath, rotatedPath); err != nil {
		// If rename fails, just continue with a new file
		os.Remove(currentPath)
	}

	// Open new log file
	return fo.openLogFile()
}

// getCurrentFilePath returns the current log file path
func (fo *FileOutput) getCurrentFilePath() string {
	var extension string
	switch fo.format {
	case JSONFormat:
		extension = ".json"
	default:
		extension = ".log"
	}
	return filepath.Join(fo.baseDir, fo.filename+extension)
}

// openLogFile opens a new log file
func (fo *FileOutput) openLogFile() error {
	filePath := fo.getCurrentFilePath()
	
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}

	fo.currentFile = file
	
	// Get current file size
	info, err := file.Stat()
	if err != nil {
		fo.currentSize = 0
	} else {
		fo.currentSize = info.Size()
	}

	return nil
}

// Close implements the Output interface
func (fo *FileOutput) Close() error {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	if fo.currentFile != nil {
		return fo.currentFile.Close()
	}
	return nil
}

