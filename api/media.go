package api

import (
	// "errors"
	"net/http"
	"strconv"
	"time"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nathfavour/noplacelike.go/config"
)

// MediaAPI handles media streaming operations
type MediaAPI struct {
	config     *config.Config
	wsUpgrader websocket.Upgrader
}

// NewMediaAPI creates a new media API handler
func NewMediaAPI(cfg *config.Config) *MediaAPI {
	return &MediaAPI{
		config: cfg,
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin
			},
		},
	}
}

// AudioDevice represents an audio device on the system
type AudioDevice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsInput     bool   `json:"isInput"`
	IsOutput    bool   `json:"isOutput"`
	IsDefault   bool   `json:"isDefault"`
	SampleRate  int    `json:"sampleRate,omitempty"`
	Channels    int    `json:"channels,omitempty"`
	Description string `json:"description,omitempty"`
}

// GetAudioDevices returns a list of audio devices on the system
func (m *MediaAPI) GetAudioDevices(c *gin.Context) {
	// This is a mock implementation
	// TODO: Implement actual audio device detection based on platform
	// For example, using a library like:
	// - go-portaudio for cross-platform support
	// - or platform-specific libraries

	devices := []AudioDevice{
		{
			ID:          "default",
			Name:        "System Default",
			IsOutput:    true,
			IsDefault:   true,
			SampleRate:  44100,
			Channels:    2,
			Description: "Default system audio output",
		},
		{
			ID:          "default-input",
			Name:        "System Default Input",
			IsInput:     true,
			IsDefault:   true,
			SampleRate:  44100,
			Channels:    1,
			Description: "Default system audio input",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

// StreamAudio streams audio over WebSocket
func (m *MediaAPI) StreamAudio(c *gin.Context) {
	// Check if audio streaming is enabled
	if (!m.config.EnableAudioStreaming) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Audio streaming is disabled",
		})
		return
	}

	// Get device ID from query parameter
	deviceID := c.DefaultQuery("device", "default")

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := m.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upgrade connection: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	// Send initial message
	conn.WriteJSON(map[string]string{
		"status": "Connected",
		"device": deviceID,
	})

	// TODO: Implement actual audio capture and streaming
	// This would typically involve:
	// 1. Setting up an audio capture from the specified device
	// 2. Processing the audio (e.g., encoding to a suitable format like Opus)
	// 3. Streaming the packets over the WebSocket connection

	// For now, just keep the connection alive
	for {
		// Read from WebSocket (client messages)
		_, _, err := conn.ReadMessage()
		if err != nil {
			break // Exit on connection close or error
		}
		
		// Send a ping every 5 seconds to keep connection alive
		time.Sleep(5 * time.Second)
		if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
			break
		}
	}
}

// StreamScreen streams screen content over WebSocket
func (m *MediaAPI) StreamScreen(c *gin.Context) {
	// Check if screen streaming is enabled
	if (!m.config.EnableScreenStreaming) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Screen streaming is disabled",
		})
		return
	}

	// Get streaming parameters
	quality := c.DefaultQuery("quality", "medium")
	fpsStr := c.DefaultQuery("fps", "15")
	
	fps, err := strconv.Atoi(fpsStr)
	if err != nil || fps < 1 || fps > 30 {
		fps = 15 // Default to 15 FPS if invalid
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := m.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upgrade connection: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	// Send initial message
	conn.WriteJSON(map[string]interface{}{
		"status":  "Connected",
		"quality": quality,
		"fps":     fps,
	})

	// TODO: Implement actual screen capture and streaming
	// This would typically involve:
	// 1. Capturing screen frames at the specified FPS
	// 2. Encoding the frames to a suitable format (e.g., JPEG, VP8)
	// 3. Streaming the encoded frames over the WebSocket connection

	// For now, just keep the connection alive
	for {
		// Read from WebSocket (client messages)
		_, _, err := conn.ReadMessage()
		if err != nil {
			break // Exit on connection close or error
		}
		
		// Send a ping every 5 seconds to keep connection alive
		time.Sleep(5 * time.Second)
		if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
			break
		}
	}
}

// MediaDirInfo represents a directory with media info
type MediaDirInfo struct {
	Path         string   `json:"path"`
	AudioCount   int      `json:"audioCount"`
	TotalCount   int      `json:"totalCount"`
	Ratio        float64  `json:"ratio"`
	SampleFiles  []string `json:"sampleFiles"`
}

// ScanMediaDirectories scans allowed paths for media-rich directories
func (m *MediaAPI) ScanMediaDirectories(c *gin.Context) {
	var results []MediaDirInfo
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true}
	for _, base := range m.config.AllowedPaths {
		_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() {
				return nil
			}
			files, _ := os.ReadDir(path)
			total, audio := 0, 0
			var samples []string
			for _, f := range files {
				if f.IsDir() { continue }
				total++
				ext := filepath.Ext(f.Name())
				if audioExts[ext] {
					audio++
					if len(samples) < 3 { samples = append(samples, f.Name()) }
				}
			}
			if total > 0 && float64(audio)/float64(total) > 0.5 && audio >= 3 {
				results = append(results, MediaDirInfo{
					Path: path, AudioCount: audio, TotalCount: total, Ratio: float64(audio)/float64(total), SampleFiles: samples,
				})
			}
			return nil
		})
	}
	c.JSON(http.StatusOK, gin.H{"mediaDirs": results})
}

// ListMediaFiles lists audio files in a directory
func (m *MediaAPI) ListMediaFiles(c *gin.Context) {
	dir := c.Query("dir")
	if dir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing dir"})
		return
	}
	files, _ := os.ReadDir(dir)
	var audioFiles []string
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true}
	for _, f := range files {
		if !f.IsDir() && audioExts[filepath.Ext(f.Name())] {
			audioFiles = append(audioFiles, f.Name())
		}
	}
	c.JSON(http.StatusOK, gin.H{"files": audioFiles})
}

// GetMediaMetadata returns basic metadata for an audio file
func (m *MediaAPI) GetMediaMetadata(c *gin.Context) {
	file := c.Query("file")
	if file == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
		return
	}
	info, err := os.Stat(file)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name": info.Name(),
		"size": info.Size(),
		"modTime": info.ModTime(),
		// TODO: Add duration/metadata extraction if needed
	})
}
