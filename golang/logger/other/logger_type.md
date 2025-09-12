// logger/types.go - Additional types and interfaces (non-core)
package logger

import (
	"english-ai-full/logger/core"
	"fmt"
	"sync"
)

// LogBuffer interface for async processing (additional implementation)
type LogBuffer interface {
	Add(entry *core.LogEntry) error
	Flush() error
	Close() error
}

// SimpleLogBuffer - basic in-memory buffer implementation
type SimpleLogBuffer struct {
	entries []*core.LogEntry
	maxSize int
	mu      sync.Mutex
}

func NewSimpleLogBuffer(maxSize int) *SimpleLogBuffer {
	return &SimpleLogBuffer{
		entries: make([]*core.LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (slb *SimpleLogBuffer) Add(entry *core.LogEntry) error {
	slb.mu.Lock()
	defer slb.mu.Unlock()
	
	if len(slb.entries) >= slb.maxSize {
		// Remove oldest entry to make room
		slb.entries = slb.entries[1:]
	}
	
	slb.entries = append(slb.entries, entry)
	return nil
}

func (slb *SimpleLogBuffer) Flush() error {
	slb.mu.Lock()
	defer slb.mu.Unlock()
	
	// In a real implementation, you would flush entries to storage
	// For now, just clear the buffer
	slb.entries = slb.entries[:0]
	return nil
}

func (slb *SimpleLogBuffer) Close() error {
	return slb.Flush()
}

// ChannelLogBuffer - channel-based buffer for async processing
type ChannelLogBuffer struct {
	entries   chan *core.LogEntry
	done      chan struct{}
	processor func(*core.LogEntry) error
}

func NewChannelLogBuffer(bufferSize int, processor func(*core.LogEntry) error) *ChannelLogBuffer {
	clb := &ChannelLogBuffer{
		entries:   make(chan *core.LogEntry, bufferSize),
		done:      make(chan struct{}),
		processor: processor,
	}
	
	// Start processing goroutine
	go clb.process()
	
	return clb
}

func (clb *ChannelLogBuffer) Add(entry *core.LogEntry) error {
	select {
	case clb.entries <- entry:
		return nil
	case <-clb.done:
		return fmt.Errorf("buffer is closed")
	default:
		return fmt.Errorf("buffer is full")
	}
}

func (clb *ChannelLogBuffer) process() {
	for {
		select {
		case entry := <-clb.entries:
			if clb.processor != nil {
				clb.processor(entry)
			}
		case <-clb.done:
			// Process remaining entries
			for {
				select {
				case entry := <-clb.entries:
					if clb.processor != nil {
						clb.processor(entry)
					}
				default:
					return
				}
			}
		}
	}
}

func (clb *ChannelLogBuffer) Flush() error {
	// For channel-based buffer, flushing means waiting for processing
	// In a real implementation, you might want to add a flush signal
	return nil
}

func (clb *ChannelLogBuffer) Close() error {
	close(clb.done)
	close(clb.entries)
	return nil
}

// Additional utility types for advanced logging scenarios

// LogFilter interface for filtering log entries
type LogFilter interface {
	ShouldLog(entry *core.LogEntry) bool
}

// LevelFilter filters by log level
type LevelFilter struct {
	minLevel core.Level
}

func NewLevelFilter(minLevel core.Level) *LevelFilter {
	return &LevelFilter{minLevel: minLevel}
}

func (lf *LevelFilter) ShouldLog(entry *core.LogEntry) bool {
	return entry.Level >= lf.minLevel
}

// ComponentFilter filters by component
type ComponentFilter struct {
	allowedComponents map[string]bool
}

func NewComponentFilter(components ...string) *ComponentFilter {
	allowed := make(map[string]bool)
	for _, component := range components {
		allowed[component] = true
	}
	return &ComponentFilter{allowedComponents: allowed}
}

func (cf *ComponentFilter) ShouldLog(entry *core.LogEntry) bool {
	return cf.allowedComponents[entry.Component]
}

// LayerFilter filters by layer
type LayerFilter struct {
	allowedLayers map[string]bool
}

func NewLayerFilter(layers ...string) *LayerFilter {
	allowed := make(map[string]bool)
	for _, layer := range layers {
		allowed[layer] = true
	}
	return &LayerFilter{allowedLayers: allowed}
}

func (lf *LayerFilter) ShouldLog(entry *core.LogEntry) bool {
	return lf.allowedLayers[entry.Layer]
}

// CompositeFilter combines multiple filters
type CompositeFilter struct {
	filters []LogFilter
	mode    FilterMode
}

type FilterMode int

const (
	FilterModeAND FilterMode = iota // All filters must pass
	FilterModeOR                    // At least one filter must pass
)

func NewCompositeFilter(mode FilterMode, filters ...LogFilter) *CompositeFilter {
	return &CompositeFilter{
		filters: filters,
		mode:    mode,
	}
}

func (cf *CompositeFilter) ShouldLog(entry *core.LogEntry) bool {
	if len(cf.filters) == 0 {
		return true
	}
	
	switch cf.mode {
	case FilterModeAND:
		for _, filter := range cf.filters {
			if !filter.ShouldLog(entry) {
				return false
			}
		}
		return true
	case FilterModeOR:
		for _, filter := range cf.filters {
			if filter.ShouldLog(entry) {
				return true
			}
		}
		return false
	default:
		return true
	}
}