package platform

import (
	"fmt"
	"sync"
)

// Service represents a generic service interface
type Service interface {
	Name() string
	Start() error
	Stop() error
}

// ServiceManagerImpl is the implementation of the ServiceManager
type ServiceManagerImpl struct {
	services map[string]Service
	mu       sync.Mutex
}

// NewServiceManager creates a new instance of ServiceManagerImpl
func NewServiceManager() *ServiceManagerImpl {
	return &ServiceManagerImpl{
		services: make(map[string]Service),
	}
}

// RegisterService registers a new service with the manager
func (sm *ServiceManagerImpl) RegisterService(service Service) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.services[service.Name()]; exists {
		return fmt.Errorf("service %s already registered", service.Name())
	}

	sm.services[service.Name()] = service
	return nil
}

// StartAll starts all registered services
func (sm *ServiceManagerImpl) StartAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for name, service := range sm.services {
		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", name, err)
		}
	}
	return nil
}

// StopAll stops all registered services
func (sm *ServiceManagerImpl) StopAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for name, service := range sm.services {
		if err := service.Stop(); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", name, err)
		}
	}
	return nil
}
