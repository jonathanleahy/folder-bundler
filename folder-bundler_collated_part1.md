# Project Files Summary - Part 1

Generated on: 2024-11-29T22:33:04Z

Root Directory: /home/jon/folder-bundler

---

## Directory: .

## File: README.md

Size: 1619 bytes

Last Modified: 2024-11-29T22:23:08Z

```markdown
# folder-bundler

folder-bundler is a Go tool that helps you document and recreate project file structures. It creates detailed documentation of your project files and allows you to rebuild the structure elsewhere.

## Quick Start

Install the tool:
```bash
git clone [repository-url]
cd folder-bundler
go build -o bundler
```

Document your project structure:
```bash
./bundler collect ./myproject
```

Recreate the structure elsewhere:
```bash
./bundler reconstruct project_collated.md
```

## Core Features

The tool creates comprehensive project documentation including file contents, directory structures, and metadata. It handles text and binary files appropriately, supports syntax highlighting for major programming languages, and manages large projects through automatic file splitting.

When reconstructing projects, it accurately recreates the original structure while preserving file contents, metadata, and timestamps.

## Configuration Options

Customize collection with these parameters:
```bash
./bundler collect -max-file 5M -exclude-dirs "logs,temp" ./myproject
```

Common settings:
- `-max-file`: Set maximum file size (default: 1MB)
- `-exclude-dirs`: Skip specific directories
- `-include-hidden`: Include hidden files
- `-preserve-time`: Keep original timestamps during reconstruction

The tool automatically excludes common directories like node_modules, dist, and build, as well as binary files (.exe, .dll, etc.) and lock files.

## Use Cases

folder-bundler works well for:
- Project documentation and archiving
- Deployment preparation
- Code review preparation
- Project structure analysis

```

## File: folder-bundler_collated_part1.md

Size: 1846 bytes

Last Modified: 2024-11-29T22:33:04Z

```markdown
# Project Files Summary - Part 1

Generated on: 2024-11-29T22:33:04Z

Root Directory: /home/jon/folder-bundler

---

## Directory: .

## File: README.md

Size: 1619 bytes

Last Modified: 2024-11-29T22:23:08Z

```markdown
# folder-bundler

folder-bundler is a Go tool that helps you document and recreate project file structures. It creates detailed documentation of your project files and allows you to rebuild the structure elsewhere.

## Quick Start

Install the tool:
```bash
git clone [repository-url]
cd folder-bundler
go build -o bundler
```

Document your project structure:
```bash
./bundler collect ./myproject
```

Recreate the structure elsewhere:
```bash
./bundler reconstruct project_collated.md
```

## Core Features

The tool creates comprehensive project documentation including file contents, directory structures, and metadata. It handles text and binary files appropriately, supports syntax highlighting for major programming languages, and manages large projects through automatic file splitting.

When reconstructing projects, it accurately recreates the original structure while preserving file contents, metadata, and timestamps.

## Configuration Options

Customize collection with these parameters:
```bash
./bundler collect -max-file 5M -exclude-dirs "logs,temp" ./myproject
```

Common settings:
- `-max-file`: Set maximum file size (default: 1MB)
- `-exclude-dirs`: Skip specific directories
- `-include-hidden`: Include hidden files
- `-preserve-time`: Keep original timestamps during reconstruction

The tool automatically excludes common directories like node_modules, dist, and build, as well as binary files (.exe, .dll, etc.) and lock files.

## Use Cases

folder-bundler works well for:
- Project documentation and archiving
- Deployment preparation
- Code review preparation
- Project structure analysis

```

## File: folder-bundler_collated_part1.md

Size: 1846 bytes

Last Modified: 2024-11-29T22:33:04Z


```

## File: go.mod

Size: 110 bytes

Last Modified: 2024-11-29T22:16:51Z

```
module folder-bundler

go 1.23.3

require github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06

```

## File: go.sum

Size: 1162 bytes

Last Modified: 2024-11-29T21:54:23Z

