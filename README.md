# NoPlaceLike 🚀💻

A distributed virtual OS for sharing resources across your network.

## Overview

NoPlaceLike is your virtual distributed operating system for effortlessly streaming clipboard data, files, music, and more across devices—wirelessly and seamlessly! 🌐✨

Built in Go for excellent performance, stability, and low resource consumption, NoPlaceLike provides a complete network resource sharing solution with a clean web interface accessible from any browser.

## Features

- **Clipboard Sharing** ✂️📋  
  Copy and paste between devices with a simple click. Your clipboard travels with you!

- **File Sharing** 📂🚀  
  Effortlessly upload and download files across your network with a sleek, intuitive web interface.

- **Audio Streaming** 🎵📡  
  Stream music live from your server to any client device. Enjoy real-time controls and seamless playback without downloading.

- **Distributed Resources** 🤖🌐  
  Share computational and network resources across devices, turning your local network into a collaborative powerhouse.

- **Cross-Platform** 💻📱  
  Run on Windows, macOS, Linux. Access from any device with a browser.

## Installation

### From Source

1. Clone the repository:
```bash
git clone https://github.com/nathfavour/noplacelike.go.git
cd noplacelike.go
```

2. Build the application:
```bash
go build -o noplacelike
```

3. Run the application:
```bash
./noplacelike
```

### Using Go Install

```bash
go install github.com/nathfavour/noplacelike.go@latest
```

## Usage

### Basic Usage

Simply run the application:

```bash
noplacelike
```

This will start the server on all network interfaces (0.0.0.0) on port 8000.

### Command Line Options

```
noplacelike [flags]
```

Available flags:
- `--host`: Host address to bind to (default "0.0.0.0")
- `-p, --port`: Port to listen on (default 8000)
- `-u, --upload-folder`: Custom folder for uploads
- `-d, --download-folder`: Custom folder for downloads

### Configuration

NoPlaceLike stores its configuration in `~/.noplacelike.json`. You can modify this file directly or use the configuration commands:

```bash
# View current configuration
noplacelike config

# Add an audio directory
noplacelike config -a "/path/to/music"

# Clear all audio directories
noplacelike config -c
```

### Accessing the Web Interface

After starting the server, you'll see QR codes and URLs in the terminal. Scan a QR code or enter the URL in your browser to access the web interface.

The interface provides several key sections:
- **Clipboard Sharing**: Share text between devices
- **File Sharing**: Upload and download files
- **Audio Streaming**: Stream audio files from configured directories

## Architecture

NoPlaceLike is built with a modular architecture:

- **CLI Interface**: Built with [Cobra](https://github.com/spf13/cobra)
- **Web Server**: Powered by [Gin](https://github.com/gin-gonic/gin)
- **Configuration Management**: JSON-based persistent configuration
- **Resource Sharing**: Multiple specialized modules for different resource types

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

NoPlaceLike was inspired by the Python-based noplacelikehomely project, reimplemented in Go for better performance and network capabilities.
