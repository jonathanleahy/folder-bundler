package adapters

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// TemplateCompression implements template-based compression
type TemplateCompression struct {
	minSimilarity float64
	minInstances  int
}

// NewTemplateCompression creates a new template compression strategy
func NewTemplateCompression() *TemplateCompression {
	return &TemplateCompression{
		minSimilarity: 0.8, // 80% similarity threshold
		minInstances:  3,   // minimum 3 instances to create template
	}
}

// Name returns the strategy name
func (t *TemplateCompression) Name() string {
	return "template"
}

// Template represents a parameterized pattern
type template struct {
	pattern     string
	params      []string
	instances   []map[string]string
	occurrences int
}

// Compress compresses content using template replacement
func (t *TemplateCompression) Compress(content []byte) ([]byte, string, error) {
	text := string(content)
	lines := strings.Split(text, "\n")
	
	// Find templates
	templates := t.findTemplates(lines)
	
	if len(templates) == 0 {
		return content, "template:0", nil
	}
	
	// Build template section and calculate savings
	var templateBuilder strings.Builder
	var compressed strings.Builder
	templateBuilder.WriteString("===TEMPLATES_START===\n")
	
	savings := 0
	templateNum := 1
	usedTemplates := make(map[int]bool)
	templateMap := make(map[string]int)
	
	// Process templates
	for idx, tmpl := range templates {
		ref := fmt.Sprintf("T%d", templateNum)
		
		// Calculate if this template saves space
		originalSize := len(tmpl.pattern) * tmpl.occurrences
		templateSize := len(fmt.Sprintf("%s=%s\n", ref, tmpl.pattern))
		instancesSize := 0
		
		for _, instance := range tmpl.instances {
			instanceLine := ref + "{"
			for i, param := range tmpl.params {
				if i > 0 {
					instanceLine += ","
				}
				instanceLine += instance[param]
			}
			instanceLine += "}\n"
			instancesSize += len(instanceLine)
		}
		
		if originalSize > templateSize+instancesSize {
			templateBuilder.WriteString(fmt.Sprintf("%s=%s\n", ref, tmpl.pattern))
			savings += originalSize - (templateSize + instancesSize)
			usedTemplates[idx] = true
			templateMap[tmpl.pattern] = templateNum
			templateNum++
		}
	}
	
	// Only compress if we save space
	if savings <= 0 {
		return content, "template:0", nil
	}
	
	templateBuilder.WriteString("===TEMPLATES_END===\n\n")
	
	// Create a map to track which lines belong to which instances
	lineToInstance := make(map[int]map[string]string)
	lineToTemplate := make(map[int]int)
	
	// Replace templates in content
	for i, tmpl := range templates {
		if !usedTemplates[i] {
			continue
		}
		
		// Check each instance for multi-line patterns
		for _, instance := range tmpl.instances {
			expandedPattern := t.expandTemplate(tmpl.pattern, instance)
			expandedLines := strings.Split(expandedPattern, "\n")
			
			// Find matching sequences in the original lines
			for lineNum := 0; lineNum <= len(lines)-len(expandedLines); lineNum++ {
				match := true
				for j, expLine := range expandedLines {
					if lines[lineNum+j] != expLine {
						match = false
						break
					}
				}
				
				if match {
					// Store the instance info for the first line
					lineToInstance[lineNum] = instance
					lineToTemplate[lineNum] = templateMap[tmpl.pattern]
					
					// Mark other lines for deletion
					for j := 1; j < len(expandedLines); j++ {
						lineToInstance[lineNum+j] = nil
					}
					
					// Move to next instance
					break
				}
			}
		}
	}
	
	// Build the compressed lines
	var compressedLines []string
	for i := 0; i < len(lines); i++ {
		if instance, exists := lineToInstance[i]; exists {
			if instance != nil {
				// This is the start of a template instance
				ref := fmt.Sprintf("T%d", lineToTemplate[i])
				instanceLine := ref + "{"
				
				// Get the params in order from the template
				tmplIdx := -1
				for j, tmpl := range templates {
					if usedTemplates[j] && templateMap[tmpl.pattern] == lineToTemplate[i] {
						tmplIdx = j
						break
					}
				}
				
				if tmplIdx >= 0 {
					for j, param := range templates[tmplIdx].params {
						if j > 0 {
							instanceLine += ","
						}
						instanceLine += instance[param]
					}
				}
				
				instanceLine += "}"
				compressedLines = append(compressedLines, instanceLine)
			}
			// else: this line is part of a multi-line template, skip it
		} else {
			// Regular line, keep as-is
			compressedLines = append(compressedLines, lines[i])
		}
	}
	
	// Build final compressed content
	compressed.WriteString(templateBuilder.String())
	compressed.WriteString(strings.Join(compressedLines, "\n"))
	
	metadata := fmt.Sprintf("template:%d", templateNum-1)
	return []byte(compressed.String()), metadata, nil
}

