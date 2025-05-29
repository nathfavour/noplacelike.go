// Package core defines the foundational interfaces and types for the NoPlaceLike platform
package core

import (
	"context"
	"net/http"
	"time"

	"github.com/nathfavour/noplacelike.go/internal/logger"
)

// Service represents a platform service that can be started and stopped
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsHealthy() bool
	Name() string
	Health() HealthStatus
	Configuration() ConfigSchema
}

// Plugin represents a platform plugin
type Plugin interface {
	Service

	// Plugin metadata
	ID() string
	Version() string
	Dependencies() []string

	// Plugin lifecycle
	Initialize(platform PlatformAPI) error
	Configure(config map[string]interface{}) error

	// HTTP routes
	Routes() []Route

	// Event handling
	HandleEvent(event Event) error

	// Plugin health
	Health() HealthStatus
}

// PlatformAPI provides access to platform services for plugins
type PlatformAPI interface {
	GetLogger() logger.Logger
	GetEventBus() EventBus
	GetResourceManager() ResourceManager
	GetNetworkManager() NetworkManager
	GetSecurityManager() SecurityManager
	GetMetrics() MetricsCollector
	GetHealthChecker() HealthChecker
}

// Logger interface for structured logging - use logger.Logger instead
// Keeping this for backward compatibility
type Logger = logger.Logger

// EventBus handles event publishing and subscription
type EventBus interface {
	Service

	Publish(event Event) error
	PublishToTopic(ctx context.Context, topic string, event Event) error
	Subscribe(eventType string, handler EventHandler) error
	SubscribeWithContext(ctx context.Context, eventType string, handler func(context.Context, Event) error) error
	Unsubscribe(eventType string, handler EventHandler) error
	Configuration() ConfigSchema
}

// Field is a key-value pair for structured logging
// This is a stub for compatibility with platform.go
// Replace with your actual implementation as needed
type Field struct {
	Key   string
	Value interface{}
}

// ResourceManager manages platform resources
type ResourceManager interface {
	Service

	RegisterResource(resource Resource) error
	UnregisterResource(id string) error
	GetResource(ctx context.Context, id string) (Resource, error)
	ListResources(ctx context.Context, filter ResourceFilter) ([]Resource, error)
	StreamResource(ctx context.Context, id string) (ResourceStream, error)
	Configuration() ConfigSchema
}

// ResourceFilter for filtering resources
type ResourceFilter struct {
	Type  string `json:"type,omitempty"`
	Owner string `json:"owner,omitempty"`
}

// NetworkManager handles network operations and peer management
type NetworkManager interface {
	Service

	DiscoverPeers() ([]Peer, error)
	GetPeers() []Peer
	ConnectToPeer(address string) (Peer, error)
	ListPeers() []Peer
	SendMessage(peerID string, message []byte) error
	BroadcastMessage(message []byte) error
	Configuration() ConfigSchema
}

// SecurityManager handles authentication and authorization
type SecurityManager interface {
	Service

	Authenticate(token string) (*User, error)
	Authorize(user *User, resource string, action string) bool
	GenerateToken(user *User) (string, error)
	ValidatePermissions(userID string, permissions []string) bool
	ValidateToken(ctx context.Context, token string) (*TokenInfo, error)
	Configuration() ConfigSchema
}

// TokenInfo for authentication
type TokenInfo struct {
	Valid       bool         `json:"valid"`
	PeerID      string       `json:"peerId"`
	Permissions []Permission `json:"permissions"`
}

// Permission represents a user permission
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// MetricsCollector collects and exports metrics
type MetricsCollector interface {
	Service

	Counter(name string) Counter
	Gauge(name string) Gauge
	Histogram(name string) Histogram
	Timer(name string) Timer
	Export(format string) ([]byte, error)
	Configuration() ConfigSchema
}

// HealthChecker monitors component health
type HealthChecker interface {
	Service

	RegisterCheck(name string, check HealthCheck) error
	GetStatus() HealthStatus
	IsHealthy() bool
	Configuration() ConfigSchema
}

// HTTPService provides HTTP server functionality
type HTTPService interface {
	Service

	RegisterRoute(route Route) error
	RegisterMiddleware(middleware func(http.Handler) http.Handler)
	GetRouter() http.Handler
	Configuration() ConfigSchema
}

// PluginManager manages platform plugins
type PluginManager interface {
	Service

	LoadPlugin(name string) error
	UnloadPlugin(name string) error
	GetPlugin(name string) (Plugin, error)
	ListPlugins() []Plugin
	IsPluginLoaded(name string) bool
	Configuration() ConfigSchema
}

// ServiceManager stub
// Replace with your actual implementation as needed
type ServiceManager interface {
	StartAll(ctx context.Context) error
	StopAll(ctx context.Context) error
	HealthCheck() map[string]HealthStatus
	GetService(name string) (Service, error)
	Configuration() ConfigSchema
}

// ConfigManager stub
type ConfigManager interface{}

// Supporting types

// Route represents an HTTP route
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middleware  []func(http.Handler) http.Handler
	Auth        AuthRequirement
	Description string
}

// AuthRequirement specifies authentication requirements for a route
type AuthRequirement struct {
	Required    bool
	Permissions []string
	Roles       []string
}

// Event represents a platform event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// EventHandler handles events
type EventHandler func(event Event) error

// Resource represents a platform resource
type Resource struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Provider    string                 `json:"provider"`
	CreatedAt   int64                  `json:"createdAt"`
	UpdatedAt   int64                  `json:"updatedAt"`
}

// ResourceStream represents a streamable resource
type ResourceStream interface {
	Read(p []byte) (n int, err error)
	Close() error
	ContentType() string
	Size() int64
}

// Peer represents a network peer
type Peer struct {
	ID          string                 `json:"id"`
	Address     string                 `json:"address"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
	ConnectedAt int64                  `json:"connectedAt"`
	LastSeen    int64                  `json:"lastSeen"`
}

// User represents a platform user
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   int64             `json:"createdAt"`
	LastLogin   int64             `json:"lastLogin"`
}

// Metrics interfaces
type Counter interface {
	Inc()
	Add(delta float64)
	Get() float64
}

type Gauge interface {
	Set(value float64)
	Inc()
	Dec()
	Add(delta float64)
	Sub(delta float64)
	Get() float64
}

type Histogram interface {
	Observe(value float64)
	Reset()
}

type Timer interface {
	Start() TimerInstance
	Observe(duration float64)
}

type TimerInstance interface {
	Stop()
}

// Health check types
type HealthCheck func() error

// HealthStatus constants and struct
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusDegraded  = "degraded"
)

type ComponentHealth struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HealthStatus struct {
	Status    string                     `json:"status"`
	Timestamp time.Time                  `json:"timestamp"`
	Error     string                     `json:"error,omitempty"`
	Details   map[string]interface{}     `json:"details,omitempty"`
	Checks    map[string]ComponentHealth `json:"checks,omitempty"`
}

// ConfigSchema and PropertySchema for service configuration
type ConfigSchema struct {
	Properties map[string]PropertySchema `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

type PropertySchema struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required,omitempty"`
}
