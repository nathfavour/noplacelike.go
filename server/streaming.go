package server

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// streamAudio streams an audio file
func (s *Server) streamAudio(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing file parameter",
		})
		return
	}
	
	// Clean the filename to prevent path traversal
	safeFilename := filepath.Base(filename)
	if safeFilename != filename {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}
	
	// Check all configured audio folders for the file
	var filePath string
	found := false
	
	for _, folder := range s.config.AudioFolders {
		expandedFolder := expandPath(folder)
		candidatePath := filepath.Join(expandedFolder, safeFilename)
		if _, err := os.Stat(candidatePath); err == nil {
			filePath = candidatePath
			found = true
			break
		}
	}
	
	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}
	
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
		})
		return
	}
	defer file.Close()
	
	// Get file info for size
	info, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get file info",
		})
		return
	}
	
	// Set content type based on file extension
	contentType := "audio/mpeg"
	if strings.HasSuffix(strings.ToLower(safeFilename), ".ogg") {
		contentType = "audio/ogg"
	} else if strings.HasSuffix(strings.ToLower(safeFilename), ".wav") {
		contentType = "audio/wav"
	} else if strings.HasSuffix(strings.ToLower(safeFilename), ".flac") {
		contentType = "audio/flac"
	}
	
	// Set response headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", string(info.Size()))
	c.Header("Accept-Ranges", "bytes")
	
	// Stream the file
	c.Status(http.StatusOK)
	io.Copy(c.Writer, file)
}

// listAudio lists audio files from all configured folders
func (s *Server) listAudio(c *gin.Context) {
	result := make(map[string][]string)
	
	for _, folder := range s.config.AudioFolders {
		expandedFolder := expandPath(folder)
		
		// Try to read directory
		files, err := os.ReadDir(expandedFolder)
		if err != nil {
			// Skip if folder doesn't exist or can't be read
			result[expandedFolder] = []string{}
			continue
		}
		
		fileList := []string{}
		for _, file := range files {
			if !file.IsDir() {
				// Simple extension check for audio files
				name := file.Name()
				ext := strings.ToLower(filepath.Ext(name))
				if ext == ".mp3" || ext == ".ogg" || ext == ".wav" || ext == ".flac" {
					fileList = append(fileList, name)
				}
			}
		}
		
		result[expandedFolder] = fileList
	}
	
	c.JSON(http.StatusOK, gin.H{
		"files": result,
	})
}
