# Contributing to NoPlaceLike

Thank you for your interest in contributing to NoPlaceLike! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Architecture Overview](#architecture-overview)
- [Plugin Development](#plugin-development)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Code Style Guidelines](#code-style-guidelines)
- [Documentation](#documentation)
- [Security](#security)

## Code of Conduct

This project adheres to a code of conduct adapted from the [Contributor Covenant](https://www.contributor-covenant.org/). By participating, you are expected to uphold this code.

### Our Pledge

We pledge to make participation in our project and community a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- Make (optional, for convenience commands)

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/noplacelike.go.git
   cd noplacelike.go
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Build the Application**
   ```bash
   go build -o noplacelike
   ```

4. **Run Tests**
   ```bash
   go test ./...
   ```

5. **Start Development Server**
   ```bash
   ./noplacelike --host localhost --port 8080
   ```

### Development Tools

We recommend installing these tools for development:

```bash
# Code formatting
go install golang.org/x/tools/cmd/goimports@latest

# Linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security scanning
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Testing tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install github.com/onsi/gomega/...@latest
```

## Architecture Overview

NoPlaceLike follows a modular, service-oriented architecture:

```
internal/
├── core/           # Core interfaces and types
├── platform/       # Platform management and coordination
├── plugins/        # Plugin system and built-in plugins
├── network/        # Networking and peer discovery
├── security/       # Authentication and encryption
├── services/       # HTTP, WebSocket, and other services
├── storage/        # Data persistence and resource management
├── metrics/        # Monitoring and observability
└── utils/          # Utility functions and helpers
```

### Key Components

- **Platform**: Central coordinator for all services and plugins
- **Plugin System**: Extensible architecture for adding functionality
- **Network Manager**: Handles peer discovery and communication
- **Security Manager**: Manages authentication and encryption
- **Service Manager**: Coordinates various network services
- **Resource Manager**: Handles resource lifecycle and access
- **Event Bus**: Enables loose coupling between components

## Plugin Development

### Creating a New Plugin

1. **Implement the Plugin Interface**
   ```go
   type MyPlugin struct {
       *plugins.BasePlugin
   }
   
   func NewMyPlugin() *MyPlugin {
       base := plugins.NewBasePlugin("my-plugin", "1.0.0", []string{})
       return &MyPlugin{BasePlugin: base}
   }
   ```

2. **Add HTTP Routes**
   ```go
   func (p *MyPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
       p.AddRoute(core.Route{
           Method:  "GET",
           Path:    "/my-endpoint",
           Handler: p.handleMyEndpoint,
           Auth:    core.AuthRequirement{Required: false},
       })
       return p.BasePlugin.Initialize(ctx, config)
   }
   ```

3. **Implement Route Handlers**
   ```go
   func (p *MyPlugin) handleMyEndpoint(w http.ResponseWriter, r *http.Request) {
       response := map[string]interface{}{
           "message": "Hello from my plugin!",
           "status":  "success",
       }
       
       w.Header().Set("Content-Type", "application/json")
       json.NewEncoder(w).Encode(response)
   }
   ```

### Plugin Guidelines

- Follow the single responsibility principle
- Implement proper error handling
- Use structured logging
- Include comprehensive tests
- Document all public APIs
- Handle graceful shutdown

## Testing

### Testing Strategy

We use a multi-layered testing approach:

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete user workflows
4. **Performance Tests**: Benchmark critical paths
5. **Security Tests**: Validate security measures

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/plugins/...

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Writing Tests

Follow these conventions:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

## Pull Request Process

### Before Submitting

1. **Create an Issue**: For new features or significant changes
2. **Fork the Repository**: Work in your own fork
3. **Create a Feature Branch**: Use descriptive branch names
4. **Write Tests**: Ensure good test coverage
5. **Update Documentation**: Include relevant documentation updates

### PR Requirements

- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] PR description explains changes
- [ ] Linked to relevant issues

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

## Code Style Guidelines

### Go Code Style

We follow the official Go style guidelines with some additions:

1. **Formatting**: Use `gofmt` and `goimports`
2. **Naming**: Follow Go naming conventions
3. **Comments**: Public APIs must have doc comments
4. **Error Handling**: Always handle errors explicitly
5. **Context**: Use context.Context for cancellation

### Linting

Run the linter before submitting:

```bash
golangci-lint run
```

### Security

Run security checks:

```bash
gosec ./...
```

## Documentation

### Types of Documentation

1. **Code Comments**: Explain complex logic
2. **API Documentation**: Document all public APIs
3. **README Updates**: Keep README current
4. **Examples**: Provide usage examples
5. **Architecture Docs**: Document design decisions

### Documentation Standards

- Use clear, concise language
- Include code examples
- Keep documentation current
- Follow markdown standards
- Use consistent terminology

## Security

### Security Guidelines

1. **Input Validation**: Validate all inputs
2. **Authentication**: Implement proper auth checks
3. **Authorization**: Enforce access controls
4. **Encryption**: Use strong encryption
5. **Secrets**: Never commit secrets
6. **Dependencies**: Keep dependencies updated

### Reporting Security Issues

Please report security vulnerabilities privately by emailing [security@noplacelike.dev]. Do not create public issues for security vulnerabilities.

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Discord**: Real-time chat (link in README)

### Development Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [WebSocket Guide](https://github.com/gorilla/websocket)
- [Plugin Architecture](./docs/plugins.md)

## Recognition

Contributors are recognized in several ways:

- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- GitHub contributor statistics
- Special recognition for significant contributions

Thank you for contributing to NoPlaceLike! 🚀