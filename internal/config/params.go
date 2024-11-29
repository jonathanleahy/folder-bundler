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
