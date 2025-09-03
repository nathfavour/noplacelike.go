// Package platform provides the main platform implementation
package platform

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nathfavour/noplacelike.go/internal/core"
)

// Platform represents the main NoPlaceLike platform instance
type Platform struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core managers
	serviceManager  core.ServiceManager
	networkManager  core.NetworkManager
	resourceManager core.ResourceManager
	securityManager core.SecurityManager
	configManager   core.ConfigManager
	eventBus        core.EventBus
	metrics         core.MetricsCollector
	logger          core.Logger

	// Plugin system
	plugins    map[string]core.Plugin
	pluginDeps map[string][]string

	// Platform state
	started   bool
	startTime time.Time
	version   string
	buildInfo BuildInfo
}

// BuildInfo contains build-time information
type BuildInfo struct {
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildTime time.Time `json:"buildTime"`
	GoVersion string    `json:"goVersion"`
}

// PlatformConfig holds platform-wide configuration
type PlatformConfig struct {
	// Core settings
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`

	// Network settings
	Network NetworkConfig `json:"network"`

	// Security settings
	Security SecurityConfig `json:"security"`

	// Performance settings
	Performance PerformanceConfig `json:"performance"`

	// Plugin settings
	Plugins PluginsConfig `json:"plugins"`

	// Logging settings
	Logging LoggingConfig `json:"logging"`

	// Metrics settings
	Metrics MetricsConfig `json:"metrics"`
}

// NetworkConfig contains network-related settings
type NetworkConfig struct {
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	EnableDiscovery   bool          `json:"enableDiscovery"`
	DiscoveryPort     int           `json:"discoveryPort"`
	DiscoveryInterval time.Duration `json:"discoveryInterval"`
	MaxPeers          int           `json:"maxPeers"`
	Timeout           time.Duration `json:"timeout"`
	KeepAliveInterval time.Duration `json:"keepAliveInterval"`
	EnableTLS         bool          `json:"enableTLS"`
	TLSCertFile       string        `json:"tlsCertFile"`
	TLSKeyFile        string        `json:"tlsKeyFile"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	EnableAuth       bool          `json:"enableAuth"`
	AuthMethod       string        `json:"authMethod"`
	TokenExpiry      time.Duration `json:"tokenExpiry"`
	EnableEncryption bool          `json:"enableEncryption"`
	EncryptionAlgo   string        `json:"encryptionAlgo"`
	MaxLoginAttempts int           `json:"maxLoginAttempts"`
	LockoutDuration  time.Duration `json:"lockoutDuration"`
	AllowedPeers     []string      `json:"allowedPeers"`
	BlockedPeers     []string      `json:"blockedPeers"`
	// JWT settings (HS256)
	JWTSecret   string   `json:"jwtSecret"`
	JWTIssuer   string   `json:"jwtIssuer"`
	JWTAudience []string `json:"jwtAudience"`
}

// PerformanceConfig contains performance-related settings
type PerformanceConfig struct {
	MaxConcurrentConnections int           `json:"maxConcurrentConnections"`
	MaxRequestSize           int64         `json:"maxRequestSize"`
	MaxResponseSize          int64         `json:"maxResponseSize"`
	RequestTimeout           time.Duration `json:"requestTimeout"`
	IdleTimeout              time.Duration `json:"idleTimeout"`
	ReadTimeout              time.Duration `json:"readTimeout"`
	WriteTimeout             time.Duration `json:"writeTimeout"`
	MaxMemoryUsage           int64         `json:"maxMemoryUsage"`
	GCInterval               time.Duration `json:"gcInterval"`
}

// PluginsConfig contains plugin-related settings
type PluginsConfig struct {
	EnablePlugins bool     `json:"enablePlugins"`
	PluginDirs    []string `json:"pluginDirs"`
	AutoLoad      []string `json:"autoLoad"`
	Disabled      []string `json:"disabled"`
	Sandbox       bool     `json:"sandbox"`
}

// LoggingConfig contains logging-related settings
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	MaxAge     int    `json:"maxAge"`
	Compress   bool   `json:"compress"`
}

// MetricsConfig contains metrics-related settings
type MetricsConfig struct {
	Enabled         bool          `json:"enabled"`
	Endpoint        string        `json:"endpoint"`
	Interval        time.Duration `json:"interval"`
	RetentionTime   time.Duration `json:"retentionTime"`
	ExportFormat    string        `json:"exportFormat"`
	EnableProfiling bool          `json:"enableProfiling"`
}

