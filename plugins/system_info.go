package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/nathfavour/noplacelike.go/internal/core"
	"github.com/nathfavour/noplacelike.go/internal/logger"
)

// SystemInfoPlugin provides system information and monitoring
type SystemInfoPlugin struct {
	id       string
	version  string
	logger   logger.Logger
	platform core.PlatformAPI
	running  bool
}

// NewSystemInfoPlugin creates a new system info plugin
func NewSystemInfoPlugin() core.Plugin {
	return &SystemInfoPlugin{
		id:      "system-info",
		version: "1.0.0",
	}
}

// Plugin interface implementation
func (p *SystemInfoPlugin) ID() string {
	return p.id
}

func (p *SystemInfoPlugin) Version() string {
	return p.version
}

func (p *SystemInfoPlugin) Dependencies() []string {
	return []string{} // No dependencies
}

func (p *SystemInfoPlugin) Name() string {
	return "System Info Plugin"
}

func (p *SystemInfoPlugin) Initialize(platform core.PlatformAPI) error {
	p.platform = platform
	p.logger = platform.GetLogger().WithFields(map[string]interface{}{
		"plugin": p.id,
	})

	p.logger.Info("System info plugin initialized")
	return nil
}

func (p *SystemInfoPlugin) Configure(config map[string]interface{}) error {
	// Plugin-specific configuration can be handled here
	p.logger.Info("System info plugin configured")
	return nil
}

func (p *SystemInfoPlugin) Start(ctx context.Context) error {
	p.running = true
	p.logger.Info("System info plugin started")

	// Register health check
	if healthChecker := p.platform.GetHealthChecker(); healthChecker != nil {
		healthChecker.RegisterCheck("system-info", func() error {
			if !p.running {
				return fmt.Errorf("system info plugin is not running")
			}
			return nil
		})
	}

	return nil
}

func (p *SystemInfoPlugin) Stop(ctx context.Context) error {
	p.running = false
	p.logger.Info("System info plugin stopped")
	return nil
}

func (p *SystemInfoPlugin) IsHealthy() bool {
	return p.running
}

func (p *SystemInfoPlugin) Routes() []core.Route {
	return []core.Route{
		{
			Method:      "GET",
			Path:        "/plugins/system-info/system/info",
			Handler:     p.handleSystemInfo,
			Description: "Get detailed system information",
		},
		{
			Method:      "GET",
			Path:        "/plugins/system-info/system/health",
			Handler:     p.handleSystemHealth,
			Description: "Get system health metrics",
		},
		{
			Method:      "GET",
			Path:        "/plugins/system-info/runtime/info",
			Handler:     p.handleRuntimeInfo,
			Description: "Get Go runtime information",
		},
	}
}

func (p *SystemInfoPlugin) HandleEvent(event core.Event) error {
	// Handle platform events if needed
	p.logger.Debug("Received event", "type", event.Type, "source", event.Source)
	return nil
}

// HTTP handlers
func (p *SystemInfoPlugin) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := p.getSystemInfo()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		p.logger.Error("Error encoding system info", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (p *SystemInfoPlugin) handleSystemHealth(w http.ResponseWriter, r *http.Request) {
	health := p.getSystemHealth()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health); err != nil {
		p.logger.Error("Error encoding system health", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (p *SystemInfoPlugin) handleRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	info := p.getRuntimeInfo()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		p.logger.Error("Error encoding runtime info", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Data collection methods
func (p *SystemInfoPlugin) getSystemInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()

	return map[string]interface{}{
		"hostname":              hostname,
		"platform":              runtime.GOOS,
		"architecture":          runtime.GOARCH,
		"working_directory":     wd,
		"environment_variables": len(os.Environ()),
		"timestamp":             time.Now().Unix(),
	}
}

func (p *SystemInfoPlugin) getSystemHealth() map[string]interface{} {
	// Basic health metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"status": "healthy",
		"uptime": time.Since(time.Now()).Seconds(), // This would be actual uptime in real implementation
		"memory": map[string]interface{}{
			"allocated":       memStats.Alloc,
			"total_allocated": memStats.TotalAlloc,
			"system":          memStats.Sys,
			"gc_runs":         memStats.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"timestamp":  time.Now().Unix(),
	}
}

func (p *SystemInfoPlugin) getRuntimeInfo() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"go_version":    runtime.Version(),
		"compiler":      runtime.Compiler,
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"memory_stats": map[string]interface{}{
			"alloc":           memStats.Alloc,
			"total_alloc":     memStats.TotalAlloc,
			"sys":             memStats.Sys,
			"lookups":         memStats.Lookups,
			"mallocs":         memStats.Mallocs,
			"frees":           memStats.Frees,
			"heap_alloc":      memStats.HeapAlloc,
			"heap_sys":        memStats.HeapSys,
			"heap_idle":       memStats.HeapIdle,
			"heap_inuse":      memStats.HeapInuse,
			"heap_released":   memStats.HeapReleased,
			"heap_objects":    memStats.HeapObjects,
			"stack_inuse":     memStats.StackInuse,
			"stack_sys":       memStats.StackSys,
			"next_gc":         memStats.NextGC,
			"last_gc":         memStats.LastGC,
			"gc_cpu_fraction": memStats.GCCPUFraction,
			"num_gc":          memStats.NumGC,
		},
		"timestamp": time.Now().Unix(),
	}
}
