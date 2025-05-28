// Package core defines the foundational interfaces and types for the NoPlaceLike platform
package core

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Plugin represents a loadable plugin that extends platform functionality
type Plugin interface {
	// Name returns the unique identifier for this plugin
	Name() string

	// Version returns the plugin version
	Version() string

	// Initialize sets up the plugin with the given configuration
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Start begins plugin execution
	Start(ctx context.Context) error

	// Stop gracefully shuts down the plugin
	Stop(ctx context.Context) error

	// Health returns the current health status
	Health() HealthStatus

	// Routes returns HTTP routes this plugin provides
	Routes() []Route

	// Dependencies returns other plugins this plugin depends on
	Dependencies() []string
}

// ServiceManager manages the lifecycle of all platform services
type ServiceManager interface {
	// RegisterService adds a new service to the manager
	RegisterService(name string, service Service) error

	// GetService retrieves a service by name
	GetService(name string) (Service, error)

	// StartAll starts all registered services
	StartAll(ctx context.Context) error

	// StopAll gracefully stops all services
	StopAll(ctx context.Context) error

	// HealthCheck returns the health status of all services
	HealthCheck() map[string]HealthStatus
}

// Service represents a core platform service
type Service interface {
	// Name returns the service identifier
	Name() string

	// Start begins service execution
	Start(ctx context.Context) error

	// Stop gracefully shuts down the service
	Stop(ctx context.Context) error

	// Health returns current service health
	Health() HealthStatus

	// Configuration returns service config schema
	Configuration() ConfigSchema
}

// NetworkManager handles all network-related operations
type NetworkManager interface {
	// DiscoverPeers finds other NoPlaceLike instances on the network
	DiscoverPeers(ctx context.Context) ([]Peer, error)

	// RegisterPeer adds a new peer to the network
	RegisterPeer(peer Peer) error

	// GetPeers returns all known peers
	GetPeers() []Peer

	// SendMessage sends a message to a specific peer
	SendMessage(ctx context.Context, peerID string, message Message) error

	// BroadcastMessage sends a message to all peers
	BroadcastMessage(ctx context.Context, message Message) error

	// CreateSecureChannel establishes an encrypted connection with a peer
	CreateSecureChannel(ctx context.Context, peerID string) (SecureChannel, error)
}

// ResourceManager handles shared resources across the network
type ResourceManager interface {
	// RegisterResource makes a local resource available to the network
	RegisterResource(resource Resource) error

	// GetResource retrieves a resource by ID
	GetResource(ctx context.Context, resourceID string) (Resource, error)

	// ListResources returns all available resources
	ListResources(ctx context.Context, filter ResourceFilter) ([]Resource, error)

	// StreamResource provides streaming access to a resource
	StreamResource(ctx context.Context, resourceID string) (io.ReadCloser, error)

	// SubscribeToUpdates notifies when resources change
	SubscribeToUpdates(ctx context.Context) (<-chan ResourceEvent, error)
}

// SecurityManager handles authentication, authorization, and encryption
type SecurityManager interface {
	// Authenticate verifies peer identity
	Authenticate(ctx context.Context, credentials Credentials) (AuthResult, error)

	// Authorize checks if a peer can access a resource
	Authorize(ctx context.Context, peerID string, resource string, action string) (bool, error)

	// Encrypt encrypts data for transmission
	Encrypt(data []byte, recipientID string) ([]byte, error)

	// Decrypt decrypts received data
	Decrypt(data []byte, senderID string) ([]byte, error)

	// GenerateToken creates an access token
	GenerateToken(ctx context.Context, peerID string, permissions []Permission) (string, error)

	// ValidateToken verifies an access token
	ValidateToken(ctx context.Context, token string) (TokenInfo, error)
}

// EventBus handles inter-service communication
type EventBus interface {
	// Publish sends an event to all subscribers
	Publish(ctx context.Context, topic string, event Event) error

	// Subscribe registers a handler for events on a topic
	Subscribe(ctx context.Context, topic string, handler EventHandler) error

	// Unsubscribe removes a handler from a topic
	Unsubscribe(ctx context.Context, topic string, handler EventHandler) error
}

// ConfigManager handles configuration across the platform
type ConfigManager interface {
	// Get retrieves a configuration value
	Get(key string) (interface{}, error)

	// Set updates a configuration value
	Set(key string, value interface{}) error

	// Watch monitors configuration changes
	Watch(ctx context.Context, key string) (<-chan ConfigChange, error)

	// Validate checks configuration against schema
	Validate(config map[string]interface{}, schema ConfigSchema) error

	// Reload reloads configuration from storage
	Reload(ctx context.Context) error
}

// MetricsCollector gathers performance and usage metrics
type MetricsCollector interface {
	// Counter increments a counter metric
	Counter(name string, tags map[string]string) error

	// Gauge sets a gauge metric value
	Gauge(name string, value float64, tags map[string]string) error

	// Histogram records a histogram value
	Histogram(name string, value float64, tags map[string]string) error

	// Timer records timing information
	Timer(name string, duration time.Duration, tags map[string]string) error

	// Export returns metrics in the specified format
	Export(format string) ([]byte, error)
}

// Logger provides structured logging across the platform
type Logger interface {
	// Debug logs debug-level information
	Debug(msg string, fields ...Field)

	// Info logs info-level information
	Info(msg string, fields ...Field)

	// Warn logs warning-level information
	Warn(msg string, fields ...Field)

	// Error logs error-level information
	Error(msg string, fields ...Field)

	// Fatal logs fatal-level information and exits
	Fatal(msg string, fields ...Field)

	// WithFields returns a logger with additional fields
	WithFields(fields ...Field) Logger

	// WithContext returns a logger with context
	WithContext(ctx context.Context) Logger
}

