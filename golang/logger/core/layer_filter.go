// logger/core/layer_filter.go - Layer filtering system
package core

import (
	"strings"
	"sync"
)

// LayerFilterConfig represents the configuration for layer filtering
type LayerFilterConfig struct {
	// FilterMode determines how the filter works
	// "whitelist" - only allow specified layers
	// "blacklist" - block specified layers
	// "disabled" - no filtering (log all layers)
	FilterMode string `json:"filter_mode"`
	
	// Layers contains the list of layers for filtering
	Layers []string `json:"layers"`
	
	// LevelOverrides allows different log levels for specific layers
	// e.g., {"handler": "ERROR", "database": "DEBUG"}
	LevelOverrides map[string]Level `json:"level_overrides"`
}

// LayerFilter manages layer-based filtering for logging
type LayerFilter struct {
	config LayerFilterConfig
	mu     sync.RWMutex
}

// NewLayerFilter creates a new layer filter
func NewLayerFilter() *LayerFilter {
	return &LayerFilter{
		config: LayerFilterConfig{
			FilterMode:     "disabled",
			Layers:         []string{},
			LevelOverrides: make(map[string]Level),
		},
	}
}

// SetFilterMode sets the filtering mode
func (lf *LayerFilter) SetFilterMode(mode string) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	switch strings.ToLower(mode) {
	case "whitelist", "blacklist", "disabled":
		lf.config.FilterMode = strings.ToLower(mode)
	default:
		lf.config.FilterMode = "disabled"
	}
}

// EnableOnlyLayers enables logging only for specified layers (whitelist mode)
func (lf *LayerFilter) EnableOnlyLayers(layers ...string) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	lf.config.FilterMode = "whitelist"
	lf.config.Layers = make([]string, len(layers))
	copy(lf.config.Layers, layers)
}

// DisableLayers disables logging for specified layers (blacklist mode)
func (lf *LayerFilter) DisableLayers(layers ...string) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	lf.config.FilterMode = "blacklist"
	lf.config.Layers = make([]string, len(layers))
	copy(lf.config.Layers, layers)
}

// EnableAllLayers enables logging for all layers (no filtering)
func (lf *LayerFilter) EnableAllLayers() {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	lf.config.FilterMode = "disabled"
	lf.config.Layers = []string{}
}

// SetLayerLevel sets a specific log level for a layer
func (lf *LayerFilter) SetLayerLevel(layer string, level Level) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	if lf.config.LevelOverrides == nil {
		lf.config.LevelOverrides = make(map[string]Level)
	}
	lf.config.LevelOverrides[layer] = level
}

// RemoveLayerLevel removes level override for a layer
func (lf *LayerFilter) RemoveLayerLevel(layer string) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	delete(lf.config.LevelOverrides, layer)
}

// ClearLayerLevels clears all layer-specific level overrides
func (lf *LayerFilter) ClearLayerLevels() {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	lf.config.LevelOverrides = make(map[string]Level)
}

// ShouldLog determines if a log entry should be output based on layer filtering
func (lf *LayerFilter) ShouldLog(layer string, level Level) bool {
	lf.mu.RLock()
	defer lf.mu.RUnlock()
	
	// Check layer-specific level override first
	if layerLevel, exists := lf.config.LevelOverrides[layer]; exists {
		if level < layerLevel {
			return false
		}
	}
	
	// Apply layer filtering based on mode
	switch lf.config.FilterMode {
	case "whitelist":
		return lf.isLayerInList(layer)
	case "blacklist":
		return !lf.isLayerInList(layer)
	case "disabled":
		return true
	default:
		return true
	}
}

// isLayerInList checks if a layer is in the configured layer list
func (lf *LayerFilter) isLayerInList(layer string) bool {
	for _, l := range lf.config.Layers {
		if l == layer {
			return true
		}
	}
	return false
}

// GetConfig returns a copy of the current filter configuration
func (lf *LayerFilter) GetConfig() LayerFilterConfig {
	lf.mu.RLock()
	defer lf.mu.RUnlock()
	
	config := LayerFilterConfig{
		FilterMode:     lf.config.FilterMode,
		Layers:         make([]string, len(lf.config.Layers)),
		LevelOverrides: make(map[string]Level),
	}
	
	copy(config.Layers, lf.config.Layers)
	for k, v := range lf.config.LevelOverrides {
		config.LevelOverrides[k] = v
	}
	
	return config
}

// SetConfig sets the filter configuration
func (lf *LayerFilter) SetConfig(config LayerFilterConfig) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	
	lf.config = LayerFilterConfig{
		FilterMode:     config.FilterMode,
		Layers:         make([]string, len(config.Layers)),
		LevelOverrides: make(map[string]Level),
	}
	
	copy(lf.config.Layers, config.Layers)
	for k, v := range config.LevelOverrides {
		lf.config.LevelOverrides[k] = v
	}
}

// GetFilterStatus returns a human-readable status of the filter
func (lf *LayerFilter) GetFilterStatus() string {
	lf.mu.RLock()
	defer lf.mu.RUnlock()
	
	var status strings.Builder
	
	switch lf.config.FilterMode {
	case "whitelist":
		status.WriteString("Whitelist mode - Only logging layers: ")
		status.WriteString(strings.Join(lf.config.Layers, ", "))
	case "blacklist":
		status.WriteString("Blacklist mode - Blocking layers: ")
		status.WriteString(strings.Join(lf.config.Layers, ", "))
	case "disabled":
		status.WriteString("Filter disabled - Logging all layers")
	}
	
	if len(lf.config.LevelOverrides) > 0 {
		status.WriteString("\nLayer level overrides:")
		for layer, level := range lf.config.LevelOverrides {
			status.WriteString("\n  ")
			status.WriteString(layer)
			status.WriteString(": ")
			status.WriteString(level.String())
		}
	}
	
	return status.String()
}