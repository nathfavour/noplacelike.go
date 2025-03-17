package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	// "fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nathfavour/noplacelike.go/config"
)

// ShellRequest represents a shell command execution request
type ShellRequest struct {
	Command string `json:"command" binding:"required"`
	Timeout int    `json:"timeout"` // in seconds
	Dir     string `json:"dir"`     // working directory
}

// ShellResponse represents the response from a shell command
type ShellResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
	Error    string `json:"error,omitempty"`
}

// ShellAPI handles shell command execution
type ShellAPI struct {
	config     *config.Config
	wsUpgrader websocket.Upgrader
}

// NewShellAPI creates a new shell API handler
func NewShellAPI(cfg *config.Config) *ShellAPI {
	return &ShellAPI{
		config: cfg,
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin
			},
		},
	}
}

// ExecuteCommand executes a shell command and returns the result
func (s *ShellAPI) ExecuteCommand(c *gin.Context) {
	// Check if shell execution is enabled
	if !s.config.EnableShell {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Shell command execution is disabled",
		})
		return
	}

	var req ShellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Security check: Only allow commands that are in the allowlist if configured
	if len(s.config.AllowedCommands) > 0 {
		cmdName := strings.Fields(req.Command)[0]
		allowed := false
		for _, allowedCmd := range s.config.AllowedCommands {
			if cmdName == allowedCmd {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Command not in allowed list",
			})
			return
		}
	}

	// Set default timeout if not specified
	if req.Timeout <= 0 {
		req.Timeout = 30 // Default to 30 seconds
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	defer cancel()

	// Prepare command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", req.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", req.Command)
	}

	// Set working directory if specified
	if req.Dir != "" {
		cmd.Dir = expandPath(req.Dir)
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Prepare response
	resp := ShellResponse{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			resp.Error = err.Error()
		}
	}

	c.JSON(http.StatusOK, resp)
}

// StreamCommand streams the output of a command through WebSocket
func (s *ShellAPI) StreamCommand(c *gin.Context) {
	// Check if shell execution is enabled
	if !s.config.EnableShell {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Shell command execution is disabled",
		})
		return
	}

	// Get command from query parameter
	command := c.Query("command")
	if command == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Command parameter is required",
		})
		return
	}

	// Security check: Only allow commands that are in the allowlist if configured
	if len(s.config.AllowedCommands) > 0 {
		cmdName := strings.Fields(command)[0]
		allowed := false
		for _, allowedCmd := range s.config.AllowedCommands {
			if cmdName == allowedCmd {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Command not in allowed list",
			})
			return
		}
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upgrade connection: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	// Prepare command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "Failed to create stdout pipe: " + err.Error()})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "Failed to create stderr pipe: " + err.Error()})
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		conn.WriteJSON(map[string]string{"error": "Failed to start command: " + err.Error()})
		return
	}

	// Send initial message
	conn.WriteJSON(map[string]string{"status": "Command started"})

	// Create done channel
	done := make(chan struct{})
	defer close(done)

	// Stream stdout to client
	go streamPipeToWebsocket(stdout, conn, "stdout", done)
	go streamPipeToWebsocket(stderr, conn, "stderr", done)

	// Handle command completion
	go func() {
		err := cmd.Wait()
		var exitCode int
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				conn.WriteJSON(map[string]string{"error": "Command failed: " + err.Error()})
				return
			}
		}

		// Wait a moment for streaming to complete
		time.Sleep(100 * time.Millisecond)
		conn.WriteJSON(map[string]interface{}{
			"status":   "Command completed",
			"exitCode": exitCode,
		})
	}()

	// Handle client messages (like send input to command)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// Client disconnected or connection error
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			break
		}

		var clientMsg struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			continue
		}

		// Handle client message
		switch clientMsg.Type {
		case "interrupt":
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
		// Future: could add support for stdin input
	}
}

// streamPipeToWebsocket reads from a pipe and sends the data to a WebSocket
func streamPipeToWebsocket(pipe io.ReadCloser, conn *websocket.Conn, streamType string, done chan struct{}) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		select {
		case <-done:
			return
		default:
			line := scanner.Text()
			conn.WriteJSON(map[string]string{
				"type":    streamType,
				"content": line,
			})
		}
	}
}
