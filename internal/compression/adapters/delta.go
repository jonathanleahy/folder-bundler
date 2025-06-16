package adapters

import (
	"fmt"
	"strings"
)

// DeltaCompression implements delta-based compression
// It stores files as differences from a base file or previous similar file
type DeltaCompression struct {
	minSimilarity float64
}

// NewDeltaCompression creates a new delta compression strategy
func NewDeltaCompression() *DeltaCompression {
	return &DeltaCompression{
		minSimilarity: 0.5, // 50% similarity threshold
	}
}

// Name returns the strategy name
func (d *DeltaCompression) Name() string {
	return "delta"
}

// DeltaOperation represents a single delta operation
type DeltaOperation struct {
	Op    string // "keep", "add", "remove", "replace"
	Line  int    // Line number (1-based)
	Count int    // Number of lines affected
	Text  string // Text for add/replace operations
}

// Compress compresses content using delta encoding
func (d *DeltaCompression) Compress(content []byte) ([]byte, string, error) {
	text := string(content)
	lines := strings.Split(text, "\n")
	
	// For delta compression to work, we need to identify files within the content
	// and find similar ones to use as base files
	files := d.extractFiles(lines)
	
	
	if len(files) < 2 {
		// Not enough files for delta compression
		return content, "delta:0", nil
	}
	
	// Find the best base file and compress others as deltas
	var compressed strings.Builder
	compressed.WriteString("===DELTA_START===\n")
	
	savings := 0
	deltaCount := 0
	processedFiles := make(map[int]bool)
	
	// Process each file
	for i, file := range files {
		if processedFiles[i] {
			continue
		}
		
		// Find the best match for this file
		bestMatch := -1
		bestSimilarity := 0.0
		var bestOps []DeltaOperation
		
		// Only use files that have already been processed as base files
		for j := 0; j < i; j++ {
			candidate := files[j]
			
			similarity, ops := d.computeDelta(candidate.lines, file.lines)
			if similarity > bestSimilarity && similarity >= d.minSimilarity {
				bestSimilarity = similarity
				bestMatch = j
				bestOps = ops
			}
		}
		
		if bestMatch >= 0 && len(bestOps) > 0 {
			// Calculate if delta saves space
			deltaSize := d.estimateDeltaSize(bestOps)
			originalSize := len(strings.Join(file.lines, "\n"))
			
			if deltaSize < int(float64(originalSize)*0.8) { // Save at least 20%
				// Write base file info
				compressed.WriteString(fmt.Sprintf("BASE:%s\n", files[bestMatch].path))
				compressed.WriteString(fmt.Sprintf("TARGET:%s\n", file.path))
				
				// Write delta operations
				for _, op := range bestOps {
					switch op.Op {
					case "keep":
						compressed.WriteString(fmt.Sprintf("K%d,%d\n", op.Line, op.Count))
					case "add":
						compressed.WriteString(fmt.Sprintf("A%d\n", op.Line))
						for _, line := range strings.Split(op.Text, "\n") {
							compressed.WriteString(">" + line + "\n")
						}
					case "remove":
						compressed.WriteString(fmt.Sprintf("R%d,%d\n", op.Line, op.Count))
					case "replace":
						compressed.WriteString(fmt.Sprintf("C%d,%d\n", op.Line, op.Count))
						for _, line := range strings.Split(op.Text, "\n") {
							compressed.WriteString(">" + line + "\n")
						}
					}
				}
				compressed.WriteString("---\n")
				
				savings += originalSize - deltaSize
				deltaCount++
				processedFiles[i] = true
			}
		}
		
		// If not processed as delta, keep original
		if !processedFiles[i] {
			compressed.WriteString(fmt.Sprintf("FILE:%s\n", file.path))
			compressed.WriteString(strings.Join(file.lines, "\n"))
			compressed.WriteString("\n---\n")
			processedFiles[i] = true
		}
	}
	
	compressed.WriteString("===DELTA_END===\n")
	
	// Only use delta compression if we saved space
	if savings <= 0 {
		return content, "delta:0", nil
	}
	
	// For now, return the delta content directly
	// In a full implementation, we'd merge this with the file structure
	metadata := fmt.Sprintf("delta:%d", deltaCount)
	
	return []byte(compressed.String()), metadata, nil
}

// FileContent represents a file's content within the bundle
type FileContent struct {
	path      string
	startLine int
	endLine   int
	lines     []string
}

