package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	ignore "github.com/sabhiram/go-gitignore"
)

// Parameters holds all configurable settings
type Parameters struct {
	maxFileSize       int64
	maxOutputSize     int64
	excludedDirs      map[string]bool
	excludedFiles     map[string]bool
	excludedExts      map[string]bool
	includeHidden     bool
	skipGitignore     bool
	preserveTimestamp bool
}

// FileInfo stores information about a file or directory
type FileInfo struct {
	path         string
	content      strings.Builder
	size         int64
	lastModified time.Time
	isDirectory  bool
}

// FileCollator manages the output file splitting
type FileCollator struct {
	currentSize  int64
	currentPart  int
	currentFile  *os.File
	baseFileName string
	rootDir      string
	params       *Parameters
}

func NewFileCollator(rootDir string, params *Parameters) *FileCollator {
	return &FileCollator{
		currentPart:  1,
		baseFileName: fmt.Sprintf("%s_collated", filepath.Base(rootDir)),
		rootDir:      rootDir,
		params:       params,
	}
}

func printUsage() {
	fmt.Printf(`File Structure Manager - Tool for collecting and reconstructing file structures

Usage:
  %s <command> [flags] [arguments]

Commands:
  collect     Create a detailed summary of a directory structure
  reconstruct Rebuild directory structure from a summary file

Global Flags:
  -h, --help    Display help information for any command

Examples:
  %s collect ./myproject
  %s reconstruct project_collated.md
  %s collect -h
  %s reconstruct -h

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func printCollectHelp() {
	fmt.Printf(`Usage: %s collect [flags] [directory]

Create a detailed summary of a directory structure and its contents.
If no directory is specified, the current directory will be used.

Flags:
  -h, --help          Display this help message
  -max-file int       Maximum size of individual files to process (bytes) (default: 1MB)
  -max-output int     Maximum size of output files (bytes) (default: 2MB)
  -exclude-dirs       Comma-separated list of directories to exclude
                     (default: "node_modules,vendor,venv,dist,build")
  -exclude-files      Comma-separated list of files to exclude
                     (default: "package-lock.json,yarn.lock")
  -exclude-exts       Comma-separated list of file extensions to exclude
                     (default: ".exe,.dll,.so,.dylib,.bin,.pkl,.pyc,.bak")
  -include-hidden     Include hidden files and directories (default: false)
  -skip-gitignore     Skip .gitignore processing (default: false)

Examples:
  %s collect                           # Collect current directory
  %s collect ./myproject              # Collect specific directory
  %s collect -max-file 5242880 ./src  # Set custom file size limit (5MB)
  %s collect -exclude-dirs "logs,temp" # Custom directory exclusions

Note: Size values can use suffixes: K, M, G (e.g., 5M for 5 megabytes)
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func printReconstructHelp() {
	fmt.Printf(`Usage: %s reconstruct [flags] <input_file>

Rebuild a directory structure from a previously created summary file.
The input file should be a summary file created by the collect command.

Flags:
  -h, --help         Display this help message
  -preserve-time     Preserve original timestamps (default: true)

Examples:
  %s reconstruct project_collated.md
  %s reconstruct -preserve-time=false project_collated.md

Note: If the summary was split into multiple files, you can specify any part 
and the tool will automatically locate and process all related parts.
`, os.Args[0], os.Args[0], os.Args[0])
}

func parseParameters() *Parameters {
	if len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(0)
	}

	collect := flag.NewFlagSet("collect", flag.ExitOnError)
	reconstruct := flag.NewFlagSet("reconstruct", flag.ExitOnError)

	collect.Usage = printCollectHelp
	reconstruct.Usage = printReconstructHelp

	var params Parameters
	var excludeDirs, excludeFiles, excludeExts string

	if len(os.Args) > 1 && os.Args[1] == "collect" {
		collect.Int64Var((*int64)(&params.maxFileSize), "max-file", 1*1024*1024, "")
		collect.Int64Var((*int64)(&params.maxOutputSize), "max-output", 2*1024*1024, "")
		collect.StringVar(&excludeDirs, "exclude-dirs", "node_modules,vendor,venv,dist,build", "")
		collect.StringVar(&excludeFiles, "exclude-files", "package-lock.json,yarn.lock", "")
		collect.StringVar(&excludeExts, "exclude-exts", ".exe,.dll,.so,.dylib,.bin,.pkl,.pyc,.bak", "")
		collect.BoolVar(&params.includeHidden, "include-hidden", false, "")
		collect.BoolVar(&params.skipGitignore, "skip-gitignore", false, "")

		err := collect.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error parsing parameters: %v\n", err)
			os.Exit(1)
		}
	}

	if len(os.Args) > 1 && os.Args[1] == "reconstruct" {
		reconstruct.BoolVar(&params.preserveTimestamp, "preserve-time", true, "")

		err := reconstruct.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error parsing parameters: %v\n", err)
			os.Exit(1)
		}
	}

	params.excludedDirs = stringToMap(excludeDirs)
	params.excludedFiles = stringToMap(excludeFiles)
	params.excludedExts = stringToMap(excludeExts)

	return &params
}

