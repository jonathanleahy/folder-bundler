package adapters

import (
	"fmt"
	"strings"
)

// localStrategy interface to avoid import cycle
type localStrategy interface {
	Name() string
	Compress(content []byte) ([]byte, string, error)
	CanCompress(content []byte) bool
	EstimateRatio(content []byte) float64
}

// localDecompressor interface to avoid import cycle
type localDecompressor interface {
	Decompress(compressed []byte, metadata string) ([]byte, error)
	CanDecompress(metadata string) bool
}

// CombinedCompression implements layered compression using multiple strategies
type CombinedCompression struct {
	strategies []localStrategy
}

// NewCombinedCompression creates a new combined compression strategy
func NewCombinedCompression(strategies ...localStrategy) *CombinedCompression {
	return &CombinedCompression{
		strategies: strategies,
	}
}

// NewTemplateDeltaCompression creates a template+delta combined strategy
func NewTemplateDeltaCompression() *CombinedCompression {
	return &CombinedCompression{
		strategies: []localStrategy{
			NewTemplateCompression(),
			NewDeltaCompression(),
		},
	}
}

// Name returns the strategy name
func (c *CombinedCompression) Name() string {
	names := make([]string, len(c.strategies))
	for i, s := range c.strategies {
		names[i] = s.Name()
	}
	return strings.Join(names, "+")
}

// Compress applies multiple compression strategies in sequence
func (c *CombinedCompression) Compress(content []byte) ([]byte, string, error) {
	if len(c.strategies) == 0 {
		return content, "combined:0", nil
	}
	
	// Track compression results
	var layers []compressionLayer
	currentContent := content
	
	// Apply each compression strategy in sequence
	for i, strategy := range c.strategies {
		compressed, metadata, err := strategy.Compress(currentContent)
		if err != nil {
			return nil, "", fmt.Errorf("layer %d (%s) failed: %w", i, strategy.Name(), err)
		}
		
		// Check if compression was beneficial
		if len(compressed) < len(currentContent) {
			layers = append(layers, compressionLayer{
				strategy: strategy.Name(),
				metadata: metadata,
			})
			currentContent = compressed
		}
	}
	
	// If no layers provided benefit, return original
	if len(layers) == 0 {
		return content, "combined:0", nil
	}
	
	// Build combined output with layer information
	var result strings.Builder
	result.WriteString("===COMBINED_COMPRESSION_START===\n")
	result.WriteString(fmt.Sprintf("LAYERS:%d\n", len(layers)))
	
	for i, layer := range layers {
		result.WriteString(fmt.Sprintf("L%d:%s:%s\n", i+1, layer.strategy, layer.metadata))
	}
	
	result.WriteString("===COMBINED_CONTENT_START===\n")
	result.Write(currentContent)
	result.WriteString("\n===COMBINED_CONTENT_END===\n")
	
	metadata := fmt.Sprintf("combined:%d", len(layers))
	return []byte(result.String()), metadata, nil
}

// CanCompress checks if content is suitable for combined compression
func (c *CombinedCompression) CanCompress(content []byte) bool {
	// Content is suitable if at least one strategy can compress it
	for _, strategy := range c.strategies {
		if strategy.CanCompress(content) {
			return true
		}
	}
	return false
}

// EstimateRatio estimates compression ratio
func (c *CombinedCompression) EstimateRatio(content []byte) float64 {
	if len(c.strategies) == 0 {
		return 1.0
	}
	
	// Estimate by applying each strategy's ratio in sequence
	ratio := 1.0
	for _, strategy := range c.strategies {
		if strategy.CanCompress(content) {
			strategyRatio := strategy.EstimateRatio(content)
			ratio *= strategyRatio
		}
	}
	
	// Add overhead for layer metadata
	overhead := float64(100 + 50*len(c.strategies)) / float64(len(content))
	return ratio + overhead
}

// Decompress restores original content from combined compressed data
func (c *CombinedCompression) Decompress(compressed []byte, metadata string) ([]byte, error) {
	text := string(compressed)
	
	// Check for combined compression markers
	if !strings.HasPrefix(text, "===COMBINED_COMPRESSION_START===\n") {
		return compressed, nil
	}
	
	// Parse layer information
	lines := strings.Split(text, "\n")
	var layers []compressionLayer
	var contentStart int
	
	for i, line := range lines {
		if strings.HasPrefix(line, "LAYERS:") {
			// Parse number of layers
			var numLayers int
			fmt.Sscanf(line, "LAYERS:%d", &numLayers)
		} else if strings.HasPrefix(line, "L") {
			// Parse layer info: L1:strategy:metadata
			parts := strings.SplitN(line, ":", 3)
			if len(parts) == 3 {
				layers = append(layers, compressionLayer{
					strategy: parts[1],
					metadata: parts[2],
				})
			}
		} else if line == "===COMBINED_CONTENT_START===" {
			contentStart = i + 1
			break
		}
	}
	
	// Extract compressed content
	contentEnd := -1
	for i := contentStart; i < len(lines); i++ {
		if lines[i] == "===COMBINED_CONTENT_END===" {
			contentEnd = i
			break
		}
	}
	
	if contentEnd == -1 {
		return nil, fmt.Errorf("combined content end marker not found")
	}
	
	compressedContent := []byte(strings.Join(lines[contentStart:contentEnd], "\n"))
	
	// Decompress in reverse order
	for i := len(layers) - 1; i >= 0; i-- {
		layer := layers[i]
		
		// Find appropriate decompressor
		var decompressor localDecompressor
		switch layer.strategy {
		case "template":
			decompressor = NewTemplateCompression()
		case "delta":
			decompressor = NewDeltaCompression()
		case "dictionary":
			decompressor = NewDictionaryCompression()
		default:
			return nil, fmt.Errorf("unknown compression strategy: %s", layer.strategy)
		}
		
		// Decompress this layer
		decompressed, err := decompressor.Decompress(compressedContent, layer.metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress layer %d (%s): %w", i+1, layer.strategy, err)
		}
		
		compressedContent = decompressed
	}
	
	return compressedContent, nil
}

// CanDecompress checks if metadata indicates combined compression
func (c *CombinedCompression) CanDecompress(metadata string) bool {
	return strings.HasPrefix(metadata, "combined:")
}

// compressionLayer represents a single compression layer
type compressionLayer struct {
	strategy string
	metadata string
}