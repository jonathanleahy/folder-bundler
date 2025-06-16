package main

import (
	"fmt"
	"os"
	"strings"
	
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
	
	// Parse command-specific flags
	switch command {
	case "collect":
		// Extract path and reorder arguments
		args := os.Args[2:]
		var path string
		var flags []string
		
		for i := 0; i < len(args); i++ {
			if strings.HasPrefix(args[i], "-") {
				flags = append(flags, args[i])
				// Check if this flag has a value
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					if args[i] == "-compress" || args[i] == "-skip-dirs" || args[i] == "-skip-files" || args[i] == "-skip-ext" || args[i] == "-max" || args[i] == "-out-max" {
						i++
						flags = append(flags, args[i])
					}
				}
			} else if path == "" {
				path = args[i]
			}
		}
		
		// Reconstruct os.Args with flags first, then path
		os.Args = append([]string{os.Args[0]}, flags...)
		if path != "" {
			os.Args = append(os.Args, path)
		}
		
		params, err := config.ParseParameters()
		if err != nil {
			fmt.Printf("Error parsing parameters: %v\n", err)
			os.Exit(1)
		}
		
		// Set the root directory
		if path != "" {
			params.RootDir = path
		}
		
		if err := collect.ProcessDirectory(params); err != nil {
			fmt.Printf("Error during collection: %v\n", err)
			os.Exit(1)
		}
		
	case "reconstruct":
		// Extract path and reorder arguments
		args := os.Args[2:]
		var path string
		var flags []string
		
		for i := 0; i < len(args); i++ {
			if strings.HasPrefix(args[i], "-") {
				flags = append(flags, args[i])
				// Check if this flag has a value
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					if args[i] == "-time" {
						// -time is a boolean flag, don't consume next arg
					}
				}
			} else if path == "" {
				path = args[i]
			}
		}
		
		// Reconstruct os.Args with flags first, then path
		os.Args = append([]string{os.Args[0]}, flags...)
		if path != "" {
			os.Args = append(os.Args, path)
		}
		
		params, err := config.ParseParameters()
		if err != nil {
			fmt.Printf("Error parsing parameters: %v\n", err)
			os.Exit(1)
		}
		
		if path == "" {
			config.PrintReconstructHelp()
			os.Exit(1)
		}
		
		if err := reconstruct.FromFile(path, params); err != nil {
			fmt.Printf("Error during reconstruction: %v\n", err)
			os.Exit(1)
		}
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		config.PrintUsage()
		os.Exit(1)
	}
}
