package api

import (
	"os"
	"path/filepath"
	"strings"
)

// expandPath expands the ~ in a path to the user's home directory
func expandPath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}
	return path
}

// isSubPath checks if path is a subpath of basePath
func isSubPath(path, basePath string) bool {
	rel, err := filepath.Rel(basePath, path)
	if err != nil {
		return false
	}
	return !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..")
}