```
github.com/davecgh/go-spew v1.1.0 h1:ZDRjVQ15GmhC3fiQ8ni8+OwkZQO4DARzQgrnXU1Liz8=
github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 h1:OkMGxebDjyw0ULyrTYWeN0UNCCkmCWfjPnIA2W6oviI=
github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06/go.mod h1:+ePHsJ1keEjQtpvf9HHw0f4ZeJ0TLRsxhunSI2hYJSs=
github.com/stretchr/objx v0.1.0/go.mod h1:HFkY916IF+rwdDfMAkV7OtwuqBVzrE8GR6GFx+wExME=
github.com/stretchr/testify v1.6.1 h1:hDPOHmpOpP40lSULcqw7IrRb/u7w6RpDC9399XyoNd0=
github.com/stretchr/testify v1.6.1/go.mod h1:6Fq8oRcR53rry900zMqJjRRixrwX3KX962/h/Wwjteg=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c h1:dUUwHk2QECo/6vqA44rthZ8ie2QXMNeKRTHCNY2nXvo=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=

```

## Directory: internal

## Directory: internal/collect

## File: internal/collect/collector.go

Size: 4023 bytes

Last Modified: 2024-11-29T22:27:41Z

```go
package collect

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"folder-bundler/internal/config"
	"folder-bundler/internal/fileutils"

	ignore "github.com/sabhiram/go-gitignore"
)

type FileCollator struct {
	currentSize  int64
	currentPart  int
	currentFile  *os.File
	baseFileName string
	rootDir      string
	params       *config.Parameters
}

func NewFileCollator(rootDir string, params *config.Parameters) *FileCollator {
	return &FileCollator{
		currentPart:  1,
		baseFileName: fmt.Sprintf("%s_collated", filepath.Base(rootDir)),
		rootDir:      rootDir,
		params:       params,
	}
}

func ProcessDirectory(params *config.Parameters) error {
	collator := NewFileCollator(params.RootDir, params)
	defer collator.Close()

	if err := collator.createNewFile(); err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}

	var ignorer *ignore.GitIgnore
	if !params.SkipGitignore {
		gitignoreFile := filepath.Join(params.RootDir, ".gitignore")
		if _, err := os.Stat(gitignoreFile); err == nil {
			ignorer, _ = ignore.CompileIgnoreFile(gitignoreFile)
		}
	}

	return filepath.Walk(params.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(params.RootDir, path)
		if err != nil {
			return err
		}

		if !params.IncludeHidden && strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if ignorer != nil && ignorer.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		return collator.processPath(relPath, info)
	})
}

func (fc *FileCollator) processPath(relPath string, info os.FileInfo) error {
	if info.IsDir() {
		if fc.params.ExcludedDirs[info.Name()] {
			return filepath.SkipDir
		}
		return fc.writeContent(fmt.Sprintf("## Directory: %s\n\n", relPath))
	}

	if fc.params.ExcludedFiles[info.Name()] {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(relPath))
	if fc.params.ExcludedExts[ext] {
		return nil
	}

	return fc.processFile(relPath, info)
}

func (fc *FileCollator) processFile(relPath string, info os.FileInfo) error {
	if info.Size() > fc.params.MaxFileSize {
		return fc.writeContent(fmt.Sprintf("## File: %s (Skipped - Size: %d bytes)\n\n", relPath, info.Size()))
	}

	metadata := fmt.Sprintf("## File: %s\n\nSize: %d bytes\n\nLast Modified: %s\n\n",
		relPath, info.Size(), info.ModTime().Format(time.RFC3339))

	if err := fc.writeContent(metadata); err != nil {
		return err
	}

	content, err := ioutil.ReadFile(filepath.Join(fc.rootDir, relPath))
	if err != nil {
		return err
	}

	if fileutils.IsTextFile(content) {
		language := fileutils.GetLanguage(ext)
		return fc.writeContent(fmt.Sprintf("```%s\n%s\n```\n\n", language, string(content)))
	}

	return fc.writeContent("Unable to parse: File appears to be in a non-text format.\n\n")
}

func (fc *FileCollator) createNewFile() error {
	if fc.currentFile != nil {
		fc.currentFile.Close()
	}

	fileName := fmt.Sprintf("%s_part%d.md", fc.baseFileName, fc.currentPart)
	if err := fileutils.BackupExistingFile(fileName); err != nil {
		return err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create new file part: %v", err)
	}

	fc.currentFile = file
	fc.currentSize = 0

	header := fmt.Sprintf("# Project Files Summary - Part %d\n\nGenerated on: %s\n\nRoot Directory: %s\n\n---\n\n",
		fc.currentPart, time.Now().Format(time.RFC3339), fc.rootDir)

	_, err = fc.currentFile.WriteString(header)
	return err
}

func (fc *FileCollator) writeContent(content string) error {
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

func (fc *FileCollator) Close() {
	if fc.currentFile != nil {
		fc.currentFile.Close()
	}
}

```

