package config

import (
	"flag"
	"fmt"
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
	fmt.Printf(`Folder Bundler v2.1

Usage: bundler <command> [flags] [path]

Commands:
  collect     Create directory structure summary
  reconstruct Build from summary file

Flags:
  -max          Maximum file size (default: 2MB)
  -out-max      Maximum output file size (default: 2MB)
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
  bundler collect -compress dictionary -max 5242880 myproject
  bundler reconstruct myproject_collated_part1.md
`)
}

func PrintReconstructHelp() {
	fmt.Printf(`Folder Bundler v2.1

Usage: bundler reconstruct [flags] <input_file>

Flags:
  -time  Preserve timestamps (default: true)

Example:
  bundler reconstruct myproject_collated_part1.md
`)
}

func ParseParameters() (*Parameters, error) {
	var params Parameters
	var excludeDirs, excludeFiles, excludeExts string

	defaultExcludeDirs := strings.Join([]string{
		"node_modules", "dist", "build", "coverage", "tmp",
		".git", ".next", ".idea", ".vscode", ".cache", ".build",
		".vercel", ".turbo", ".yarn", ".npm",
	}, ",")

	defaultExcludeFiles := "package-lock.json,yarn.lock,.DS_Store,.env"
	defaultExcludeExts := ".exe,.dll,.so,.dylib,.bin,.pkl,.pyc,.bak"

	flag.Int64Var(&params.MaxFileSize, "max", 2*1024*1024, "Maximum file size")
	flag.Int64Var(&params.MaxOutputSize, "out-max", 2*1024*1024, "Maximum output size")
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

	// Parse flags and set compression
	params.EnableCompression = params.CompressionStrategy != "none"
	
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
