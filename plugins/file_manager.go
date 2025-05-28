package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nathfavour/noplacelike.go/internal/core"
	"github.com/nathfavour/noplacelike.go/internal/logger"
)

// FileManagerPlugin provides comprehensive file management capabilities
type FileManagerPlugin struct {
	id       string
	version  string
	logger   logger.Logger
	platform core.PlatformAPI
	config   FileManagerConfig
	running  bool
}

type FileManagerConfig struct {
	BaseDir     string   `json:"baseDir"`
	MaxFileSize int64    `json:"maxFileSize"`
	AllowedExts []string `json:"allowedExts"`
	EnableCORS  bool     `json:"enableCors"`
}

// NewFileManagerPlugin creates a new file manager plugin
func NewFileManagerPlugin() core.Plugin {
	return &FileManagerPlugin{
		id:      "file-manager",
		version: "1.0.0",
		config: FileManagerConfig{
			BaseDir:     "./files",
			MaxFileSize: 100 * 1024 * 1024, // 100MB
			AllowedExts: []string{},        // Empty means all extensions allowed
			EnableCORS:  true,
		},
	}
}

// Plugin interface implementation
func (p *FileManagerPlugin) ID() string {
	return p.id
}

func (p *FileManagerPlugin) Version() string {
	return p.version
}

func (p *FileManagerPlugin) Dependencies() []string {
	return []string{}
}

func (p *FileManagerPlugin) Name() string {
	return "File Manager Plugin"
}

