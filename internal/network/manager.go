// Package network provides robust networking capabilities for the platform
package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nathfavour/noplacelike.go/internal/core"
)

// NetworkManager implements distributed networking capabilities
type NetworkManager struct {
	mu       sync.RWMutex
	config   NetworkConfig
	security core.SecurityManager
	eventBus core.EventBus
	logger   core.Logger

	// Peer management
	peers     map[string]*core.Peer
	localPeer *core.Peer

	// Network services
	server          *http.Server
	discoveryServer *DiscoveryServer

	// Communication channels
	channels        map[string]core.SecureChannel
	messageHandlers map[string]MessageHandler

	// State
	started bool
}

// NetworkConfig contains network configuration
type NetworkConfig struct {
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	EnableDiscovery   bool          `json:"enableDiscovery"`
	DiscoveryPort     int           `json:"discoveryPort"`
	DiscoveryInterval time.Duration `json:"discoveryInterval"`
	MaxPeers          int           `json:"maxPeers"`
	Timeout           time.Duration `json:"timeout"`
	KeepAliveInterval time.Duration `json:"keepAliveInterval"`
	EnableTLS         bool          `json:"enableTLS"`
	TLSCertFile       string        `json:"tlsCertFile"`
	TLSKeyFile        string        `json:"tlsKeyFile"`
}

// MessageHandler processes incoming messages
type MessageHandler func(ctx context.Context, message core.Message) error

// DiscoveryServer handles peer discovery
type DiscoveryServer struct {
	port     int
	interval time.Duration
	peers    map[string]*core.Peer
	mu       sync.RWMutex
}

// SecureChannelImpl implements encrypted communication
type SecureChannelImpl struct {
	conn     *websocket.Conn
	peerID   string
	security core.SecurityManager
	mu       sync.Mutex
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(config NetworkConfig, security core.SecurityManager, eventBus core.EventBus, logger core.Logger) (*NetworkManager, error) {
	nm := &NetworkManager{
		config:          config,
		security:        security,
		eventBus:        eventBus,
		logger:          logger,
		peers:           make(map[string]*core.Peer),
		channels:        make(map[string]core.SecureChannel),
		messageHandlers: make(map[string]MessageHandler),
	}

	// Create local peer identity
	if err := nm.initializeLocalPeer(); err != nil {
		return nil, fmt.Errorf("failed to initialize local peer: %w", err)
	}

	// Initialize discovery server if enabled
	if config.EnableDiscovery {
		nm.discoveryServer = &DiscoveryServer{
			port:     config.DiscoveryPort,
			interval: config.DiscoveryInterval,
			peers:    make(map[string]*core.Peer),
		}
	}

	return nm, nil
}

// DiscoverPeers finds other instances on the network
func (nm *NetworkManager) DiscoverPeers(ctx context.Context) ([]core.Peer, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if !nm.config.EnableDiscovery {
		return []core.Peer{}, nil
	}

	nm.logger.Info("Starting peer discovery")

	// Start discovery server
	if err := nm.startDiscoveryServer(ctx); err != nil {
		return nil, fmt.Errorf("failed to start discovery server: %w", err)
	}

	// Broadcast discovery request
	peers, err := nm.broadcastDiscovery(ctx)
	if err != nil {
		nm.logger.Warn("Discovery broadcast failed", core.Field{Key: "error", Value: err})
	}

	// Add discovered peers
	for _, peer := range peers {
		nm.addPeer(&peer)
	}

	result := make([]core.Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		result = append(result, *peer)
	}

	return result, nil
}

// RegisterPeer adds a peer to the network
func (nm *NetworkManager) RegisterPeer(peer core.Peer) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if len(nm.peers) >= nm.config.MaxPeers {
		return fmt.Errorf("maximum peers (%d) reached", nm.config.MaxPeers)
	}

	nm.addPeer(&peer)
	return nil
}

// GetPeers returns all known peers
func (nm *NetworkManager) GetPeers() []core.Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	peers := make([]core.Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, *peer)
	}

	return peers
}

