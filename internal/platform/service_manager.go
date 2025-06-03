package platform

// Service represents a generic service interface
type Service interface {
	Name() string
	Start() error
	Stop() error
}

// This file is intentionally left blank. NewServiceManager is defined in platform.go.
