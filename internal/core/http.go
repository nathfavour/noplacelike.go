package core

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/internal/logger"
)

// HTTPService implementation
type httpService struct {
	config   NetworkConfig
	logger   logger.Logger
	platform PlatformAPI
	router   *gin.Engine
	server   *http.Server
	routes   []Route
	running  bool
	mu       sync.RWMutex
}

func NewHTTPService(config NetworkConfig, log logger.Logger, platform PlatformAPI) (HTTPService, error) {
	// Set gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	service := &httpService{
		config:   config,
		logger:   log,
		platform: platform,
		router:   router,
		routes:   make([]Route, 0),
	}

	// Setup default routes
	service.setupDefaultRoutes()

	return service, nil
}

func (h *httpService) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return ErrAlreadyRunning
	}

	// Create HTTP server
	h.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", h.config.Host, h.config.Port),
		Handler:        h.router,
		ReadTimeout:    h.config.ReadTimeout,
		WriteTimeout:   h.config.WriteTimeout,
		IdleTimeout:    h.config.IdleTimeout,
		MaxHeaderBytes: h.config.MaxHeaderBytes,
	}

	// Start server in goroutine
	go func() {
		h.logger.Info("Starting HTTP server", "addr", h.server.Addr)

		var err error
		if h.config.EnableTLS {
			err = h.server.ListenAndServeTLS(h.config.TLSCertFile, h.config.TLSKeyFile)
		} else {
			err = h.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			h.logger.Error("HTTP server error", "error", err)
		}
	}()

	h.running = true
	h.logger.Info("HTTP service started")
	return nil
}

func (h *httpService) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	if h.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := h.server.Shutdown(shutdownCtx); err != nil {
			h.logger.Error("Error shutting down HTTP server", "error", err)
			return err
		}
	}

	h.running = false
	h.logger.Info("HTTP service stopped")
	return nil
}

func (h *httpService) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

func (h *httpService) Name() string {
	return "HTTPService"
}

func (h *httpService) RegisterRoute(route Route) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.routes = append(h.routes, route)

	// Register with gin router
	switch route.Method {
	case "GET":
		h.router.GET(route.Path, gin.WrapF(route.Handler))
	case "POST":
		h.router.POST(route.Path, gin.WrapF(route.Handler))
	case "PUT":
		h.router.PUT(route.Path, gin.WrapF(route.Handler))
	case "DELETE":
		h.router.DELETE(route.Path, gin.WrapF(route.Handler))
	case "PATCH":
		h.router.PATCH(route.Path, gin.WrapF(route.Handler))
	default:
		return fmt.Errorf("unsupported HTTP method: %s", route.Method)
	}

	h.logger.Info("Route registered", "method", route.Method, "path", route.Path)
	return nil
}

func (h *httpService) RegisterMiddleware(middleware func(http.Handler) http.Handler) {
	// Convert to gin middleware
	ginMiddleware := func(c *gin.Context) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		}))
		handler.ServeHTTP(c.Writer, c.Request)
	}

	h.router.Use(ginMiddleware)
}

func (h *httpService) GetRouter() http.Handler {
	return h.router
}

func (h *httpService) setupDefaultRoutes() {
	// Health check endpoint
	h.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Platform info endpoint
	h.router.GET("/info", func(c *gin.Context) {
		version, buildTime, gitCommit := GetBuildInfo()
		c.JSON(http.StatusOK, gin.H{
			"name":      "NoPlaceLike",
			"version":   version,
			"buildTime": buildTime,
			"gitCommit": gitCommit,
		})
	})

	// API routes group
	api := h.router.Group("/api")
	{
		// Platform routes
		platform := api.Group("/platform")
		{
			platform.GET("/metrics", h.handleMetrics)
			platform.GET("/health", h.handleHealthCheck)
		}

		// Plugin routes
		plugins := api.Group("/plugins")
		{
			plugins.GET("", h.handleListPlugins)
			plugins.GET("/:name", h.handleGetPlugin)
			plugins.POST("/:name/start", h.handleStartPlugin)
			plugins.POST("/:name/stop", h.handleStopPlugin)
		}

		// Network routes
		network := api.Group("/network")
		{
			network.GET("/peers", h.handleListPeers)
			network.POST("/peers/discover", h.handleDiscoverPeers)
		}

		// Resource routes
		resources := api.Group("/resources")
		{
			resources.GET("", h.handleListResources)
			resources.GET("/:id", h.handleGetResource)
			resources.POST("", h.handleCreateResource)
			resources.DELETE("/:id", h.handleDeleteResource)
		}
	}
}

// Handler implementations
func (h *httpService) handleMetrics(c *gin.Context) {
	// TODO: Implement metrics endpoint
	c.JSON(http.StatusOK, gin.H{"metrics": "not implemented"})
}

func (h *httpService) handleHealthCheck(c *gin.Context) {
	// TODO: Implement health check
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *httpService) handleListPlugins(c *gin.Context) {
	// TODO: Implement plugin listing
	c.JSON(http.StatusOK, gin.H{"plugins": []interface{}{}})
}

func (h *httpService) handleGetPlugin(c *gin.Context) {
	name := c.Param("name")
	// TODO: Implement plugin details
	c.JSON(http.StatusOK, gin.H{"plugin": name})
}

func (h *httpService) handleStartPlugin(c *gin.Context) {
	name := c.Param("name")
	// TODO: Implement plugin start
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Plugin %s started", name)})
}

func (h *httpService) handleStopPlugin(c *gin.Context) {
	name := c.Param("name")
	// TODO: Implement plugin stop
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Plugin %s stopped", name)})
}

func (h *httpService) handleListPeers(c *gin.Context) {
	// TODO: Implement peer listing
	c.JSON(http.StatusOK, gin.H{"peers": []interface{}{}})
}

func (h *httpService) handleDiscoverPeers(c *gin.Context) {
	// TODO: Implement peer discovery
	c.JSON(http.StatusOK, gin.H{"message": "Peer discovery initiated"})
}

func (h *httpService) handleListResources(c *gin.Context) {
	// TODO: Implement resource listing
	c.JSON(http.StatusOK, gin.H{"resources": []interface{}{}})
}

func (h *httpService) handleGetResource(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement resource details
	c.JSON(http.StatusOK, gin.H{"resource": id})
}

func (h *httpService) handleCreateResource(c *gin.Context) {
	// TODO: Implement resource creation
	c.JSON(http.StatusCreated, gin.H{"message": "Resource created"})
}

func (h *httpService) handleDeleteResource(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement resource deletion
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Resource %s deleted", id)})
}

func (s *httpService) Configuration() ConfigSchema {
	return ConfigSchema{
		Properties: map[string]PropertySchema{
			"host": {
				Type:        "string",
				Description: "HTTP server host",
				Default:     "localhost",
			},
			"port": {
				Type:        "integer",
				Description: "HTTP server port",
				Default:     8080,
			},
		},
	}
}

func (s *httpService) Health() HealthStatus {
	return HealthStatus{
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
	}
}
