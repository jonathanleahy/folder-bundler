package adapters

import "strings"

// NoneCompression is a passthrough strategy that doesn't compress
type NoneCompression struct{}

// NewNoneCompression creates a new no-compression strategy
func NewNoneCompression() *NoneCompression {
	return &NoneCompression{}
}

// Name returns the strategy name
func (n *NoneCompression) Name() string {
	return "none"
}

// Compress returns the content unchanged
func (n *NoneCompression) Compress(content []byte) ([]byte, string, error) {
	return content, "none", nil
}

// CanCompress always returns true as this handles all content
func (n *NoneCompression) CanCompress(content []byte) bool {
	return true
}

// EstimateRatio always returns 1.0 (no compression)
func (n *NoneCompression) EstimateRatio(content []byte) float64 {
	return 1.0
}

// Decompress returns the content unchanged
func (n *NoneCompression) Decompress(compressed []byte, metadata string) ([]byte, error) {
	return compressed, nil
}

// CanDecompress checks if metadata indicates no compression
func (n *NoneCompression) CanDecompress(metadata string) bool {
	return metadata == "none" || metadata == "" || !strings.Contains(metadata, ":")
}