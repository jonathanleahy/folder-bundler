package main

import (
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