// extractFiles extracts individual files from the bundled content
func (d *DeltaCompression) extractFiles(lines []string) []FileContent {
	var files []FileContent
	var currentFile *FileContent
	inContent := false
	
	for i, line := range lines {
		if strings.HasPrefix(line, "## File: ") {
			if currentFile != nil && currentFile.lines != nil {
				currentFile.endLine = i - 1
				files = append(files, *currentFile)
			}
			
			path := strings.TrimPrefix(line, "## File: ")
			currentFile = &FileContent{
				path:      path,
				startLine: i,
				lines:     []string{},
			}
			inContent = false
		} else if line == "===FILE_CONTENT_START===" {
			inContent = true
		} else if line == "===FILE_CONTENT_END===" {
			inContent = false
		} else if inContent && currentFile != nil && line != "__CONTENT_END_MARKER__" {
			currentFile.lines = append(currentFile.lines, line)
		}
	}
	
	// Add the last file
	if currentFile != nil && currentFile.lines != nil {
		currentFile.endLine = len(lines) - 1
		files = append(files, *currentFile)
	}
	
	return files
}

// computeDelta computes the delta operations to transform base into target
func (d *DeltaCompression) computeDelta(base, target []string) (float64, []DeltaOperation) {
	// Simplified approach: just track changed lines
	var ops []DeltaOperation
	commonLines := 0
	
	minLen := len(base)
	if len(target) < minLen {
		minLen = len(target)
	}
	
	i := 0
	for i < minLen {
		if base[i] == target[i] {
			// Count consecutive matching lines
			keepStart := i
			keepCount := 0
			for i < minLen && base[i] == target[i] {
				keepCount++
				commonLines++
				i++
			}
			ops = append(ops, DeltaOperation{
				Op:    "keep",
				Line:  keepStart + 1,
				Count: keepCount,
			})
		} else {
			// Lines differ - replace with target line
			ops = append(ops, DeltaOperation{
				Op:    "replace",
				Line:  i + 1,
				Count: 1,
				Text:  target[i],
			})
			i++
		}
	}
	
	// Handle remaining lines
	if i < len(target) {
		// Add remaining target lines
		remainingLines := target[i:]
		ops = append(ops, DeltaOperation{
			Op:   "add",
			Line: i + 1,
			Text: strings.Join(remainingLines, "\n"),
		})
	} else if i < len(base) {
		// Remove remaining base lines
		ops = append(ops, DeltaOperation{
			Op:    "remove",
			Line:  i + 1,
			Count: len(base) - i,
		})
	}
	
	// Calculate similarity
	totalLines := max(len(base), len(target))
	similarity := float64(commonLines) / float64(totalLines)
	
	return similarity, ops
}

// estimateDeltaSize estimates the size of delta operations
func (d *DeltaCompression) estimateDeltaSize(ops []DeltaOperation) int {
	size := 0
	for _, op := range ops {
		switch op.Op {
		case "keep":
			size += 10 // K line,count
		case "add":
			size += 5 + len(op.Text) + strings.Count(op.Text, "\n")*2 // A line + >text
		case "remove":
			size += 10 // R line,count
		case "replace":
			size += 10 + len(op.Text) + strings.Count(op.Text, "\n")*2 // C line,count + >text
		}
	}
	return size
}

// reconstructWithDeltas reconstructs the content with delta compression
func (d *DeltaCompression) reconstructWithDeltas(originalLines []string, files []FileContent, deltaContent string) string {
	// This is a simplified version - in practice, we'd need to carefully
	// merge the delta content with the original structure
	return deltaContent
}

// CanCompress checks if content is suitable for delta compression
func (d *DeltaCompression) CanCompress(content []byte) bool {
	// Skip binary files
	for _, b := range content {
		if b == 0 {
			return false
		}
	}
	
	// Need multiple files for delta compression
	lines := strings.Split(string(content), "\n")
	files := d.extractFiles(lines)
	return len(files) >= 2
}

// EstimateRatio estimates compression ratio
func (d *DeltaCompression) EstimateRatio(content []byte) float64 {
	lines := strings.Split(string(content), "\n")
	files := d.extractFiles(lines)
	
	if len(files) < 2 {
		return 1.0
	}
	
	// Simple estimation based on file similarity
	totalSize := len(content)
	estimatedSize := totalSize
	
	// Check pairs of files for similarity
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			similarity, _ := d.computeDelta(files[i].lines, files[j].lines)
			if similarity >= d.minSimilarity {
				// Estimate savings
				savings := int(float64(len(strings.Join(files[j].lines, "\n"))) * similarity * 0.5)
				estimatedSize -= savings
				break // Each file can only be delta-compressed once
			}
		}
	}
	
	return float64(estimatedSize) / float64(totalSize)
}

