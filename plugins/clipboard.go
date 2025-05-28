package plugins

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nathfavour/noplacelike.go/internal/core"
	"github.com/nathfavour/noplacelike.go/internal/logger"
)

// ClipboardPlugin provides cross-device clipboard sharing
type ClipboardPlugin struct {
	id         string
	version    string
	logger     logger.Logger
	platform   core.PlatformAPI
	config     ClipboardConfig
	clipboard  ClipboardData
	history    []ClipboardEntry
	mu         sync.RWMutex
	running    bool
	maxHistory int
}

type ClipboardConfig struct {
	MaxContentSize int  `json:"maxContentSize"`
	EnableHistory  bool `json:"enableHistory"`
	MaxHistory     int  `json:"maxHistory"`
	EnableCORS     bool `json:"enableCors"`
}

type ClipboardData struct {
	Content   string `json:"content"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	UpdatedAt int64  `json:"updatedAt"`
	Hash      string `json:"hash"`
}

type ClipboardEntry struct {
	ClipboardData
	ID        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
}

// NewClipboardPlugin creates a new clipboard plugin
func NewClipboardPlugin() core.Plugin {
	return &ClipboardPlugin{
		id:      "clipboard",
		version: "1.0.0",
		config: ClipboardConfig{
			MaxContentSize: 1024 * 1024, // 1MB
			EnableHistory:  true,
			MaxHistory:     50,
			EnableCORS:     true,
		},
		history:    make([]ClipboardEntry, 0),
		maxHistory: 50,
	}
}

// Plugin interface implementation
func (p *ClipboardPlugin) ID() string {
	return p.id
}

func (p *ClipboardPlugin) Version() string {
	return p.version
}

func (p *ClipboardPlugin) Dependencies() []string {
	return []string{}
}

func (p *ClipboardPlugin) Name() string {
	return "Clipboard Sharing Plugin"
}

func (p *ClipboardPlugin) Initialize(platform core.PlatformAPI) error {
	p.platform = platform
	p.logger = platform.GetLogger().WithFields(map[string]interface{}{
		"plugin": p.id,
	})

	p.logger.Info("Clipboard plugin initialized")
	return nil
}

func (p *ClipboardPlugin) Configure(config map[string]interface{}) error {
	if configBytes, err := json.Marshal(config); err == nil {
		if err := json.Unmarshal(configBytes, &p.config); err != nil {
			p.logger.Warn("Failed to parse configuration", "error", err)
		}
	}

	p.maxHistory = p.config.MaxHistory
	p.logger.Info("Clipboard plugin configured", "config", p.config)
	return nil
}

func (p *ClipboardPlugin) Start(ctx context.Context) error {
	p.running = true
	p.logger.Info("Clipboard plugin started")

	// Register as a resource provider
	if resourceMgr := p.platform.GetResourceManager(); resourceMgr != nil {
		resource := core.Resource{
			ID:          p.id,
			Type:        "clipboard",
			Name:        "Clipboard Sharing",
			Description: "Provides cross-device clipboard sharing with history",
			Provider:    p.id,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		resourceMgr.RegisterResource(resource)
	}

	// Subscribe to network events for clipboard sync
	if eventBus := p.platform.GetEventBus(); eventBus != nil {
		eventBus.Subscribe("clipboard.sync", p.handleSyncEvent)
		eventBus.Subscribe("peer.connected", p.handlePeerConnected)
	}

	return nil
}

func (p *ClipboardPlugin) Stop(ctx context.Context) error {
	p.running = false

	// Unregister resource
	if resourceMgr := p.platform.GetResourceManager(); resourceMgr != nil {
		resourceMgr.UnregisterResource(p.id)
	}

	// Unsubscribe from events
	if eventBus := p.platform.GetEventBus(); eventBus != nil {
		eventBus.Unsubscribe("clipboard.sync", p.handleSyncEvent)
		eventBus.Unsubscribe("peer.connected", p.handlePeerConnected)
	}

	p.logger.Info("Clipboard plugin stopped")
	return nil
}

func (p *ClipboardPlugin) IsHealthy() bool {
	return p.running
}

func (p *ClipboardPlugin) Routes() []core.Route {
	return []core.Route{
		{
			Method:      "GET",
			Path:        "/plugins/clipboard/clipboard",
			Handler:     p.handleGetClipboard,
			Description: "Get current clipboard content",
		},
		{
			Method:      "POST",
			Path:        "/plugins/clipboard/clipboard",
			Handler:     p.handleSetClipboard,
			Description: "Set clipboard content",
		},
		{
			Method:      "GET",
			Path:        "/plugins/clipboard/history",
			Handler:     p.handleGetHistory,
			Description: "Get clipboard history",
		},
		{
			Method:      "DELETE",
			Path:        "/plugins/clipboard/history",
			Handler:     p.handleClearHistory,
			Description: "Clear clipboard history",
		},
		{
			Method:      "GET",
			Path:        "/plugins/clipboard/history/:id",
			Handler:     p.handleGetHistoryEntry,
			Description: "Get specific history entry",
		},
		{
			Method:      "POST",
			Path:        "/plugins/clipboard/sync",
			Handler:     p.handleSyncClipboard,
			Description: "Sync clipboard across devices",
		},
	}
}

func (p *ClipboardPlugin) HandleEvent(event core.Event) error {
	switch event.Type {
	case "clipboard.changed":
		p.logger.Debug("Clipboard changed", "source", event.Source)
	case "peer.connected":
		// Sync clipboard when new peer connects
		go p.syncToNewPeer(event.Data)
	}
	return nil
}

// HTTP handlers
func (p *ClipboardPlugin) handleGetClipboard(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	p.mu.RLock()
	clipboard := p.clipboard
	p.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clipboard)
}

func (p *ClipboardPlugin) handleSetClipboard(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	if r.Method == "OPTIONS" {
		return
	}

	var request struct {
		Content string `json:"content"`
		Type    string `json:"type"`
		Source  string `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.logger.Error("Error decoding request", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate content size
	if len(request.Content) > p.config.MaxContentSize {
		http.Error(w, "Content too large", http.StatusRequestEntityTooLarge)
		return
	}

	// Set default values
	if request.Type == "" {
		request.Type = "text/plain"
	}
	if request.Source == "" {
		request.Source = "unknown"
	}

	// Update clipboard
	p.setClipboardContent(request.Content, request.Type, request.Source)

	// Broadcast to peers
	if eventBus := p.platform.GetEventBus(); eventBus != nil {
		event := core.Event{
			Type:   "clipboard.changed",
			Source: p.id,
			Data: map[string]interface{}{
				"content": request.Content,
				"type":    request.Type,
				"source":  request.Source,
			},
		}
		eventBus.Publish(event)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Clipboard updated successfully",
		"hash":    p.clipboard.Hash,
	})
}

