// OllamaAPI proxies requests to a local Ollama server (default: http://localhost:11434)
package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type OllamaAPI struct {
	BaseURL string
}

func NewOllamaAPI(baseURL string) *OllamaAPI {
	return &OllamaAPI{BaseURL: baseURL}
}

// Proxy all requests to Ollama
func (o *OllamaAPI) Proxy(c *gin.Context) {
	// Extract path without the /api/v1/ollama prefix
	path := c.Param("proxyPath")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Make sure we're hitting the right Ollama API endpoints
	// Path comes in as "/tags", "/chat", etc. but Ollama expects "/api/tags", "/api/chat"
	url := o.BaseURL + "/api" + path

	method := c.Request.Method
	client := &http.Client{}

	var body io.Reader
	if c.Request.Body != nil {
		body = c.Request.Body
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Copy headers - ensure Content-Type is preserved
	for k, v := range c.Request.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	// Ensure Content-Type is set for JSON requests if not already present
	if req.Header.Get("Content-Type") == "" && (method == "POST" || method == "PUT") {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Copy response headers back to client
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Writer.Header().Add(k, vv)
		}
	}

	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