// SendMessage sends a message to a specific peer
func (nm *NetworkManager) SendMessage(ctx context.Context, peerID string, message core.Message) error {
	nm.mu.RLock()
	peer, exists := nm.peers[peerID]
	nm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}

	// Get or create secure channel
	channel, err := nm.getOrCreateChannel(ctx, peerID)
	if err != nil {
		return fmt.Errorf("failed to get channel for peer %s: %w", peerID, err)
	}

	// Serialize message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Send message
	if err := channel.Send(data); err != nil {
		return fmt.Errorf("failed to send message to peer %s: %w", peerID, err)
	}

	nm.logger.Debug("Message sent",
		core.Field{Key: "peer", Value: peerID},
		core.Field{Key: "messageType", Value: message.Type},
	)

	// Update peer last seen
	peer.LastSeen = time.Now()

	return nil
}

// BroadcastMessage sends a message to all peers
func (nm *NetworkManager) BroadcastMessage(ctx context.Context, message core.Message) error {
	nm.mu.RLock()
	peers := make([]*core.Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, peer)
	}
	nm.mu.RUnlock()

	errors := make([]error, 0)

	for _, peer := range peers {
		if err := nm.SendMessage(ctx, peer.ID, message); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to peer %s: %w", peer.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d peers", len(errors))
	}

	nm.logger.Info("Message broadcasted",
		core.Field{Key: "peers", Value: len(peers)},
		core.Field{Key: "messageType", Value: message.Type},
	)

	return nil
}

// CreateSecureChannel establishes an encrypted connection
func (nm *NetworkManager) CreateSecureChannel(ctx context.Context, peerID string) (core.SecureChannel, error) {
	nm.mu.RLock()
	peer, exists := nm.peers[peerID]
	nm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	// Create WebSocket connection
	addr := fmt.Sprintf("ws://%s:%d/ws", peer.Address, peer.Port)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %w", peerID, err)
	}

	channel := &SecureChannelImpl{
		conn:     conn,
		peerID:   peerID,
		security: nm.security,
	}

	nm.mu.Lock()
	nm.channels[peerID] = channel
	nm.mu.Unlock()

	nm.logger.Info("Secure channel established", core.Field{Key: "peer", Value: peerID})

	return channel, nil
}

// RegisterMessageHandler registers a handler for a message type
func (nm *NetworkManager) RegisterMessageHandler(messageType string, handler MessageHandler) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.messageHandlers[messageType] = handler
	nm.logger.Debug("Message handler registered", core.Field{Key: "type", Value: messageType})
}

// Start begins network operations
func (nm *NetworkManager) Start(ctx context.Context) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.started {
		return fmt.Errorf("network manager already started")
	}

	// Start HTTP server for peer communications
	if err := nm.startHTTPServer(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Start keep-alive routine
	go nm.keepAliveRoutine(ctx)

	nm.started = true
	nm.logger.Info("Network manager started",
		core.Field{Key: "host", Value: nm.config.Host},
		core.Field{Key: "port", Value: nm.config.Port},
	)

	return nil
}

// Stop gracefully shuts down network operations
func (nm *NetworkManager) Stop(ctx context.Context) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if !nm.started {
		return fmt.Errorf("network manager not started")
	}

	// Close all channels
	for peerID, channel := range nm.channels {
		if err := channel.Close(); err != nil {
			nm.logger.Warn("Failed to close channel",
				core.Field{Key: "peer", Value: peerID},
				core.Field{Key: "error", Value: err},
			)
		}
	}

	// Stop HTTP server
	if nm.server != nil {
		if err := nm.server.Shutdown(ctx); err != nil {
			nm.logger.Warn("Failed to shutdown HTTP server", core.Field{Key: "error", Value: err})
		}
	}

	nm.started = false
	nm.logger.Info("Network manager stopped")

	return nil
}