## Directory: internal/config

## File: internal/config/params.go

Size: 2978 bytes

Last Modified: 2024-11-29T22:26:38Z

```go
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Parameters struct {
	MaxFileSize       int64
	MaxOutputSize     int64
	ExcludedDirs      map[string]bool
	ExcludedFiles     map[string]bool
	ExcludedExts      map[string]bool
	IncludeHidden     bool
	SkipGitignore     bool
	PreserveTimestamp bool
	RootDir           string
}

func ParseParameters() (*Parameters, error) {
	var params Parameters
	var excludeDirs, excludeFiles, excludeExts string

	collect := flag.NewFlagSet("collect", flag.ExitOnError)
	reconstruct := flag.NewFlagSet("reconstruct", flag.ExitOnError)

	// Configure collect flags
	collect.Int64Var(&params.MaxFileSize, "max-file", 1*1024*1024, "Maximum size of individual files")
	collect.Int64Var(&params.MaxOutputSize, "max-output", 2*1024*1024, "Maximum size of output files")
	collect.StringVar(&excludeDirs, "exclude-dirs", "node_modules,vendor,venv,dist,build", "Directories to exclude")
	collect.StringVar(&excludeFiles, "exclude-files", "package-lock.json,yarn.lock", "Files to exclude")
	collect.StringVar(&excludeExts, "exclude-exts", ".exe,.dll,.so,.dylib,.bin,.pkl,.pyc,.bak", "Extensions to exclude")
	collect.BoolVar(&params.IncludeHidden, "include-hidden", false, "Include hidden files")
	collect.BoolVar(&params.SkipGitignore, "skip-gitignore", false, "Skip .gitignore processing")

	// Configure reconstruct flags
	reconstruct.BoolVar(&params.PreserveTimestamp, "preserve-time", true, "Preserve original timestamps")

	// Parse appropriate flag set
	var err error
	switch os.Args[1] {
	case "collect":
		err = collect.Parse(os.Args[2:])
	case "reconstruct":
		err = reconstruct.Parse(os.Args[2:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %v", err)
	}

	// Convert string lists to maps
	params.ExcludedDirs = stringToMap(excludeDirs)
	params.ExcludedFiles = stringToMap(excludeFiles)
	params.ExcludedExts = stringToMap(excludeExts)

	// Set root directory
	params.RootDir = "."
	if flag.NArg() > 0 {
		params.RootDir = flag.Arg(0)
	}

	return &params, nil
}

func stringToMap(s string) map[string]bool {
	result := make(map[string]bool)
	if s == "" {
		return result
	}
	for _, item := range strings.Split(s, ",") {
		result[strings.TrimSpace(item)] = true
	}
	return result
}

func PrintUsage() {
	fmt.Printf(`folder-bundler - Tool for collecting and reconstructing file structures

Usage:
  bundler <command> [flags] [arguments]

Commands:
  collect     Create a detailed summary of a directory structure
  reconstruct Rebuild directory structure from a summary file

