package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nathfavour/noplacelike.go/cmd"
	"github.com/nathfavour/noplacelike.go/config"
)

func main() {
	// Load config at startup
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Ensure default upload folder exists
	if cfg.UploadFolder == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		cfg.UploadFolder = filepath.Join(homeDir, ".noplacelike", "uploads")
	}

	// Don't create directory at startup, let the handlers create it on demand
	// Only validate the path is properly formatted
	if strings.HasPrefix(cfg.UploadFolder, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			cfg.UploadFolder = filepath.Join(homeDir, cfg.UploadFolder[1:])
		}
	}

	// Same for download folder
	if cfg.DownloadFolder == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		cfg.DownloadFolder = filepath.Join(homeDir, ".noplacelike", "downloads")
	}

	// Don't create directory at startup, just validate/expand path
	if strings.HasPrefix(cfg.DownloadFolder, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			cfg.DownloadFolder = filepath.Join(homeDir, cfg.DownloadFolder[1:])
		}
	}

	// Initialize the root command
	rootCmd := cmd.NewRootCmd(cfg)
	
	// Execute command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
