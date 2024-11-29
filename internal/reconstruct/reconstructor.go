package reconstruct

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"folder-bundler/internal/config"
)

type FileInfo struct {
	path         string
	content      strings.Builder
	size         int64
	lastModified time.Time
	isDirectory  bool
}

func FromFile(inputFile string, params *config.Parameters) error {
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

	var allFiles []FileInfo
	var rootDir string

	for _, match := range matches {
		fmt.Printf("Processing file: %s\n", match)
		currentRootDir, files, err := parseInputFile(match)
		if err != nil {
			return fmt.Errorf("error parsing input file %s: %v", match, err)
		}

		if rootDir == "" {
			rootDir = currentRootDir
		} else if rootDir != currentRootDir {
			fmt.Printf("Warning: Inconsistent root directories found. Using %s\n", rootDir)
		}

		allFiles = append(allFiles, files...)
	}

	return reconstructFiles(rootDir, allFiles, params)
}

func parseInputFile(filename string) (string, []FileInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	var files []FileInfo
	var currentFile *FileInfo
	var rootDir string
	isReadingCode := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "Root Directory: "):
			rootDir = strings.TrimPrefix(line, "Root Directory: ")

		case strings.HasPrefix(line, "## Directory: "):
			dirPath := strings.TrimPrefix(line, "## Directory: ")
			if dirPath == "." {
				dirPath = ""
			}
			files = append(files, FileInfo{
				path:        dirPath,
				isDirectory: true,
			})

		case strings.HasPrefix(line, "## File: "):
			if currentFile != nil {
				files = append(files, *currentFile)
			}
			path := strings.TrimPrefix(line, "## File: ")
			if idx := strings.Index(path, " (Skipped - Size:"); idx != -1 {
				path = path[:idx]
			}
			currentFile = &FileInfo{
				path:        path,
				isDirectory: false,
			}

		case strings.HasPrefix(line, "Size: "):
			size := strings.TrimPrefix(line, "Size: ")
			size = strings.TrimSuffix(size, " bytes")
			fmt.Sscanf(size, "%d", &currentFile.size)

		case strings.HasPrefix(line, "Last Modified: "):
			timeStr := strings.TrimPrefix(line, "Last Modified: ")
			currentFile.lastModified, _ = time.Parse(time.RFC3339, timeStr)

		case line == "```":
			isReadingCode = !isReadingCode

		default:
			if isReadingCode && currentFile != nil {
				currentFile.content.WriteString(line + "\n")
			}
		}
	}

	if currentFile != nil {
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
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return fmt.Errorf("error creating root directory: %v", err)
	}

	if err := os.Chdir(rootDir); err != nil {
		return fmt.Errorf("error changing to root directory: %v", err)
	}

	// First create all directories
	for _, f := range files {
		if f.isDirectory && f.path != "" {
			if err := os.MkdirAll(f.path, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", f.path, err)
			}
			fmt.Printf("Created directory: %s\n", f.path)
		}
	}

	// Then create all files
	for _, f := range files {
		if !f.isDirectory {
			if err := reconstructFile(f, params.PreserveTimestamp); err != nil {
				return fmt.Errorf("error reconstructing file %s: %v", f.path, err)
			}
		}
	}

	return nil
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

	if _, err := file.WriteString(f.content.String()); err != nil {
		return fmt.Errorf("error writing content: %v", err)
	}

	if preserveTimestamp && !f.lastModified.IsZero() {
		if err := os.Chtimes(f.path, f.lastModified, f.lastModified); err != nil {
			return fmt.Errorf("error setting file time: %v", err)
		}
	}

	fmt.Printf("Created file: %s\n", f.path)
	return nil
}