// Private methods
func (nm *NetworkManager) initializeLocalPeer() error {
	hostname, err := getHostname()
	if err != nil {
		hostname = "unknown"
	}

	nm.localPeer = &core.Peer{
		ID:           generatePeerID(),
		Name:         hostname,
		Address:      nm.config.Host,
		Port:         nm.config.Port,
		Version:      "1.0.0",
		Capabilities: []string{"file-sharing", "clipboard", "messaging"},
		LastSeen:     time.Now(),
		Metadata: map[string]string{
			"platform": "noplacelike-go",
			"hostname": hostname,
		},
	}

	return nil
}

func (nm *NetworkManager) addPeer(peer *core.Peer) {
	existing, exists := nm.peers[peer.ID]
	if exists {
		// Update existing peer
		existing.LastSeen = time.Now()
		existing.Address = peer.Address
		existing.Port = peer.Port
	} else {
		// Add new peer
		nm.peers[peer.ID] = peer

		// Publish peer joined event
		event := core.Event{
			ID:        generateID(),
			Type:      core.EventPeerJoined,
			Source:    "network",
			Data:      *peer,
			Timestamp: time.Now(),
		}

		if err := nm.eventBus.Publish(context.Background(), "network", event); err != nil {
			nm.logger.Warn("Failed to publish peer joined event", core.Field{Key: "error", Value: err})
		}

		nm.logger.Info("Peer added",
			core.Field{Key: "peerID", Value: peer.ID},
			core.Field{Key: "address", Value: peer.Address},
		)
	}
}

func (nm *NetworkManager) getOrCreateChannel(ctx context.Context, peerID string) (core.SecureChannel, error) {
	nm.mu.RLock()
	channel, exists := nm.channels[peerID]
	nm.mu.RUnlock()

	if exists {
		return channel, nil
	}

	return nm.CreateSecureChannel(ctx, peerID)
}

func (nm *NetworkManager) startHTTPServer(ctx context.Context) error {
	mux := http.NewServeMux()

	// WebSocket endpoint for peer communication
	mux.HandleFunc("/ws", nm.handleWebSocket)

	// Peer discovery endpoint
	mux.HandleFunc("/discover", nm.handleDiscovery)

	// Peer information endpoint
	mux.HandleFunc("/peer", nm.handlePeerInfo)

	addr := fmt.Sprintf("%s:%d", nm.config.Host, nm.config.Port)
	nm.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		var err error
		if nm.config.EnableTLS {
			err = nm.server.ListenAndServeTLS(nm.config.TLSCertFile, nm.config.TLSKeyFile)
		} else {
			err = nm.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			nm.logger.Error("HTTP server error", core.Field{Key: "error", Value: err})
		}
	}()

	return nil
}

func (nm *NetworkManager) startDiscoveryServer(ctx context.Context) error {
	if nm.discoveryServer == nil {
		return nil
	}

	// Start UDP discovery server
	go func() {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", nm.discoveryServer.port))
		if err != nil {
			nm.logger.Error("Failed to resolve UDP address", core.Field{Key: "error", Value: err})
			return
		}

		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			nm.logger.Error("Failed to listen on UDP", core.Field{Key: "error", Value: err})
			return
		}
		defer conn.Close()

		nm.logger.Info("Discovery server started", core.Field{Key: "port", Value: nm.discoveryServer.port})

		buffer := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, addr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					continue
				}

				// Handle discovery request
				go nm.handleDiscoveryRequest(conn, addr, buffer[:n])
			}
		}
	}()

	return nil
}

