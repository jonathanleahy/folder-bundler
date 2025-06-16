package reconstruct

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonathanleahy/folder-bundler/internal/compression"
	"github.com/jonathanleahy/folder-bundler/internal/config"
)

type FileInfo struct {
	path         string
	content      strings.Builder
	size         int64
	lastModified time.Time
	isDirectory  bool
	isSymlink    bool
	symlinkTarget string
}

func FromFile(inputFile string, params *config.Parameters) error {
	fmt.Printf("Starting reconstruction from: %s\n", inputFile)
	
	basePath := strings.TrimSuffix(inputFile, "_part1.md")
	basePath = strings.TrimSuffix(basePath, ".md")
	pattern := basePath + "*.md"

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("error finding collated files: %v", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no collated files found matching pattern: %s", pattern)
	}

	fmt.Printf("Found %d file(s) to process\n", len(matches))

	var allFiles []FileInfo
	var rootDir string

	for _, match := range matches {
		fmt.Printf("  Processing: %s\n", match)
		currentRootDir, files, err := parseInputFile(match)
		if err != nil {
			return fmt.Errorf("error parsing input file %s: %v", match, err)
		}

		if rootDir == "" {
			rootDir = currentRootDir
		} else if rootDir != currentRootDir {
			fmt.Printf("  Warning: Inconsistent root directories found. Using %s\n", rootDir)
		}

		allFiles = append(allFiles, files...)
	}

	return reconstructFiles(rootDir, allFiles, params)
}

func parseInputFile(filename string) (string, []FileInfo, error) {
	// Read entire file content first
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", nil, err
	}

	// Check for compression headers and decompress if needed
	decompressedContent, err := handleCompression(content)
	if err != nil {
		return "", nil, fmt.Errorf("error handling compression: %v", err)
	}


	// Parse the (decompressed) content
	return parseContent(decompressedContent)
}

func handleCompression(content []byte) ([]byte, error) {
	// Check if content starts with compression header
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	
	
	if len(lines) < 1 || !strings.HasPrefix(lines[0], "# Compression: ") {
		// Not compressed, return as-is
		return content, nil
	}

	// Extract compression metadata
	var compressionType string
	var compressedStart int
	currentLine := 0

	// Parse compression headers
	for currentLine < len(lines) {
		line := lines[currentLine]
		if strings.HasPrefix(line, "# Compression: ") {
			compressionType = strings.TrimPrefix(line, "# Compression: ")
		} else if strings.HasPrefix(line, "# Original Size: ") {
			// Skip
		} else if strings.HasPrefix(line, "# Compressed Size: ") {
			// Skip
		} else if strings.HasPrefix(line, "# Ratio: ") {
			// Skip
		} else if line == "" {
			// Empty line after headers
			compressedStart = currentLine + 1
			break
		} else {
			// Non-header line, content starts here
			compressedStart = currentLine
			break
		}
		currentLine++
	}

	if compressionType == "" || compressedStart == 0 {
		// No valid compression header found
		return content, nil
	}

	// Get compressed content
	compressedLines := lines[compressedStart:]
	compressedContent := []byte(strings.Join(compressedLines, "\n"))

	// Initialize compression if not already done
	if err := compression.InitializeStrategies(); err != nil {
		return nil, fmt.Errorf("failed to initialize compression strategies: %v", err)
	}

	// Create selector and decompress
	selector := compression.NewSelector(compression.DefaultRegistry)
	decompressed, err := selector.DecompressContent(compressedContent, compressionType)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %v", err)
	}

	return decompressed, nil
}

