package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathfavour/noplacelike.go/config"
	"github.com/nathfavour/noplacelike.go/server"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for the noplacelike CLI application
func NewRootCmd(cfg *config.Config) *cobra.Command {
	var host string
	var port int
	var uploadFolder string

	rootCmd := &cobra.Command{
		Use:   "noplacelike",
		Short: "NoPlaceLike is a network resource sharing application",
		Long: `NoPlaceLike is your virtual distributed operating system for effortlessly
streaming clipboard data, files, music, and more across devices—wirelessly and seamlessly!`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Ensure upload directory exists
			uploadDir := cfg.UploadFolder // Changed from cfg.DownloadFolder
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				if err := os.MkdirAll(uploadDir, 0755); err != nil {
					cmd.PrintErrf("Error creating upload directory: %v\n", err)
					os.Exit(1)
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Update config with flag values if provided
			if cmd.Flags().Changed("host") {
				cfg.Host = host
			}
			if cmd.Flags().Changed("port") {
				cfg.Port = port
			}
			if cmd.Flags().Changed("upload-folder") {
				cfg.UploadFolder = uploadFolder
			}

			// Save updated configuration
			if err := config.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save configuration: %v\n", err)
			}

			// Start the server
			srv := server.NewServer(cfg)
			go func() {
				if err := srv.Start(); err != nil {
					fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
					os.Exit(1)
				}
			}()

			// Display access information
			server.DisplayAccessInfo(cfg.Host, cfg.Port)

			// Wait for interrupt signal to gracefully shutdown the server
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			fmt.Println("\nShutting down server...")
			srv.Shutdown()
		},
	}

	// Define flags
	rootCmd.Flags().StringVarP(&host, "host", "", "0.0.0.0", "Host address to bind to")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8000, "Port to listen on")
	rootCmd.Flags().StringVarP(&uploadFolder, "upload-folder", "u", "", "Custom folder for uploads")

	// Add sub-commands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newConfigCmd(cfg))

	return rootCmd
}