// NewPlatform creates a new platform instance
func NewPlatform(config *PlatformConfig, logger core.Logger) (*Platform, error) {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Platform{
		ctx:        ctx,
		cancel:     cancel,
		plugins:    make(map[string]core.Plugin),
		pluginDeps: make(map[string][]string),
		version:    config.Version,
		buildInfo:  getBuildInfo(),
		logger:     logger,
	}

	// Initialize core managers (implementations would be in separate files)
	var err error

	if p.configManager, err = NewConfigManager(config); err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	if p.eventBus, err = NewEventBus(p.logger); err != nil {
		return nil, fmt.Errorf("failed to initialize event bus: %w", err)
	}

	if p.metrics, err = NewMetricsCollector(config.Metrics, p.logger); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics collector: %w", err)
	}

	if p.securityManager, err = NewSecurityManager(config.Security, p.logger); err != nil {
		return nil, fmt.Errorf("failed to initialize security manager: %w", err)
	}

	if p.networkManager, err = NewNetworkManager(config.Network, p.securityManager, p.eventBus, p.logger); err != nil {
		return nil, fmt.Errorf("failed to initialize network manager: %w", err)
	}

	if p.resourceManager, err = NewResourceManager(p.networkManager, p.securityManager, p.eventBus, p.logger); err != nil {
		return nil, fmt.Errorf("failed to initialize resource manager: %w", err)
	}

	if p.serviceManager, err = NewServiceManager(p.eventBus, p.logger); err != nil {
		return nil, fmt.Errorf("failed to initialize service manager: %w", err)
	}

	return p, nil
}

// Start initializes and starts the platform
func (p *Platform) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("platform already started")
	}

	p.logger.Info("Starting NoPlaceLike platform",
		core.Field{Key: "version", Value: p.version},
		core.Field{Key: "buildTime", Value: p.buildInfo.BuildTime},
	)

	// Start core services in order
	if err := p.serviceManager.StartAll(ctx); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Mark platform as started before plugin loading so preloaded and discovered plugins auto-start
	p.started = true
	p.startTime = time.Now()

	// Start any preloaded plugins
	for name, plugin := range p.plugins {
		if err := plugin.Start(ctx); err != nil {
			p.logger.Warn("Failed to start preloaded plugin",
				core.Field{Key: "plugin", Value: name},
				core.Field{Key: "error", Value: err},
			)
		}
	}

	// Load and start plugins from configured directories
	if err := p.loadPlugins(ctx); err != nil {
		p.logger.Warn("Failed to load some plugins", core.Field{Key: "error", Value: err})
	}

	// Start network discovery
	if _, err := p.networkManager.DiscoverPeers(); err != nil {
		p.logger.Warn("Failed to start peer discovery", core.Field{Key: "error", Value: err})
	}

	// Publish platform started event
	event := core.Event{
		ID:     generateID(),
		Type:   "platform.started",
		Source: "platform",
		Data: map[string]interface{}{
			"version":   p.buildInfo.Version,
			"commit":    p.buildInfo.Commit,
			"buildTime": p.buildInfo.BuildTime,
			"goVersion": p.buildInfo.GoVersion,
		},
		Timestamp: time.Now().Unix(),
	}

	if err := p.eventBus.Publish(event); err != nil {
		p.logger.Warn("Failed to publish platform started event", core.Field{Key: "error", Value: err})
	}

	p.logger.Info("NoPlaceLike platform started successfully")
	return nil
}

// Stop gracefully shuts down the platform
func (p *Platform) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return fmt.Errorf("platform not started")
	}

	p.logger.Info("Stopping NoPlaceLike platform")

	// Stop plugins first
	for name, plugin := range p.plugins {
		if err := plugin.Stop(ctx); err != nil {
			p.logger.Warn("Failed to stop plugin",
				core.Field{Key: "plugin", Value: name},
				core.Field{Key: "error", Value: err},
			)
		}
	}

	// Stop core services
	if err := p.serviceManager.StopAll(ctx); err != nil {
		p.logger.Warn("Failed to stop all services", core.Field{Key: "error", Value: err})
	}

	p.started = false
	p.cancel()

	p.logger.Info("NoPlaceLike platform stopped")
	return nil
}

// LoadPlugin loads a plugin into the platform
func (p *Platform) LoadPlugin(ctx context.Context, plugin core.Plugin) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	name := plugin.Name()

	if _, exists := p.plugins[name]; exists {
		return fmt.Errorf("plugin %s already loaded", name)
	}

	// Check dependencies
	deps := plugin.Dependencies()
	for _, dep := range deps {
		if _, exists := p.plugins[dep]; !exists {
			return fmt.Errorf("plugin %s depends on %s which is not loaded", name, dep)
		}
	}

	// Initialize plugin
	if err := plugin.Initialize(p); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}

	// Start plugin if platform is running
	if p.started {
		if err := plugin.Start(ctx); err != nil {
			return fmt.Errorf("failed to start plugin %s: %w", name, err)
		}
	}

	p.plugins[name] = plugin
	p.pluginDeps[name] = deps

	p.logger.Info("Plugin loaded successfully",
		core.Field{Key: "plugin", Value: name},
		core.Field{Key: "version", Value: plugin.Version()},
	)

	// Publish plugin loaded event
	event := core.Event{
		ID:        generateID(),
		Type:      "plugin.loaded",
		Source:    "platform",
		Data:      map[string]interface{}{"name": name, "version": plugin.Version()},
		Timestamp: time.Now().Unix(),
	}

	if err := p.eventBus.Publish(event); err != nil {
		p.logger.Warn("Failed to publish plugin loaded event", core.Field{Key: "error", Value: err})
	}

	return nil
}