func parseContent(content []byte) (string, []FileInfo, error) {
	var files []FileInfo
	var currentFile *FileInfo
	var rootDir string
	isReadingCode := false
	isFirstContentLine := true

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "Root Directory: "):
			rootDir = strings.TrimPrefix(line, "Root Directory: ")

		case strings.HasPrefix(line, "## Directory: "):
			if currentFile != nil && !isReadingCode {
				files = append(files, *currentFile)
				currentFile = nil
			}
			dirPath := strings.TrimPrefix(line, "## Directory: ")
			if dirPath == "." {
				dirPath = ""
			}
			// Convert forward slashes to OS-specific path separator
			dirPath = filepath.FromSlash(dirPath)
			files = append(files, FileInfo{
				path:        dirPath,
				isDirectory: true,
			})

		case strings.HasPrefix(line, "## Symlink: "):
			if currentFile != nil && !isReadingCode {
				files = append(files, *currentFile)
			}
			path := strings.TrimPrefix(line, "## Symlink: ")
			if idx := strings.Index(path, " (Error"); idx != -1 {
				path = path[:idx]
			}
			// Convert forward slashes to OS-specific path separator
			path = filepath.FromSlash(path)
			currentFile = &FileInfo{
				path:      path,
				isSymlink: true,
			}

		case strings.HasPrefix(line, "## File: "):
			if currentFile != nil && !isReadingCode {
				files = append(files, *currentFile)
			}
			path := strings.TrimPrefix(line, "## File: ")
			if idx := strings.Index(path, " (Skipped - Size:"); idx != -1 {
				path = path[:idx]
			}
			// Convert forward slashes to OS-specific path separator
			path = filepath.FromSlash(path)
			currentFile = &FileInfo{
				path:        path,
				isDirectory: false,
			}

		case strings.HasPrefix(line, "Size: "):
			if currentFile != nil {
				size := strings.TrimPrefix(line, "Size: ")
				size = strings.TrimSuffix(size, " bytes")
				fmt.Sscanf(size, "%d", &currentFile.size)
			}

		case strings.HasPrefix(line, "Last Modified: "):
			if currentFile != nil {
				timeStr := strings.TrimPrefix(line, "Last Modified: ")
				currentFile.lastModified, _ = time.Parse(time.RFC3339, timeStr)
			}

		case strings.HasPrefix(line, "Target: "):
			if currentFile != nil && currentFile.isSymlink {
				currentFile.symlinkTarget = strings.TrimPrefix(line, "Target: ")
			}

		case line == "===FILE_CONTENT_START===":
			if !isReadingCode {
				isReadingCode = true
				isFirstContentLine = true
			}

		case line == "===FILE_CONTENT_END===":
			if isReadingCode {
				isReadingCode = false
			}

		default:
			if isReadingCode && currentFile != nil {
				// Skip our content end marker
				if line == "__CONTENT_END_MARKER__" {
					// Don't add this line to content
				} else {
					if !isFirstContentLine {
						currentFile.content.WriteString("\n")
					}
					currentFile.content.WriteString(line)
					isFirstContentLine = false
				}
			}
		}
	}

	if currentFile != nil && !isReadingCode {
		files = append(files, *currentFile)
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	if rootDir == "" {
		return "", nil, fmt.Errorf("root directory not found in input file")
	}

	return rootDir, files, nil
}

func reconstructFiles(rootDir string, files []FileInfo, params *config.Parameters) error {
	fmt.Printf("\nReconstructing project structure:\n")
	fmt.Printf("  Root directory: %s\n", rootDir)
	fmt.Printf("  Total items: %d\n", len(files))
	
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return fmt.Errorf("error creating root directory: %v", err)
	}

	if err := os.Chdir(rootDir); err != nil {
		return fmt.Errorf("error changing to root directory: %v", err)
	}

	dirCount := 0
	fileCount := 0
	symlinkCount := 0
	totalSize := int64(0)

	// First create all directories
	for _, f := range files {
		if f.isDirectory && f.path != "" {
			if err := os.MkdirAll(f.path, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", f.path, err)
			}
			dirCount++
		}
	}

	// Then create all files and symlinks
	for _, f := range files {
		if !f.isDirectory {
			if f.isSymlink {
				if err := reconstructSymlink(f); err != nil {
					return fmt.Errorf("error reconstructing symlink %s: %v", f.path, err)
				}
				symlinkCount++
			} else {
				if err := reconstructFile(f, params.PreserveTimestamp); err != nil {
					return fmt.Errorf("error reconstructing file %s: %v", f.path, err)
				}
				fileCount++
				totalSize += int64(len(f.content.String()))
			}
		}
	}

	fmt.Printf("\nReconstruction complete:\n")
	fmt.Printf("  Directories created: %d\n", dirCount)
	fmt.Printf("  Files created: %d\n", fileCount)
	if symlinkCount > 0 {
		fmt.Printf("  Symlinks created: %d\n", symlinkCount)
	}
	fmt.Printf("  Total size: %s\n", formatSize(totalSize))

	return nil
}

// formatSize formats bytes into human readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func reconstructFile(f FileInfo, preserveTimestamp bool) error {
	dir := filepath.Dir(f.path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating parent directory: %v", err)
		}
	}

	file, err := os.Create(f.path)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	content := f.content.String()
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("error writing content: %v", err)
	}

	if preserveTimestamp && !f.lastModified.IsZero() {
		if err := os.Chtimes(f.path, f.lastModified, f.lastModified); err != nil {
			return fmt.Errorf("error setting file time: %v", err)
		}
	}

	return nil
}

func reconstructSymlink(f FileInfo) error {
	dir := filepath.Dir(f.path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating parent directory: %v", err)
		}
	}

	// Remove existing symlink if it exists
	if _, err := os.Lstat(f.path); err == nil {
		if err := os.Remove(f.path); err != nil {
			return fmt.Errorf("error removing existing symlink: %v", err)
		}
	}

	if err := os.Symlink(f.symlinkTarget, f.path); err != nil {
		return fmt.Errorf("error creating symlink: %v", err)
	}

	return nil
}
