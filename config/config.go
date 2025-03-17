package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	UploadFolder   string   `json:"upload_folder"`
	DownloadFolder string   `json:"download_folder"`
	AudioFolders   []string `json:"audio_folders"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return &Config{
		Host:           "0.0.0.0",
		Port:           8000,
		UploadFolder:   filepath.Join(homeDir, "noplacelike", "uploads"),
		DownloadFolder: filepath.Join(homeDir, "Downloads"),
		AudioFolders:   []string{},
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
