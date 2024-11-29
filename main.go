package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	ignore "github.com/sabhiram/go-gitignore"
)

// Constants for file size limits
const (
	maxFileSize   = 1 * 1024 * 1024 // Maximum size of individual files to process (1 MB)
	maxOutputSize = 2 * 1024 * 1024 // Maximum size of output files (2 MB)
)

// Configuration maps for exclusions
var (
	excludedDirs = map[string]bool{
		"node_modules": true,
		"vendor":       true,
		"venv":         true,
		"dist":         true,
		"build":        true,
	}

	excludedExtensions = map[string]bool{
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".bin":   true,
		".pkl":   true,
		".pyc":   true,
		".bak":   true,
	}

	excludedFiles = map[string]bool{
		"package-lock.json": true,
		"yarn.lock":         true,
	}
)

// FileCollator manages the output file splitting
type FileCollator struct {
	currentSize  int64
	currentPart  int
	currentFile  *os.File
	baseFileName string
	rootDir      string
}

func NewFileCollator(rootDir string) *FileCollator {
	return &FileCollator{
		currentPart:  1,
		baseFileName: fmt.Sprintf("%s_collated", filepath.Base(rootDir)),
		rootDir:      rootDir,
	}
}

func (fc *FileCollator) createNewFile() error {
	if fc.currentFile != nil {
		fc.currentFile.Close()
	}

	fileName := fmt.Sprintf("%s_part%d.md", fc.baseFileName, fc.currentPart)

	// Backup existing file if it exists
	if err := backupExistingFile(fileName); err != nil {
		return err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create new file part: %v", err)
	}

	fc.currentFile = file
	fc.currentSize = 0

	// Write header for new file part
	header := fmt.Sprintf("# Project Files Summary - Part %d\n\nGenerated on: %s\n\nRoot Directory: %s\n\n---\n\n",
		fc.currentPart, time.Now().Format(time.RFC3339), fc.rootDir)

	_, err = fc.currentFile.WriteString(header)
	return err
}

func (fc *FileCollator) writeContent(content string) error {
	contentSize := int64(len(content))

	// Check if adding this content would exceed size limit
	if fc.currentSize+contentSize > maxOutputSize {
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

func (fc *FileCollator) Close() {
	if fc.currentFile != nil {
		fc.currentFile.Close()
	}
}

func main() {
	// Determine the root directory for processing
	var rootDir string

	// If no argument is provided, or it's "." or "", use the current working directory
	if len(os.Args) < 2 || os.Args[1] == "." || os.Args[1] == "" {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			return
		}
	} else {
		rootDir = os.Args[1]
	}

	// Initialize the file collator
	collator := NewFileCollator(rootDir)
	defer collator.Close()

	// Create the first output file
	if err := collator.createNewFile(); err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}

	// Check for and parse .gitignore file if it exists
	gitignoreFile := filepath.Join(rootDir, ".gitignore")
	var ignorer *ignore.GitIgnore
	if _, err := os.Stat(gitignoreFile); err == nil {
		ignorer, err = ignore.CompileIgnoreFile(gitignoreFile)
		if err != nil {
			fmt.Printf("Error parsing .gitignore: %v\n", err)
		}
	}

	// Walk through the directory structure
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			fmt.Printf("Error getting relative path for %q: %v\n", path, err)
			return err
		}

		fmt.Printf("Processing file: %s\n", relPath)

		// Check if the path matches any patterns in .gitignore
		if ignorer != nil && ignorer.MatchesPath(relPath) {
			fmt.Printf("Skipping ignored file: %s\n", relPath)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			fmt.Printf("Skipping hidden file/directory: %s\n", relPath)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Process directories
		if info.IsDir() {
			if excludedDirs[info.Name()] {
				fmt.Printf("Skipping excluded directory: %s\n", relPath)
				return filepath.SkipDir
			}
			return collator.writeContent(fmt.Sprintf("## Directory: %s\n\n", relPath))
		}

		// Check for excluded extensions
		ext := strings.ToLower(filepath.Ext(path))
		if excludedExtensions[ext] {
			fmt.Printf("Skipping excluded file extension: %s\n", relPath)
			return nil
		}

		// Check for excluded files
		if excludedFiles[filepath.Base(path)] {
			fmt.Printf("Skipping excluded file: %s\n", relPath)
			return nil
		}

		// Handle large files
		if info.Size() > maxFileSize {
			fmt.Printf("Skipping large file: %s (Size: %d bytes)\n", relPath, info.Size())
			return collator.writeContent(fmt.Sprintf("## File: %s (Skipped - Size: %d bytes)\n\n", relPath, info.Size()))
		}

		// Write file metadata
		metadata := fmt.Sprintf("## File: %s\n\nSize: %d bytes\n\nLast Modified: %s\n\n",
			relPath, info.Size(), info.ModTime().Format(time.RFC3339))

		if err := collator.writeContent(metadata); err != nil {
			return err
		}

		// Handle previously collated files
		if strings.HasSuffix(info.Name(), "_collated.md") {
			return collator.writeContent("This is a collated file from a previous run. Contents not displayed.\n\n")
		}

		// Read file content
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %q: %v\n", path, err)
			return err
		}

		// Process and write file content
		isText := isTextFile(content)
		fmt.Printf("File %s is text: %v\n", relPath, isText)

		if isText {
			language := getLanguage(ext)
			return collator.writeContent(fmt.Sprintf("```%s\n%s\n```\n\n", language, string(content)))
		} else {
			return collator.writeContent("Unable to parse: File appears to be in a non-text format.\n\n")
		}
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", rootDir, err)
		return
	}

	fmt.Printf("Processing complete. Output files generated with prefix %s\n", collator.baseFileName)
}

// backupExistingFile renames an existing file to .bak if it exists
func backupExistingFile(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		backupName := filename + ".bak"
		err = os.Rename(filename, backupName)
		if err != nil {
			return fmt.Errorf("failed to rename existing file: %v", err)
		}
		fmt.Printf("Existing file backed up to %s\n", backupName)
	}
	return nil
}

// getLanguage determines the programming language based on file extension
func getLanguage(extension string) string {
	switch extension {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".cpp", ".hpp", ".h":
		return "cpp"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".md":
		return "markdown"
	default:
		return ""
	}
}

// isTextFile checks if the file content is likely to be text
func isTextFile(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	return utf8.Valid(content) && !containsNullByte(content)
}

// containsNullByte checks if the byte slice contains any null bytes
func containsNullByte(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}