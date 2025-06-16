package config

import (
	"flag"
	"fmt"
	"strconv"
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
	// Compression settings
	CompressionStrategy string
	EnableCompression   bool
}

func PrintUsage() {
	fmt.Printf(`Folder Bundler v2.8

Usage: bundler <command> [flags] [path]

Commands:
  collect     Create directory structure summary
  reconstruct Build from summary file

Flags:
  -max          Maximum file size (default: 2M, accepts: 500K, 1M, 2G, etc.)
  -out-max      Maximum output file size (default: 2M)
  -skip-dirs    Skip directories (default: node_modules,.git,...)
  -skip-files   Skip files (default: .DS_Store,.env,...)
  -skip-ext     Skip extensions (default: .exe,.dll,...)
  -hidden       Include hidden files (default: false)
  -no-gitignore Skip .gitignore (default: false)
  -time         Preserve timestamps (default: true)
  -compress     Compression: none|auto|dictionary|template|delta|template+delta (default: none)

Examples:
  bundler collect myproject
  bundler collect -compress auto myproject
  bundler collect -compress dictionary -max 5M myproject
  bundler collect myproject -max 1G -out-max 10M
  bundler reconstruct myproject_collated_part1.fb
`)
}

func PrintReconstructHelp() {
	fmt.Printf(`Folder Bundler v2.8

Usage: bundler reconstruct [flags] <input_file>

Flags:
  -time  Preserve timestamps (default: true)

Example:
  bundler reconstruct myproject_collated_part1.fb
`)
}

func ParseParameters() (*Parameters, error) {
	var params Parameters
	var excludeDirs, excludeFiles, excludeExts string
	var maxFileSizeStr, maxOutputSizeStr string

	defaultExcludeDirs := strings.Join([]string{
		"node_modules", "dist", "build", "coverage", "tmp",
		".git", ".next", ".idea", ".vscode", ".cache", ".build",
		".vercel", ".turbo", ".yarn", ".npm",
	}, ",")

	defaultExcludeFiles := "package-lock.json,yarn.lock,.DS_Store,.env"
	defaultExcludeExts := ".exe,.dll,.so,.dylib,.bin,.pkl,.pyc,.bak"

	flag.StringVar(&maxFileSizeStr, "max", "2M", "Maximum file size (e.g. 2M, 500K, 1G)")
	flag.StringVar(&maxOutputSizeStr, "out-max", "2M", "Maximum output size (e.g. 2M, 500K, 1G)")
	flag.StringVar(&excludeDirs, "skip-dirs", defaultExcludeDirs, "Skip directories")
	flag.StringVar(&excludeFiles, "skip-files", defaultExcludeFiles, "Skip files")
	flag.StringVar(&excludeExts, "skip-ext", defaultExcludeExts, "Skip extensions")
	flag.BoolVar(&params.IncludeHidden, "hidden", false, "Include hidden files")
	flag.BoolVar(&params.SkipGitignore, "no-gitignore", false, "Skip .gitignore")
	flag.BoolVar(&params.PreserveTimestamp, "time", true, "Preserve timestamps")
	flag.StringVar(&params.CompressionStrategy, "compress", "none", "Compression (none|auto|dictionary|template|delta|template+delta)")

	flag.Parse()

	params.ExcludedDirs = stringToMap(excludeDirs)
	params.ExcludedFiles = stringToMap(excludeFiles)
	params.ExcludedExts = stringToMap(excludeExts)
	params.RootDir = "."

	// Parse size strings
	maxFileSize, err := parseSize(maxFileSizeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid max file size '%s': %v", maxFileSizeStr, err)
	}
	params.MaxFileSize = maxFileSize

	maxOutputSize, err := parseSize(maxOutputSizeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid max output size '%s': %v", maxOutputSizeStr, err)
	}
	params.MaxOutputSize = maxOutputSize

	// Parse flags and set compression
	// Check if compress flag was explicitly set by looking at Visit
	compressSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "compress" {
			compressSet = true
		}
	})
	
	// Enable compression if flag was set (even if value is "none")
	params.EnableCompression = compressSet
	
	// Validate compression settings
	validStrategies := map[string]bool{
		"none":           true,
		"auto":           true,
		"dictionary":     true,
		"template":       true,
		"delta":          true,
		"template+delta": true,
	}

	if !validStrategies[params.CompressionStrategy] {
		return nil, fmt.Errorf("invalid compression '%s'. Valid options: none, auto, dictionary, template, delta, template+delta", params.CompressionStrategy)
	}

	return &params, nil
}

func stringToMap(s string) map[string]bool {
	result := make(map[string]bool)
	for _, item := range strings.Split(s, ",") {
		result[strings.TrimSpace(item)] = true
	}
	return result
}

// parseSize parses human-readable size strings like "1M", "500K", "2G"
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Check if it's just a number (bytes)
	if size, err := strconv.ParseInt(s, 10, 64); err == nil {
		return size, nil
	}

	// Extract number and unit
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	
	// Handle optional 'B' suffix (e.g., "2MB" vs "2M")
	if unit == 'B' && len(numStr) > 0 {
		unit = numStr[len(numStr)-1]
		numStr = numStr[:len(numStr)-1]
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", numStr)
	}

	var multiplier float64
	switch unit {
	case 'K':
		multiplier = 1024
	case 'M':
		multiplier = 1024 * 1024
	case 'G':
		multiplier = 1024 * 1024 * 1024
	case 'T':
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid unit: %c (use K, M, G, or T)", unit)
	}

	return int64(num * multiplier), nil
}
