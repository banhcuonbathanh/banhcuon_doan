// internal/logger/outputs/outputs.go - Output system implementation
package logger

import (
	"english-ai-full/logger/core"
	"fmt"
	"os"
	"sync"
)

// Output interface for different output destinations
type Output interface {
	Write(entry *core.LogEntry) error
	Close() error
}

// OutputManager manages multiple output destinations
type OutputManager struct {
	outputs map[string]Output
	mu      sync.RWMutex
}

func NewOutputManager() *OutputManager {
	return &OutputManager{
		outputs: make(map[string]Output),
	}
}

func (om *OutputManager) AddOutput(name string, output Output) error {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.outputs[name] = output
	return nil
}

func (om *OutputManager) RemoveOutput(name string) error {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	if output, exists := om.outputs[name]; exists {
		output.Close()
		delete(om.outputs, name)
	}
	return nil
}

func (om *OutputManager) WriteToOutput(name string, entry *core.LogEntry) error {
	om.mu.RLock()
	output, exists := om.outputs[name]
	om.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("output %s not found", name)
	}
	
	return output.Write(entry)
}

func (om *OutputManager) WriteToAll(entry *core.LogEntry) error {
	om.mu.RLock()
	outputs := make([]Output, 0, len(om.outputs))
	for _, output := range om.outputs {
		outputs = append(outputs, output)
	}
	om.mu.RUnlock()
	
	var errors []error
	for _, output := range outputs {
		if err := output.Write(entry); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to write to %d outputs: %v", len(errors), errors)
	}
	
	return nil
}

func (om *OutputManager) Close() error {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	var errors []error
	for name, output := range om.outputs {
		if err := output.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close output %s: %w", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to close outputs: %v", errors)
	}
	
	return nil
}

// ConsoleOutput writes logs to console with optional colors
type ConsoleOutput struct {
	formatter formatters.Formatter
	colors    bool
	mu        sync.Mutex
}

func NewConsoleOutput(formatter formatters.Formatter, colors bool) *ConsoleOutput {
	return &ConsoleOutput{
		formatter: formatter,
		colors:    colors,
	}
}

func (co *ConsoleOutput) Write(entry *core.LogEntry) error {
	co.mu.Lock()
	defer co.mu.Unlock()
	
	formatted, err := co.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}
	
	var output string
	if co.colors {
		output = co.addColor(entry.Level, formatted)
	} else {
		output = formatted
	}
	
	// Write to appropriate stream based on level
	if entry.Level >= core.ErrorLevel {
		fmt.Fprintf(os.Stderr, "%s\n", output)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	
	return nil
}

func (co *ConsoleOutput) addColor(level core.Level, message string) string {
	if !co.colors {
		return message
	}
	
	var colorCode string
	switch level {
	case core.DebugLevel:
		colorCode = "\033[36m" // Cyan
	case core.InfoLevel:
		colorCode = "\033[32m" // Green
	case core.WarnLevel:
		colorCode = "\033[33m" // Yellow
	case core.ErrorLevel:
		colorCode = "\033[31m" // Red
	case core.FatalLevel:
		colorCode = "\033[35m" // Magenta
	default:
		return message
	}
	
	return fmt.Sprintf("%s%s\033[0m", colorCode, message)
}

func (co *ConsoleOutput) Close() error {
	return nil // Nothing to close for console output
}

// SimpleFileOutput writes logs to a file (basic implementation)
type SimpleFileOutput struct {
	formatter formatters.Formatter
	file      *os.File
	mu        sync.Mutex
}

func NewSimpleFileOutput(formatter formatters.Formatter, filename string) (*SimpleFileOutput, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	return &SimpleFileOutput{
		formatter: formatter,
		file:      file,
	}, nil
}

func (sfo *SimpleFileOutput) Write(entry *core.LogEntry) error {
	sfo.mu.Lock()
	defer sfo.mu.Unlock()
	
	formatted, err := sfo.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}
	
	_, err = fmt.Fprintf(sfo.file, "%s\n", formatted)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	
	return sfo.file.Sync()
}

func (sfo *SimpleFileOutput) Close() error {
	sfo.mu.Lock()
	defer sfo.mu.Unlock()
	
	if sfo.file != nil {
		return sfo.file.Close()
	}
	return nil
}

// MultiOutput writes to multiple outputs
type MultiOutput struct {
	outputs []Output
	mu      sync.RWMutex
}

func NewMultiOutput(outputs ...Output) *MultiOutput {
	return &MultiOutput{
		outputs: outputs,
	}
}

func (mo *MultiOutput) AddOutput(output Output) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.outputs = append(mo.outputs, output)
}

func (mo *MultiOutput) Write(entry *core.LogEntry) error {
	mo.mu.RLock()
	outputs := make([]Output, len(mo.outputs))
	copy(outputs, mo.outputs)
	mo.mu.RUnlock()
	
	var errors []error
	for _, output := range outputs {
		if err := output.Write(entry); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to write to %d outputs: %v", len(errors), errors)
	}
	
	return nil
}

func (mo *MultiOutput) Close() error {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	
	var errors []error
	for _, output := range mo.outputs {
		if err := output.Close(); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d outputs: %v", len(errors), errors)
	}
	
	return nil
}

// FilteredOutput wraps an output with level filtering
type FilteredOutput struct {
	output   Output
	minLevel core.Level
}

func NewFilteredOutput(output Output, minLevel core.Level) *FilteredOutput {
	return &FilteredOutput{
		output:   output,
		minLevel: minLevel,
	}
}

func (fo *FilteredOutput) Write(entry *core.LogEntry) error {
	if entry.Level < fo.minLevel {
		return nil // Skip this entry
	}
	return fo.output.Write(entry)
}

func (fo *FilteredOutput) Close() error {
	return fo.output.Close()
}

// BufferedOutput buffers log entries and flushes periodically
type BufferedOutput struct {
	output  Output
	buffer  []*core.LogEntry
	maxSize int
	mu      sync.Mutex
}

func NewBufferedOutput(output Output, maxSize int) *BufferedOutput {
	return &BufferedOutput{
		output:  output,
		buffer:  make([]*core.LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (bo *BufferedOutput) Write(entry *core.LogEntry) error {
	bo.mu.Lock()
	defer bo.mu.Unlock()
	
	bo.buffer = append(bo.buffer, entry)
	
	if len(bo.buffer) >= bo.maxSize {
		return bo.flushLocked()
	}
	
	return nil
}

func (bo *BufferedOutput) Flush() error {
	bo.mu.Lock()
	defer bo.mu.Unlock()
	return bo.flushLocked()
}

func (bo *BufferedOutput) flushLocked() error {
	if len(bo.buffer) == 0 {
		return nil
	}
	
	var errors []error
	for _, entry := range bo.buffer {
		if err := bo.output.Write(entry); err != nil {
			errors = append(errors, err)
		}
	}
	
	bo.buffer = bo.buffer[:0] // Clear buffer
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to flush %d entries: %v", len(errors), errors)
	}
	
	return nil
}

func (bo *BufferedOutput) Close() error {
	if err := bo.Flush(); err != nil {
		return err
	}
	return bo.output.Close()
}