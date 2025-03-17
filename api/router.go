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

// NewAPI creates a new API instance
func NewAPI(cfg *config.Config) *API {
	return &API{
		config: cfg,
	}
}

// CreateRoutes creates all the API routes
func (a *API) CreateRoutes(router *gin.Engine) {
	// Initialize API documentation
	InitDocs()

	// Base API route group
	api := router.Group("/api")
	{
		// API documentation
		api.GET("/", a.redirectToDocumentation)
	// Base API route groupAPIDocsUI)
	api := router.Group("/api")
	{// Version 1 API
		// API documentation")
		api.GET("/", a.redirectToDocumentation)
		api.GET("/docs", ServeAPIDocsUI)
			v1.GET("/docs", ServeAPIDocsUI)
		// Version 1 APIson", ServeAPIDocsJSON)
		v1 := api.Group("/v1")
		{// Clipboard endpoints
			// API docs= v1.Group("/clipboard")
			v1.GET("/docs", ServeAPIDocsUI)
			v1.GET("/docs/json", ServeAPIDocsJSON)board)
				clipboard.POST("", a.clipboard.SetClipboard)
			// Clipboard endpointsry", a.clipboard.GetClipboardHistory)
			clipboard := v1.Group("/clipboard")pboard.ClearClipboardHistory)
			{
				clipboard.GET("", a.clipboard.GetClipboard)
				clipboard.POST("", a.clipboard.SetClipboard)
				clipboard.GET("/history", a.clipboard.GetClipboardHistory)
				clipboard.DELETE("/history", a.clipboard.ClearClipboardHistory)
			}files.GET("", a.listFiles)
				files.POST("", a.uploadFile)
			// File operationsname", a.downloadFile)
			files := v1.Group("/files") a.deleteFile)
			{
				files.GET("", a.listFiles)
				files.POST("", a.uploadFile)
				files.GET("/:filename", a.downloadFile)
				files.DELETE("/:filename", a.deleteFile)
			}filesystem.GET("/list", a.filesystem.ListDirectory)
				filesystem.GET("/content", a.filesystem.GetFileContent)
			// Filesystem operationsm endpoints could be added here
			filesystem := v1.Group("/filesystem")
			{
				filesystem.GET("/list", a.filesystem.ListDirectory)
				filesystem.GET("/content", a.filesystem.GetFileContent)
				// Additional filesystem endpoints could be added here
			}shell.POST("/exec", a.shell.ExecuteCommand)
				shell.GET("/stream", a.shell.StreamCommand)
			// Shell command execution
			shell := v1.Group("/shell")
			{/ System information
				shell.POST("/exec", a.shell.ExecuteCommand)
				shell.GET("/stream", a.shell.StreamCommand)
			}system.GET("/info", a.system.GetSystemInfo)
				system.GET("/processes", a.system.GetProcesses)
			// System information", a.system.SendNotification)
			system := v1.Group("/system")
			{
				system.GET("/info", a.system.GetSystemInfo)
				system.GET("/processes", a.system.GetProcesses)
				system.POST("/notify", a.system.SendNotification)
			}audio := media.Group("/audio")
				{
			// Media streamingces", a.media.GetAudioDevices)
			media := v1.Group("/media")dia.StreamAudio)
			{}
				audio := media.Group("/audio")
				{edia.GET("/screen", a.media.StreamScreen)
					audio.GET("/devices", a.media.GetAudioDevices)
					audio.GET("/stream", a.media.StreamAudio)
				}
				
				media.GET("/screen", a.media.StreamScreen)
			}We maintain these for backward compatibility
		}i.GET("/clipboard", a.clipboard.GetClipboard)
	}pi.POST("/clipboard", a.clipboard.SetClipboard)
	api.GET("/files", a.listFiles)
	// Compatibility with existing endpoints
	// We maintain these for backward compatibility
	api.GET("/clipboard", a.clipboard.GetClipboard)
	api.POST("/clipboard", a.clipboard.SetClipboard)
	api.GET("/files", a.listFiles)rects to API documentation
	api.POST("/files", a.uploadFile)tion(c *gin.Context) {
	api.GET("/files/:filename", a.downloadFile))
}

// redirectToDocumentation redirects to API documentation
func (a *API) redirectToDocumentation(c *gin.Context) {
	c.Redirect(http.StatusFound, "/api/v1/docs")r)
}
	// Create directory if it doesn't exist, instead of failing