// UnloadPlugin removes a plugin from the platform
func (p *Platform) UnloadPlugin(ctx context.Context, name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	plugin, exists := p.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Check if other plugins depend on this one
	for pluginName, deps := range p.pluginDeps {
		for _, dep := range deps {
			if dep == name {
				return fmt.Errorf("cannot unload plugin %s: plugin %s depends on it", name, pluginName)
			}
		}
	}

	// Stop plugin
	if err := plugin.Stop(ctx); err != nil {
		p.logger.Warn("Failed to stop plugin",
			core.Field{Key: "plugin", Value: name},
			core.Field{Key: "error", Value: err},
		)
	}

	delete(p.plugins, name)
	delete(p.pluginDeps, name)

	p.logger.Info("Plugin unloaded", core.Field{Key: "plugin", Value: name})

	// Publish plugin unloaded event
	event := core.Event{
		ID:        generateID(),
		Type:      "plugin.unloaded",
		Source:    "platform",
		Data:      map[string]interface{}{"name": name},
		Timestamp: time.Now().Unix(),
	}

	if err := p.eventBus.Publish(event); err != nil {
		p.logger.Warn("Failed to publish plugin unloaded event", core.Field{Key: "error", Value: err})
	}

	return nil
}

// GetPlugin retrieves a loaded plugin by name
func (p *Platform) GetPlugin(name string) (core.Plugin, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	plugin, exists := p.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// ListPlugins returns all loaded plugins
func (p *Platform) ListPlugins() map[string]core.Plugin {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]core.Plugin)
	for name, plugin := range p.plugins {
		result[name] = plugin
	}

	return result
}

// Health returns the overall platform health
func (p *Platform) Health() core.HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return core.HealthStatus{
			Status:    core.HealthStatusUnhealthy,
			Timestamp: time.Now(),
			Error:     "platform not started",
		}
	}

	// Check service health
	serviceHealth := p.serviceManager.HealthCheck()
	unhealthyServices := 0

	for _, health := range serviceHealth {
		if health.Status != core.HealthStatusHealthy {
			unhealthyServices++
		}
	}

	// Check plugin health
	unhealthyPlugins := 0
	for _, plugin := range p.plugins {
		health := plugin.Health()
		if health.Status != core.HealthStatusHealthy {
			unhealthyPlugins++
		}
	}

	status := core.HealthStatusHealthy
	if unhealthyServices > 0 || unhealthyPlugins > 0 {
		if unhealthyServices > len(serviceHealth)/2 || unhealthyPlugins > len(p.plugins)/2 {
			status = core.HealthStatusUnhealthy
		} else {
			status = core.HealthStatusDegraded
		}
	}

	return core.HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"uptime":            time.Since(p.startTime).String(),
			"servicesTotal":     len(serviceHealth),
			"servicesUnhealthy": unhealthyServices,
			"pluginsTotal":      len(p.plugins),
			"pluginsUnhealthy":  unhealthyPlugins,
			"version":           p.version,
		},
	}
}

// Managers provide access to core platform managers
func (p *Platform) ServiceManager() core.ServiceManager   { return p.serviceManager }
func (p *Platform) NetworkManager() core.NetworkManager   { return p.networkManager }
func (p *Platform) ResourceManager() core.ResourceManager { return p.resourceManager }
func (p *Platform) SecurityManager() core.SecurityManager { return p.securityManager }
func (p *Platform) ConfigManager() core.ConfigManager     { return p.configManager }
func (p *Platform) EventBus() core.EventBus               { return p.eventBus }
func (p *Platform) Metrics() core.MetricsCollector        { return p.metrics }
func (p *Platform) Logger() core.Logger                   { return p.logger }

// Implement core.PlatformAPI interface
func (p *Platform) GetEventBus() core.EventBus {
	return p.eventBus
}

func (p *Platform) GetLogger() core.Logger {
	return p.logger
}

func (p *Platform) GetConfigManager() core.ConfigManager {
	return p.configManager
}

func (p *Platform) GetMetrics() core.MetricsCollector {
	return p.metrics
}

func (p *Platform) GetNetworkManager() core.NetworkManager {
	return p.networkManager
}

func (p *Platform) GetResourceManager() core.ResourceManager {
	return p.resourceManager
}

func (p *Platform) GetSecurityManager() core.SecurityManager {
	return p.securityManager
}

func (p *Platform) GetHealthChecker() core.HealthChecker {
	return nil // TODO: implement if you have a health checker in your platform
}

// loadPlugins loads plugins from configured directories
func (p *Platform) loadPlugins(ctx context.Context) error {
	// Plugin loading implementation would go here
	// This would scan plugin directories, load plugin files, and register them
	return nil
}

// generateID generates a unique identifier
func generateID() string {
	// Implementation would generate a UUID or similar unique ID
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

// getBuildInfo returns build information
func getBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   "1.0.0", // Would be set at build time
		Commit:    "dev",   // Would be set at build time
		BuildTime: time.Now(),
		GoVersion: "go1.22.2",
	}
}

