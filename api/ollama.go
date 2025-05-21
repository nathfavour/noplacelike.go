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
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1/ollama")
	url := o.BaseURL + path
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
	// Copy headers
	for k, v := range c.Request.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Writer.Header().Add(k, vv)
		}
	}
	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
