package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	// Server settings
	Host           string `json:"host"`
	Port           int    `json:"port"`
	
	// Directory settings
	UploadFolder   string   `json:"uploadFolder"`
	AudioFolders   []string `json:"audioFolders"`
	AllowedPaths   []string `json:"allowedPaths"`
	ShowHidden     bool     `json:"showHidden"`
	
	// Feature flags
	EnableShell           bool `json:"enableShell"`
	EnableAudioStreaming  bool `json:"enableAudioStreaming"`
	EnableScreenStreaming bool `json:"enableScreenStreaming"`
	
	// Security settings
	AllowedCommands     []string `json:"allowedCommands"`
	MaxFileContentSize  int      `json:"maxFileContentSize"` // in bytes
	ClipboardHistorySize int     `json:"clipboardHistorySize"`
	
	// API version
	APIVersion string `json:"apiVersion"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	uploadDir := filepath.Join(homeDir, "Downloads", "noplacelike-uploads")
	
	return &Config{
		Host:                "0.0.0.0",
		Port:                8080,
		UploadFolder:        uploadDir,
		AudioFolders:        []string{},
		AllowedPaths:        []string{homeDir},
		ShowHidden:          false,
		EnableShell:         true,
		EnableAudioStreaming: false,
		EnableScreenStreaming: false,
		AllowedCommands:     []string{},
		MaxFileContentSize:   1024 * 1024, // 1MB
		ClipboardHistorySize: 50,
		APIVersion:          "v1",
	}
}

// configPath returns the path to the config file
func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".noplacelike.json"), nil
}

// Load loads configuration from the config file
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// If config file doesn't exist, create it with default values
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}

	// Read and parse the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	return &cfg, nil
}

// Save saves the configuration to the config file
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal and write config
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
