package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// listFiles lists all files in the upload directory
func (s *Server) listFiles(c *gin.Context) {
	uploadDir := expandPath(s.config.UploadFolder)
	
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read directory",
		})
		return
	}
	
	fileNames := []string{}
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"files": fileNames,
	})
}

// uploadFile handles file uploads
func (s *Server) uploadFile(c *gin.Context) {
	uploadDir := expandPath(s.config.UploadFolder)

	// Create upload directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error creating upload directory: " + err.Error(),
			})
			return
		}
	}
	
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	
	// Ensure filename is safe
	filename := filepath.Base(file.Filename)
	
	// Save the file
	dst := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"filename": filename,
	})
}

// downloadFile serves a file for download
func (s *Server) downloadFile(c *gin.Context) {
	uploadDir := expandPath(s.config.UploadFolder)
	filename := c.Param("filename")
	
	// Ensure no path traversal
	if filepath.Base(filename) != filename {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}
	
	filePath := filepath.Join(uploadDir, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}
	
	// Serve the file
	c.FileAttachment(filePath, filename)
}