// Placeholder functions for manager creation (these would be implemented in separate files)
func NewLogger(config LoggingConfig) (core.Logger, error) { return nil, fmt.Errorf("not implemented") }

// Minimal stub config manager

type stubConfigManager struct{}

// Implement core.ConfigManager methods as no-ops or defaults
// Add more methods if your interface requires them

func (s *stubConfigManager) Reload() error                     { return nil }
func (s *stubConfigManager) Save() error                       { return nil }
func (s *stubConfigManager) Get(key string) interface{}        { return nil }
func (s *stubConfigManager) Set(key string, value interface{}) {}

func NewConfigManager(config *PlatformConfig) (core.ConfigManager, error) {
	return &stubConfigManager{}, nil
}

// --- Implementations for core managers and services ---

// EventBus implementation
type eventBusImpl struct {
	mu      sync.RWMutex
	subs    map[string][]func(context.Context, core.Event) error
	started bool
	logger  core.Logger
}

func (e *eventBusImpl) Name() string { return "event-bus" }

func (e *eventBusImpl) Start(ctx context.Context) error {
	e.mu.Lock()
	e.started = true
	if e.subs == nil {
		e.subs = make(map[string][]func(context.Context, core.Event) error)
	}
	e.mu.Unlock()
	return nil
}

func (e *eventBusImpl) Stop(ctx context.Context) error {
	e.mu.Lock()
	e.started = false
	e.mu.Unlock()
	return nil
}

func (e *eventBusImpl) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.started
}

func (e *eventBusImpl) Health() core.HealthStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	status := core.HealthStatusHealthy
	if !e.started {
		status = core.HealthStatusUnhealthy
	}
	return core.HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
	}
}

func (e *eventBusImpl) Configuration() core.ConfigSchema {
	return core.ConfigSchema{Properties: map[string]core.PropertySchema{}}
}

func (e *eventBusImpl) Publish(event core.Event) error {
	e.mu.RLock()
	handlers := append([]func(context.Context, core.Event) error{}, e.subs[event.Type]...)
	starHandlers := append([]func(context.Context, core.Event) error{}, e.subs["*"]...)
	e.mu.RUnlock()

	for _, h := range handlers {
		_ = h(context.Background(), event)
	}
	for _, h := range starHandlers {
		_ = h(context.Background(), event)
	}
	return nil
}

func (e *eventBusImpl) PublishToTopic(ctx context.Context, topic string, event core.Event) error {
	// Treat topic as event type channel
	e.mu.RLock()
	handlers := append([]func(context.Context, core.Event) error{}, e.subs[topic]...)
	starHandlers := append([]func(context.Context, core.Event) error{}, e.subs["*"]...)
	e.mu.RUnlock()

	for _, h := range handlers {
		_ = h(ctx, event)
	}
	for _, h := range starHandlers {
		_ = h(ctx, event)
	}
	return nil
}

func (e *eventBusImpl) Subscribe(eventType string, handler core.EventHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.subs == nil {
		e.subs = make(map[string][]func(context.Context, core.Event) error)
	}
	wrapped := func(ctx context.Context, ev core.Event) error { return handler(ev) }
	e.subs[eventType] = append(e.subs[eventType], wrapped)
	return nil
}

func (e *eventBusImpl) SubscribeWithContext(ctx context.Context, eventType string, handler func(context.Context, core.Event) error) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.subs == nil {
		e.subs = make(map[string][]func(context.Context, core.Event) error)
	}
	e.subs[eventType] = append(e.subs[eventType], handler)
	return nil
}

func (e *eventBusImpl) Unsubscribe(eventType string, handler core.EventHandler) error {
	// Minimal implementation: clear all subscribers for the eventType
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.subs, eventType)
	return nil
}

// Metrics implementation
type counterImpl struct {
	mu    sync.RWMutex
	value float64
}

func (c *counterImpl) Inc()               { c.Add(1) }
func (c *counterImpl) Add(delta float64)  { c.mu.Lock(); c.value += delta; c.mu.Unlock() }
func (c *counterImpl) Get() float64       { c.mu.RLock(); defer c.mu.RUnlock(); return c.value }

type gaugeImpl struct {
	mu    sync.RWMutex
	value float64
}

func (g *gaugeImpl) Set(v float64)        { g.mu.Lock(); g.value = v; g.mu.Unlock() }
func (g *gaugeImpl) Inc()                 { g.Add(1) }
func (g *gaugeImpl) Dec()                 { g.Add(-1) }
func (g *gaugeImpl) Add(delta float64)    { g.mu.Lock(); g.value += delta; g.mu.Unlock() }
func (g *gaugeImpl) Sub(delta float64)    { g.Add(-delta) }
func (g *gaugeImpl) Get() float64         { g.mu.RLock(); defer g.mu.RUnlock(); return g.value }

type histogramImpl struct {
	mu      sync.RWMutex
	values  []float64
}

