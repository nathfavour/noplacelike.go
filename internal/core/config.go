package core

import (
	"time"
)

// Config represents the platform configuration
type Config struct {
	Name        string           `json:"name" yaml:"name"`
	Version     string           `json:"version" yaml:"version"`
	Environment string           `json:"environment" yaml:"environment"`
	Network     NetworkConfig    `json:"network" yaml:"network"`
	Security    SecurityConfig   `json:"security" yaml:"security"`
	Plugins     PluginsConfig    `json:"plugins" yaml:"plugins"`
	Storage     StorageConfig    `json:"storage" yaml:"storage"`
	Monitoring  MonitoringConfig `json:"monitoring" yaml:"monitoring"`
}

// NetworkConfig holds network-related configuration
type NetworkConfig struct {
	Host              string        `json:"host" yaml:"host"`
	Port              int           `json:"port" yaml:"port"`
	EnableDiscovery   bool          `json:"enableDiscovery" yaml:"enableDiscovery"`
	MaxPeers          int           `json:"maxPeers" yaml:"maxPeers"`
	EnableTLS         bool          `json:"enableTLS" yaml:"enableTLS"`
	TLSCertFile       string        `json:"tlsCertFile" yaml:"tlsCertFile"`
	TLSKeyFile        string        `json:"tlsKeyFile" yaml:"tlsKeyFile"`
	ReadTimeout       time.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout      time.Duration `json:"writeTimeout" yaml:"writeTimeout"`
	IdleTimeout       time.Duration `json:"idleTimeout" yaml:"idleTimeout"`
	MaxHeaderBytes    int           `json:"maxHeaderBytes" yaml:"maxHeaderBytes"`
	EnableCompression bool          `json:"enableCompression" yaml:"enableCompression"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableAuth       bool          `json:"enableAuth" yaml:"enableAuth"`
	EnableEncryption bool          `json:"enableEncryption" yaml:"enableEncryption"`
	JWTSecret        string        `json:"jwtSecret" yaml:"jwtSecret"`
	JWTExpiry        time.Duration `json:"jwtExpiry" yaml:"jwtExpiry"`
	EnableRBAC       bool          `json:"enableRBAC" yaml:"enableRBAC"`
	EnableAuditLog   bool          `json:"enableAuditLog" yaml:"enableAuditLog"`
	TrustedProxies   []string      `json:"trustedProxies" yaml:"trustedProxies"`
	CORSOrigins      []string      `json:"corsOrigins" yaml:"corsOrigins"`
}

// PluginsConfig holds plugin-related configuration
type PluginsConfig struct {
	EnablePlugins bool     `json:"enablePlugins" yaml:"enablePlugins"`
	PluginDir     string   `json:"pluginDir" yaml:"pluginDir"`
	AutoLoad      []string `json:"autoLoad" yaml:"autoLoad"`
	MaxPlugins    int      `json:"maxPlugins" yaml:"maxPlugins"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	DataDir     string `json:"dataDir" yaml:"dataDir"`
	TempDir     string `json:"tempDir" yaml:"tempDir"`
	MaxFileSize int64  `json:"maxFileSize" yaml:"maxFileSize"`
	EnableCache bool   `json:"enableCache" yaml:"enableCache"`
	CacheSize   int64  `json:"cacheSize" yaml:"cacheSize"`
}

// MonitoringConfig holds monitoring-related configuration
type MonitoringConfig struct {
	EnableMetrics   bool          `json:"enableMetrics" yaml:"enableMetrics"`
	MetricsPort     int           `json:"metricsPort" yaml:"metricsPort"`
	MetricsPath     string        `json:"metricsPath" yaml:"metricsPath"`
	EnableProfiling bool          `json:"enableProfiling" yaml:"enableProfiling"`
	HealthCheckPath string        `json:"healthCheckPath" yaml:"healthCheckPath"`
	LogLevel        string        `json:"logLevel" yaml:"logLevel"`
	EnableTracing   bool          `json:"enableTracing" yaml:"enableTracing"`
	SampleRate      float64       `json:"sampleRate" yaml:"sampleRate"`
	FlushInterval   time.Duration `json:"flushInterval" yaml:"flushInterval"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Name:        "NoPlaceLike",
		Version:     "2.0.0",
		Environment: "development",
		Network: NetworkConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			EnableDiscovery:   true,
			MaxPeers:          50,
			EnableTLS:         false,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    1 << 20, // 1MB
			EnableCompression: true,
		},
		Security: SecurityConfig{
			EnableAuth:       false,
			EnableEncryption: false,
			JWTExpiry:        24 * time.Hour,
			EnableRBAC:       false,
			EnableAuditLog:   false,
			CORSOrigins:      []string{"*"},
		},
		Plugins: PluginsConfig{
			EnablePlugins: true,
			PluginDir:     "./plugins",
			AutoLoad:      []string{"file-manager", "clipboard", "system-info"},
			MaxPlugins:    20,
		},
		Storage: StorageConfig{
			DataDir:     "./data",
			TempDir:     "./tmp",
			MaxFileSize: 100 * 1024 * 1024, // 100MB
			EnableCache: true,
			CacheSize:   50 * 1024 * 1024, // 50MB
		},
		Monitoring: MonitoringConfig{
			EnableMetrics:   true,
			MetricsPort:     9090,
			MetricsPath:     "/metrics",
			EnableProfiling: false,
			HealthCheckPath: "/health",
			LogLevel:        "info",
			EnableTracing:   false,
			SampleRate:      0.1,
			FlushInterval:   10 * time.Second,
		},
	}
}
