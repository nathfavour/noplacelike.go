package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/config"
)

// dirRequest is the request structure for adding/removing directories
type dirRequest struct {
	Dir string `json:"dir"`
}

// getAudioDirs returns all configured audio directories
func (s *Server) getAudioDirs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"dirs": s.config.AudioFolders,
	})
}

// addAudioDir adds a new audio directory to the configuration
func (s *Server) addAudioDir(c *gin.Context) {
	var req dirRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Dir == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid directory path",
		})
		return
	}
	
	// Check if directory already exists in config
	dirExists := false
	for _, dir := range s.config.AudioFolders {
		if dir == req.Dir {
			dirExists = true
			break
		}
	}
	
	// Add directory if it doesn't exist
	if !dirExists {
		s.config.AudioFolders = append(s.config.AudioFolders, req.Dir)
		
		// Save updated config
		if err := config.Save(s.config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save configuration",
			})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// removeAudioDir removes an audio directory from the configuration
func (s *Server) removeAudioDir(c *gin.Context) {
	var req dirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}
	
	// Filter out the directory to remove
	newDirs := []string{}
	for _, dir := range s.config.AudioFolders {
		if dir != req.Dir {
			newDirs = append(newDirs, dir)
		}
	}
	
	// Update config
	s.config.AudioFolders = newDirs
	
	// Save updated config
	if err := config.Save(s.config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save configuration",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