func (h *histogramImpl) Observe(v float64) { h.mu.Lock(); h.values = append(h.values, v); h.mu.Unlock() }
func (h *histogramImpl) Reset()            { h.mu.Lock(); h.values = nil; h.mu.Unlock() }

type timerInstanceImpl struct {
	start time.Time
	rec   func(duration time.Duration)
}

func (t *timerInstanceImpl) Stop() {
	if t.rec != nil {
		t.rec(time.Since(t.start))
	}
}

type timerImpl struct {
	h *histogramImpl
}

func (t *timerImpl) Start() core.TimerInstance {
	return &timerInstanceImpl{
		start: time.Now(),
		rec: func(d time.Duration) {
			if t.h != nil {
				t.h.Observe(float64(d) / float64(time.Millisecond))
			}
		},
	}
}
func (t *timerImpl) Observe(duration float64) {
	if t.h != nil {
		t.h.Observe(duration)
	}
}

type metricsCollectorImpl struct {
	mu         sync.RWMutex
	started    bool
	logger     core.Logger
	counters   map[string]*counterImpl
	gauges     map[string]*gaugeImpl
	histograms map[string]*histogramImpl
	timers     map[string]*timerImpl
}

func (m *metricsCollectorImpl) Name() string { return "metrics" }
func (m *metricsCollectorImpl) Start(ctx context.Context) error {
	m.mu.Lock()
	m.started = true
	if m.counters == nil {
		m.counters = map[string]*counterImpl{}
	}
	if m.gauges == nil {
		m.gauges = map[string]*gaugeImpl{}
	}
	if m.histograms == nil {
		m.histograms = map[string]*histogramImpl{}
	}
	if m.timers == nil {
		m.timers = map[string]*timerImpl{}
	}
	m.mu.Unlock()
	return nil
}
func (m *metricsCollectorImpl) Stop(ctx context.Context) error {
	m.mu.Lock()
	m.started = false
	m.mu.Unlock()
	return nil
}
func (m *metricsCollectorImpl) IsHealthy() bool {
	m.mu.RLock(); defer m.mu.RUnlock()
	return m.started
}
func (m *metricsCollectorImpl) Health() core.HealthStatus {
	m.mu.RLock(); defer m.mu.RUnlock()
	status := core.HealthStatusHealthy
	if !m.started { status = core.HealthStatusUnhealthy }
	return core.HealthStatus{ Status: status, Timestamp: time.Now() }
}
func (m *metricsCollectorImpl) Configuration() core.ConfigSchema {
	return core.ConfigSchema{Properties: map[string]core.PropertySchema{}}
}
func (m *metricsCollectorImpl) Counter(name string) core.Counter {
	m.mu.Lock(); defer m.mu.Unlock()
	if c, ok := m.counters[name]; ok { return c }
	c := &counterImpl{}
	m.counters[name] = c
	return c
}
func (m *metricsCollectorImpl) Gauge(name string) core.Gauge {
	m.mu.Lock(); defer m.mu.Unlock()
	if g, ok := m.gauges[name]; ok { return g }
	g := &gaugeImpl{}
	m.gauges[name] = g
	return g
}
func (m *metricsCollectorImpl) Histogram(name string) core.Histogram {
	m.mu.Lock(); defer m.mu.Unlock()
	if h, ok := m.histograms[name]; ok { return h }
	h := &histogramImpl{}
	m.histograms[name] = h
	return h
}
func (m *metricsCollectorImpl) Timer(name string) core.Timer {
	m.mu.Lock(); defer m.mu.Unlock()
	if t, ok := m.timers[name]; ok { return t }
	h := &histogramImpl{}
	t := &timerImpl{h: h}
	m.histograms[name+"_duration_ms"] = h
	m.timers[name] = t
	return t
}
func (m *metricsCollectorImpl) Export(format string) ([]byte, error) {
	// Minimal text/JSON-like export without extra imports
	m.mu.RLock()
	defer m.mu.RUnlock()

	if format == "json" {
		// Build a simple JSON string
		s := "{"
		// counters
		s += "\"counters\":{"
		first := true
		for k, v := range m.counters {
			if !first { s += "," } ; first = false
			s += fmt.Sprintf("%q:%v", k, v.Get())
		}
		s += "},"
		// gauges
		s += "\"gauges\":{"
		first = true
		for k, v := range m.gauges {
			if !first { s += "," } ; first = false
			s += fmt.Sprintf("%q:%v", k, v.Get())
		}
		s += "},"
		// histograms (export count only)
		s += "\"histograms\":{"
		first = true
		for k, v := range m.histograms {
			if !first { s += "," } ; first = false
			count := 0
			if v.values != nil { count = len(v.values) }
			s += fmt.Sprintf("%q:{\"count\":%d}", k, count)
		}
		s += "}"
		s += "}"
		return []byte(s), nil
	}

	// Plain text
	out := "metrics:\n"
	out += " counters:\n"
	for k, v := range m.counters {
		out += fmt.Sprintf("  - %s=%v\n", k, v.Get())
	}
	out += " gauges:\n"
	for k, v := range m.gauges {
		out += fmt.Sprintf("  - %s=%v\n", k, v.Get())
	}
	out += " histograms:\n"
	for k, v := range m.histograms {
		count := 0
		if v.values != nil { count = len(v.values) }
		out += fmt.Sprintf("  - %s count=%d\n", k, count)
	}
	return []byte(out), nil
}