func (nm *NetworkManager) broadcastDiscovery(ctx context.Context) ([]core.Peer, error) {
	// Broadcast UDP discovery message
	conn, err := net.Dial("udp", fmt.Sprintf("255.255.255.255:%d", nm.config.DiscoveryPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	message := map[string]interface{}{
		"type": "discovery",
		"peer": nm.localPeer,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	// Wait for responses (simplified implementation)
	time.Sleep(time.Second * 2)

	nm.discoveryServer.mu.RLock()
	defer nm.discoveryServer.mu.RUnlock()

	peers := make([]core.Peer, 0, len(nm.discoveryServer.peers))
	for _, peer := range nm.discoveryServer.peers {
		peers = append(peers, *peer)
	}

	return peers, nil
}

func (nm *NetworkManager) keepAliveRoutine(ctx context.Context) {
	ticker := time.NewTicker(nm.config.KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nm.performKeepAlive(ctx)
		}
	}
}

func (nm *NetworkManager) performKeepAlive(ctx context.Context) {
	nm.mu.RLock()
	peers := make([]*core.Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, peer)
	}
	nm.mu.RUnlock()

	// Remove stale peers
	staleThreshold := time.Now().Add(-nm.config.KeepAliveInterval * 3)

	for _, peer := range peers {
		if peer.LastSeen.Before(staleThreshold) {
			nm.removePeer(peer.ID)
		}
	}
}

func (nm *NetworkManager) removePeer(peerID string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	peer, exists := nm.peers[peerID]
	if !exists {
		return
	}

	// Close channel if exists
	if channel, exists := nm.channels[peerID]; exists {
		channel.Close()
		delete(nm.channels, peerID)
	}

	delete(nm.peers, peerID)

	// Publish peer left event
	event := core.Event{
		ID:        generateID(),
		Type:      core.EventPeerLeft,
		Source:    "network",
		Data:      *peer,
		Timestamp: time.Now(),
	}

	if err := nm.eventBus.Publish(context.Background(), "network", event); err != nil {
		nm.logger.Warn("Failed to publish peer left event", core.Field{Key: "error", Value: err})
	}

	nm.logger.Info("Peer removed", core.Field{Key: "peerID", Value: peerID})
}

// HTTP handlers
func (nm *NetworkManager) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		nm.logger.Error("Failed to upgrade WebSocket", core.Field{Key: "error", Value: err})
		return
	}
	defer conn.Close()

	// Handle WebSocket messages
	for {
		var message core.Message
		if err := conn.ReadJSON(&message); err != nil {
			break
		}

		// Process message
		go nm.processMessage(r.Context(), message)
	}
}

func (nm *NetworkManager) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nm.localPeer)
}

func (nm *NetworkManager) handlePeerInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"peer":  nm.localPeer,
		"peers": nm.GetPeers(),
	})
}

func (nm *NetworkManager) handleDiscoveryRequest(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	var request map[string]interface{}
	if err := json.Unmarshal(data, &request); err != nil {
		return
	}

	if request["type"] == "discovery" {
		// Respond with our peer info
		response := map[string]interface{}{
			"type": "discovery_response",
			"peer": nm.localPeer,
		}

		responseData, err := json.Marshal(response)
		if err != nil {
			return
		}

		conn.WriteToUDP(responseData, addr)
	}
}

func (nm *NetworkManager) processMessage(ctx context.Context, message core.Message) {
	nm.mu.RLock()
	handler, exists := nm.messageHandlers[message.Type]
	nm.mu.RUnlock()

	if exists {
		if err := handler(ctx, message); err != nil {
			nm.logger.Error("Message handler error",
				core.Field{Key: "messageType", Value: message.Type},
				core.Field{Key: "error", Value: err},
			)
		}
	} else {
		nm.logger.Debug("No handler for message type", core.Field{Key: "type", Value: message.Type})
	}
}

// SecureChannelImpl methods
func (c *SecureChannelImpl) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Encrypt data if security manager is available
	if c.security != nil {
		encrypted, err := c.security.Encrypt(data, c.peerID)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		data = encrypted
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *SecureChannelImpl) Receive() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// Decrypt data if security manager is available
	if c.security != nil {
		decrypted, err := c.security.Decrypt(data, c.peerID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt data: %w", err)
		}
		data = decrypted
	}

	return data, nil
}

func (c *SecureChannelImpl) Close() error {
	return c.conn.Close()
}

// Helper functions
func generatePeerID() string {
	return fmt.Sprintf("peer-%d", time.Now().UnixNano())
}

func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

func getHostname() (string, error) {
	// This would get the actual hostname
	return "localhost", nil
}