// CanCompress checks if content is suitable for template compression
func (t *TemplateCompression) CanCompress(content []byte) bool {
	// Skip binary files
	for _, b := range content {
		if b == 0 {
			return false
		}
	}
	return len(content) > 500 // Need minimum size to benefit
}

// EstimateRatio estimates compression ratio
func (t *TemplateCompression) EstimateRatio(content []byte) float64 {
	lines := strings.Split(string(content), "\n")
	templates := t.findTemplates(lines)
	
	if len(templates) == 0 {
		return 1.0
	}
	
	totalSavings := 0
	overhead := len("===TEMPLATES_START===\n===TEMPLATES_END===\n\n")
	
	for _, tmpl := range templates {
		if tmpl.occurrences >= t.minInstances {
			// Rough estimate of savings
			originalSize := len(tmpl.pattern) * tmpl.occurrences
			compressedSize := len(tmpl.pattern) + 10 + (20 * tmpl.occurrences) // rough estimate
			if originalSize > compressedSize {
				totalSavings += originalSize - compressedSize
				overhead += len(tmpl.pattern) + 10
			}
		}
	}
	
	if totalSavings <= overhead {
		return 1.0
	}
	
	return float64(len(content)-totalSavings+overhead) / float64(len(content))
}

// Decompress restores original content from template-compressed data
func (t *TemplateCompression) Decompress(compressed []byte, metadata string) ([]byte, error) {
	text := string(compressed)
	
	// Extract templates
	if !strings.HasPrefix(text, "===TEMPLATES_START===\n") {
		return compressed, nil // No templates found
	}
	
	templatesEnd := strings.Index(text, "===TEMPLATES_END===\n")
	if templatesEnd == -1 {
		return nil, fmt.Errorf("templates end marker not found")
	}
	
	templatesStart := len("===TEMPLATES_START===\n")
	templatesContent := text[templatesStart:templatesEnd]
	content := text[templatesEnd+len("===TEMPLATES_END===\n\n"):]
	
	// Parse templates
	templates := make(map[string]string)
	for _, line := range strings.Split(templatesContent, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid template entry: %s", line)
		}
		templates[parts[0]] = parts[1]
	}
	
	// Replace template instances
	lines := strings.Split(content, "\n")
	expandedLines := []string{}
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		
		// Check if this is a template instance
		if match := regexp.MustCompile(`^(T\d+)\{([^}]+)\}$`).FindStringSubmatch(line); match != nil {
			ref := match[1]
			paramsStr := match[2]
			
			if tmplPattern, exists := templates[ref]; exists {
				// Extract parameters
				params := strings.Split(paramsStr, ",")
				
				// Extract parameter names from template
				paramNames := t.extractParamNames(tmplPattern)
				
				if len(params) == len(paramNames) {
					// Build parameter map
					paramMap := make(map[string]string)
					for j, name := range paramNames {
						paramMap[name] = params[j]
					}
					
					// Expand template - it might be multi-line
					expanded := t.expandTemplate(tmplPattern, paramMap)
					expandedTemplateLines := strings.Split(expanded, "\n")
					
					// Add all lines from the expanded template
					for _, expLine := range expandedTemplateLines {
						expandedLines = append(expandedLines, expLine)
					}
				} else {
					// Mismatch in parameters, keep original
					expandedLines = append(expandedLines, line)
				}
			} else {
				// Template not found, keep original
				expandedLines = append(expandedLines, line)
			}
		} else {
			// Not a template instance, keep as-is
			expandedLines = append(expandedLines, line)
		}
	}
	
	return []byte(strings.Join(expandedLines, "\n")), nil
}

// CanDecompress checks if metadata indicates template compression
func (t *TemplateCompression) CanDecompress(metadata string) bool {
	return strings.HasPrefix(metadata, "template:")
}

// findTemplates finds similar patterns that can be templatized
func (t *TemplateCompression) findTemplates(lines []string) []template {
	var templates []template
	
	// Group similar lines
	groups := make(map[string][]int)
	for i, line := range lines {
		if len(line) < 20 { // Skip short lines
			continue
		}
		
		// Find lines with similar structure
		normalized := t.normalizeForGrouping(line)
		groups[normalized] = append(groups[normalized], i)
	}
	
	// Process groups to find templates
	for _, lineNums := range groups {
		if len(lineNums) < t.minInstances {
			continue
		}
		
		// Extract common pattern
		tmpl := t.extractTemplate(lines, lineNums)
		if tmpl != nil && len(tmpl.params) > 0 {
			templates = append(templates, *tmpl)
		}
	}
	
	// Sort by potential savings
	sort.Slice(templates, func(i, j int) bool {
		savingsI := (len(templates[i].pattern) - 20) * templates[i].occurrences
		savingsJ := (len(templates[j].pattern) - 20) * templates[j].occurrences
		return savingsI > savingsJ
	})
	
	return templates
}

