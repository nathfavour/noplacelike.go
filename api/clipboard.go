package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/config"
)

// ClipboardEntry represents a single clipboard history entry
type ClipboardEntry struct {
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// ClipboardAPI handles clipboard operations
type ClipboardAPI struct {
	config         *config.Config
	currentText    string
	history        []ClipboardEntry
	historyMaxSize int
	mu             sync.RWMutex
}

// NewClipboardAPI creates a new clipboard API handler
func NewClipboardAPI(cfg *config.Config) *ClipboardAPI {
	maxSize := 50
	if cfg.ClipboardHistorySize > 0 {
		maxSize = cfg.ClipboardHistorySize
	}

	api := &ClipboardAPI{
		config:         cfg,
		history:        make([]ClipboardEntry, 0, maxSize),
		historyMaxSize: maxSize,
	}

	// Initialize with current clipboard content if available
	if text, err := clipboard.ReadAll(); err == nil && text != "" {
		api.currentText = text
		api.history = append(api.history, ClipboardEntry{
			Text:      text,
			Timestamp: time.Now(),
		})
	}

	return api
}

// GetClipboard returns the current clipboard content
func (c *ClipboardAPI) GetClipboard(ctx *gin.Context) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Try to read from system clipboard first
	if text, err := clipboard.ReadAll(); err == nil {
		// Update our internal state if system clipboard changed
		if text != c.currentText {
			c.mu.RUnlock()
			c.mu.Lock()
			c.currentText = text
			c.addToHistory(text)
			c.mu.Unlock()
			c.mu.RLock()
		}
		
		ctx.JSON(http.StatusOK, gin.H{
			"text": text,
		})
		return
	}

	// Fall back to our stored value
	ctx.JSON(http.StatusOK, gin.H{
		"text": c.currentText,
	})
}

// SetClipboard sets the clipboard content
func (c *ClipboardAPI) SetClipboard(ctx *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update system clipboard
	if err := clipboard.WriteAll(req.Text); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to set clipboard: " + err.Error(),
		})
		return
	}

	// Update our internal state
	c.currentText = req.Text
	c.addToHistory(req.Text)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"text":   req.Text,
	})
}

// GetClipboardHistory returns the clipboard history
func (c *ClipboardAPI) GetClipboardHistory(ctx *gin.Context) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx.JSON(http.StatusOK, gin.H{
		"history": c.history,
	})
}

// ClearClipboardHistory clears the clipboard history
func (c *ClipboardAPI) ClearClipboardHistory(ctx *gin.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Keep only the current entry if it exists
	if len(c.history) > 0 {
		current := c.history[0]
		c.history = []ClipboardEntry{current}
	} else {
		c.history = []ClipboardEntry{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// addToHistory adds an entry to the clipboard history
func (c *ClipboardAPI) addToHistory(text string) {
	// Skip if text is empty or same as last entry
	if text == "" || (len(c.history) > 0 && c.history[0].Text == text) {
		return
	}

	// Create new entry
	entry := ClipboardEntry{
		Text:      text,
		Timestamp: time.Now(),
	}

	// Add to front of history
	c.history = append([]ClipboardEntry{entry}, c.history...)

	// Trim if exceeding max size
	if len(c.history) > c.historyMaxSize {
		c.history = c.history[:c.historyMaxSize]
	}
}