Global Flags:
  -h, --help    Display help information

Examples:
  bundler collect ./myproject
  bundler reconstruct project_collated.md
`)
}

func PrintReconstructHelp() {
	fmt.Printf(`Usage: bundler reconstruct [flags] <input_file>

Rebuild a directory structure from a previously created summary file.

Flags:
  -h, --help         Display this help message
  -preserve-time     Preserve original timestamps (default: true)
`)
}

```

## Directory: internal/fileutils

## File: internal/fileutils/utils.go

Size: 1900 bytes

Last Modified: 2024-11-29T22:29:23Z

```go
package fileutils

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func IsTextFile(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	return utf8.Valid(content) && !containsNullByte(content)
}

func containsNullByte(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func GetLanguage(extension string) string {
	languageMap := map[string]string{
		".go":         "go",
		".js":         "javascript",
		".jsx":        "javascript",
		".ts":         "typescript",
		".tsx":        "typescript",
		".py":         "python",
		".java":       "java",
		".cpp":        "cpp",
		".hpp":        "cpp",
		".h":          "cpp",
		".cc":         "cpp",
		".c":          "c",
		".cs":         "csharp",
		".html":       "html",
		".css":        "css",
		".scss":       "scss",
		".sass":       "scss",
		".md":         "markdown",
		".sh":         "bash",
		".bash":       "bash",
		".json":       "json",
		".xml":        "xml",
		".yaml":       "yaml",
		".yml":        "yaml",
		".sql":        "sql",
		".rs":         "rust",
		".rb":         "ruby",
		".php":        "php",
		".swift":      "swift",
		".kt":         "kotlin",
		".scala":      "scala",
		".r":          "r",
		".dart":       "dart",
		".lua":        "lua",
		".pl":         "perl",
		".gradle":     "gradle",
		".dockerfile": "dockerfile",
		".tf":         "terraform",
		".vue":        "vue",
		".ini":        "ini",
		".conf":       "ini",
		".toml":       "toml",
		".proto":      "protobuf",
	}

	return languageMap[strings.ToLower(extension)]
}

func BackupExistingFile(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		backupName := filename + ".bak"
		if err := os.Rename(filename, backupName); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
		fmt.Printf("Existing file backed up to %s\n", backupName)
	}
	return nil
}

```

## Directory: internal/reconstruct

## File: internal/reconstruct/reconstructor.go

Size: 4653 bytes

Last Modified: 2024-11-29T22:29:00Z

```go
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

```

## File: main.go

Size: 854 bytes

Last Modified: 2024-11-29T22:24:51Z

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"folder-bundler/internal/collect"
	"folder-bundler/internal/config"
	"folder-bundler/internal/reconstruct"
)

