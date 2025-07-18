// Package services contains core platform services
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/internal/core"
	"github.com/nathfavour/noplacelike.go/internal/platform"
)

// HTTPService provides HTTP API capabilities
type HTTPService struct {
	mu       sync.RWMutex
	name     string
	config   HTTPConfig
	server   *http.Server
	router   *gin.Engine
	platform *platform.Platform
	logger   core.Logger
	started  bool
}

// HTTPConfig contains HTTP service configuration
type HTTPConfig struct {
	Host           string        `json:"host"`
	Port           int           `json:"port"`
	EnableTLS      bool          `json:"enableTLS"`
	TLSCertFile    string        `json:"tlsCertFile"`
	TLSKeyFile     string        `json:"tlsKeyFile"`
	ReadTimeout    time.Duration `json:"readTimeout"`
	WriteTimeout   time.Duration `json:"writeTimeout"`
	IdleTimeout    time.Duration `json:"idleTimeout"`
	MaxRequestSize int64         `json:"maxRequestSize"`
	EnableCORS     bool          `json:"enableCORS"`
	EnableMetrics  bool          `json:"enableMetrics"`
	EnableDocs     bool          `json:"enableDocs"`
	RateLimitRPS   int           `json:"rateLimitRPS"`
	EnableGzip     bool          `json:"enableGzip"`
}

// NewHTTPService creates a new HTTP service
func NewHTTPService(config HTTPConfig, platform *platform.Platform) *HTTPService {
	// Set gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	return &HTTPService{
		name:     "http",
		config:   config,
		router:   gin.New(),
		platform: platform,
		logger:   platform.Logger(),
	}
}

// Name returns the service name
func (s *HTTPService) Name() string {
	return s.name
}

// Start begins the HTTP service
func (s *HTTPService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("HTTP service already started")
	}

	// Setup middleware
	s.setupMiddleware()

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		s.logger.Info("Starting HTTP server",
			core.Field{Key: "address", Value: addr},
			core.Field{Key: "tls", Value: s.config.EnableTLS},
		)

		var err error
		if s.config.EnableTLS {
			err = s.server.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", core.Field{Key: "error", Value: err})
		}
	}()

	s.started = true
	s.logger.Info("HTTP service started successfully")
	return nil
}

// Stop gracefully shuts down the HTTP service
func (s *HTTPService) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("HTTP service not started")
	}

	s.logger.Info("Stopping HTTP service")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.started = false
	s.logger.Info("HTTP service stopped")
	return nil
}

// Health returns the service health status
func (s *HTTPService) Health() core.HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := core.HealthStatusHealthy
	if !s.started {
		status = core.HealthStatusUnhealthy
	}

	return core.HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"started": s.started,
			"address": fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		},
	}
}

// IsHealthy returns true if the HTTP service is running
func (s *HTTPService) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// Configuration returns the service configuration schema
func (s *HTTPService) Configuration() core.ConfigSchema {
	return core.ConfigSchema{
		Properties: map[string]core.PropertySchema{
			"host": {
				Type:        "string",
				Description: "Host address to bind to",
				Default:     "0.0.0.0",
			},
			"port": {
				Type:        "integer",
				Description: "Port to listen on",
				Default:     8080,
			},
			"enableTLS": {
				Type:        "boolean",
				Description: "Enable TLS/HTTPS",
				Default:     false,
			},
		},
		Required: []string{"host", "port"},
	}
}

// setupMiddleware configures HTTP middleware
func (s *HTTPService) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logging middleware
	s.router.Use(s.loggingMiddleware())

	// CORS middleware
	if s.config.EnableCORS {
		s.router.Use(s.corsMiddleware())
	}

	// Rate limiting middleware
	if s.config.RateLimitRPS > 0 {
		s.router.Use(s.rateLimitMiddleware())
	}

	// Gzip compression middleware
	if s.config.EnableGzip {
		// Would implement gzip middleware
	}

	// Security headers middleware
	s.router.Use(s.securityHeadersMiddleware())

	// Request size limit middleware
	s.router.Use(s.requestSizeLimitMiddleware())
}