// Security manager implementation
type securityManagerImpl struct {
	mu          sync.RWMutex
	started     bool
	logger      core.Logger
	tokenExpiry time.Duration
	secret      []byte
	issuer      string
	audience    []string
}

func (s *securityManagerImpl) Name() string { return "security" }
func (s *securityManagerImpl) Start(ctx context.Context) error { s.mu.Lock(); s.started = true; s.mu.Unlock(); return nil }
func (s *securityManagerImpl) Stop(ctx context.Context) error  { s.mu.Lock(); s.started = false; s.mu.Unlock(); return nil }
func (s *securityManagerImpl) IsHealthy() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.started }
func (s *securityManagerImpl) Health() core.HealthStatus {
	s.mu.RLock(); defer s.mu.RUnlock()
	status := core.HealthStatusHealthy
	if !s.started { status = core.HealthStatusUnhealthy }
	return core.HealthStatus{ Status: status, Timestamp: time.Now() }
}
func (s *securityManagerImpl) Configuration() core.ConfigSchema {
	return core.ConfigSchema{Properties: map[string]core.PropertySchema{}}
}

func (s *securityManagerImpl) Authenticate(token string) (*core.User, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}
	return &core.User{ID: token, Username: token, CreatedAt: time.Now().Unix()}, nil
}

func (s *securityManagerImpl) Authorize(user *core.User, resource string, action string) bool {
	// Minimal implementation: allow all
	_ = user
	_ = resource
	_ = action
	return true
}

func (s *securityManagerImpl) GenerateToken(user *core.User) (string, error) {
	if user == nil || user.ID == "" {
		return "", fmt.Errorf("invalid user")
	}
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	now := time.Now()
	exp := now.Add(s.tokenExpiry)
	claims := map[string]interface{}{
		"sub": user.ID,
		"iat": now.Unix(),
		"exp": exp.Unix(),
	}
	if s.issuer != "" {
		claims["iss"] = s.issuer
	}
	if len(s.audience) > 0 {
		if len(s.audience) == 1 {
			claims["aud"] = s.audience[0]
		} else {
			claims["aud"] = s.audience
		}
	}

	hb, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	cb, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	enc := base64.RawURLEncoding
	h64 := enc.EncodeToString(hb)
	c64 := enc.EncodeToString(cb)
	signingInput := h64 + "." + c64

	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)
	s64 := enc.EncodeToString(sig)

	return signingInput + "." + s64, nil
}

func (s *securityManagerImpl) ValidatePermissions(userID string, permissions []string) bool {
	_ = userID
	_ = permissions
	return true
}

func (s *securityManagerImpl) ValidateToken(ctx context.Context, token string) (*core.TokenInfo, error) {
	if token == "" {
		return &core.TokenInfo{Valid: false}, nil
	}
	// Expect "header.payload.signature"
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return &core.TokenInfo{Valid: false}, nil
	}

	enc := base64.RawURLEncoding
	headerJSON, err := enc.DecodeString(parts[0])
	if err != nil {
		return &core.TokenInfo{Valid: false}, nil
	}
	var header map[string]interface{}
	_ = json.Unmarshal(headerJSON, &header)
	if alg, _ := header["alg"].(string); alg != "HS256" {
		return &core.TokenInfo{Valid: false}, nil
	}

	payloadJSON, err := enc.DecodeString(parts[1])
	if err != nil {
		return &core.TokenInfo{Valid: false}, nil
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(signingInput))
	expected := mac.Sum(nil)
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		return &core.TokenInfo{Valid: false}, nil
	}
	if !hmac.Equal(sig, expected) {
		return &core.TokenInfo{Valid: false}, nil
	}

	// Parse claims
	var claims map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return &core.TokenInfo{Valid: false}, nil
	}

	now := time.Now().Unix()
	// exp
	if v, ok := claims["exp"]; ok {
		switch t := v.(type) {
		case float64:
			if int64(t) < now {
				return &core.TokenInfo{Valid: false}, nil
			}
		case int64:
			if t < now {
				return &core.TokenInfo{Valid: false}, nil
			}
		}
	}
	// nbf
	if v, ok := claims["nbf"]; ok {
		switch t := v.(type) {
		case float64:
			if int64(t) > now {
				return &core.TokenInfo{Valid: false}, nil
			}
		case int64:
			if t > now {
				return &core.TokenInfo{Valid: false}, nil
			}
		}
	}
	// iss
	if s.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != s.issuer {
			return &core.TokenInfo{Valid: false}, nil
		}
	}
	// aud
	if len(s.audience) > 0 {
		okAud := false
		if audStr, ok := claims["aud"].(string); ok {
			for _, a := range s.audience {
				if a == audStr {
					okAud = true
					break
				}
			}
		} else if audArr, ok := claims["aud"].([]interface{}); ok {
			for _, ai := range audArr {
				if as, ok := ai.(string); ok {
					for _, a := range s.audience {
						if a == as {
							okAud = true
							break
						}
					}
				}
				if okAud {
					break
				}
			}
		} else {
			// missing aud but required
			return &core.TokenInfo{Valid: false}, nil
		}
		if !okAud {
			return &core.TokenInfo{Valid: false}, nil
		}
	}

	userID := ""
	if sub, _ := claims["sub"].(string); sub != "" {
		userID = sub
	}

	expireAt := int64(0)
	if v, ok := claims["exp"]; ok {
		switch t := v.(type) {
		case float64:
			expireAt = int64(t)
		case int64:
			expireAt = t
		}
	}

	return &core.TokenInfo{
		Valid:       true,
		UserID:      userID,
		PeerID:      userID,
		Permissions: []string{},
		ExpireAt:    expireAt,
	}, nil
}