func main() {
	if len(os.Args) < 2 {
		config.PrintUsage()
		os.Exit(1)
	}

	cfg, err := config.ParseParameters()
	if err != nil {
		fmt.Printf("Error parsing parameters: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "collect":
		if err := collect.ProcessDirectory(cfg); err != nil {
			fmt.Printf("Error during collection: %v\n", err)
			os.Exit(1)
		}
	case "reconstruct":
		if len(os.Args) < 3 {
			config.PrintReconstructHelp()
			os.Exit(1)
		}
		if err := reconstruct.FromFile(os.Args[2], cfg); err != nil {
			fmt.Printf("Error during reconstruction: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		config.PrintUsage()
		os.Exit(1)
	}
}

```

## File: main_test.go

Size: 5721 bytes

Last Modified: 2024-11-29T21:54:23Z

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirectoryFileProcessor(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createTestFiles(t, tempDir)

	// Test with explicit directory
	testExplicitDirectory(t, tempDir)

	// Test with current directory
	testCurrentDirectory(t)

	// Test backup functionality
	testBackupFunctionality(t, tempDir)
}

func testExplicitDirectory(t *testing.T, tempDir string) {
	os.Args = []string{"cmd", tempDir}
	output := captureOutput(main)
	t.Logf("Debug output:\n%s", output)

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(tempDir))
	outputPath := filepath.Join(".", outputFileName)
	verifyOutput(t, outputPath, tempDir, false)
}

func testCurrentDirectory(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(currentDir)

	tempDir, err := ioutil.TempDir("", "test_current_dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	createTestFiles(t, tempDir)

	os.Args = []string{"cmd"}
	output := captureOutput(main)
	t.Logf("Debug output:\n%s", output)

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(tempDir))
	outputPath := filepath.Join(".", outputFileName)
	verifyOutput(t, outputPath, tempDir, false)
}

func testBackupFunctionality(t *testing.T, tempDir string) {
	os.Args = []string{"cmd", tempDir}

	// Run once to create the initial file
	output := captureOutput(main)
	t.Logf("Debug output (first run):\n%s", output)

	// Run again to test backup functionality
	output = captureOutput(main)
	t.Logf("Debug output (second run):\n%s", output)

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(tempDir))
	outputPath := filepath.Join(".", outputFileName)
	backupPath := outputPath + ".bak"

	// Verify both the new file and the backup
	verifyOutput(t, outputPath, tempDir, false)
	verifyOutput(t, backupPath, tempDir, true)
}

func verifyOutput(t *testing.T, outputPath string, tempDir string, isBackup bool) {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", err)
	}

	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	t.Logf("Output file content:\n%s", contentStr)

	expectedContents := []string{
		"# Project Files Summary",
		"## Directory: subdir",
		"## File: test.go",
		"package main",
		"func main() {}",
		"## File: subdir", // We'll check for "subdir" and "subfile.txt" separately
		"subfile.txt",
		"This is a subfile",
		"## File: binary.dat",
		"Unable to parse: File appears to be in a non-text format",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Output does not contain expected content: %s", expected)
		}
	}

	// Check specifically for the subdirectory file, allowing for different path separators
	if !strings.Contains(contentStr, "## File: subdir/subfile.txt") && !strings.Contains(contentStr, "## File: subdir\\subfile.txt") {
		t.Errorf("Output does not contain the subdirectory file (subdir/subfile.txt or subdir\\subfile.txt)")
	}

	unexpectedContents := []string{
		"excluded.exe",
		".gitignore",
	}

	for _, unexpected := range unexpectedContents {
		if strings.Contains(contentStr, unexpected) {
			t.Errorf("Output contains unexpected content: %s", unexpected)
		}
	}

	// Check for backup-specific content
	if isBackup {
		expectedBackupMessage := "# Project Files Summary"
		if !strings.Contains(contentStr, expectedBackupMessage) {
			t.Errorf("Backup file does not contain the expected content: %s", expectedBackupMessage)
		}
	}

	// Print the directory structure for debugging
	t.Log("Directory structure:")
	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(tempDir, path)
		t.Logf("%s", relPath)
		return nil
	})
}

func createTestFiles(t *testing.T, root string) {
	err := ioutil.WriteFile(filepath.Join(root, "test.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test.go: %v", err)
	}

	subdir := filepath.Join(root, "subdir")
	err = os.Mkdir(subdir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(subdir, "subfile.txt"), []byte("This is a subfile"), 0644)
	if err != nil {
		t.Fatalf("Failed to create subfile.txt: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(root, "excluded.exe"), []byte("This should be excluded"), 0644)
	if err != nil {
		t.Fatalf("Failed to create excluded.exe: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.exe"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create a binary file
	binaryData := []byte{0x00, 0x01, 0x02, 0x03}
	err = ioutil.WriteFile(filepath.Join(root, "binary.dat"), binaryData, 0644)
	if err != nil {
		t.Fatalf("Failed to create binary.dat: %v", err)
	}

	t.Logf("Test files created in %s", root)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(root, path)
		t.Logf("Created: %s", relPath)
		return nil
	})
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

```

## Directory: src

## File: src/src (Skipped - Size: 3184764 bytes)
