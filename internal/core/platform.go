package core

import (
	"context"
	"sync"
	"time"

	"github.com/nathfavour/noplacelike.go/internal/logger"
)

var (
	buildVersion = "unknown"
	buildTime    = "unknown"
	gitCommit    = "unknown"
)

// SetBuildInfo sets the build information
func SetBuildInfo(version, buildT, commit string) {
	buildVersion = version
	buildTime = buildT
	gitCommit = commit
}

// GetBuildInfo returns build information
func GetBuildInfo() (version, buildT, commit string) {
	return buildVersion, buildTime, gitCommit
}

// Platform represents the core NoPlaceLike platform
type Platform struct {
	config        *Config
	logger        logger.Logger
	pluginMgr     PluginManager
	networkMgr    NetworkManager
	resourceMgr   ResourceManager
	securityMgr   SecurityManager
	httpService   HTTPService
	eventBus      EventBus
	healthChecker HealthChecker
	metrics       MetricsCollector

	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
}

// NewPlatform creates a new platform instance
func NewPlatform(config *Config) *Platform {
	log := logger.New()

	return &Platform{
		config:   config,
		logger:   log,
		stopChan: make(chan struct{}),
	}
}

// Start initializes and starts all platform services
func (p *Platform) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return ErrAlreadyRunning
	}

	p.logger.Info("Starting NoPlaceLike Platform",
		"version", buildVersion,
		"commit", gitCommit)

	// Initialize core components
	if err := p.initializeComponents(ctx); err != nil {
		return err
	}

	// Start services in order
	if err := p.startServices(ctx); err != nil {
		return err
	}

	p.running = true
	p.logger.Info("Platform started successfully")

	return nil
}

// Stop gracefully shuts down the platform
func (p *Platform) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.logger.Info("Stopping NoPlaceLike Platform")

	// Stop services in reverse order
	if err := p.stopServices(ctx); err != nil {
		p.logger.Error("Error stopping services", "error", err)
	}

	close(p.stopChan)
	p.running = false

	p.logger.Info("Platform stopped successfully")
	return nil
}

// Wait blocks until the platform is stopped
func (p *Platform) Wait() {
	<-p.stopChan
}

// IsRunning returns true if the platform is currently running
func (p *Platform) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// GetLogger returns the platform logger
func (p *Platform) GetLogger() logger.Logger {
	return p.logger
}

// GetConfig returns the platform configuration
func (p *Platform) GetConfig() *Config {
	return p.config
}

// GetPluginManager returns the plugin manager
func (p *Platform) GetPluginManager() PluginManager {
	return p.pluginMgr
}

// GetNetworkManager returns the network manager
func (p *Platform) GetNetworkManager() NetworkManager {
	return p.networkMgr
}

// GetResourceManager returns the resource manager
func (p *Platform) GetResourceManager() ResourceManager {
	return p.resourceMgr
}

// GetSecurityManager returns the security manager
func (p *Platform) GetSecurityManager() SecurityManager {
	return p.securityMgr
}

// GetEventBus returns the event bus
func (p *Platform) GetEventBus() EventBus {
	return p.eventBus
}

// GetMetrics returns the metrics collector
func (p *Platform) GetMetrics() MetricsCollector {
	return p.metrics
}

// GetHTTPService returns the HTTP service
func (p *Platform) GetHTTPService() HTTPService {
	return p.httpService
}

// GetHealthChecker returns the health checker
func (p *Platform) GetHealthChecker() HealthChecker {
	return p.healthChecker
}

// initializeComponents initializes all platform components
func (p *Platform) initializeComponents(ctx context.Context) error {
	var err error

	// Initialize event bus first (other components depend on it)
	p.eventBus = NewEventBus(p.logger)

	// Initialize metrics collector
	p.metrics = NewMetricsCollector()

	// Initialize health checker
	p.healthChecker = NewHealthChecker(p.logger, p.metrics)

	// Initialize security manager
	p.securityMgr, err = NewSecurityManager(p.config.Security, p.logger)
	if err != nil {
		return err
	}

	// Initialize resource manager
	p.resourceMgr = NewResourceManager(p.logger, p.eventBus)

	// Initialize network manager
	p.networkMgr, err = NewNetworkManager(p.config.Network, p.logger, p.eventBus)
	if err != nil {
		return err
	}

	// Initialize plugin manager
	p.pluginMgr, err = NewPluginManager(p.config.Plugins, p.logger, p)
	if err != nil {
		return err
	}

	// Initialize HTTP service
	p.httpService, err = NewHTTPService(p.config.Network, p.logger, p)
	if err != nil {
		return err
	}

	return nil
}

// startServices starts all platform services
func (p *Platform) startServices(ctx context.Context) error {
	// Start core services
	if err := p.eventBus.Start(ctx); err != nil {
		return err
	}

	if err := p.metrics.Start(ctx); err != nil {
		return err
	}

	if err := p.healthChecker.Start(ctx); err != nil {
		return err
	}

	if err := p.securityMgr.Start(ctx); err != nil {
		return err
	}

	if err := p.resourceMgr.Start(ctx); err != nil {
		return err
	}

	if err := p.networkMgr.Start(ctx); err != nil {
		return err
	}

	if err := p.pluginMgr.Start(ctx); err != nil {
		return err
	}

	// Start HTTP service last
	if err := p.httpService.Start(ctx); err != nil {
		return err
	}

	return nil
}

// stopServices stops all platform services
func (p *Platform) stopServices(ctx context.Context) error {
	// Create timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Stop services in reverse order
	if p.httpService != nil {
		if err := p.httpService.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping HTTP service", "error", err)
		}
	}

	if p.pluginMgr != nil {
		if err := p.pluginMgr.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping plugin manager", "error", err)
		}
	}

	if p.networkMgr != nil {
		if err := p.networkMgr.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping network manager", "error", err)
		}
	}

	if p.resourceMgr != nil {
		if err := p.resourceMgr.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping resource manager", "error", err)
		}
	}

	if p.securityMgr != nil {
		if err := p.securityMgr.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping security manager", "error", err)
		}
	}

	if p.healthChecker != nil {
		if err := p.healthChecker.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping health checker", "error", err)
		}
	}

	if p.metrics != nil {
		if err := p.metrics.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping metrics collector", "error", err)
		}
	}

	if p.eventBus != nil {
		if err := p.eventBus.Stop(shutdownCtx); err != nil {
			p.logger.Error("Error stopping event bus", "error", err)
		}
	}

	return nil
}
