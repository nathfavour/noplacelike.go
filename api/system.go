package api

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathfavour/noplacelike.go/config"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemAPI handles system information and operations
type SystemAPI struct {
	config *config.Config
}

// NewSystemAPI creates a new system API handler
func NewSystemAPI(cfg *config.Config) *SystemAPI {
	return &SystemAPI{
		config: cfg,
	}
}

// NotificationRequest represents a system notification request
type NotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Type    string `json:"type"` // info, warning, error
}

// GetSystemInfo returns basic system information
func (s *SystemAPI) GetSystemInfo(c *gin.Context) {
	info := map[string]interface{}{}

	// Get hostname
	hostname, err := os.Hostname()
	if err == nil {
		info["hostname"] = hostname
	}

	// Platform info
	info["platform"] = runtime.GOOS
	info["architecture"] = runtime.GOARCH
	info["goVersion"] = runtime.Version()
	info["numCPU"] = runtime.NumCPU()

	// Host info
	if hostInfo, err := host.Info(); err == nil {
		info["os"] = hostInfo.OS
		info["platform"] = hostInfo.Platform
		info["platformFamily"] = hostInfo.PlatformFamily
		info["platformVersion"] = hostInfo.PlatformVersion
		info["kernelVersion"] = hostInfo.KernelVersion
		info["kernelArch"] = hostInfo.KernelArch
		
		// Format uptime
		uptime := time.Duration(hostInfo.Uptime) * time.Second
		days := int(uptime.Hours() / 24)
		hours := int(uptime.Hours()) % 24
		minutes := int(uptime.Minutes()) % 60
		info["uptime"] = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}

	// CPU info
	if cpuPercents, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercents) > 0 {
		info["cpuUsage"] = fmt.Sprintf("%.1f%%", cpuPercents[0])
	}

	// Memory info
	if memInfo, err := mem.VirtualMemory(); err == nil {
		info["memoryTotal"] = memInfo.Total
		info["memoryAvailable"] = memInfo.Available
		info["memoryUsed"] = memInfo.Used
		info["memoryUsage"] = fmt.Sprintf("%.1f%%", memInfo.UsedPercent)
	}

	// Disk info
	if diskInfo, err := disk.Usage("/"); err == nil {
		info["diskTotal"] = diskInfo.Total
		info["diskFree"] = diskInfo.Free
		info["diskUsed"] = diskInfo.Used
		info["diskUsage"] = fmt.Sprintf("%.1f%%", diskInfo.UsedPercent)
	}

	c.JSON(http.StatusOK, info)
}

// GetProcesses returns a list of running processes
func (s *SystemAPI) GetProcesses(c *gin.Context) {
	// Get list of processes
	processes, err := process.Processes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to get processes: " + err.Error(),
		})
		return
	}

	// Collect process info
	processInfos := make([]map[string]interface{}, 0, len(processes))
	for _, p := range processes {
		info := make(map[string]interface{})

		// Get process ID
		info["pid"] = p.Pid

		// Try to get other info, continue if any fails
		if name, err := p.Name(); err == nil {
			info["name"] = name
		}

		if status, err := p.Status(); err == nil {
			info["status"] = status
		}

		if cmdline, err := p.Cmdline(); err == nil {
			info["cmdline"] = cmdline
		}

		if createTime, err := p.CreateTime(); err == nil {
			info["createTime"] = time.Unix(createTime/1000, 0).Format(time.RFC3339)
		}

		if cpuPercent, err := p.CPUPercent(); err == nil {
			info["cpuPercent"] = fmt.Sprintf("%.1f%%", cpuPercent)
		}

		if memPercent, err := p.MemoryPercent(); err == nil {
			info["memPercent"] = fmt.Sprintf("%.1f%%", memPercent)
		}

		processInfos = append(processInfos, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(processInfos),
		"processes": processInfos,
	})
}

// SendNotification sends a system notification
func (s *SystemAPI) SendNotification(c *gin.Context) {
	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Set default notification type if not provided
	if req.Type == "" {
		req.Type = "info" // Default to info
	}

	// Set custom type if not one of the standard types
	if req.Type != "info" && req.Type != "warning" && req.Type != "error" {
		req.Type = "info"
	}

	// Here would go platform-specific notification code
	// For now, just print to console and return success
	fmt.Printf("[%s] %s: %s\n", req.Type, req.Title, req.Message)
	
	// TODO: Implement actual notification using platform-specific libraries
	// For Linux: github.com/esiqveland/notify
	// For macOS: github.com/deckarep/gosx-notifier
	// For Windows: github.com/go-toast/toast

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Notification sent",
	})
}
