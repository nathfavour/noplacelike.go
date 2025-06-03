// NoPlaceLike - Professional Distributed Network Resource Sharing Platform
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nathfavour/noplacelike.go/config"
	"github.com/nathfavour/noplacelike.go/internal/core"
	"github.com/nathfavour/noplacelike.go/internal/logger"
	"github.com/nathfavour/noplacelike.go/internal/platform"
	"github.com/nathfavour/noplacelike.go/internal/plugins"
	"github.com/nathfavour/noplacelike.go/internal/services"
	"github.com/nathfavour/noplacelike.go/server"
)

var (
	Version   = "2.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set build info
	core.SetBuildInfo(Version, BuildTime, GitCommit)

	// Load legacy config
	legacy, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Convert legacy config to platform config
	platformConfig := convertLegacyConfig(legacy)

	// Initialize platform
	p, err := platform.NewPlatform(*platformConfig, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize platform: %v\n", err)
		os.Exit(1)
	}

	// Display QR codes and access info first
	displayAccessInfo(legacy.Host, legacy.Port)

	// Load core plugins
	if err := loadCorePlugins(ctx, p, legacy); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load core plugins: %v\n", err)
		os.Exit(1)
	}

	// Start HTTP service and keep reference for shutdown
	httpConfig := services.HTTPConfig{
		Host:           legacy.Host,
		Port:           legacy.Port,
		EnableTLS:      false,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxRequestSize: int64(legacy.MaxFileContentSize),
		EnableCORS:     true,
		EnableMetrics:  true,
		EnableDocs:     true,
		RateLimitRPS:   100,
		EnableGzip:     true,
	}
	httpService := services.NewHTTPService(httpConfig, p)
	if err := p.ServiceManager().RegisterService(httpService); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register HTTP service: %v\n", err)
		os.Exit(1)
	}
	if err := httpService.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start HTTP service: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal, gracefully shutting down...")
		// Stop HTTP service
		httpService.Stop(context.Background())
		os.Exit(0)
	}()

	// Block main goroutine until context is cancelled
	<-ctx.Done()
}

// convertLegacyConfig converts legacy config to new platform config
func convertLegacyConfig(legacy *config.Config) *platform.PlatformConfig {
	return &platform.PlatformConfig{
		Name:        "NoPlaceLike",
		Version:     "2.0.0",
		Environment: "production",

		Network: platform.NetworkConfig{
			Host:              legacy.Host,
			Port:              legacy.Port,
			EnableDiscovery:   true,
			DiscoveryPort:     legacy.Port + 1,
			DiscoveryInterval: 30 * time.Second,
			MaxPeers:          50,
			Timeout:           10 * time.Second,
			KeepAliveInterval: 30 * time.Second,
			EnableTLS:         false,
		},

		Security: platform.SecurityConfig{
			EnableAuth:       false, // Start with auth disabled for compatibility
			AuthMethod:       "token",
			TokenExpiry:      24 * time.Hour,
			EnableEncryption: false, // Start with encryption disabled
			EncryptionAlgo:   "AES-256-GCM",
			MaxLoginAttempts: 3,
			LockoutDuration:  15 * time.Minute,
		},

		Performance: platform.PerformanceConfig{
			MaxConcurrentConnections: 1000,
			MaxRequestSize:           int64(legacy.MaxFileContentSize),
			MaxResponseSize:          100 * 1024 * 1024, // 100MB
			RequestTimeout:           30 * time.Second,
			IdleTimeout:              120 * time.Second,
			ReadTimeout:              30 * time.Second,
			WriteTimeout:             30 * time.Second,
			MaxMemoryUsage:           1024 * 1024 * 1024, // 1GB
			GCInterval:               5 * time.Minute,
		},

		Plugins: platform.PluginsConfig{
			EnablePlugins: true,
			PluginDirs:    []string{"./plugins", "~/.noplacelike/plugins"},
			AutoLoad:      []string{"file-manager", "clipboard", "system-info"},
			Disabled:      []string{},
			Sandbox:       false, // Start with sandbox disabled
		},

		Logging: platform.LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     7, // days
			Compress:   true,
		},

		Metrics: platform.MetricsConfig{
			Enabled:         true,
			Endpoint:        "/metrics",
			Interval:        30 * time.Second,
			RetentionTime:   24 * time.Hour,
			ExportFormat:    "prometheus",
			EnableProfiling: false,
		},
	}
}

