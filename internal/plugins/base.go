// Package plugins provides base plugin implementations and utilities
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nathfavour/noplacelike.go/internal/core"
)

// BasePlugin provides common plugin functionality
type BasePlugin struct {
	mu           sync.RWMutex
	name         string
	version      string
	dependencies []string
	config       map[string]interface{}
	routes       []core.Route
	logger       core.Logger
	started      bool
	health       core.HealthStatus
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name, version string, dependencies []string) *BasePlugin {
	return &BasePlugin{
		name:         name,
		version:      version,
		dependencies: dependencies,
		config:       make(map[string]interface{}),
		routes:       make([]core.Route, 0),
		health: core.HealthStatus{
			Status:    core.HealthStatusHealthy,
			Timestamp: time.Now(),
		},
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *BasePlugin) Version() string {
	return p.version
}

// Dependencies returns plugin dependencies
func (p *BasePlugin) Dependencies() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	deps := make([]string, len(p.dependencies))
	copy(deps, p.dependencies)
	return deps
}

// Initialize sets up the plugin
func (p *BasePlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if config != nil {
		p.config = config
	}

	return nil
}

// Start begins plugin execution
func (p *BasePlugin) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("plugin %s already started", p.name)
	}

	p.started = true
	p.updateHealth(core.HealthStatusHealthy, "started")

	return nil
}

// Stop gracefully shuts down the plugin
func (p *BasePlugin) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return fmt.Errorf("plugin %s not started", p.name)
	}

	p.started = false
	p.updateHealth(core.HealthStatusUnhealthy, "stopped")

	return nil
}

// Health returns the current health status
func (p *BasePlugin) Health() core.HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.health
}

// Routes returns HTTP routes this plugin provides
func (p *BasePlugin) Routes() []core.Route {
	p.mu.RLock()
	defer p.mu.RUnlock()

	routes := make([]core.Route, len(p.routes))
	copy(routes, p.routes)
	return routes
}

// Protected methods for subclasses
func (p *BasePlugin) AddRoute(route core.Route) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.routes = append(p.routes, route)
}

func (p *BasePlugin) SetLogger(logger core.Logger) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger = logger
}

func (p *BasePlugin) updateHealth(status, message string) {
	p.health = core.HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"message": message,
			"started": p.started,
		},
	}
}

// FileManagerPlugin provides file management capabilities
type FileManagerPlugin struct {
	*BasePlugin
	uploadDir   string
	downloadDir string
	maxFileSize int64
}

// NewFileManagerPlugin creates a new file manager plugin
func NewFileManagerPlugin(uploadDir, downloadDir string, maxFileSize int64) *FileManagerPlugin {
	base := NewBasePlugin("file-manager", "1.0.0", []string{})

	plugin := &FileManagerPlugin{
		BasePlugin:  base,
		uploadDir:   uploadDir,
		downloadDir: downloadDir,
		maxFileSize: maxFileSize,
	}

	// Register routes
	plugin.setupRoutes()

	return plugin
}

// Initialize sets up the file manager plugin
func (p *FileManagerPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	if err := p.BasePlugin.Initialize(ctx, config); err != nil {
		return err
	}

	// Create directories
	if err := p.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	return nil
}

func (p *FileManagerPlugin) setupRoutes() {
	p.AddRoute(core.Route{
		Method:  "GET",
		Path:    "/files",
		Handler: p.handleListFiles,
		Auth:    core.AuthRequirement{Required: false},
	})

	p.AddRoute(core.Route{
		Method:  "POST",
		Path:    "/files",
		Handler: p.handleUploadFile,
		Auth:    core.AuthRequirement{Required: false},
	})

	p.AddRoute(core.Route{
		Method:  "GET",
		Path:    "/files/:filename",
		Handler: p.handleDownloadFile,
		Auth:    core.AuthRequirement{Required: false},
	})

	p.AddRoute(core.Route{
		Method:  "DELETE",
		Path:    "/files/:filename",
		Handler: p.handleDeleteFile,
		Auth:    core.AuthRequirement{Required: false},
	})
}

func (p *FileManagerPlugin) ensureDirectories() error {
	dirs := []string{p.uploadDir, p.downloadDir}

	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *FileManagerPlugin) handleListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := p.listFiles(p.uploadDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"files": files,
		"count": len(files),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (p *FileManagerPlugin) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(p.maxFileSize)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file
	filename := p.sanitizeFilename(header.Filename)
	filePath := filepath.Join(p.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":   "success",
		"filename": filename,
		"size":     header.Size,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (p *FileManagerPlugin) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	// Extract filename from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, "No filename specified", http.StatusBadRequest)
		return
	}
	filename := parts[len(parts)-1]

	if filename == "" {
		http.Error(w, "No filename specified", http.StatusBadRequest)
		return
	}

	// Security check
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(p.uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Serve file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	http.ServeFile(w, r, filePath)
}

func (p *FileManagerPlugin) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	// Extract filename from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, "No filename specified", http.StatusBadRequest)
		return
	}
	filename := parts[len(parts)-1]

	if filename == "" {
		http.Error(w, "No filename specified", http.StatusBadRequest)
		return
	}

	// Security check
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(p.uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":   "success",
		"filename": filename,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (p *FileManagerPlugin) listFiles(dir string) ([]map[string]interface{}, error) {
	if dir == "" {
		return []map[string]interface{}{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]map[string]interface{}, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, map[string]interface{}{
			"name":     entry.Name(),
			"size":     info.Size(),
			"modified": info.ModTime(),
		})
	}

	return files, nil
}

