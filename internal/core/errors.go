package core

import "errors"

// Common errors
var (
	ErrAlreadyRunning    = errors.New("platform is already running")
	ErrNotRunning       = errors.New("platform is not running")
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrPluginNotFound   = errors.New("plugin not found")
	ErrResourceNotFound = errors.New("resource not found")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrInvalidRequest   = errors.New("invalid request")
)