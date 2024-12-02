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
	fmt.Printf(`Usage: bundler <command> [flags] [path]

Commands:
  collect     Create directory structure summary
  reconstruct Build from summary file

Flags:
  -max-file        Maximum file size (default: 1MB)
  -exclude-dirs    Skip directories (default: node_modules,vendor,...)
  -include-hidden  Include hidden files
`)
}

func PrintReconstructHelp() {
	fmt.Printf(`Usage: bundler reconstruct [flags] <input_file>

Flags:
  -preserve-time  Keep original timestamps (default: true)
`)
}

func ParseParameters() (*Parameters, error) {
	var params Parameters
	var excludeDirs, excludeFiles, excludeExts string

	flag.Int64Var(&params.MaxFileSize, "max-file", 1*1024*1024, "Maximum size of individual files")
	flag.Int64Var(&params.MaxOutputSize, "max-output", 2*1024*1024, "Maximum size of output files")
	flag.StringVar(&excludeDirs, "exclude-dirs", "node_modules,vendor,venv,dist,build", "Directories to exclude")
	flag.StringVar(&excludeFiles, "exclude-files", "package-lock.json,yarn.lock", "Files to exclude")
	flag.StringVar(&excludeExts, "exclude-exts", ".exe,.dll,.so,.dylib,.bin,.pkl,.pyc,.bak", "Extensions to exclude")
	flag.BoolVar(&params.IncludeHidden, "include-hidden", false, "Include hidden files")
	flag.BoolVar(&params.SkipGitignore, "skip-gitignore", false, "Skip .gitignore processing")
	flag.BoolVar(&params.PreserveTimestamp, "preserve-time", true, "Preserve original timestamps")

	flag.Parse()

	params.ExcludedDirs = stringToMap(excludeDirs)
	params.ExcludedFiles = stringToMap(excludeFiles)
	params.ExcludedExts = stringToMap(excludeExts)
	params.RootDir = "."

	return &params, nil
}