// Network manager implementation
type networkManagerImpl struct {
	mu      sync.RWMutex
	started bool
	logger  core.Logger
	peers   map[string]core.Peer
}

func (n *networkManagerImpl) Name() string { return "network" }
func (n *networkManagerImpl) Start(ctx context.Context) error { n.mu.Lock(); n.started = true; if n.peers == nil { n.peers = map[string]core.Peer{} }; n.mu.Unlock(); return nil }
func (n *networkManagerImpl) Stop(ctx context.Context) error  { n.mu.Lock(); n.started = false; n.mu.Unlock(); return nil }
func (n *networkManagerImpl) IsHealthy() bool { n.mu.RLock(); defer n.mu.RUnlock(); return n.started }
func (n *networkManagerImpl) Health() core.HealthStatus {
	n.mu.RLock(); defer n.mu.RUnlock()
	status := core.HealthStatusHealthy
	if !n.started { status = core.HealthStatusUnhealthy }
	return core.HealthStatus{ Status: status, Timestamp: time.Now() }
}
func (n *networkManagerImpl) Configuration() core.ConfigSchema {
	return core.ConfigSchema{Properties: map[string]core.PropertySchema{}}
}

func (n *networkManagerImpl) DiscoverPeers() ([]core.Peer, error) {
	return n.GetPeers(), nil
}
func (n *networkManagerImpl) GetPeers() []core.Peer {
	n.mu.RLock(); defer n.mu.RUnlock()
	out := make([]core.Peer, 0, len(n.peers))
	for _, p := range n.peers {
		out = append(out, p)
	}
	return out
}
func (n *networkManagerImpl) ConnectToPeer(address string) (core.Peer, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.peers == nil {
		n.peers = map[string]core.Peer{}
	}
	id := fmt.Sprintf("peer-%d", time.Now().UnixNano())
	p := core.Peer{
		ID:       id,
		Address:  address,
		Name:     address,
		Status:   "connected",
		Metadata: map[string]interface{}{},
		LastSeen: time.Now().Unix(),
	}
	n.peers[id] = p
	return p, nil
}
func (n *networkManagerImpl) ListPeers() []core.Peer { return n.GetPeers() }
func (n *networkManagerImpl) SendMessage(peerID string, message []byte) error { _ = peerID; _ = message; return nil }
func (n *networkManagerImpl) BroadcastMessage(message []byte) error { _ = message; return nil }

// Resource manager implementation
type resourceManagerImpl struct {
	mu        sync.RWMutex
	started   bool
	logger    core.Logger
	eventBus  core.EventBus
	resources map[string]core.Resource
}

func (r *resourceManagerImpl) Name() string { return "resources" }
func (r *resourceManagerImpl) Start(ctx context.Context) error { r.mu.Lock(); r.started = true; if r.resources == nil { r.resources = map[string]core.Resource{} }; r.mu.Unlock(); return nil }
func (r *resourceManagerImpl) Stop(ctx context.Context) error  { r.mu.Lock(); r.started = false; r.mu.Unlock(); return nil }
func (r *resourceManagerImpl) IsHealthy() bool { r.mu.RLock(); defer r.mu.RUnlock(); return r.started }
func (r *resourceManagerImpl) Health() core.HealthStatus {
	r.mu.RLock(); defer r.mu.RUnlock()
	status := core.HealthStatusHealthy
	if !r.started { status = core.HealthStatusUnhealthy }
	return core.HealthStatus{ Status: status, Timestamp: time.Now() }
}
func (r *resourceManagerImpl) Configuration() core.ConfigSchema {
	return core.ConfigSchema{Properties: map[string]core.PropertySchema{}}
}

func (r *resourceManagerImpl) RegisterResource(resource core.Resource) error {
	if resource == nil || resource.ID() == "" {
		return fmt.Errorf("invalid resource")
	}
	r.mu.Lock()
	r.resources[resource.ID()] = resource
	r.mu.Unlock()
	return nil
}

