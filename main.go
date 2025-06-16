package main

import (
	"flag"
	"fmt"
	"github.com/jonathanleahy/folder-bundler/internal/collect"
	"github.com/jonathanleahy/folder-bundler/internal/config"
	"github.com/jonathanleahy/folder-bundler/internal/reconstruct"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		config.PrintUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	
	// Parse command-specific flags
	switch command {
	case "collect":
		// Shift os.Args to exclude the command
		os.Args = os.Args[1:]
		params, err := config.ParseParameters()
		if err != nil {
			fmt.Printf("Error parsing parameters: %v\n", err)
			os.Exit(1)
		}
		
		// Get remaining arguments after flag parsing
		args := flag.Args()
		if len(args) > 0 {
			params.RootDir = args[0]
		}
		
		if err := collect.ProcessDirectory(params); err != nil {
			fmt.Printf("Error during collection: %v\n", err)
			os.Exit(1)
		}
		
	case "reconstruct":
		// Shift os.Args to exclude the command
		os.Args = os.Args[1:]
		params, err := config.ParseParameters()
		if err != nil {
			fmt.Printf("Error parsing parameters: %v\n", err)
			os.Exit(1)
		}
		
		args := flag.Args()
		if len(args) < 1 {
			config.PrintReconstructHelp()
			os.Exit(1)
		}
		
		if err := reconstruct.FromFile(args[0], params); err != nil {
			fmt.Printf("Error during reconstruction: %v\n", err)
			os.Exit(1)
		}
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		config.PrintUsage()
		os.Exit(1)
	}
}
