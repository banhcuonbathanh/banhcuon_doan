package core

var DefaultBlockedLayers = []string{
	LayerDatabase,
	LayerCache,
	LayerValidation,
		// LayerHandler,
		// LayerService,
		// LayerRepository,
			LayerConfig,
}


// setDefaultBlockedLayers initializes the blocked layers with default values
func (l *CoreLogger) setDefaultBlockedLayers() {
	for _, layer := range DefaultBlockedLayers {
		l.blockedLayers[layer] = true
	}
}

// Layer blocking management methods

// BlockLayer blocks logging for a specific layer
func (l *CoreLogger) BlockLayer(layer string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.blockedLayers[layer] = true
}

// UnblockLayer allows logging for a specific layer
func (l *CoreLogger) UnblockLayer(layer string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.blockedLayers, layer)
}

// BlockLayers blocks multiple layers at once
func (l *CoreLogger) BlockLayers(layers []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, layer := range layers {
		l.blockedLayers[layer] = true
	}
}

// UnblockLayers unblocks multiple layers at once
func (l *CoreLogger) UnblockLayers(layers []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, layer := range layers {
		delete(l.blockedLayers, layer)
	}
}

// UnblockAllLayers removes all layer blocks
func (l *CoreLogger) UnblockAllLayers() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.blockedLayers = make(map[string]bool)
}

// IsLayerBlocked checks if a layer is currently blocked
func (l *CoreLogger) IsLayerBlocked(layer string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.blockedLayers[layer]
}

// GetBlockedLayers returns a list of currently blocked layers
func (l *CoreLogger) GetBlockedLayers() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	var blocked []string
	for layer := range l.blockedLayers {
		blocked = append(blocked, layer)
	}
	return blocked
}

// isLayerAllowed checks if logging is allowed for the current layer
func (l *CoreLogger) isLayerAllowed() bool {
	// If no layer is set, allow logging
	if l.layer == "" {
		return true
	}
	
	// Check if current layer is blocked
	return !l.blockedLayers[l.layer]
}
// Convenience methods for quick layer management

// EnableDebugLayers unblocks commonly needed layers for debugging
func (l *CoreLogger) EnableDebugLayers() {
	debugLayers := []string{LayerDatabase, LayerValidation, LayerRepository}
	l.UnblockLayers(debugLayers)
}

// EnableProductionLayers blocks verbose layers for production
func (l *CoreLogger) EnableProductionLayers() {
	verboseLayers := []string{LayerDatabase, LayerCache, LayerValidation}
	l.BlockLayers(verboseLayers)
}

// EnableAllLayers unblocks all layers (useful for debugging)
func (l *CoreLogger) EnableAllLayers() {
	l.UnblockAllLayers()
}