// loadCorePlugins loads essential plugins
func loadCorePlugins(ctx context.Context, p *platform.Platform, legacy *config.Config) error {
	// File Manager Plugin
	filePlugin := plugins.NewFileManagerPlugin(
		legacy.UploadFolder,
		legacy.DownloadFolder,
		int64(legacy.MaxFileContentSize),
	)

	if err := p.LoadPlugin(ctx, filePlugin); err != nil {
		return fmt.Errorf("failed to load file manager plugin: %w", err)
	}

	// Clipboard Plugin
	clipboardPlugin := plugins.NewClipboardPlugin(legacy.ClipboardHistorySize)

	if err := p.LoadPlugin(ctx, clipboardPlugin); err != nil {
		return fmt.Errorf("failed to load clipboard plugin: %w", err)
	}

	// System Info Plugin
	systemPlugin := plugins.NewSystemInfoPlugin()

	if err := p.LoadPlugin(ctx, systemPlugin); err != nil {
		return fmt.Errorf("failed to load system info plugin: %w", err)
	}

	return nil
}

// startHTTPService starts the HTTP service
func startHTTPService(ctx context.Context, p *platform.Platform, legacy *config.Config) error {
	httpConfig := services.HTTPConfig{
		Host:           legacy.Host,
		Port:           legacy.Port,
		EnableTLS:      false,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxRequestSize: int64(legacy.MaxFileContentSize),
		EnableCORS:     true,
		EnableMetrics:  true,
		EnableDocs:     true,
		RateLimitRPS:   100,
		EnableGzip:     true,
	}

	httpService := services.NewHTTPService(httpConfig, p)

	err := p.ServiceManager().RegisterService(httpService)

	return err
}

// displayAccessInfo shows connection information
func displayAccessInfo(host string, port int) {
	// Print QR codes and network URLs first
	server.DisplayAccessInfo(host, port)

	// Then print the rest of the CLI output
	fmt.Printf("\n")
	fmt.Printf("🚀 NoPlaceLike Platform Started Successfully!\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\n")
	fmt.Printf("🛠️  API Endpoints:\n")
	fmt.Printf("   • Platform Health: /health\n")
	fmt.Printf("   • Platform Info: /info\n")
	fmt.Printf("   • API Documentation: /api/v1/docs\n")
	fmt.Printf("   • Metrics: /api/platform/metrics\n")
	fmt.Printf("   • Plugin Management: /api/plugins\n")
	fmt.Printf("   • Network Peers: /api/network/peers\n")
	fmt.Printf("   • Resource Management: /api/resources\n")
	fmt.Printf("   • Event Stream: /api/events/stream\n")
	fmt.Printf("\n")
	fmt.Printf("🔌 Plugin APIs:\n")
	fmt.Printf("   • File Manager: /plugins/file-manager/files\n")
	fmt.Printf("   • Clipboard: /plugins/clipboard/clipboard\n")
	fmt.Printf("   • System Info: /plugins/system-info/system/info\n")
	fmt.Printf("\n")
	fmt.Printf("📚 Features:\n")
	fmt.Printf("   ✅ Distributed peer discovery\n")
	fmt.Printf("   ✅ File sharing & management\n")
	fmt.Printf("   ✅ Clipboard synchronization\n")
	fmt.Printf("   ✅ System monitoring\n")
	fmt.Printf("   ✅ Plugin system\n")
	fmt.Printf("   ✅ Real-time events\n")
	fmt.Printf("   ✅ RESTful APIs\n")
	fmt.Printf("   ✅ Health monitoring\n")
	fmt.Printf("   ✅ Metrics collection\n")
	fmt.Printf("\n")
	fmt.Printf("Press Ctrl+C to stop the platform\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}
