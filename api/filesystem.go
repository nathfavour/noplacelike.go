package api

import (
	// "errors"
	"fmt"
	// "io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/config"
	// "strings" // Import strings package
)

// FileInfo represents information about a file
type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	IsDir        bool      `json:"isDir"`
	ModTime      time.Time `json:"modifiedTime"`
	LastAccessed time.Time `json:"lastAccessed,omitempty"`
	Mode         string    `json:"mode"`
}

// DirContents represents the contents of a directory
type DirContents struct {
	Path        string     `json:"path"`
	Directories []string   `json:"directories"`
	Files       []FileInfo `json:"files"`
}

// FileSystemAPI handles filesystem operations
type FileSystemAPI struct {
	config *config.Config
}

// NewFileSystemAPI creates a new filesystem API handler
func NewFileSystemAPI(cfg *config.Config) *FileSystemAPI {
	return &FileSystemAPI{
		config: cfg,
	}
}

// ListDirectory lists contents of a directory
func (f *FileSystemAPI) ListDirectory(c *gin.Context) {
	// Reload configuration on each request
	if cfg, err := config.Load(); err == nil {
		f.config = cfg
	}
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path parameter is required",
		})
		return
	}

	// Security check: If not in allowed paths, reject
	if !f.isPathAllowed(path) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access to this path is not allowed",
		})
		return
	}

	// Expand path if needed
	expandedPath := expandPath(path)

	// Read directory contents
	entries, err := os.ReadDir(expandedPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Unable to read directory: %v", err),
		})
		return
	}

	// Process directory contents
	contents := DirContents{
		Path:        path,
		Directories: []string{},
		Files:       []FileInfo{},
	}

	for _, entry := range entries {
		// Skip hidden files by default, unless explicitly requested
		if !f.config.ShowHidden && entry.Name()[0] == '.' {
			continue
		}

		if entry.IsDir() {
			contents.Directories = append(contents.Directories, entry.Name())
		} else {
			info, err := entry.Info()
			if err != nil {
				continue // Skip if can't get file info
			}

			// Basic file info
			fileInfo := FileInfo{
				Name:    entry.Name(),
				Size:    info.Size(),
				IsDir:   info.IsDir(),
				ModTime: info.ModTime(),
				Mode:    info.Mode().String(),
			}

			// Try to get additional info on supported platforms
			contents.Files = append(contents.Files, fileInfo)
		}
	}

	// Sort directories and files alphabetically
	sort.Strings(contents.Directories)
	sort.Slice(contents.Files, func(i, j int) bool {
		return contents.Files[i].Name < contents.Files[j].Name
	})

	c.JSON(http.StatusOK, contents)
}

// GetFileContent retrieves the content of a file
func (f *FileSystemAPI) GetFileContent(c *gin.Context) {
	// Reload configuration on each request
	if cfg, err := config.Load(); err == nil {
		f.config = cfg
	}
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path parameter is required",
		})
		return
	}

	// Security check
	if !f.isPathAllowed(path) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access to this file is not allowed",
		})
		return
	}

	// Expand path if needed
	expandedPath := expandPath(path)

	// Check if it's a file
	info, err := os.Stat(expandedPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("File not found: %v", err),
		})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path is a directory, not a file",
		})
		return
	}

	// Only enforce size limit if MaxFileContentSize > 0 (0 means unlimited)
	if f.config.MaxFileContentSize > 0 && info.Size() > int64(f.config.MaxFileContentSize) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("File too large (max %d bytes)", f.config.MaxFileContentSize),
		})
		return
	}

	// Read file content
	content, err := os.ReadFile(expandedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Unable to read file: %v", err),
		})
		return
	}

	// Detect if it's likely a text file or binary
	contentType := detectContentType(content, path)

	// If binary, return error unless force flag is set
	if contentType == "application/octet-stream" && c.Query("force") != "true" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File appears to be binary. Set force=true to read anyway",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":        path,
		"contentType": contentType,
		"size":        info.Size(),
		"content":     string(content),
		"modTime":     info.ModTime(),
	})
}

// isPathAllowed checks if a path is allowed for access
func (f *FileSystemAPI) isPathAllowed(path string) bool {
	// If no allowed paths are specified, use a safe default
	if len(f.config.AllowedPaths) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		return isSubPath(expandPath(path), filepath.Join(homeDir, "Downloads"))
	}

	// Otherwise check if path is within any allowed path
	for _, allowedPath := range f.config.AllowedPaths {
		if isSubPath(expandPath(path), expandPath(allowedPath)) {
			return true
		}
	}

	return false
}

// detectContentType tries to determine if a file is text or binary
func detectContentType(content []byte, path string) string {
	// First check file extension
	switch filepath.Ext(path) {
	case ".txt", ".md", ".json", ".xml", ".html", ".htm", ".css", ".js", ".go", ".py", ".c", ".cpp", ".h", ".java":
		return "text/plain"
	}

	// Then try http.DetectContentType
	return http.DetectContentType(content)
}

// ServeFile serves raw file content for download or streaming
func (f *FileSystemAPI) ServeFile(c *gin.Context) {
	// Reload config on each request
	if cfg, err := config.Load(); err == nil {
		f.config = cfg
	}
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}
	if !f.isPathAllowed(path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access to this file is not allowed"})
		return
	}
	expandedPath := expandPath(path)
	// Serve file with proper headers (supports Range). Use attachment when download=true
	if c.Query("download") == "true" {
		c.FileAttachment(expandedPath, filepath.Base(expandedPath))
		return
	}
	c.File(expandedPath)
}