func (p *FileManagerPlugin) sanitizeFilename(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)

	// Replace problematic characters
	replacer := strings.NewReplacer(
		"../", "",
		"./", "",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	return replacer.Replace(filename)
}

// ClipboardPlugin provides clipboard sharing capabilities
type ClipboardPlugin struct {
	*BasePlugin
	clipboard  []ClipboardEntry
	maxHistory int
}

// ClipboardEntry represents a clipboard entry
type ClipboardEntry struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

// NewClipboardPlugin creates a new clipboard plugin
func NewClipboardPlugin(maxHistory int) *ClipboardPlugin {
	base := NewBasePlugin("clipboard", "1.0.0", []string{})

	plugin := &ClipboardPlugin{
		BasePlugin: base,
		clipboard:  make([]ClipboardEntry, 0),
		maxHistory: maxHistory,
	}

	plugin.setupRoutes()

	return plugin
}

func (p *ClipboardPlugin) setupRoutes() {
	p.AddRoute(core.Route{
		Method:  "GET",
		Path:    "/clipboard",
		Handler: p.handleGetClipboard,
		Auth:    core.AuthRequirement{Required: false},
	})

	p.AddRoute(core.Route{
		Method:  "POST",
		Path:    "/clipboard",
		Handler: p.handleSetClipboard,
		Auth:    core.AuthRequirement{Required: false},
	})

	p.AddRoute(core.Route{
		Method:  "GET",
		Path:    "/clipboard/history",
		Handler: p.handleGetHistory,
		Auth:    core.AuthRequirement{Required: false},
	})

	p.AddRoute(core.Route{
		Method:  "DELETE",
		Path:    "/clipboard/history",
		Handler: p.handleClearHistory,
		Auth:    core.AuthRequirement{Required: false},
	})
}

func (p *ClipboardPlugin) handleGetClipboard(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var latest *ClipboardEntry
	if len(p.clipboard) > 0 {
		latest = &p.clipboard[len(p.clipboard)-1]
	}

	response := map[string]interface{}{
		"content": latest,
		"count":   len(p.clipboard),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (p *ClipboardPlugin) handleSetClipboard(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Content string `json:"content"`
		Type    string `json:"type"`
		Source  string `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	entry := ClipboardEntry{
		ID:        fmt.Sprintf("clip-%d", time.Now().UnixNano()),
		Content:   request.Content,
		Type:      request.Type,
		Source:    request.Source,
		Timestamp: time.Now(),
	}

	p.mu.Lock()
	p.clipboard = append(p.clipboard, entry)

	// Trim history if needed
	if len(p.clipboard) > p.maxHistory {
		p.clipboard = p.clipboard[1:]
	}
	p.mu.Unlock()

	response := map[string]interface{}{
		"status": "success",
		"id":     entry.ID,
		"count":  len(p.clipboard),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (p *ClipboardPlugin) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	response := map[string]interface{}{
		"history": p.clipboard,
		"count":   len(p.clipboard),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (p *ClipboardPlugin) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	p.clipboard = make([]ClipboardEntry, 0)
	p.mu.Unlock()

	response := map[string]interface{}{
		"status": "success",
		"count":  0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SystemInfoPlugin provides system information
type SystemInfoPlugin struct {
	*BasePlugin
}

// NewSystemInfoPlugin creates a new system info plugin
func NewSystemInfoPlugin() *SystemInfoPlugin {
	base := NewBasePlugin("system-info", "1.0.0", []string{})

	plugin := &SystemInfoPlugin{
		BasePlugin: base,
	}

	plugin.setupRoutes()

	return plugin
}

func (p *SystemInfoPlugin) setupRoutes() {
	p.AddRoute(core.Route{
		Method:  "GET",
		Path:    "/system/info",
		Handler: p.handleSystemInfo,
		Auth:    core.AuthRequirement{Required: false},
	})

	p.AddRoute(core.Route{
		Method:  "GET",
		Path:    "/system/health",
		Handler: p.handleSystemHealth,
		Auth:    core.AuthRequirement{Required: false},
	})
}

func (p *SystemInfoPlugin) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"hostname": getHostname(),
		"platform": "go",
		"uptime":   time.Since(time.Now()).String(), // Placeholder
		"memory":   getMemoryInfo(),
		"cpu":      getCPUInfo(),
		"network":  getNetworkInfo(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (p *SystemInfoPlugin) handleSystemHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"checks": map[string]string{
			"memory": "ok",
			"disk":   "ok",
			"cpu":    "ok",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Helper functions (these would be properly implemented)
func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func getMemoryInfo() map[string]interface{} {
	return map[string]interface{}{
		"total": "8GB", // Placeholder
		"used":  "4GB", // Placeholder
		"free":  "4GB", // Placeholder
	}
}

func getCPUInfo() map[string]interface{} {
	return map[string]interface{}{
		"cores":       8,     // Placeholder
		"usage":       "25%", // Placeholder
		"temperature": "45C", // Placeholder
	}
}

func getNetworkInfo() map[string]interface{} {
	return map[string]interface{}{
		"interfaces": []string{"eth0", "wlan0"}, // Placeholder
		"active":     "wlan0",                   // Placeholder
	}
}
