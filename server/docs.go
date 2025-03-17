package server

import (
	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/docs"
)

// registerDocRoutes adds API documentation routes
func (s *Server) registerDocRoutes() {
	// API Documentation
	s.router.GET("/api/docs", docs.DocumentationHandler())
	s.router.GET("/api/docs/json", docs.JSONDocumentationHandler())
}