func (r *resourceManagerImpl) UnregisterResource(id string) error {
	r.mu.Lock()
	delete(r.resources, id)
	r.mu.Unlock()
	return nil
}

func (r *resourceManagerImpl) GetResource(ctx context.Context, id string) (core.Resource, error) {
	r.mu.RLock()
	res, ok := r.resources[id]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("resource not found")
	}
	return res, nil
}

func (r *resourceManagerImpl) ListResources(ctx context.Context, filter core.ResourceFilter) ([]core.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]core.Resource, 0, len(r.resources))
	for _, res := range r.resources {
		if filter.Type != "" && res.Type() != filter.Type {
			continue
		}
		if filter.Name != "" {
			if name, ok := res.GetMetadata()["name"].(string); ok {
				if name != filter.Name {
					continue
				}
			}
		}
		out = append(out, res)
	}
	return out, nil
}

type memoryResourceStream struct {
	sent bool
}

func (m *memoryResourceStream) Read() ([]byte, error) {
	if m.sent {
		return nil, fmt.Errorf("eof")
	}
	m.sent = true
	return []byte("stream not available for this resource"), nil
}

func (m *memoryResourceStream) Close() error { return nil }

func (r *resourceManagerImpl) StreamResource(ctx context.Context, id string) (core.ResourceStream, error) {
	// Minimal streaming: return a single-chunk stream
	if _, err := r.GetResource(ctx, id); err != nil {
		return nil, err
	}
	return &memoryResourceStream{}, nil
}

// Service manager implementation
type serviceManagerImpl struct {
	mu       sync.RWMutex
	services map[string]core.Service
}

func (s *serviceManagerImpl) StartAll(ctx context.Context) error {
	s.mu.RLock()
	services := make([]core.Service, 0, len(s.services))
	for _, svc := range s.services {
		services = append(services, svc)
	}
	s.mu.RUnlock()
	for _, svc := range services {
		if err := svc.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *serviceManagerImpl) StopAll(ctx context.Context) error {
	s.mu.RLock()
	services := make([]core.Service, 0, len(s.services))
	for _, svc := range s.services {
		services = append(services, svc)
	}
	s.mu.RUnlock()
	for _, svc := range services {
		if err := svc.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *serviceManagerImpl) HealthCheck() map[string]core.HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[string]core.HealthStatus{}
	for name, svc := range s.services {
		out[name] = svc.Health()
	}
	return out
}

func (s *serviceManagerImpl) GetService(name string) (core.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if svc, ok := s.services[name]; ok {
		return svc, nil
	}
	return nil, fmt.Errorf("service %s not found", name)
}

func (s *serviceManagerImpl) Configuration() core.ConfigSchema {
	return core.ConfigSchema{Properties: map[string]core.PropertySchema{}}
}

func (s *serviceManagerImpl) RegisterService(service core.Service) error {
	if service == nil || service.Name() == "" {
		return fmt.Errorf("invalid service")
	}
	s.mu.Lock()
	if s.services == nil {
		s.services = map[string]core.Service{}
	}
	s.services[service.Name()] = service
	s.mu.Unlock()
	return nil
}

func NewEventBus(logger core.Logger) (core.EventBus, error) {
	return &eventBusImpl{
		logger: logger,
		subs:   map[string][]func(context.Context, core.Event) error{},
	}, nil
}
func NewMetricsCollector(config MetricsConfig, logger core.Logger) (core.MetricsCollector, error) {
	return &metricsCollectorImpl{
		logger:     logger,
		counters:   map[string]*counterImpl{},
		gauges:     map[string]*gaugeImpl{},
		histograms: map[string]*histogramImpl{},
		timers:     map[string]*timerImpl{},
	}, nil
}
func NewSecurityManager(config SecurityConfig, logger core.Logger) (core.SecurityManager, error) {
	sm := &securityManagerImpl{
		logger:      logger,
		tokenExpiry: config.TokenExpiry,
		secret:      []byte(config.JWTSecret),
		issuer:      config.JWTIssuer,
		audience:    config.JWTAudience,
	}
	return sm, nil
}
func NewNetworkManager(config NetworkConfig, security core.SecurityManager, eventBus core.EventBus, logger core.Logger) (core.NetworkManager, error) {
	return &networkManagerImpl{
		logger: logger,
		peers:  map[string]core.Peer{},
	}, nil
}
func NewResourceManager(network core.NetworkManager, security core.SecurityManager, eventBus core.EventBus, logger core.Logger) (core.ResourceManager, error) {
	return &resourceManagerImpl{
		logger:    logger,
		eventBus:  eventBus,
		resources: map[string]core.Resource{},
	}, nil
}
func NewServiceManager(eventBus core.EventBus, logger core.Logger) (core.ServiceManager, error) {
	return &serviceManagerImpl{
		services: map[string]core.Service{},
	}, nil
}

// func NewEventBus(logger core.Logger) (core.EventBus, error) {
// 	// return nothing for now haha
// 	return nil, fmt.Errorf("not implemented yet lol")
// }
