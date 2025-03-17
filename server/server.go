package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdp/qrterminal/v3"
	"github.com/nathfavour/noplacelike.go/api"
	"github.com/nathfavour/noplacelike.go/config"
)

// Server represents the NoPlaceLike server
type Server struct {
	config *config.Config
	router *gin.Engine
	server *http.Server
	clipboard string // In-memory clipboard storage
}

// NewServer creates a new HTTP server
func NewServer(config *config.Config) *Server {
    // Initialize server without creating directories
    server := &Server{
        config: config,
        router: gin.Default(),
    }
    
    // Initialize routes
    server.setupRoutes()
    
    return server
}

// Start starts the server
func (s *Server) Start() {
    // Create address string
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    
    // Start the server
    fmt.Printf("🚀 Server running at http://%s\n", addr)
    if err := s.router.Run(addr); err != nil {
        fmt.Printf("❌ Server failed to start: %v\n", err)
        os.Exit(1)
    }
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
	}
}

// setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	// Initialize API and create its routes on the router
	apiHandler := api.NewAPI(s.config)
	apiHandler.CreateRoutes(s.router) // Changed from SetupRoutes to CreateRoutes
	
	// Redirect root to UI
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui")
	})
	
	// UI routes - Web interface
	s.router.GET("/ui", s.uiHome)
	s.router.GET("/admin", s.adminPanel)
	
	// Serve static files
	s.router.Static("/static", "./static")
}

// ensureDirExists creates a directory if it doesn't exist
func ensureDirExists(path string) error {
	path = expandPath(path)
	return os.MkdirAll(path, 0755)
}

// expandPath expands the ~ in a path to the user's home directory
func expandPath(path string) string {
	if path == "" || !strings.HasPrefix(path, "~") {
		return path
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	
	return filepath.Join(homeDir, path[1:])
}

// DisplayAccessInfo displays QR codes and URLs for accessing the server
func DisplayAccessInfo(host string, port int) {
	fmt.Println("\nNoPlaceLike Server is running!")
	fmt.Println("==================================")
	
	// Get all IP addresses
	ips := getAllIPs()
	
	// Print access URLs with QR codes
	for _, ip := range ips {
		url := fmt.Sprintf("http://%s:%d", ip, port)
		
		// Categorize the IP
		ipType := "OTHER"
		if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
			ipType = "LOCAL NETWORK (PREFERRED)"
		} else if ip == "127.0.0.1" {
			ipType = "LOCALHOST"
		}
		
		fmt.Printf("\n=== %s ACCESS ===\n", ipType)
		fmt.Printf("URL: %s\n\n", url)
		
		// Generate QR code
		config := qrterminal.Config{
			Level: qrterminal.M,
			Writer: os.Stdout,
			BlackChar: qrterminal.BLACK,
			WhiteChar: qrterminal.WHITE,
			QuietZone: 1,
		}
		qrterminal.GenerateWithConfig(url, config)
		fmt.Println(strings.Repeat("-", 50))
	}
}

// getAllIPs returns all available IP addresses sorted by preference
func getAllIPs() []string {
	ips := make(map[string]bool)
	
	// Get hostname-based IP
	hostIP, err := getOutboundIP()
	if err == nil && !strings.HasPrefix(hostIP, "127.") {
		ips[hostIP] = true
	}
	
	// Get all network interface IPs
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			// Skip loopback and non-up interfaces
			if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
				continue
			}
			
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				
				// Only include IPv4 addresses
				if ip == nil || ip.To4() == nil {
					continue
				}
				
				ips[ip.String()] = true
			}
		}
	}
	
	// Always include localhost
	ips["127.0.0.1"] = true
	
	// Convert map to slice
	var result []string
	for ip := range ips {
		result = append(result, ip)
	}
	
	// Sort IPs to prioritize local network
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		
		// Define priority function
		getPriority := func(ip string) int {
			if strings.HasPrefix(ip, "192.168.") {
				return 0
			}
			if strings.HasPrefix(ip, "10.") {
				return 1
			}
			if strings.HasPrefix(ip, "172.") {
				return 2
			}
			if ip == "127.0.0.1" {
				return 3
			}
			return 4
		}
		
		pa, pb := getPriority(a), getPriority(b)
		if pa != pb {
			return pa < pb
		}
		return a < b
	})
	
	return result
}

// getOutboundIP gets the preferred outbound IP address
func getOutboundIP() (string, error) {
	// This UDP connection doesn't actually establish a connection,
	// but it does cause the OS to determine the outbound IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
