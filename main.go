package main

import (
	"fmt"
	"os"

	"github.com/nathfavour/noplacelike.go/cmd"
	"github.com/nathfavour/noplacelike.go/config"
)

func main() {
	// Load config at startup
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
	}

	// Initialize the root command
	rootCmd := cmd.NewRootCmd(cfg)
	
	// Execute command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
