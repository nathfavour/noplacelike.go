package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/config"
)

// API represents the main API handler
type API struct {
	config     *config.Config
	clipboard  *ClipboardAPI
	filesystem *FileSystemAPI
	shell      *ShellAPI
	system     *SystemAPI
	media      *MediaAPI
}

// New creates a new API instance
func New(cfg *config.Config) *API {
	return &API{
		config:     cfg,
		clipboard:  NewClipboardAPI(cfg),
		filesystem: NewFileSystemAPI(cfg),
		shell:      NewShellAPI(cfg),
		system:     NewSystemAPI(cfg),
		media:      NewMediaAPI(cfg),
	}
}

// SetupRoutes configures all API routes
func (a *API) SetupRoutes(router *gin.Engine) {
	// Initialize API documentation
	InitDocs()

	// Base API route group
	api := router.Group("/api")
	{
		// API documentation
		api.GET("/", a.redirectToDocumentation)
		api.GET("/docs", ServeAPIDocsUI)
		
		// Version 1 API
		v1 := api.Group("/v1")
		{
			// API docs
			v1.GET("/docs", ServeAPIDocsUI)
			v1.GET("/docs/json", ServeAPIDocsJSON)

			// Clipboard endpoints
			clipboard := v1.Group("/clipboard")
			{
				clipboard.GET("", a.clipboard.GetClipboard)
				clipboard.POST("", a.clipboard.SetClipboard)
				clipboard.GET("/history", a.clipboard.GetClipboardHistory)
				clipboard.DELETE("/history", a.clipboard.ClearClipboardHistory)
			}

			// File operations
			files := v1.Group("/files")
			{
				files.GET("", a.listFiles)
				files.POST("", a.uploadFile)
				files.GET("/:filename", a.downloadFile)
				files.DELETE("/:filename", a.deleteFile)
			}

			// Filesystem operations
			filesystem := v1.Group("/filesystem")
			{
				filesystem.GET("/list", a.filesystem.ListDirectory)
				filesystem.GET("/content", a.filesystem.GetFileContent)
				// Additional filesystem endpoints could be added here
			}

			// Shell command execution
			shell := v1.Group("/shell")
			{
				shell.POST("/exec", a.shell.ExecuteCommand)
				shell.GET("/stream", a.shell.StreamCommand)
			}

			// System information
			system := v1.Group("/system")
			{
				system.GET("/info", a.system.GetSystemInfo)
				system.GET("/processes", a.system.GetProcesses)
				system.POST("/notify", a.system.SendNotification)
			}

			// Media streaming
			media := v1.Group("/media")
			{
				audio := media.Group("/audio")
				{
					audio.GET("/devices", a.media.GetAudioDevices)
					audio.GET("/stream", a.media.StreamAudio)
				}
				
				media.GET("/screen", a.media.StreamScreen)
			}
		}
	}

	// Compatibility with existing endpoints
	// We maintain these for backward compatibility
	api.GET("/clipboard", a.clipboard.GetClipboard)
	api.POST("/clipboard", a.clipboard.SetClipboard)
	api.GET("/files", a.listFiles)
	api.POST("/files", a.uploadFile)
	api.GET("/files/:filename", a.downloadFile)
}

// redirectToDocumentation redirects to API documentation
func (a *API) redirectToDocumentation(c *gin.Context) {
	c.Redirect(http.StatusFound, "/api/v1/docs")
}

// listFiles lists all files in the uploads directory
func (a *API) listFiles(c *gin.Context) {
	uploadDir := expandPath(a.config.UploadFolder)
	files, err := listFilesInDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// uploadFile handles file upload
func (a *API) uploadFile(c *gin.Context) {
	uploadDir := expandPath(a.config.UploadFolder)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	
	// Ensure filename is safe
	filename := getSafeFilename(file.Filename)
	
	// Save the file
	dst := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"filename": filename,
	})
}

// downloadFile serves a file for download
func (a *API) downloadFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No filename specified",
		})
		return
	}
	
	// Ensure the filename doesn't contain path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}
	
	filepath := filepath.Join(expandPath(a.config.UploadFolder), filename)
	
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}
	
	// Serve the file
	c.File(filepath)
}

// deleteFile deletes a file
func (a *API) deleteFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No filename specified",
		})
		return
	}
	
	// Ensure the filename doesn't contain path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}
	
	filepath := filepath.Join(expandPath(a.config.UploadFolder), filename)
	
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}
	
	// Delete the file
	if err := os.Remove(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// listFilesInDir returns a list of files in a directory
func listFilesInDir(dir string) ([]string, error) {
	// Ensure directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		return []string{}, nil // Return empty list for new directory
	}
	
	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	// Extract filenames
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// getSafeFilename ensures a filename is safe for use
func getSafeFilename(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)
	
	// Replace potentially problematic characters
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
