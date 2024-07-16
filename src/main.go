package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"
)

const maxFileSize = 1 * 1024 * 1024 // 1 MB

var excludedDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"venv":         true,
	"dist":         true,
	"build":        true,
}

var excludedExtensions = map[string]bool{
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".bin":   true,
	".dat":   true,
	".pkl":   true,
	".pyc":   true,
}

var excludedFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <directory_path>")
		return
	}

	rootDir := os.Args[1]

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(rootDir))
	outputFileName = strings.ReplaceAll(outputFileName, string(os.PathSeparator), "_")

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	writeToFile(outputFile, "# Project Files Summary\n\n")
	writeToFile(outputFile, "Generated on: %s\n\n", time.Now().Format(time.RFC3339))
	writeToFile(outputFile, "Root Directory: %s\n\n", rootDir)
	writeToFile(outputFile, "---\n\n")

	gitignoreFile := filepath.Join(rootDir, ".gitignore")
	var ignorer *ignore.GitIgnore
	if _, err := os.Stat(gitignoreFile); err == nil {
		ignorer, err = ignore.CompileIgnoreFile(gitignoreFile)
		if err != nil {
			fmt.Printf("Error parsing .gitignore: %v\n", err)
		}
	}

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

		if ignorer != nil && ignorer.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			if excludedDirs[info.Name()] {
				return filepath.SkipDir
			}
			writeToFile(outputFile, "## Directory: %s\n\n", relPath)
			return nil
		}

		if excludedExtensions[strings.ToLower(filepath.Ext(path))] || excludedFiles[filepath.Base(path)] {
			return nil
		}

		if info.Size() > maxFileSize {
			writeToFile(outputFile, "## File: %s (Skipped - Size: %d bytes)\n\n", relPath, info.Size())
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %q: %v\n", path, err)
			return err
		}

		writeToFile(outputFile, "## File: %s\n\n", relPath)
		writeToFile(outputFile, "Size: %d bytes\n\n", info.Size())
		writeToFile(outputFile, "Last Modified: %s\n\n", info.ModTime().Format(time.RFC3339))

		extension := strings.ToLower(filepath.Ext(path))
		language := getLanguage(extension)

		writeToFile(outputFile, "```%s\n%s\n```\n\n", language, string(content))

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", rootDir, err)
		return
	}

	fmt.Printf("Processing complete. Output written to %s\n", outputFileName)
}

func writeToFile(file *os.File, format string, args ...interface{}) {
	_, err := fmt.Fprintf(file, format, args...)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}
}

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