// listFiles handles file listing os.IsNotExist(err) {
func (a *API) listFiles(c *gin.Context) { err != nil {
	uploadDir := expandPath(a.config.UploadFolder){
				"error": "Error accessing files directory: " + err.Error(),
	// Create directory if it doesn't exist, instead of failing
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error accessing files directory: " + err.Error(),
			})iles": []string{},
			return
		}eturn
		// Return empty list for new directory
		c.JSON(http.StatusOK, gin.H{
			"files": []string{},l file listing
		})es, err := listFilesInDir(uploadDir)
		return!= nil {
	}c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files: " + err.Error(),
	// Continue with normal file listing
	files, err := listFilesInDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files: " + err.Error(),
		})iles": files,
		return
	}
	
	c.JSON(http.StatusOK, gin.H{load
		"files": files,oadFile(c *gin.Context) {
	})loadDir := expandPath(a.config.UploadFolder)
}
	// Create upload directory if it doesn't exist
// uploadFile handles file upload os.IsNotExist(err) {
func (a *API) uploadFile(c *gin.Context) {err != nil {
	uploadDir := expandPath(a.config.UploadFolder){
				"error": "Failed to save file: " + err.Error(),
	// Create upload directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save file: " + err.Error(),
			}), err := c.FormFile("file")
			return= nil {
		}.JSON(http.StatusBadRequest, gin.H{
	}	"error": "No file provided",
		})
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})ename := getSafeFilename(file.Filename)
		return
	}/ Save the file
	dst := filepath.Join(uploadDir, filename)
	// Ensure filename is safele(file, dst); err != nil {
	filename := getSafeFilename(file.Filename)n.H{
			"error": "Failed to save file: " + err.Error(),
	// Save the file
	dst := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file: " + err.Error(),
		})tatus":   "success",
		returname": filename,
	})
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success", file for download
		"filename": filename,ile(c *gin.Context) {
	})lename := c.Param("filename")
}if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
// downloadFile serves a file for download
func (a *API) downloadFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No filename specified",ain path traversal
		})strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return(http.StatusBadRequest, gin.H{
	}	"error": "Invalid filename",
		})
	// Ensure the filename doesn't contain path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",ndPath(a.config.UploadFolder), filename)
		})
		returnk if file exists
	}f _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
	filepath := filepath.Join(expandPath(a.config.UploadFolder), filename)
		})
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})ile(filepath)
		return
	}
	/ deleteFile deletes a file
	// Serve the fileteFile(c *gin.Context) {
	c.File(filepath)ram("filename")
}if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
// deleteFile deletes a filecified",
func (a *API) deleteFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No filename specified",ain path traversal
		})strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return(http.StatusBadRequest, gin.H{
	}	"error": "Invalid filename",
		})
	// Ensure the filename doesn't contain path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",ndPath(a.config.UploadFolder), filename)
		})
		returnk if file exists
	}f _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
	filepath := filepath.Join(expandPath(a.config.UploadFolder), filename)
		})
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})err := os.Remove(filepath); err != nil {
		return(http.StatusInternalServerError, gin.H{
	}	"error": "Failed to delete file: " + err.Error(),
		})
	// Delete the file
	if err := os.Remove(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file: " + err.Error(),
		})tatus": "success",
		return
	}
	
	c.JSON(http.StatusOK, gin.H{ist of files in a directory
		"status": "success",r string) ([]string, error) {
	}) Ensure directory exists
}if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
// listFilesInDir returns a list of files in a directory
func listFilesInDir(dir string) ([]string, error) {
	// Ensure directory exists Return empty list for new directory
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err contents
		}tries, err := os.ReadDir(dir)
		return []string{}, nil // Return empty list for new directory
	}return nil, err
	}
	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {ing
		return nil, errrange entries {
	}if !entry.IsDir() {
			files = append(files, entry.Name())
	// Extract filenames
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	} getSafeFilename ensures a filename is safe for use
	unc getSafeFilename(filename string) string {
	return files, nilmponents
}filename = filepath.Base(filename)
	
// getSafeFilename ensures a filename is safe for use
func getSafeFilename(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)
		"/", "_",
	// Replace potentially problematic characters
	replacer := strings.NewReplacer(
		"../", "",
		"./", "",
		"/", "_",,
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",lacer.Replace(filename)
		">", "_",
		"|", "_",	)		return replacer.Replace(filename)}
// initDirectories initializes directories ONLY when they're needed
func initDirectory(path string) error {
    if path == "" {
        return fmt.Errorf("empty path provided")
    }
    
    expandedPath := expandPath(path)
    if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
        if err := os.MkdirAll(expandedPath, 0755); err != nil {
            return err
        }
    }
    
    return nil
}

// expandPath expands the ~ in a path to the user's home directory
func expandPath(path string) string {
    if path == "~" || strings.HasPrefix(path, "~/") {
        homeDir, err := os.UserHomeDir()
        if err == nil {
            return filepath.Join(homeDir, path[1:])
        }
    }
    return path
}
