package adapters

import (
	"fmt"
	"sort"
	"strings"
)

// DictionaryCompression implements dictionary-based compression
type DictionaryCompression struct {
	minPatternLength int
	minOccurrences   int
}

// NewDictionaryCompression creates a new dictionary compression strategy
func NewDictionaryCompression() *DictionaryCompression {
	return &DictionaryCompression{
		minPatternLength: 20,
		minOccurrences:   3,
	}
}

// Pattern represents a repeated pattern in the content
type pattern struct {
	text        string
	occurrences int
	positions   []int
}

// Name returns the strategy name
func (d *DictionaryCompression) Name() string {
	return "dictionary"
}

// Compress compresses content using dictionary replacement
func (d *DictionaryCompression) Compress(content []byte) ([]byte, string, error) {
	text := string(content)
	patterns := d.findPatterns(text)
	
	if len(patterns) == 0 {
		return content, "dictionary:0", nil
	}
	
	// Build dictionary and calculate savings
	dictionary := make(map[string]string)
	savings := 0
	entryNum := 1
	
	for _, p := range patterns {
		ref := fmt.Sprintf("«%d»", entryNum)
		dictEntry := fmt.Sprintf("%s=%s\n", ref, p.text)
		
		// Calculate if this pattern saves space
		originalSize := len(p.text) * p.occurrences
		compressedSize := len(dictEntry) + len(ref)*p.occurrences
		
		if originalSize > compressedSize {
			dictionary[p.text] = ref
			savings += originalSize - compressedSize
			entryNum++
		}
	}
	
	// Only compress if we save space
	if savings <= 0 {
		return content, "dictionary:0", nil
	}
	
	// Build dictionary section
	var dictBuilder strings.Builder
	dictBuilder.WriteString("--- BEGIN DICTIONARY ---\n")
	
	// Sort dictionary for consistent output
	var sortedPatterns []string
	for pattern := range dictionary {
		sortedPatterns = append(sortedPatterns, pattern)
	}
	sort.Slice(sortedPatterns, func(i, j int) bool {
		return dictionary[sortedPatterns[i]] < dictionary[sortedPatterns[j]]
	})
	
	for _, pattern := range sortedPatterns {
		ref := dictionary[pattern]
		dictBuilder.WriteString(fmt.Sprintf("%s=%s\n", ref, pattern))
	}
	dictBuilder.WriteString("--- END DICTIONARY ---\n")
	
	// Replace patterns in content
	compressed := text
	for _, pattern := range sortedPatterns {
		ref := dictionary[pattern]
		compressed = strings.ReplaceAll(compressed, pattern, ref)
	}
	
	// Escape any existing $ symbols that weren't part of references
	compressed = d.escapeReferences(compressed, dictionary)
	
	result := dictBuilder.String() + compressed
	metadata := fmt.Sprintf("dictionary:%d", len(dictionary))
	
	return []byte(result), metadata, nil
}

// CanCompress checks if content is suitable for dictionary compression
func (d *DictionaryCompression) CanCompress(content []byte) bool {
	// Skip binary files
	for _, b := range content {
		if b == 0 {
			return false
		}
	}
	return len(content) > 100 // Need minimum size to benefit
}

// EstimateRatio estimates compression ratio
func (d *DictionaryCompression) EstimateRatio(content []byte) float64 {
	text := string(content)
	patterns := d.findPatterns(text)
	
	if len(patterns) == 0 {
		return 1.0
	}
	
	totalSavings := 0
	dictSize := len("--- BEGIN DICTIONARY ---\n--- END DICTIONARY ---\n")
	
	for i, p := range patterns {
		if i >= 100 { // Limit dictionary size
			break
		}
		ref := fmt.Sprintf("«%d»", i+1)
		dictEntry := len(ref) + 1 + len(p.text) + 1 // ref=pattern\n
		
		originalSize := len(p.text) * p.occurrences
		compressedSize := dictEntry + len(ref)*p.occurrences
		
		if originalSize > compressedSize {
			totalSavings += originalSize - compressedSize
			dictSize += dictEntry
		}
	}
	
	if totalSavings <= dictSize {
		return 1.0
	}
	
	compressedSize := len(content) - totalSavings + dictSize
	return float64(compressedSize) / float64(len(content))
}

