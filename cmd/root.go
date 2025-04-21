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
	// CLI flag variables for all config fields
	var host string
	var port int
	var uploadFolder string
	var downloadFolder string
	var audioFolders []string
	var allowedPaths []string
	var showHidden bool
	var enableShell bool
	var enableAudioStreaming bool
	var enableScreenStreaming bool
	var allowedCommands []string
	var maxFileContentSize int
	var clipboardHistorySize int

	rootCmd := &cobra.Command{
		Use:   "noplacelike",
		Short: "NoPlaceLike is a network resource sharing application",
		Long: `NoPlaceLike is your virtual distributed operating system for effortlessly
streaming clipboard data, files, music, and more across devices—wirelessly and seamlessly!`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Ensure upload directory exists
			uploadDir := cfg.UploadFolder
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
			if cmd.Flags().Changed("download-folder") {
				cfg.DownloadFolder = downloadFolder
			}
			if cmd.Flags().Changed("audio-folders") {
				cfg.AudioFolders = audioFolders
			}
			if cmd.Flags().Changed("allowed-paths") {
				cfg.AllowedPaths = allowedPaths
			}
			if cmd.Flags().Changed("show-hidden") {
				cfg.ShowHidden = showHidden
			}
			if cmd.Flags().Changed("enable-shell") {
				cfg.EnableShell = enableShell
			}
			if cmd.Flags().Changed("enable-audio-streaming") {
				cfg.EnableAudioStreaming = enableAudioStreaming
			}
			if cmd.Flags().Changed("enable-screen-streaming") {
				cfg.EnableScreenStreaming = enableScreenStreaming
			}
			if cmd.Flags().Changed("allowed-commands") {
				cfg.AllowedCommands = allowedCommands
			}
			if cmd.Flags().Changed("max-file-content-size") {
				cfg.MaxFileContentSize = maxFileContentSize
			}
			if cmd.Flags().Changed("clipboard-history-size") {
				cfg.ClipboardHistorySize = clipboardHistorySize
			}

			// Save updated configuration
			if err := config.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save configuration: %v\n", err)
			}

			// Start the server
			srv := server.NewServer(cfg)
			go func() {
				srv.Start()
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

	// Define flags for all config fields
	rootCmd.Flags().StringVar(&host, "host", cfg.Host, "Host address to bind to")
	rootCmd.Flags().IntVarP(&port, "port", "p", cfg.Port, "Port to listen on")
	rootCmd.Flags().StringVarP(&uploadFolder, "upload-folder", "u", cfg.UploadFolder, "Custom folder for uploads")
	rootCmd.Flags().StringVar(&downloadFolder, "download-folder", cfg.DownloadFolder, "Custom folder for downloads")
	rootCmd.Flags().StringSliceVar(&audioFolders, "audio-folders", cfg.AudioFolders, "Comma-separated list of audio folders")
	rootCmd.Flags().StringSliceVar(&allowedPaths, "allowed-paths", cfg.AllowedPaths, "Comma-separated list of allowed paths for browsing")
	rootCmd.Flags().BoolVar(&showHidden, "show-hidden", cfg.ShowHidden, "Show hidden files in directory listings")
	rootCmd.Flags().BoolVar(&enableShell, "enable-shell", cfg.EnableShell, "Enable shell command execution API")
	rootCmd.Flags().BoolVar(&enableAudioStreaming, "enable-audio-streaming", cfg.EnableAudioStreaming, "Enable audio streaming API")
	rootCmd.Flags().BoolVar(&enableScreenStreaming, "enable-screen-streaming", cfg.EnableScreenStreaming, "Enable screen streaming API")
	rootCmd.Flags().StringSliceVar(&allowedCommands, "allowed-commands", cfg.AllowedCommands, "Comma-separated list of allowed shell commands")
	rootCmd.Flags().IntVar(&maxFileContentSize, "max-file-content-size", cfg.MaxFileContentSize, "Maximum file content size for reading (bytes)")
	rootCmd.Flags().IntVar(&clipboardHistorySize, "clipboard-history-size", cfg.ClipboardHistorySize, "Clipboard history size")

	// Add sub-commands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newConfigCmd(cfg))

	return rootCmd
}
