package compression

import (
	"fmt"
	"sync"
)

// Registry manages available compression strategies
type Registry struct {
	strategies map[string]CompressionStrategy
	mu         sync.RWMutex
}

// NewRegistry creates a new compression strategy registry
func NewRegistry() *Registry {
	return &Registry{
		strategies: make(map[string]CompressionStrategy),
	}
}

// Register adds a new compression strategy to the registry
func (r *Registry) Register(strategy CompressionStrategy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := strategy.Name()
	if _, exists := r.strategies[name]; exists {
		return fmt.Errorf("strategy %s already registered", name)
	}
	
	r.strategies[name] = strategy
	return nil
}

// Get returns a compression strategy by name
func (r *Registry) Get(name string) (CompressionStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	strategy, exists := r.strategies[name]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", name)
	}
	
	return strategy, nil
}

// List returns all registered strategy names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		names = append(names, name)
	}
	
	return names
}

// SelectBest analyzes content and selects the best compression strategy
func (r *Registry) SelectBest(content []byte) (CompressionStrategy, float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var bestStrategy CompressionStrategy
	bestRatio := 1.0
	
	// Always check "none" strategy first as baseline
	if noneStrategy, exists := r.strategies["none"]; exists {
		bestStrategy = noneStrategy
	}
	
	// Try each strategy and find the best compression ratio
	for _, strategy := range r.strategies {
		if strategy.Name() == "none" {
			continue // Skip none strategy in comparison
		}
		
		if strategy.CanCompress(content) {
			ratio := strategy.EstimateRatio(content)
			if ratio < bestRatio && ratio < 0.9 { // Only use if saves at least 10%
				bestStrategy = strategy
				bestRatio = ratio
			}
		}
	}
	
	return bestStrategy, bestRatio
}

// DefaultRegistry is the global registry instance
var DefaultRegistry = NewRegistry()