func (p *FileManagerPlugin) Initialize(platform core.PlatformAPI) error {
	p.platform = platform
	p.logger = platform.GetLogger().WithFields(map[string]interface{}{
		"plugin": p.id,
	})

	// Ensure base directory exists
	if err := os.MkdirAll(p.config.BaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	p.logger.Info("File manager plugin initialized", "baseDir", p.config.BaseDir)
	return nil
}

func (p *FileManagerPlugin) Configure(config map[string]interface{}) error {
	if configBytes, err := json.Marshal(config); err == nil {
		if err := json.Unmarshal(configBytes, &p.config); err != nil {
			p.logger.Warn("Failed to parse configuration", "error", err)
		}
	}

	p.logger.Info("File manager plugin configured", "config", p.config)
	return nil
}

func (p *FileManagerPlugin) Start(ctx context.Context) error {
	p.running = true
	p.logger.Info("File manager plugin started")

	// Register as a resource provider
	if resourceMgr := p.platform.GetResourceManager(); resourceMgr != nil {
		resource := core.Resource{
			ID:          p.id,
			Type:        "file-manager",
			Name:        "File Manager",
			Description: "Provides file upload, download, and management capabilities",
			Provider:    p.id,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		resourceMgr.RegisterResource(resource)
	}

	return nil
}

func (p *FileManagerPlugin) Stop(ctx context.Context) error {
	p.running = false

	// Unregister resource
	if resourceMgr := p.platform.GetResourceManager(); resourceMgr != nil {
		resourceMgr.UnregisterResource(p.id)
	}

	p.logger.Info("File manager plugin stopped")
	return nil
}

func (p *FileManagerPlugin) IsHealthy() bool {
	return p.running && p.isBaseDirAccessible()
}

func (p *FileManagerPlugin) Routes() []core.Route {
	return []core.Route{
		{
			Method:      "GET",
			Path:        "/plugins/file-manager/files",
			Handler:     p.handleListFiles,
			Description: "List all files in the managed directory",
		},
		{
			Method:      "POST",
			Path:        "/plugins/file-manager/files",
			Handler:     p.handleUploadFile,
			Description: "Upload a new file",
		},
		{
			Method:      "GET",
			Path:        "/plugins/file-manager/files/:filename",
			Handler:     p.handleDownloadFile,
			Description: "Download a specific file",
		},
		{
			Method:      "DELETE",
			Path:        "/plugins/file-manager/files/:filename",
			Handler:     p.handleDeleteFile,
			Description: "Delete a specific file",
		},
		{
			Method:      "GET",
			Path:        "/plugins/file-manager/info/:filename",
			Handler:     p.handleFileInfo,
			Description: "Get file information and metadata",
		},
	}
}

func (p *FileManagerPlugin) HandleEvent(event core.Event) error {
	switch event.Type {
	case "file.uploaded":
		p.logger.Info("File uploaded", "filename", event.Data["filename"])
	case "file.deleted":
		p.logger.Info("File deleted", "filename", event.Data["filename"])
	}
	return nil
}

// HTTP handlers
func (p *FileManagerPlugin) handleListFiles(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	files, err := p.listFiles()
	if err != nil {
		p.logger.Error("Error listing files", "error", err)
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

func (p *FileManagerPlugin) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	if r.Method == "OPTIONS" {
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(p.config.MaxFileSize); err != nil {
		p.logger.Error("Error parsing multipart form", "error", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		p.logger.Error("Error getting file from form", "error", err)
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file
	if !p.isFileAllowed(header.Filename) {
		http.Error(w, "File type not allowed", http.StatusBadRequest)
		return
	}

	if header.Size > p.config.MaxFileSize {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	// Save file
	filename, err := p.saveFile(file, header)
	if err != nil {
		p.logger.Error("Error saving file", "error", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Publish event
	if eventBus := p.platform.GetEventBus(); eventBus != nil {
		event := core.Event{
			Type:   "file.uploaded",
			Source: p.id,
			Data: map[string]interface{}{
				"filename": filename,
				"size":     header.Size,
			},
		}
		eventBus.Publish(event)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "File uploaded successfully",
		"filename": filename,
		"size":     header.Size,
	})
}

func (p *FileManagerPlugin) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	filename := p.extractFilename(r.URL.Path)
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(p.config.BaseDir, filename)

	// Security check - ensure file is within base directory
	if !p.isPathSafe(filePath) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if err != nil {
		p.logger.Error("Error checking file", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		p.logger.Error("Error opening file", "error", err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	// Stream file
	io.Copy(w, file)
}

func (p *FileManagerPlugin) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	filename := p.extractFilename(r.URL.Path)
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(p.config.BaseDir, filename)

	// Security check
	if !p.isPathSafe(filePath) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			p.logger.Error("Error deleting file", "error", err)
			http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		}
		return
	}

	// Publish event
	if eventBus := p.platform.GetEventBus(); eventBus != nil {
		event := core.Event{
			Type:   "file.deleted",
			Source: p.id,
			Data: map[string]interface{}{
				"filename": filename,
			},
		}
		eventBus.Publish(event)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "File deleted successfully",
		"filename": filename,
	})
}

func (p *FileManagerPlugin) handleFileInfo(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	filename := p.extractFilename(r.URL.Path)
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(p.config.BaseDir, filename)

	if !p.isPathSafe(filePath) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if err != nil {
		p.logger.Error("Error getting file info", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":      info.Name(),
		"size":      info.Size(),
		"modified":  info.ModTime().Unix(),
		"is_dir":    info.IsDir(),
		"mode":      info.Mode().String(),
		"extension": filepath.Ext(filename),
	})
}

// Helper methods
func (p *FileManagerPlugin) listFiles() ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(p.config.BaseDir)
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
			"name":      entry.Name(),
			"size":      info.Size(),
			"modified":  info.ModTime().Unix(),
			"extension": filepath.Ext(entry.Name()),
		})
	}

	return files, nil
}

func (p *FileManagerPlugin) saveFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Generate safe filename
	filename := p.sanitizeFilename(header.Filename)
	filePath := filepath.Join(p.config.BaseDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return "", err
	}

	return filename, nil
}

func (p *FileManagerPlugin) isFileAllowed(filename string) bool {
	if len(p.config.AllowedExts) == 0 {
		return true // All extensions allowed
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowedExt := range p.config.AllowedExts {
		if ext == strings.ToLower(allowedExt) {
			return true
		}
	}

	return false
}

func (p *FileManagerPlugin) isPathSafe(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absBaseDir, err := filepath.Abs(p.config.BaseDir)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absPath, absBaseDir)
}

func (p *FileManagerPlugin) sanitizeFilename(filename string) string {
	// Remove path separators and other unsafe characters
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "..", "")
	return filename
}

func (p *FileManagerPlugin) extractFilename(urlPath string) string {
	parts := strings.Split(urlPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func (p *FileManagerPlugin) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (p *FileManagerPlugin) isBaseDirAccessible() bool {
	_, err := os.Stat(p.config.BaseDir)
	return err == nil
}