// Common types and structures
type (
	// HealthStatus represents the health of a component
	HealthStatus struct {
		Status    string                 `json:"status"`
		Timestamp time.Time              `json:"timestamp"`
		Details   map[string]interface{} `json:"details,omitempty"`
		Error     string                 `json:"error,omitempty"`
	}

	// Route represents an HTTP route provided by a plugin
	Route struct {
		Method     string           `json:"method"`
		Path       string           `json:"path"`
		Handler    http.HandlerFunc `json:"-"`
		Middleware []MiddlewareFunc `json:"-"`
		Auth       AuthRequirement  `json:"auth"`
	}

	// Peer represents another NoPlaceLike instance
	Peer struct {
		ID           string            `json:"id"`
		Name         string            `json:"name"`
		Address      string            `json:"address"`
		Port         int               `json:"port"`
		Version      string            `json:"version"`
		Capabilities []string          `json:"capabilities"`
		LastSeen     time.Time         `json:"lastSeen"`
		Metadata     map[string]string `json:"metadata"`
	}

	// Resource represents a shared resource
	Resource struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Type        string            `json:"type"`
		Size        int64             `json:"size"`
		Owner       string            `json:"owner"`
		Permissions []Permission      `json:"permissions"`
		Metadata    map[string]string `json:"metadata"`
		CreatedAt   time.Time         `json:"createdAt"`
		UpdatedAt   time.Time         `json:"updatedAt"`
	}

	// Message represents inter-peer communication
	Message struct {
		ID        string            `json:"id"`
		Type      string            `json:"type"`
		From      string            `json:"from"`
		To        string            `json:"to"`
		Payload   []byte            `json:"payload"`
		Metadata  map[string]string `json:"metadata"`
		Timestamp time.Time         `json:"timestamp"`
	}

	// Event represents system events
	Event struct {
		ID        string            `json:"id"`
		Type      string            `json:"type"`
		Source    string            `json:"source"`
		Data      interface{}       `json:"data"`
		Metadata  map[string]string `json:"metadata"`
		Timestamp time.Time         `json:"timestamp"`
	}

	// ConfigSchema defines configuration structure
	ConfigSchema struct {
		Properties map[string]PropertySchema `json:"properties"`
		Required   []string                  `json:"required"`
	}

	// PropertySchema defines a configuration property
	PropertySchema struct {
		Type        string      `json:"type"`
		Description string      `json:"description"`
		Default     interface{} `json:"default"`
		Validation  []Validator `json:"validation"`
	}

	// Permission represents access permissions
	Permission struct {
		Resource string   `json:"resource"`
		Actions  []string `json:"actions"`
	}

	// Credentials for authentication
	Credentials struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	// AuthResult contains authentication results
	AuthResult struct {
		Success     bool              `json:"success"`
		PeerID      string            `json:"peerId"`
		Permissions []Permission      `json:"permissions"`
		Token       string            `json:"token"`
		ExpiresAt   time.Time         `json:"expiresAt"`
		Metadata    map[string]string `json:"metadata"`
	}

	// TokenInfo contains token validation results
	TokenInfo struct {
		Valid       bool              `json:"valid"`
		PeerID      string            `json:"peerId"`
		Permissions []Permission      `json:"permissions"`
		ExpiresAt   time.Time         `json:"expiresAt"`
		Metadata    map[string]string `json:"metadata"`
	}

	// Field represents a structured log field
	Field struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}

	// ResourceFilter for filtering resources
	ResourceFilter struct {
		Type    string            `json:"type"`
		Owner   string            `json:"owner"`
		Tags    map[string]string `json:"tags"`
		MinSize int64             `json:"minSize"`
		MaxSize int64             `json:"maxSize"`
	}

	// ResourceEvent represents resource changes
	ResourceEvent struct {
		Type      string    `json:"type"` // created, updated, deleted
		Resource  Resource  `json:"resource"`
		Timestamp time.Time `json:"timestamp"`
	}

	// ConfigChange represents configuration updates
	ConfigChange struct {
		Key       string      `json:"key"`
		OldValue  interface{} `json:"oldValue"`
		NewValue  interface{} `json:"newValue"`
		Timestamp time.Time   `json:"timestamp"`
	}

	// SecureChannel represents an encrypted communication channel
	SecureChannel interface {
		Send(data []byte) error
		Receive() ([]byte, error)
		Close() error
	}

	// EventHandler processes events
	EventHandler func(ctx context.Context, event Event) error

	// MiddlewareFunc for HTTP middleware
	MiddlewareFunc func(http.Handler) http.Handler

	// AuthRequirement specifies authentication needs
	AuthRequirement struct {
		Required    bool     `json:"required"`
		Permissions []string `json:"permissions"`
	}

	// Validator for configuration validation
	Validator interface {
		Validate(value interface{}) error
	}
)

// Standard event types
const (
	EventPeerJoined      = "peer.joined"
	EventPeerLeft        = "peer.left"
	EventResourceAdded   = "resource.added"
	EventResourceUpdated = "resource.updated"
	EventResourceRemoved = "resource.removed"
	EventConfigChanged   = "config.changed"
	EventServiceStarted  = "service.started"
	EventServiceStopped  = "service.stopped"
	EventPluginLoaded    = "plugin.loaded"
	EventPluginUnloaded  = "plugin.unloaded"
)

// Standard health statuses
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)

// Standard resource types
const (
	ResourceTypeFile      = "file"
	ResourceTypeDirectory = "directory"
	ResourceTypeClipboard = "clipboard"
	ResourceTypeAudio     = "audio"
	ResourceTypeVideo     = "video"
	ResourceTypeScreen    = "screen"
	ResourceTypeProcess   = "process"
	ResourceTypeService   = "service"
)
