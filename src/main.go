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

// maxFileSize defines the maximum size of files to be fully processed (1 MB)
const maxFileSize = 1 * 1024 * 1024

// excludedDirs is a map of directory names to be excluded from processing
var excludedDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"venv":         true,
	"dist":         true,
	"build":        true,
}

// excludedExtensions is a map of file extensions to be excluded from processing
var excludedExtensions = map[string]bool{
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".bin":   true,
	".pkl":   true,
	".pyc":   true,
	".bak":   true,
}

// excludedFiles is a map of specific filenames to be excluded from processing
var excludedFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
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
		// Use the provided directory path
		rootDir = os.Args[1]
	}

	// Create the output filename based on the root directory name
	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(rootDir))
	outputFileName = strings.ReplaceAll(outputFileName, string(os.PathSeparator), "_")

	// Check if the output file already exists and rename it to .bak if it does
	if err := backupExistingFile(outputFileName); err != nil {
		fmt.Printf("Error backing up existing file: %v\n", err)
		return
	}

	// Create the output file
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	// Write the header information to the output file
	writeToFile(outputFile, "# Project Files Summary\n\n")
	writeToFile(outputFile, "Generated on: %s\n\n", time.Now().Format(time.RFC3339))
	writeToFile(outputFile, "Root Directory: %s\n\n", rootDir)
	writeToFile(outputFile, "---\n\n")

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
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			fmt.Printf("Error getting relative path for %q: %v\n", path, err)
			return err
		}

		fmt.Printf("Processing file: %s\n", relPath) // Debug log

		// Check if the path matches any patterns in .gitignore
		if ignorer != nil && ignorer.MatchesPath(relPath) {
			fmt.Printf("Skipping ignored file: %s\n", relPath) // Debug log
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			fmt.Printf("Skipping hidden file/directory: %s\n", relPath) // Debug log
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Process directories
		if info.IsDir() {
			if excludedDirs[info.Name()] {
				fmt.Printf("Skipping excluded directory: %s\n", relPath) // Debug log
				return filepath.SkipDir
			}
			writeToFile(outputFile, "## Directory: %s\n\n", relPath)
			return nil
		}

		// Check for excluded extensions
		ext := strings.ToLower(filepath.Ext(path))
		if excludedExtensions[ext] {
			fmt.Printf("Skipping excluded file extension: %s\n", relPath) // Debug log
			return nil
		}

		// Check for excluded files
		if excludedFiles[filepath.Base(path)] {
			fmt.Printf("Skipping excluded file: %s\n", relPath) // Debug log
			return nil
		}

		// Handle large files
		if info.Size() > maxFileSize {
			fmt.Printf("Skipping large file: %s (Size: %d bytes)\n", relPath, info.Size()) // Debug log
			writeToFile(outputFile, "## File: %s (Skipped - Size: %d bytes)\n\n", relPath, info.Size())
			return nil
		}

		// Write file metadata
		writeToFile(outputFile, "## File: %s\n\n", relPath)
		writeToFile(outputFile, "Size: %d bytes\n\n", info.Size())
		writeToFile(outputFile, "Last Modified: %s\n\n", info.ModTime().Format(time.RFC3339))

		// Handle previously collated files
		if strings.HasSuffix(info.Name(), "_collated.md") {
			writeToFile(outputFile, "This is a collated file from a previous run. Contents not displayed.\n\n")
			return nil
		}

		// Read file content
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %q: %v\n", path, err)
			return err
		}

		// Process and write file content
		isText := isTextFile(content)
		fmt.Printf("File %s is text: %v\n", relPath, isText) // Debug log
		if isText {
			language := getLanguage(ext)
			writeToFile(outputFile, "```%s\n%s\n```\n\n", language, string(content))
		} else {
			writeToFile(outputFile, "Unable to parse: File appears to be in a non-text format.\n\n")
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", rootDir, err)
		return
	}

	fmt.Printf("Processing complete. Output written to %s\n", outputFileName)
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

// writeToFile is a helper function to write formatted text to a file
func writeToFile(file *os.File, format string, args ...interface{}) {
	_, err := fmt.Fprintf(file, format, args...)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}
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
