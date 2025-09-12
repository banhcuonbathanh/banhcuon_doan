// logger/core/output_manager.go - Default output manager implementation
package core

import (
	"fmt"
	"sync"
)

// Default output manager implementation
type defaultOutputManager struct {
	outputs map[string]Output
	mu      sync.RWMutex
}

func NewDefaultOutputManager() OutputManager {
	manager := &defaultOutputManager{
		outputs: make(map[string]Output),
	}
	
	// Add a rich console output by default
	consoleOutput := NewRichConsoleOutput(true) // Enable colors
	manager.AddOutput("console", consoleOutput)
	
	return manager
}

func (dom *defaultOutputManager) AddOutput(name string, output Output) error {
	dom.mu.Lock()
	defer dom.mu.Unlock()
	dom.outputs[name] = output
	return nil
}

func (dom *defaultOutputManager) RemoveOutput(name string) error {
	dom.mu.Lock()
	defer dom.mu.Unlock()
	
	if output, exists := dom.outputs[name]; exists {
		output.Close()
		delete(dom.outputs, name)
	}
	return nil
}

func (dom *defaultOutputManager) WriteToOutput(name string, entry *LogEntry) error {
	dom.mu.RLock()
	output, exists := dom.outputs[name]
	dom.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("output %s not found", name)
	}
	
	return output.Write(entry)
}

func (dom *defaultOutputManager) WriteToAll(entry *LogEntry) error {
	dom.mu.RLock()
	outputs := make([]Output, 0, len(dom.outputs))
	for _, output := range dom.outputs {
		outputs = append(outputs, output)
	}
	dom.mu.RUnlock()
	
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

func (dom *defaultOutputManager) Close() error {
	dom.mu.Lock()
	defer dom.mu.Unlock()
	
	var errors []error
	for name, output := range dom.outputs {
		if err := output.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close output %s: %w", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to close outputs: %v", errors)
	}
	
	return nil
}