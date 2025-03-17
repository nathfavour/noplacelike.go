package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nathfavour/noplacelike.go/config"
	"github.com/spf13/cobra"
)

func newConfigCmd(cfg *config.Config) *cobra.Command {
	var audioFolder string
	var clearAudioFolders bool

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "View or update configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Handle clearing audio folders
			if clearAudioFolders {
				cfg.AudioFolders = []string{}
				fmt.Println("Cleared audio folders list")
			}

			// Add audio folder if provided
			if audioFolder != "" {
				cfg.AudioFolders = append(cfg.AudioFolders, audioFolder)
				fmt.Printf("Added audio folder: %s\n", audioFolder)
			}

			// If any changes were made, save them
			if clearAudioFolders || audioFolder != "" {
				if err := config.Save(cfg); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update configuration: %v\n", err)
					return
				}
			}

			// Print current config
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to marshal config: %v\n", err)
				return
			}
			fmt.Println("Current configuration:")
			fmt.Println(string(data))
		},
	}

	// Define flags
	configCmd.Flags().StringVarP(&audioFolder, "add-audio-folder", "a", "", "Add a folder to audio directories list")
	configCmd.Flags().BoolVarP(&clearAudioFolders, "clear-audio-folders", "c", false, "Clear all audio folders from config")

	return configCmd
}
