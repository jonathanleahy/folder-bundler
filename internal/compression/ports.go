package compression

// Compressor defines the interface for compressing content
type Compressor interface {
	// Name returns the name of this compression strategy
	Name() string
	
	// Compress attempts to compress the content
	// Returns compressed content, metadata needed for decompression, and any error
	Compress(content []byte) (compressed []byte, metadata string, err error)
	
	// CanCompress checks if this strategy can compress the given content
	CanCompress(content []byte) bool
	
	// EstimateRatio estimates the compression ratio (compressed/original)
	// Returns value between 0 and 1, where lower is better compression
	EstimateRatio(content []byte) float64
}

// Decompressor defines the interface for decompressing content
type Decompressor interface {
	// Decompress restores the original content using the compressed data and metadata
	Decompress(compressed []byte, metadata string) ([]byte, error)
	
	// CanDecompress checks if this strategy can decompress based on metadata
	CanDecompress(metadata string) bool
}

// CompressionStrategy combines both compression and decompression capabilities
type CompressionStrategy interface {
	Compressor
	Decompressor
}

// CompressionResult holds the result of a compression attempt
type CompressionResult struct {
	Strategy   string
	Compressed []byte
	Metadata   string
	Ratio      float64
}