func (p *ClipboardPlugin) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	if !p.config.EnableHistory {
		http.Error(w, "History is disabled", http.StatusNotFound)
		return
	}

	p.mu.RLock()
	history := make([]ClipboardEntry, len(p.history))
	copy(history, p.history)
	p.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

func (p *ClipboardPlugin) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	if !p.config.EnableHistory {
		http.Error(w, "History is disabled", http.StatusNotFound)
		return
	}

	p.mu.Lock()
	p.history = make([]ClipboardEntry, 0)
	p.mu.Unlock()

	p.logger.Info("Clipboard history cleared")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "History cleared successfully",
	})
}

func (p *ClipboardPlugin) handleGetHistoryEntry(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	if !p.config.EnableHistory {
		http.Error(w, "History is disabled", http.StatusNotFound)
		return
	}

	id := p.extractIDFromPath(r.URL.Path)
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	p.mu.RLock()
	var entry *ClipboardEntry
	for _, h := range p.history {
		if h.ID == id {
			entry = &h
			break
		}
	}
	p.mu.RUnlock()

	if entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (p *ClipboardPlugin) handleSyncClipboard(w http.ResponseWriter, r *http.Request) {
	if p.config.EnableCORS {
		p.setCORSHeaders(w)
	}

	// Trigger clipboard sync across all peers
	if networkMgr := p.platform.GetNetworkManager(); networkMgr != nil {
		peers := networkMgr.ListPeers()

		syncData := map[string]interface{}{
			"clipboard": p.clipboard,
			"action":    "sync_request",
		}

		syncMessage, _ := json.Marshal(syncData)

		for _, peer := range peers {
			if err := networkMgr.SendMessage(peer.ID, syncMessage); err != nil {
				p.logger.Error("Failed to sync to peer", "peer", peer.ID, "error", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Sync initiated",
		"peers":   "all",
	})
}

// Helper methods
func (p *ClipboardPlugin) setClipboardContent(content, contentType, source string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Generate content hash
	hash := fmt.Sprintf("%x", md5.Sum([]byte(content)))

	// Update clipboard
	p.clipboard = ClipboardData{
		Content:   content,
		Type:      contentType,
		Source:    source,
		UpdatedAt: time.Now().Unix(),
		Hash:      hash,
	}

	// Add to history if enabled and content is different
	if p.config.EnableHistory && (len(p.history) == 0 || p.history[0].Hash != hash) {
		entry := ClipboardEntry{
			ClipboardData: p.clipboard,
			ID:            fmt.Sprintf("clip_%d", time.Now().UnixNano()),
			CreatedAt:     time.Now().Unix(),
		}

		// Prepend to history
		p.history = append([]ClipboardEntry{entry}, p.history...)

		// Limit history size
		if len(p.history) > p.maxHistory {
			p.history = p.history[:p.maxHistory]
		}
	}

	p.logger.Info("Clipboard updated", "source", source, "type", contentType, "size", len(content))
}

func (p *ClipboardPlugin) handleSyncEvent(event core.Event) error {
	// Handle clipboard sync events from other instances
	if data, ok := event.Data["clipboard"].(map[string]interface{}); ok {
		content, _ := data["content"].(string)
		contentType, _ := data["type"].(string)
		source, _ := data["source"].(string)

		if content != "" {
			p.setClipboardContent(content, contentType, source)
		}
	}

	return nil
}

func (p *ClipboardPlugin) handlePeerConnected(event core.Event) error {
	// When a new peer connects, sync our current clipboard
	p.syncToNewPeer(event.Data)
	return nil
}

func (p *ClipboardPlugin) syncToNewPeer(peerData map[string]interface{}) {
	if networkMgr := p.platform.GetNetworkManager(); networkMgr != nil {
		peerID, _ := peerData["id"].(string)
		if peerID == "" {
			return
		}

		p.mu.RLock()
		clipboard := p.clipboard
		p.mu.RUnlock()

		syncData := map[string]interface{}{
			"clipboard": clipboard,
			"action":    "sync_response",
		}

		if syncMessage, err := json.Marshal(syncData); err == nil {
			networkMgr.SendMessage(peerID, syncMessage)
		}
	}
}

func (p *ClipboardPlugin) extractIDFromPath(urlPath string) string {
	// Extract ID from URL path like /plugins/clipboard/history/clip_123
	parts := strings.Split(urlPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func (p *ClipboardPlugin) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
