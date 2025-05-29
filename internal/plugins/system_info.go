package plugins

import (
	"github.com/example/core"
)

// SystemInfoPlugin represents the system information plugin
type SystemInfoPlugin struct{}

// Name returns the name of the plugin
func (p *SystemInfoPlugin) Name() string {
	return "SystemInfo"
}

// Configuration returns the plugin configuration schema
func (p *SystemInfoPlugin) Configuration() core.ConfigSchema {
	return core.ConfigSchema{
		Properties: map[string]core.PropertySchema{
			"refreshInterval": {
				Type:        "integer",
				Description: "System info refresh interval in seconds",
				Default:     60,
			},
			"includeProcesses": {
				Type:        "boolean",
				Description: "Include process information",
				Default:     false,
			},
		},
		Required: []string{},
	}
}
