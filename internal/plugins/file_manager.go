package plugins

import (
	"github.com/nathfavour/noplacelike.go/internal/core"
)

// FileManagerPlugin represents the file manager plugin
type FileManagerPlugin struct{}

// Configuration returns the plugin configuration schema
func (p *FileManagerPlugin) Configuration() core.ConfigSchema {
	return core.ConfigSchema{
		Properties: map[string]core.PropertySchema{
			"rootPath": {
				Type:        "string",
				Description: "Root path for file operations",
				Default:     "./files",
			},
			"maxFileSize": {
				Type:        "integer",
				Description: "Maximum file size in bytes",
				Default:     10485760, // 10MB
			},
		},
		Required: []string{"rootPath"},
	}
}