// setupRoutes configures HTTP routes
func (s *HTTPService) setupRoutes() {
	// API version info
	s.router.GET("/", s.handleRoot)
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/info", s.handleInfo)

	// API routes
	api := s.router.Group("/api")
	{
		// Platform management
		platform := api.Group("/platform")
		{
			platform.GET("/health", s.handlePlatformHealth)
			platform.GET("/info", s.handlePlatformInfo)
			platform.GET("/metrics", s.handleMetrics)
		}

		// Plugin management
		plugins := api.Group("/plugins")
		{
			plugins.GET("", s.handleListPlugins)
			plugins.GET("/:name", s.handleGetPlugin)
			plugins.POST("/:name/start", s.handleStartPlugin)
			plugins.POST("/:name/stop", s.handleStopPlugin)
			plugins.GET("/:name/health", s.handlePluginHealth)
		}

		// Service management
		services := api.Group("/services")
		{
			services.GET("", s.handleListServices)
			services.GET("/:name", s.handleGetService)
			services.GET("/:name/health", s.handleServiceHealth)
		}

		// Network management
		network := api.Group("/network")
		{
			network.GET("/peers", s.handleListPeers)
			network.GET("/peers/:id", s.handleGetPeer)
			network.POST("/peers/discover", s.handleDiscoverPeers)
		}

		// Resource management
		resources := api.Group("/resources")
		{
			resources.GET("", s.handleListResources)
			resources.GET("/:id", s.handleGetResource)
			resources.POST("", s.handleCreateResource)
			resources.DELETE("/:id", s.handleDeleteResource)
			resources.GET("/:id/stream", s.handleStreamResource)
		}

		// Events and subscriptions
		events := api.Group("/events")
		{
			events.GET("/stream", s.handleEventStream)
			events.POST("/publish", s.handlePublishEvent)
		}
	}

	// Register plugin routes
	s.registerPluginRoutes()
}

// registerPluginRoutes registers routes provided by plugins
func (s *HTTPService) registerPluginRoutes() {
	plugins := s.platform.ListPlugins()

	for name, plugin := range plugins {
		routes := plugin.Routes()

		for _, route := range routes {
			// Create a group for the plugin
			group := s.router.Group(fmt.Sprintf("/plugins/%s", name))

			// Add authentication middleware if required
			var handlers []gin.HandlerFunc
			if route.Auth.Required {
				handlers = append(handlers, s.authMiddleware(route.Auth.Permissions))
			}

			// Add custom middleware
			for _, middleware := range route.Middleware {
				handlers = append(handlers, gin.WrapH(middleware(http.HandlerFunc(route.Handler))))
			}

			// Add the main handler
			handlers = append(handlers, gin.WrapH(http.HandlerFunc(route.Handler)))

			// Register the route
			group.Handle(route.Method, route.Path, handlers...)
		}
	}
}

// HTTP Handlers
func (s *HTTPService) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":    "NoPlaceLike Platform",
		"version": s.platform.Health().Details["version"],
		"status":  "running",
		"uptime":  s.platform.Health().Details["uptime"],
	})
}

func (s *HTTPService) handleHealth(c *gin.Context) {
	health := s.platform.Health()

	statusCode := http.StatusOK
	if health.Status == core.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == core.HealthStatusDegraded {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, health)
}

func (s *HTTPService) handleInfo(c *gin.Context) {
	info := map[string]interface{}{
		"platform": s.platform.Health().Details,
		"services": s.platform.ServiceManager().HealthCheck(),
		"plugins":  len(s.platform.ListPlugins()),
		"peers":    len(s.platform.NetworkManager().GetPeers()),
	}

	c.JSON(http.StatusOK, info)
}

func (s *HTTPService) handlePlatformHealth(c *gin.Context) {
	c.JSON(http.StatusOK, s.platform.Health())
}

func (s *HTTPService) handlePlatformInfo(c *gin.Context) {
	c.JSON(http.StatusOK, s.platform.Health().Details)
}

