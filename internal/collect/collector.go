package collect

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonathanleahy/folder-bundler/internal/compression"
	"github.com/jonathanleahy/folder-bundler/internal/config"
	"github.com/jonathanleahy/folder-bundler/internal/fileutils"
)

type FileCollator struct {
	currentSize  int64
	currentPart  int
	currentFile  *os.File
	baseFileName string
	params       *config.Parameters
	// Compression support
	compressionEnabled bool
	collectedContent   []byte
	contentBuffer      strings.Builder
}

func hasHiddenComponent(path string) bool {
	parts := strings.Split(path, string(os.PathSeparator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func ProcessDirectory(params *config.Parameters) error {
	collator := &FileCollator{
		currentPart:        1,
		baseFileName:       fmt.Sprintf("%s_collated", filepath.Base(params.RootDir)),
		params:             params,
		compressionEnabled: params.EnableCompression,
	}
	defer collator.closeCurrentFile()

	// If compression is enabled, collect all content first
	if collator.compressionEnabled {
		// Write to buffer instead of file initially
		collator.contentBuffer.WriteString(fmt.Sprintf("# Project Files Summary - Part %d\n\nGenerated on: %s\n\nRoot Directory: %s\n\n---\n\n",
			collator.currentPart, time.Now().Format(time.RFC3339), params.RootDir))
	} else {
		if err := collator.createNewFile(); err != nil {
			return err
		}
	}

	// Walk directory and collect/write files
	err := filepath.Walk(params.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(params.RootDir, path)
		if err != nil || relPath == "." {
			return nil
		}

		// Skip entire hidden paths unless explicitly included
		if !params.IncludeHidden && hasHiddenComponent(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip excluded directories
		if info.IsDir() && params.ExcludedDirs[info.Name()] {
			return filepath.SkipDir
		}

		return collator.processPath(relPath, info)
	})

	if err != nil {
		return err
	}

	// If compression is enabled, compress and write the content
	if collator.compressionEnabled {
		return collator.finalizeWithCompression()
	}

	return nil
}

func (fc *FileCollator) processPath(relPath string, info os.FileInfo) error {
	// Normalize path to use forward slashes for cross-platform compatibility
	normalizedPath := filepath.ToSlash(relPath)
	
	fullPath := filepath.Join(fc.params.RootDir, relPath)
	
	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(fullPath)
		if err != nil {
			return fc.writeContent(fmt.Sprintf("## Symlink: %s (Error reading target: %v)\n\n", normalizedPath, err))
		}
		return fc.writeContent(fmt.Sprintf("## Symlink: %s\n\nTarget: %s\n\n", normalizedPath, target))
	}
	
	if info.IsDir() {
		return fc.writeContent(fmt.Sprintf("## Directory: %s\n\n", normalizedPath))
	}

	if info.Size() > fc.params.MaxOutputSize {
		return fc.writeContent(fmt.Sprintf("## File: %s (Skipped - Size exceeds 2MB)\n\n", normalizedPath))
	}

	metadata := fmt.Sprintf("## File: %s\n\nSize: %d bytes\n\nLast Modified: %s\n\n",
		normalizedPath, info.Size(), info.ModTime().Format(time.RFC3339))

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	if fileutils.IsTextFile(content) {
		contentStr := string(content)
		// Add a unique marker that won't conflict with actual content
		return fc.writeContent(fmt.Sprintf("%s===FILE_CONTENT_START===\n%s\n__CONTENT_END_MARKER__\n===FILE_CONTENT_END===\n\n", metadata, contentStr))
	}

	return fc.writeContent(fmt.Sprintf("%sBinary file - content not shown\n\n", metadata))
}

func (fc *FileCollator) writeContent(content string) error {
	// If compression is enabled, buffer content instead of writing
	if fc.compressionEnabled {
		fc.contentBuffer.WriteString(content)
		return nil
	}

	contentSize := int64(len(content))
	if fc.currentSize+contentSize > fc.params.MaxOutputSize {
		fc.currentPart++
		if err := fc.createNewFile(); err != nil {
			return err
		}
	}

	_, err := fc.currentFile.WriteString(content)
	if err != nil {
		return err
	}

	fc.currentSize += contentSize
	return nil
}

func (fc *FileCollator) createNewFile() error {
	fc.closeCurrentFile()

	fileName := fmt.Sprintf("%s_part%d.md", fc.baseFileName, fc.currentPart)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	fc.currentFile = file
	fc.currentSize = 0

	header := fmt.Sprintf("# Project Files Summary - Part %d\n\nGenerated on: %s\n\nRoot Directory: %s\n\n---\n\n",
		fc.currentPart, time.Now().Format(time.RFC3339), fc.params.RootDir)

	_, err = fc.currentFile.WriteString(header)
	return err
}

func (fc *FileCollator) closeCurrentFile() {
	if fc.currentFile != nil {
		fc.currentFile.Close()
		fc.currentFile = nil
	}
}

func (fc *FileCollator) finalizeWithCompression() error {
	// Get the buffered content
	content := fc.contentBuffer.String()
	originalSize := len(content)
	
	// Initialize compression strategies
	if err := compression.InitializeStrategies(); err != nil {
		return fmt.Errorf("failed to initialize compression strategies: %w", err)
	}
	
	// Create compression selector
	selector := compression.NewSelector(compression.DefaultRegistry)
	
	// Compress content using specified strategy
	result, err := selector.CompressContentWithStrategy([]byte(content), fc.params.CompressionStrategy)
	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}
	
	// Create output file without header for compressed content
	fc.closeCurrentFile()
	fileName := fmt.Sprintf("%s_part%d.md", fc.baseFileName, fc.currentPart)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	fc.currentFile = file
	fc.currentSize = 0
	
	// Write compression metadata if compressed
	if result.Strategy != "none" {
		header := fmt.Sprintf("# Compression: %s\n# Original Size: %d bytes\n# Compressed Size: %d bytes\n# Ratio: %.2f%%\n\n",
			result.Metadata, originalSize, len(result.Compressed), result.Ratio*100)
		if _, err := fc.currentFile.WriteString(header); err != nil {
			return err
		}
	}
	
	// Write the content (compressed or original)
	if _, err := fc.currentFile.Write(result.Compressed); err != nil {
		return err
	}
	
	// Log compression results
	if result.Strategy != "none" {
		fmt.Printf("Compressed using %s strategy: %d -> %d bytes (%.1f%% reduction)\n",
			result.Strategy, originalSize, len(result.Compressed), (1-result.Ratio)*100)
	} else {
		fmt.Printf("No compression applied (would not reduce size)\n")
	}
	
	return nil
}
