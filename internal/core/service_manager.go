package core

import (
	"fmt"
	"sync"

	"example.com/project/core"
)

// ServiceManager manages the lifecycle of services
type ServiceManager struct {
	mu       sync.Mutex
	services map[string]core.Service
}

// NewServiceManager creates a new ServiceManager instance
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]core.Service),
	}
}

// RegisterService registers a new service with the manager
func (sm *ServiceManager) RegisterService(service core.Service) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.services[service.Name()]; exists {
		return fmt.Errorf("service %s already registered", service.Name())
	}

	sm.services[service.Name()] = service
	return nil
}
