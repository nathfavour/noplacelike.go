package server

import (
	"net/http"

	"github.com/atotto/clipboard"
	"github.com/gin-gonic/gin"
)

type clipboardRequest struct {
	Text string `json:"text"`
}

// getClipboard returns the server's clipboard content
func (s *Server) getClipboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"text": s.clipboard,
	})
}

// setClipboard sets the server's clipboard content
func (s *Server) setClipboard(c *gin.Context) {
	var req clipboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Store clipboard text in memory
	s.clipboard = req.Text

	// Try to set system clipboard if available
	_ = clipboard.WriteAll(req.Text)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