func (fc *FileCollator) createNewFile() error {
	if fc.currentFile != nil {
		fc.currentFile.Close()
	}

	fileName := fmt.Sprintf("%s_part%d.md", fc.baseFileName, fc.currentPart)
	if err := backupExistingFile(fileName); err != nil {
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
	if fc.currentSize+contentSize > fc.params.maxOutputSize {
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
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	params := parseParameters()

	switch os.Args[1] {
	case "collect":
		collectFiles(params)
	case "reconstruct":
		if len(os.Args) < 3 {
			printReconstructHelp()
			os.Exit(1)
		}
		reconstructFromFile(os.Args[2], params)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
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

func collectFiles(params *Parameters) {
	rootDir := "."
	if flag.NArg() > 0 {
		rootDir = flag.Arg(0)
	}

	var err error
	if rootDir == "." {
		rootDir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			return
		}
	}

	collator := NewFileCollator(rootDir, params)
	defer collator.Close()

	if err := collator.createNewFile(); err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}

	var ignorer *ignore.GitIgnore
	if !params.skipGitignore {
		gitignoreFile := filepath.Join(rootDir, ".gitignore")
		if _, err := os.Stat(gitignoreFile); err == nil {
			ignorer, err = ignore.CompileIgnoreFile(gitignoreFile)
			if err != nil {
				fmt.Printf("Error parsing .gitignore: %v\n", err)
			}
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

		fmt.Printf("Processing: %s\n", relPath)

		if !params.includeHidden && strings.HasPrefix(filepath.Base(path), ".") {
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

		if info.IsDir() {
			if params.excludedDirs[info.Name()] {
				return filepath.SkipDir
			}
			return collator.writeContent(fmt.Sprintf("## Directory: %s\n\n", relPath))
		}

		if params.excludedFiles[info.Name()] {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if params.excludedExts[ext] {
			return nil
		}

		if info.Size() > params.maxFileSize {
			return collator.writeContent(fmt.Sprintf("## File: %s (Skipped - Size: %d bytes)\n\n", relPath, info.Size()))
		}

		metadata := fmt.Sprintf("## File: %s\n\nSize: %d bytes\n\nLast Modified: %s\n\n",
			relPath, info.Size(), info.ModTime().Format(time.RFC3339))

		if err := collator.writeContent(metadata); err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), "_collated.md") {
			return collator.writeContent("This is a collated file from a previous run. Contents not displayed.\n\n")
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %q: %v\n", path, err)
			return err
		}

		if isTextFile(content) {
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

	fmt.Printf("Collection complete. Output files generated with prefix %s\n", collator.baseFileName)
}

func reconstructFromFile(inputFile string, params *Parameters) {
	basePath := strings.TrimSuffix(inputFile, "_part1.md")
	basePath = strings.TrimSuffix(basePath, ".md")
	pattern := basePath + "*.md"

	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error finding collated files: %v\n", err)
		os.Exit(1)
	}

	if len(matches) == 0 {
		fmt.Printf("No collated files found matching pattern: %s\n", pattern)
		os.Exit(1)
	}

	var allFiles []FileInfo
	var rootDir string

	for _, match := range matches {
		fmt.Printf("Processing file: %s\n", match)
		currentRootDir, files, err := parseInputFile(match)
		if err != nil {
			fmt.Printf("Error parsing input file %s: %v\n", match, err)
			os.Exit(1)
		}

		if rootDir == "" {
			rootDir = currentRootDir
		} else if rootDir != currentRootDir {
			fmt.Printf("Warning: Inconsistent root directories found. Using %s\n", rootDir)
		}

		allFiles = append(allFiles, files...)
	}

	err = os.MkdirAll(rootDir, 0755)
	if err != nil {
		fmt.Printf("Error creating root directory: %v\n", err)
		os.Exit(1)
	}

	err = os.Chdir(rootDir)
	if err != nil {
		fmt.Printf("Error changing to root directory: %v\n", err)
		os.Exit(1)
	}

	err = reconstructFiles(allFiles, params)
	if err != nil {
		fmt.Printf("Error reconstructing files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Files and directories have been reconstructed successfully in %s!\n", rootDir)
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

		if strings.HasPrefix(line, "Root Directory: ") {
			rootDir = strings.TrimPrefix(line, "Root Directory: ")
			continue
		}

		if strings.HasPrefix(line, "# Project Files Summary") ||
			strings.HasPrefix(line, "Generated on:") ||
			strings.HasPrefix(line, "---") {
			continue
		}

		if strings.HasPrefix(line, "## Directory: ") {
			dirPath := strings.TrimPrefix(line, "## Directory: ")
			if dirPath == "." {
				dirPath = ""
			}
			files = append(files, FileInfo{
				path:        dirPath,
				isDirectory: true,
			})
			continue
		}

		if strings.HasPrefix(line, "## File: ") {
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
			continue
		}

		if currentFile != nil {
			if strings.HasPrefix(line, "Size: ") {
				size := strings.TrimPrefix(line, "Size: ")
				size = strings.TrimSuffix(size, " bytes")
				fmt.Sscanf(size, "%d", &currentFile.size)
			} else if strings.HasPrefix(line, "Last Modified: ") {
				timeStr := strings.TrimPrefix(line, "Last Modified: ")
				currentFile.lastModified, _ = time.Parse(time.RFC3339, timeStr)
			} else if line == "```" {
				isReadingCode = !isReadingCode
			} else if isReadingCode {
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

func reconstructFiles(files []FileInfo, params *Parameters) error {
	// First pass: create all directories
	for _, f := range files {
		if f.isDirectory {
			if f.path != "" {
				err := os.MkdirAll(f.path, 0755)
				if err != nil {
					return fmt.Errorf("error creating directory %s: %v", f.path, err)
				}
				fmt.Printf("Created directory: %s\n", f.path)
			}
		}
	}

	// Second pass: create all files
	for _, f := range files {
		if !f.isDirectory {
			dir := filepath.Dir(f.path)
			if dir != "." {
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					return fmt.Errorf("error creating parent directory for %s: %v", f.path, err)
				}
			}

			file, err := os.Create(f.path)
			if err != nil {
				return fmt.Errorf("error creating file %s: %v", f.path, err)
			}

			_, err = file.WriteString(f.content.String())
			file.Close()
			if err != nil {
				return fmt.Errorf("error writing to file %s: %v", f.path, err)
			}

			if params.preserveTimestamp && !f.lastModified.IsZero() {
				err = os.Chtimes(f.path, f.lastModified, f.lastModified)
				if err != nil {
					return fmt.Errorf("error setting file time for %s: %v", f.path, err)
				}
			}

			fmt.Printf("Created file: %s\n", f.path)
		}
	}

	return nil
}

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

func getLanguage(extension string) string {
	switch extension {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".jsx":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".cpp", ".hpp", ".h", ".cc":
		return "cpp"
	case ".c":
		return "c"
	case ".cs":
		return "csharp"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".scss", ".sass":
		return "scss"
	case ".md":
		return "markdown"
	case ".sh", ".bash":
		return "bash"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".yaml", ".yml":
		return "yaml"
	case ".sql":
		return "sql"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".r":
		return "r"
	case ".dart":
		return "dart"
	case ".lua":
		return "lua"
	case ".pl":
		return "perl"
	case ".gradle":
		return "gradle"
	case ".dockerfile", ".containerfile":
		return "dockerfile"
	case ".tf":
		return "terraform"
	case ".vue":
		return "vue"
	case ".ini", ".conf":
		return "ini"
	case ".toml":
		return "toml"
	case ".proto":
		return "protobuf"
	default:
		return ""
	}
}

func isTextFile(content []byte) bool {
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
