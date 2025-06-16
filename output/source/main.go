package main

import (
	"fmt"
	"os"

	"github.com/jonathanleahy/folder-bundler/internal/collect"
	"github.com/jonathanleahy/folder-bundler/internal/config"
	"github.com/jonathanleahy/folder-bundler/internal/reconstruct"
)

func main() {
	if len(os.Args) < 2 {
		config.PrintUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	os.Args = os.Args[1:] // Shift arguments for flag parsing

	params, err := config.ParseParameters()
	if err != nil {
		fmt.Printf("Error parsing parameters: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "collect":
		if len(os.Args) > 1 {
			params.RootDir = os.Args[1]
		}
		if err := collect.ProcessDirectory(params); err != nil {
			fmt.Printf("Error during collection: %v\n", err)
			os.Exit(1)
		}
	case "reconstruct":
		if len(os.Args) < 2 {
			config.PrintReconstructHelp()
			os.Exit(1)
		}
		if err := reconstruct.FromFile(os.Args[1], params); err != nil {
			fmt.Printf("Error during reconstruction: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		config.PrintUsage()
		os.Exit(1)
	}
}
