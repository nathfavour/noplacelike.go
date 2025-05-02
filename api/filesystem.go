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
	"io"
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

// CreateDirectory creates a new directory
func (f *FileSystemAPI) CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing path"})
		return
	}
	if !f.isPathAllowed(req.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
		return
	}
	if err := os.MkdirAll(expandPath(req.Path), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "created"})
}

// RenameFile renames a file or directory
func (f *FileSystemAPI) RenameFile(c *gin.Context) {
	var req struct{ OldPath, NewPath string }
	if err := c.ShouldBindJSON(&req); err != nil || req.OldPath == "" || req.NewPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing path(s)"})
		return
	}
	if !f.isPathAllowed(req.OldPath) || !f.isPathAllowed(req.NewPath) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
		return
	}
	if err := os.Rename(expandPath(req.OldPath), expandPath(req.NewPath)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "renamed"})
}

// DeletePath deletes a file or directory
func (f *FileSystemAPI) DeletePath(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing path"})
		return
	}
	if !f.isPathAllowed(req.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
		return
	}
	if err := os.RemoveAll(expandPath(req.Path)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// CopyFile copies a file
func (f *FileSystemAPI) CopyFile(c *gin.Context) {
	var req struct{ Src, Dst string }
	if err := c.ShouldBindJSON(&req); err != nil || req.Src == "" || req.Dst == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing src/dst"})
		return
	}
	if !f.isPathAllowed(req.Src) || !f.isPathAllowed(req.Dst) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
		return
	}
	src := expandPath(req.Src)
	dst := expandPath(req.Dst)
	in, err := os.Open(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "copied"})
}

// MoveFile moves a file or directory
func (f *FileSystemAPI) MoveFile(c *gin.Context) {
	var req struct{ Src, Dst string }
	if err := c.ShouldBindJSON(&req); err != nil || req.Src == "" || req.Dst == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing src/dst"})
		return
	}
	if !f.isPathAllowed(req.Src) || !f.isPathAllowed(req.Dst) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
		return
	}
	if err := os.Rename(expandPath(req.Src), expandPath(req.Dst)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "moved"})
}

// SearchFiles searches for files by name in allowed paths
func (f *FileSystemAPI) SearchFiles(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query"})
		return
	}
	var results []FileInfo
	for _, base := range f.config.AllowedPaths {
		_ = filepath.Walk(expandPath(base), func(path string, info os.FileInfo, err error) error {
			if err == nil && info != nil && !info.IsDir() && filepath.Base(path) == q {
				results = append(results, FileInfo{
					Name:    info.Name(),
					Size:    info.Size(),
					IsDir:   false,
					ModTime: info.ModTime(),
					Mode:    info.Mode().String(),
				})
			}
			return nil
		})
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