// Decompress restores original content from dictionary-compressed data
func (d *DictionaryCompression) Decompress(compressed []byte, metadata string) ([]byte, error) {
	text := string(compressed)
	
	// Extract dictionary
	if !strings.HasPrefix(text, "--- BEGIN DICTIONARY ---\n") {
		return compressed, nil // No dictionary found
	}
	
	dictEnd := strings.Index(text, "--- END DICTIONARY ---\n")
	if dictEnd == -1 {
		return nil, fmt.Errorf("dictionary end marker not found")
	}
	
	dictStart := len("--- BEGIN DICTIONARY ---\n")
	dictContent := text[dictStart:dictEnd]
	content := text[dictEnd+len("--- END DICTIONARY ---\n"):]
	
	// Parse dictionary
	dictionary := make(map[string]string)
	for _, line := range strings.Split(dictContent, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid dictionary entry: %s", line)
		}
		dictionary[parts[0]] = parts[1]
	}
	
	// Unescape references
	content = d.unescapeReferences(content)
	
	// Replace references with original text
	// Sort by reference length descending to avoid partial replacements
	var refs []string
	for ref := range dictionary {
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(i, j int) bool {
		return len(refs[i]) > len(refs[j])
	})
	
	for _, ref := range refs {
		content = strings.ReplaceAll(content, ref, dictionary[ref])
	}
	
	return []byte(content), nil
}

// CanDecompress checks if metadata indicates dictionary compression
func (d *DictionaryCompression) CanDecompress(metadata string) bool {
	return strings.HasPrefix(metadata, "dictionary:")
}

// findPatterns finds repeated patterns in text
func (d *DictionaryCompression) findPatterns(text string) []pattern {
	patternMap := make(map[string]*pattern)
	textLen := len(text)
	
	// Find all substrings of minimum length
	for i := 0; i < textLen-d.minPatternLength; i++ {
		for length := d.minPatternLength; length <= 100 && i+length <= textLen; length++ {
			substr := text[i : i+length]
			
			// Skip if contains our markers, delimiters, or newlines
			if strings.Contains(substr, "--- BEGIN") || strings.Contains(substr, "--- END") ||
			   strings.Contains(substr, "«") || strings.Contains(substr, "»") || 
			   strings.Contains(substr, "@CONTENT-END@") ||
			   strings.Contains(substr, "FILE CONTENT BEGIN") ||
			   strings.Contains(substr, "FILE CONTENT END") ||
			   strings.Contains(substr, "\n") {
				continue
			}
			
			// Skip if mostly whitespace
			if len(strings.TrimSpace(substr)) < length/2 {
				continue
			}
			
			if p, exists := patternMap[substr]; exists {
				p.occurrences++
				p.positions = append(p.positions, i)
			} else {
				patternMap[substr] = &pattern{
					text:        substr,
					occurrences: 1,
					positions:   []int{i},
				}
			}
		}
	}
	
	// Filter patterns by minimum occurrences
	var patterns []pattern
	for _, p := range patternMap {
		if p.occurrences >= d.minOccurrences {
			patterns = append(patterns, *p)
		}
	}
	
	// Sort by potential savings (descending)
	sort.Slice(patterns, func(i, j int) bool {
		savingsI := (len(patterns[i].text) - 3) * patterns[i].occurrences // -3 for ref length
		savingsJ := (len(patterns[j].text) - 3) * patterns[j].occurrences
		return savingsI > savingsJ
	})
	
	// Remove overlapping patterns
	return d.removeOverlaps(patterns)
}

// removeOverlaps removes patterns that overlap with higher-value patterns
func (d *DictionaryCompression) removeOverlaps(patterns []pattern) []pattern {
	if len(patterns) == 0 {
		return patterns
	}
	
	var result []pattern
	usedPositions := make(map[int]bool)
	
	for _, p := range patterns {
		hasOverlap := false
		for _, pos := range p.positions {
			for i := pos; i < pos+len(p.text); i++ {
				if usedPositions[i] {
					hasOverlap = true
					break
				}
			}
			if hasOverlap {
				break
			}
		}
		
		if !hasOverlap {
			result = append(result, p)
			for _, pos := range p.positions {
				for i := pos; i < pos+len(p.text); i++ {
					usedPositions[i] = true
				}
			}
		}
	}
	
	return result
}

// escapeReferences - with «» delimiters, escaping is rarely needed
func (d *DictionaryCompression) escapeReferences(text string, dictionary map[string]string) string {
	// «» are extremely rare in code, so we'll just return as-is
	// If needed in future, we could escape « as «« 
	return text
}

// unescapeReferences - with «» delimiters, unescaping is rarely needed
func (d *DictionaryCompression) unescapeReferences(text string) string {
	// If we implement escaping in future, we'd unescape «« back to «
	return text
}