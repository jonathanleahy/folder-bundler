package compression

import (
	"fmt"
)

// Selector helps choose the best compression strategy for content
type Selector struct {
	registry *Registry
}

// NewSelector creates a new compression selector
func NewSelector(registry *Registry) *Selector {
	return &Selector{
		registry: registry,
	}
}

// CompressContent compresses content using the best available strategy
func (s *Selector) CompressContent(content []byte) (*CompressionResult, error) {
	return s.CompressContentWithStrategy(content, "auto")
}

// CompressContentWithStrategy compresses content using a specific strategy
func (s *Selector) CompressContentWithStrategy(content []byte, strategyName string) (*CompressionResult, error) {
	var strategy CompressionStrategy
	var err error
	
	if strategyName == "auto" {
		// Try auto-selection
		strategy, _ = s.registry.SelectBest(content)
		if strategy == nil {
			// Fallback to none
			strategy, err = s.registry.Get("none")
			if err != nil {
				return nil, fmt.Errorf("no compression strategies available")
			}
		}
	} else {
		// Use specific strategy
		strategy, err = s.registry.Get(strategyName)
		if err != nil {
			return nil, fmt.Errorf("strategy %s not available: %w", strategyName, err)
		}
	}
	
	// Compress with selected strategy
	compressed, metadata, err := strategy.Compress(content)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}
	
	// Calculate actual ratio
	actualRatio := float64(len(compressed)) / float64(len(content))
	
	// If compression made it larger, use none strategy
	if actualRatio >= 1.0 && strategy.Name() != "none" {
		noneStrategy, err := s.registry.Get("none")
		if err == nil {
			compressed = content
			metadata = "none"
			actualRatio = 1.0
			strategy = noneStrategy
		}
	}
	
	return &CompressionResult{
		Strategy:   strategy.Name(),
		Compressed: compressed,
		Metadata:   metadata,
		Ratio:      actualRatio,
	}, nil
}

// DecompressContent decompresses content using the appropriate strategy
func (s *Selector) DecompressContent(compressed []byte, metadata string) ([]byte, error) {
	// Find strategy that can handle this metadata
	strategies := s.registry.List()
	
	for _, name := range strategies {
		strategy, err := s.registry.Get(name)
		if err != nil {
			continue
		}
		
		if strategy.CanDecompress(metadata) {
			return strategy.Decompress(compressed, metadata)
		}
	}
	
	return nil, fmt.Errorf("no strategy found for metadata: %s", metadata)
}