// OllamaAPI proxies requests to a local Ollama server (default: http://localhost:11434)
package api

import (
	"net/http"
	"net/url"
	"strings"

	ollama "github.com/JexSrs/go-ollama"
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

	parsedURL, err := url.Parse(o.BaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid Ollama base URL"})
		return
	}
	LLM := ollama.New(*parsedURL)

	switch path {
	case "/chat":
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		model, _ := req["model"].(string)
		messages, _ := req["messages"].([]interface{})
		var lastMsg map[string]interface{}
		if len(messages) > 0 {
			if msg, ok := messages[len(messages)-1].(map[string]interface{}); ok {
				lastMsg = msg
			}
		}
		if lastMsg == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no message provided"})
			return
		}
		var roleStr, contentStr string
		if v, ok := lastMsg["role"].(string); ok {
			roleStr = v
		}
		if v, ok := lastMsg["content"].(string); ok {
			contentStr = v
		}
		msg := ollama.Message{
			Role:    &roleStr,
			Content: &contentStr,
		}
		res, err := LLM.Chat(
			nil,
			LLM.Chat.WithModel(model),
			LLM.Chat.WithMessage(msg),
		)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	case "/generate":
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		model, _ := req["model"].(string)
		prompt, _ := req["prompt"].(string)
		res, err := LLM.Generate(
			LLM.Generate.WithModel(model),
			LLM.Generate.WithPrompt(prompt),
		)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	case "/tags":
		res, err := LLM.Models.List()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "unsupported endpoint"})
	}
}