// normalizeForGrouping creates a normalized version for grouping similar lines
func (t *TemplateCompression) normalizeForGrouping(line string) string {
	// Replace common variable parts with placeholders
	normalized := line
	
	// Replace quoted strings
	normalized = regexp.MustCompile(`"[^"]*"`).ReplaceAllString(normalized, `"?"`)
	normalized = regexp.MustCompile(`'[^']*'`).ReplaceAllString(normalized, `'?'`)
	
	// Replace numbers
	normalized = regexp.MustCompile(`\b\d+\b`).ReplaceAllString(normalized, "?")
	
	// Replace identifiers in common patterns
	normalized = regexp.MustCompile(`\b[a-zA-Z_]\w*\b`).ReplaceAllString(normalized, "?")
	
	return normalized
}

// extractTemplate extracts a template from similar lines
func (t *TemplateCompression) extractTemplate(lines []string, lineNums []int) *template {
	if len(lineNums) < 2 {
		return nil
	}
	
	// Use first two lines to find differences
	line1 := lines[lineNums[0]]
	line2 := lines[lineNums[1]]
	
	pattern, params := t.findDifferences(line1, line2)
	if len(params) == 0 {
		return nil
	}
	
	// Verify pattern works for all instances
	instances := []map[string]string{}
	for _, lineNum := range lineNums {
		instance := t.extractInstance(lines[lineNum], pattern, params)
		if instance != nil {
			instances = append(instances, instance)
		}
	}
	
	if len(instances) < t.minInstances {
		return nil
	}
	
	return &template{
		pattern:     pattern,
		params:      params,
		instances:   instances,
		occurrences: len(instances),
	}
}

// findDifferences finds the differences between two similar lines
func (t *TemplateCompression) findDifferences(line1, line2 string) (string, []string) {
	var pattern strings.Builder
	var params []string
	paramSet := make(map[string]bool)
	
	i, j := 0, 0
	for i < len(line1) && j < len(line2) {
		if line1[i] == line2[j] {
			pattern.WriteByte(line1[i])
			i++
			j++
		} else {
			// Find the end of the difference
			end1 := i
			end2 := j
			
			// Look ahead to find where they match again
			found := false
			for k := 1; k < 50 && i+k < len(line1); k++ {
				for l := 1; l < 50 && j+l < len(line2); l++ {
					if i+k < len(line1) && j+l < len(line2) && 
					   line1[i+k] == line2[j+l] && 
					   (i+k+5 >= len(line1) || j+l+5 >= len(line2) || 
					    line1[i+k:i+k+5] == line2[j+l:j+l+5]) {
						end1 = i + k
						end2 = j + l
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			
			if !found {
				// No common suffix found
				return "", nil
			}
			
			// Extract the different parts
			diff1 := line1[i:end1]
			diff2 := line2[j:end2]
			
			// Create parameter name based on content
			paramName := t.generateParamName(diff1, diff2)
			if !paramSet[paramName] {
				params = append(params, paramName)
				paramSet[paramName] = true
			}
			
			pattern.WriteString("{" + paramName + "}")
			i = end1
			j = end2
		}
	}
	
	// Add any remaining content
	if i < len(line1) {
		pattern.WriteString(line1[i:])
	}
	
	return pattern.String(), params
}

// generateParamName generates a parameter name based on the content
func (t *TemplateCompression) generateParamName(val1, val2 string) string {
	// Simple heuristics for parameter naming
	if strings.Contains(val1, "get") || strings.Contains(val2, "get") {
		return "method"
	}
	if strings.Contains(val1, "/") || strings.Contains(val2, "/") {
		return "path"
	}
	if regexp.MustCompile(`^\d+$`).MatchString(val1) {
		return "number"
	}
	if strings.HasPrefix(val1, "'") || strings.HasPrefix(val1, "\"") {
		return "string"
	}
	return "param"
}

// extractInstance extracts parameter values for a specific line
func (t *TemplateCompression) extractInstance(line, pattern string, params []string) map[string]string {
	instance := make(map[string]string)
	
	// Ensure both line and pattern are valid UTF-8
	if !utf8.ValidString(line) || !utf8.ValidString(pattern) {
		return nil
	}
	
	// Build regex from pattern
	regexPattern := regexp.QuoteMeta(pattern)
	for _, param := range params {
		regexPattern = strings.Replace(regexPattern, "\\{"+param+"\\}", "(.+?)", 1)
	}
	
	re, err := regexp.Compile("^" + regexPattern + "$")
	if err != nil {
		// Invalid regex pattern, skip this instance
		return nil
	}
	matches := re.FindStringSubmatch(line)
	
	if matches == nil || len(matches) != len(params)+1 {
		return nil
	}
	
	for i, param := range params {
		instance[param] = matches[i+1]
	}
	
	return instance
}

// expandTemplate expands a template with given parameters
func (t *TemplateCompression) expandTemplate(pattern string, params map[string]string) string {
	result := pattern
	for name, value := range params {
		result = strings.Replace(result, "{"+name+"}", value, -1)
	}
	return result
}

// extractParamNames extracts parameter names from a template pattern
func (t *TemplateCompression) extractParamNames(pattern string) []string {
	var params []string
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)
	
	for _, match := range matches {
		params = append(params, match[1])
	}
	
	return params
}