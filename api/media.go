package api

import (
	// "errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	if !m.config.EnableAudioStreaming {
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
	if !m.config.EnableScreenStreaming {
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
	Path        string   `json:"path"`
	AudioCount  int      `json:"audioCount"`
	TotalCount  int      `json:"totalCount"`
	Ratio       float64  `json:"ratio"`
	SampleFiles []string `json:"sampleFiles"`
}

// ScanMediaDirectories scans allowed paths for media-rich directories
func (m *MediaAPI) ScanMediaDirectories(c *gin.Context) {
	var results []MediaDirInfo
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true}
	visited := make(map[string]bool)
	for _, base := range m.config.AllowedPaths {
		_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() {
				return nil
			}
			if visited[path] {
				return nil
			}
			visited[path] = true
			files, _ := os.ReadDir(path)
			total, audio := 0, 0
			var samples []string
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				total++
				ext := filepath.Ext(f.Name())
				if audioExts[ext] {
					audio++
					if len(samples) < 3 {
						samples = append(samples, f.Name())
					}
				}
			}
			if total > 0 && float64(audio)/float64(total) > 0.5 && audio >= 3 {
				results = append(results, MediaDirInfo{
					Path: path, AudioCount: audio, TotalCount: total, Ratio: float64(audio) / float64(total), SampleFiles: samples,
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

// StreamAudioFile streams a specific audio file to the client (robust HTTP streaming)
func (m *MediaAPI) StreamAudioFile(c *gin.Context) {
	file := c.Query("file")
	if file == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
		return
	}
	// Security: Only allow files in allowed paths
	allowed := false
	for _, base := range m.config.AllowedPaths {
		if isSubPath(file, base) {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
		return
	}
	// Check file exists and is audio
	info, err := os.Stat(file)
	if err != nil || info.IsDir() {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	ext := filepath.Ext(file)
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true}
	if !audioExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not an audio file"})
		return
	}
	// Set headers for streaming
	c.Header("Content-Type", getAudioMimeType(ext))
	c.Header("Content-Disposition", "inline; filename="+filepath.Base(file))
	c.File(file)
}

// getAudioMimeType returns the MIME type for a given audio file extension
func getAudioMimeType(ext string) string {
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	default:
		return "application/octet-stream"
	}
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
		"name":    info.Name(),
		"size":    info.Size(),
		"modTime": info.ModTime(),
		// TODO: Add duration/metadata extraction if needed
	})
}

// LiveAudioHub manages live audio WebSocket clients
var liveAudioClients = make(map[*websocket.Conn]bool)
var liveAudioBroadcast = make(chan []byte, 1024)

// StartLiveAudioBroadcaster starts a goroutine to broadcast audio to all clients
func StartLiveAudioBroadcaster() {
	go func() {
		for data := range liveAudioBroadcast {
			for client := range liveAudioClients {
				if err := client.WriteMessage(websocket.BinaryMessage, data); err != nil {
					client.Close()
					delete(liveAudioClients, client)
				}
			}
		}
	}()
}

// LiveAudioWebSocket streams live audio to clients via WebSocket
func (m *MediaAPI) LiveAudioWebSocket(c *gin.Context) {
	conn, err := m.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection: " + err.Error()})
		return
	}
	defer conn.Close()
	liveAudioClients[conn] = true
	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(liveAudioClients, conn)
			break
		}
	}
}

// Mock/placeholder: StartLiveAudioCapture simulates capturing system audio and broadcasting it
func StartLiveAudioCapture() {
	go func() {
		// TODO: Replace this with actual system audio capture (e.g., using go-portaudio, ffmpeg, or platform-specific tools)
		// For now, send silence (or a sine wave) as PCM/Opus/MP3 data every 20ms
		for {
			// Example: send 20ms of silence (44100Hz, 16bit, mono = 1764 bytes for 20ms)
			// Replace with actual audio data in production
			data := make([]byte, 1764)
			liveAudioBroadcast <- data
			time.Sleep(20 * time.Millisecond)
		}
	}()
}

// LiveAudioPage serves a simple HTML page that plays the live audio
func LiveAudioPage(c *gin.Context) {
	html := `<!DOCTYPE html>
<html><head><title>Live Audio</title></head><body>
<h2>Live Audio Stream</h2>
<audio id="audio" controls autoplay></audio>
<script>
const audio = document.getElementById('audio');
const ws = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/api/v1/live/audio');
let ctx, source, queue = [];
ws.binaryType = 'arraybuffer';
ws.onmessage = function(e) {
    if (!ctx) {
        ctx = new (window.AudioContext || window.webkitAudioContext)();
        source = ctx.createBufferSource();
    }
    // For real PCM/Opus/MP3, decode and play here. For now, just ignore silence.
    // Example: decode as PCM and play (requires actual PCM data)
    // let buf = e.data; ...
};
ws.onclose = function() { audio.pause(); };
</script>
</body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
