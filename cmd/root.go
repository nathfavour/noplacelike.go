package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/nathfavour/noplacelike.go/internal/core"
	"github.com/spf13/cobra"
)

var (
	configFile string
	host       string
	port       int
	logLevel   string
	enableAuth bool
	enableTLS  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "noplacelike",
	Short: "Professional Distributed Network Resource Sharing Platform",
	Long: `NoPlaceLike is a professional distributed operating system designed for 
seamless resource sharing across networks. Built from the ground up in Go with 
a robust plugin architecture, it provides enterprise-grade performance, security, 
and extensibility for modern distributed computing environments.`,
	RunE: runPlatform,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.noplacelike.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "logging level (debug, info, warn, error)")

	// Server flags
	rootCmd.Flags().StringVar(&host, "host", "0.0.0.0", "host address to bind to")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVar(&enableAuth, "enable-auth", false, "enable authentication")
	rootCmd.Flags().BoolVar(&enableTLS, "enable-tls", false, "enable TLS/HTTPS")
}

func initConfig() {
	// Set log level from environment or flag
	if logLevel != "" {
		os.Setenv("LOG_LEVEL", logLevel)
	}
}

func runPlatform(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Create configuration
	config := core.DefaultConfig()

	// Override with command line flags
	if host != "" {
		config.Network.Host = host
	}
	if port > 0 {
		config.Network.Port = port
	}
	config.Security.EnableAuth = enableAuth
	config.Network.EnableTLS = enableTLS

	// Load config file if specified
	if configFile != "" {
		if err := loadConfigFile(config, configFile); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Create and start platform
	platform := core.NewPlatform(config)

	// Start platform
	if err := platform.Start(ctx); err != nil {
		return fmt.Errorf("failed to start platform: %w", err)
	}

	// Wait for shutdown signal
	platform.Wait()

	// Graceful shutdown
	if err := platform.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop platform: %w", err)
	}

	return nil
}

func loadConfigFile(config *core.Config, filename string) error {
	// TODO: Implement config file loading with viper
	// For now, just return nil
	return nil
}
