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
	p, err := platform.NewPlatform(platformConfig, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize platform: %v\n", err)
		os.Exit(1)
	}
	// Set logger if method exists
	if setter, ok := interface{}(p).(interface{ SetLogger(core.Logger) }); ok {
		setter.SetLogger(log)
	}

	// Display QR codes and access info first
	displayAccessInfo(legacy.Host, legacy.Port)

	// Load core plugins BEFORE starting platform so HTTP routes can register them
	if err := loadCorePlugins(ctx, p, legacy); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load core plugins: %v\n", err)
		os.Exit(1)
	}

	// Register HTTP service (platform will start it)
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

	// Start the platform (starts all registered services)
	if err := p.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start platform: %v\n", err)
		os.Exit(1)
	}

	// Plugins are preloaded before platform start; nothing to do here

	// Register a sample in-memory resource to make the resources API functional out of the box
	registerSampleResource(p)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal, gracefully shutting down...")
		// Stop the platform (stops all services/plugins)
		_ = p.Stop(context.Background())
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
			JWTSecret:        legacy.JWTSecret,
			JWTIssuer:        legacy.JWTIssuer,
			JWTAudience:      legacy.JWTAudience,
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

// memoryResource is a simple in-memory core.Resource implementation
type memoryResource struct {
	id       string
	typ      string
	data     []byte
	meta     map[string]interface{}
	started  bool
}

// Service interface methods
func (m *memoryResource) Start(ctx context.Context) error {
	m.started = true
	return nil
}
func (m *memoryResource) Stop(ctx context.Context) error {
	m.started = false
	return nil
}
func (m *memoryResource) IsHealthy() bool { return true }
func (m *memoryResource) Name() string    { return "resource:" + m.id }
func (m *memoryResource) Health() core.HealthStatus {
	return core.HealthStatus{
		Status:    core.HealthStatusHealthy,
		Timestamp: time.Now(),
	}
}
func (m *memoryResource) Configuration() core.ConfigSchema { return core.ConfigSchema{} }

// Resource interface methods
func (m *memoryResource) ID() string                         { return m.id }
func (m *memoryResource) Type() string                       { return m.typ }
func (m *memoryResource) GetMetadata() map[string]interface{} { return m.meta }
func (m *memoryResource) GetSize() int64                     { return int64(len(m.data)) }

// registerSampleResource registers a trivial in-memory resource
func registerSampleResource(p *platform.Platform) {
	res := &memoryResource{
		id:   "mem-hello",
		typ:  "memory",
		data: []byte("hello"),
		meta: map[string]interface{}{"name": "hello"},
	}
	_ = p.ResourceManager().RegisterResource(res)
}