// Decompress restores original content from delta-compressed data
func (d *DeltaCompression) Decompress(compressed []byte, metadata string) ([]byte, error) {
	text := string(compressed)
	
	// Extract delta section
	if !strings.HasPrefix(text, "===DELTA_START===\n") {
		return compressed, nil // No delta compression found
	}
	
	deltaEnd := strings.Index(text, "===DELTA_END===\n")
	if deltaEnd == -1 {
		return nil, fmt.Errorf("delta end marker not found")
	}
	
	deltaStart := len("===DELTA_START===\n")
	deltaContent := text[deltaStart:deltaEnd]
	
	// Parse and apply deltas
	var result strings.Builder
	sections := strings.Split(deltaContent, "---\n")
	
	// Keep track of processed files
	processedFiles := make(map[string][]string)
	
	for _, section := range sections {
		if section == "" {
			continue
		}
		
		lines := strings.Split(section, "\n")
		if len(lines) == 0 {
			continue
		}
		
		if strings.HasPrefix(lines[0], "FILE:") {
			// Direct file content
			path := strings.TrimPrefix(lines[0], "FILE:")
			fileContent := strings.Join(lines[1:], "\n")
			processedFiles[path] = strings.Split(fileContent, "\n")
		} else if strings.HasPrefix(lines[0], "BASE:") {
			// Delta-compressed file
			basePath := strings.TrimPrefix(lines[0], "BASE:")
			targetPath := strings.TrimPrefix(lines[1], "TARGET:")
			
			// Get base content
			baseContent, exists := processedFiles[basePath]
			if !exists {
				return nil, fmt.Errorf("base file not found: %s", basePath)
			}
			
			// Apply delta operations
			targetContent := d.applyDelta(baseContent, lines[2:])
			processedFiles[targetPath] = targetContent
		}
	}
	
	// Reconstruct the full content
	// This is simplified - in practice we'd need to maintain the original structure
	for path, content := range processedFiles {
		result.WriteString(fmt.Sprintf("## File: %s\n\n", path))
		result.WriteString("===FILE_CONTENT_START===\n")
		result.WriteString(strings.Join(content, "\n"))
		result.WriteString("\n__CONTENT_END_MARKER__\n")
		result.WriteString("===FILE_CONTENT_END===\n\n")
	}
	
	return []byte(result.String()), nil
}

// applyDelta applies delta operations to base content
func (d *DeltaCompression) applyDelta(base []string, ops []string) []string {
	var result []string
	baseIndex := 0
	
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		if op == "" {
			continue
		}
		
		switch op[0] {
		case 'K': // Keep
			// Parse K<line>,<count>
			var line, count int
			fmt.Sscanf(op, "K%d,%d", &line, &count)
			// Copy lines from base
			for j := 0; j < count && baseIndex < len(base); j++ {
				result = append(result, base[baseIndex])
				baseIndex++
			}
			
		case 'A': // Add
			// Parse A<line>
			i++ // Move to content lines
			for i < len(ops) && strings.HasPrefix(ops[i], ">") {
				result = append(result, strings.TrimPrefix(ops[i], ">"))
				i++
			}
			i-- // Back up one since loop will increment
			
		case 'R': // Remove
			// Parse R<line>,<count>
			var line, count int
			fmt.Sscanf(op, "R%d,%d", &line, &count)
			// Skip lines in base
			baseIndex += count
			
		case 'C': // Replace (Change)
			// Parse C<line>,<count>
			var line, count int
			fmt.Sscanf(op, "C%d,%d", &line, &count)
			// Skip lines in base
			baseIndex += count
			// Add replacement lines
			i++
			for i < len(ops) && strings.HasPrefix(ops[i], ">") {
				result = append(result, strings.TrimPrefix(ops[i], ">"))
				i++
			}
			i--
		}
	}
	
	// Add any remaining base lines
	for baseIndex < len(base) {
		result = append(result, base[baseIndex])
		baseIndex++
	}
	
	return result
}

// CanDecompress checks if metadata indicates delta compression
func (d *DeltaCompression) CanDecompress(metadata string) bool {
	return strings.HasPrefix(metadata, "delta:")
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}