func (s *HTTPService) handleMetrics(c *gin.Context) {
	format := c.DefaultQuery("format", "json")

	metrics, err := s.platform.Metrics().Export(format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if format == "json" {
		c.Data(http.StatusOK, "application/json", metrics)
	} else {
		c.Data(http.StatusOK, "text/plain", metrics)
	}
}

func (s *HTTPService) handleListPlugins(c *gin.Context) {
	plugins := s.platform.ListPlugins()

	result := make([]map[string]interface{}, 0, len(plugins))
	for name, plugin := range plugins {
		result = append(result, map[string]interface{}{
			"name":    name,
			"version": plugin.Version(),
			"health":  plugin.Health(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"plugins": result})
}

func (s *HTTPService) handleGetPlugin(c *gin.Context) {
	name := c.Param("name")

	plugin, err := s.platform.GetPlugin(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"name":         plugin.Name(),
		"version":      plugin.Version(),
		"health":       plugin.Health(),
		"dependencies": plugin.Dependencies(),
		"routes":       plugin.Routes(),
	})
}

func (s *HTTPService) handleStartPlugin(c *gin.Context) {
	name := c.Param("name")

	plugin, err := s.platform.GetPlugin(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := plugin.Start(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

func (s *HTTPService) handleStopPlugin(c *gin.Context) {
	name := c.Param("name")

	plugin, err := s.platform.GetPlugin(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := plugin.Stop(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

func (s *HTTPService) handlePluginHealth(c *gin.Context) {
	name := c.Param("name")

	plugin, err := s.platform.GetPlugin(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plugin.Health())
}

func (s *HTTPService) handleListServices(c *gin.Context) {
	health := s.platform.ServiceManager().HealthCheck()
	c.JSON(http.StatusOK, gin.H{"services": health})
}

func (s *HTTPService) handleGetService(c *gin.Context) {
	name := c.Param("name")

	service, err := s.platform.ServiceManager().GetService(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"name":   service.Name(),
		"health": service.Health(),
		"config": service.Configuration(),
	})
}

func (s *HTTPService) handleServiceHealth(c *gin.Context) {
	name := c.Param("name")

	service, err := s.platform.ServiceManager().GetService(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service.Health())
}

func (s *HTTPService) handleListPeers(c *gin.Context) {
	peers := s.platform.NetworkManager().GetPeers()
	c.JSON(http.StatusOK, gin.H{"peers": peers})
}

func (s *HTTPService) handleGetPeer(c *gin.Context) {
	id := c.Param("id")

	peers := s.platform.NetworkManager().GetPeers()
	for _, peer := range peers {
		if peer.ID == id {
			c.JSON(http.StatusOK, peer)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "peer not found"})
}

func (s *HTTPService) handleDiscoverPeers(c *gin.Context) {
	peers, err := s.platform.NetworkManager().DiscoverPeers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"peers": peers})
}

func (s *HTTPService) handleListResources(c *gin.Context) {
	filter := core.ResourceFilter{
		Name: "example",
		Type: "file",
		// Type:  c.Query("type"),
		// Owner: c.Query("owner"),
	}

	resources, err := s.platform.ResourceManager().ListResources(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resources": resources})
}

func (s *HTTPService) handleGetResource(c *gin.Context) {
	id := c.Param("id")

	resource, err := s.platform.ResourceManager().GetResource(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource)
}

func (s *HTTPService) handleCreateResource(c *gin.Context) {
	var resource core.Resource
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.platform.ResourceManager().RegisterResource(resource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resource)
}

func (s *HTTPService) handleDeleteResource(c *gin.Context) {
	id := c.Param("id")

	// Implementation would remove the resource
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
}

func (s *HTTPService) handleStreamResource(c *gin.Context) {
	id := c.Param("id")

	stream, err := s.platform.ResourceManager().StreamResource(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	// Stream the resource content
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Transfer-Encoding", "chunked")

	// Copy stream to response
	c.Stream(func(w io.Writer) bool {
		data, err := stream.Read()
		if err != nil {
			return false
		}
		w.Write(data)
		return true
	})
}

func (s *HTTPService) handleEventStream(c *gin.Context) {
	// Implementation for Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Subscribe to events
	err := s.platform.EventBus().Subscribe("*", core.EventHandler(func(event core.Event) error {
		data, _ := json.Marshal(event)
		c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
		c.Writer.Flush()
		return nil
	}))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Keep connection alive
	<-c.Request.Context().Done()
}

func (s *HTTPService) handlePublishEvent(c *gin.Context) {
	var event core.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// topic := c.DefaultQuery("topic", "custom")

	if err := s.platform.EventBus().Publish(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "published"})
}

// Middleware functions
func (s *HTTPService) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func (s *HTTPService) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *HTTPService) rateLimitMiddleware() gin.HandlerFunc {
	// Implementation would use a rate limiter
	return func(c *gin.Context) {
		c.Next()
	}
}

func (s *HTTPService) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

func (s *HTTPService) requestSizeLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > s.config.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request too large"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *HTTPService) authMiddleware(permissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Validate token
		tokenInfo, err := s.platform.SecurityManager().ValidateToken(c.Request.Context(), token)
		if err != nil || !tokenInfo.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Check permissions
		for _, permission := range permissions {
			hasPermission := false
			for _, userPerm := range tokenInfo.Permissions {
				if userPerm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				c.Abort()
				return
			}
		}

		// Set user context
		c.Set("userID", tokenInfo.PeerID)
		c.Set("permissions", tokenInfo.Permissions)

		c.Next()
	}
}
