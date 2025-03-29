# Code Structure

This document provides a guide to the code structure of Cloud Torrent, helping developers understand where to find specific functionality and how the code is organized.

## Directory Structure

```
cloud-torrent/
├── .git/               # Git repository data
├── .github/            # GitHub-specific files
├── engine/             # Torrent engine implementation
├── server/             # HTTP server and API
├── static/             # Web UI assets
├── .gitignore          # Git ignore patterns
├── CONTRIBUTING.md     # Contribution guidelines
├── LICENSE             # License information
├── README.md           # Project readme
├── go.mod              # Go module definition
├── go.sum              # Go dependency checksums
└── main.go             # Application entry point
```

## Key Components

### Main Application (`main.go`)

The entry point for the application. This file:
- Imports the server package
- Sets up command-line options using the `opts` package
- Initializes and runs the server

### Engine (`engine/`)

The torrent engine implementation that handles downloading torrents.

**Key Files:**
- `engine.go`: Main engine implementation with methods for managing torrents
- `torrent.go`: Defines the `Torrent` and `File` types for representing torrents and their files
- `config.go`: Configuration options for the engine

### Server (`server/`)

The HTTP server that provides the web interface and API.

**Key Files:**
- `server.go`: Main server implementation, HTTP handlers, and state management
- `server_api.go`: API endpoint implementations for controlling torrents
- `server_files.go`: File serving functionality for browsing and downloading files
- `server_search.go`: Search functionality implementation
- `server_stats.go`: System statistics tracking

### Static Files (`static/`)

The web UI assets and embedded file system.

**Key Files:**
- `static.go`: Go code for embedding and serving static files
- `files/`: The actual UI files (HTML, CSS, JavaScript)

## Code Walkthrough

### Application Initialization

1. `main.go` - Creates a server instance, configures it from command-line options, and runs it
2. `server.Run()` - Initializes the server and engine components
3. Server starts handling HTTP requests and serving the UI

### Adding and Managing Torrents

1. UI/API sends a request to add a torrent
2. `server_api.go` handles the request and calls `engine.NewMagnet()` or `engine.NewTorrent()`
3. `engine.go` adds the torrent to the client and creates a tracking struct
4. A goroutine monitors the download progress and updates state
5. The state updates are pushed to clients via WebSockets using the velox library

### File Serving

1. User requests a file via the UI
2. `server_files.go` handles the request
3. The file is served from the download directory using Go's `http.ServeContent`

### Search Functionality

1. User enters a search query in the UI
2. Request is sent to the search API
3. `server_search.go` handles the request and forwards it to the configured scraper
4. The scraper fetches search results from external sites
5. Results are returned to the UI

## Detailed Component Descriptions

### Engine Component

The engine component is built on top of the [anacrolix/torrent](https://github.com/anacrolix/torrent) library and provides:

- Torrent management (add, start, stop, delete)
- Progress tracking
- File selection within torrents
- Persistent configuration

Key methods:
- `NewMagnet()`: Add a torrent from a magnet link
- `NewTorrent()`: Add a torrent from a .torrent file
- `StartTorrent()`: Start downloading a torrent
- `StopTorrent()`: Stop downloading a torrent
- `DeleteTorrent()`: Remove a torrent

### Server Component

The server component provides:

- HTTP endpoints for the UI and API
- WebSocket updates for real-time UI updates
- File serving for downloads
- Search capabilities

Key methods:
- `Run()`: Start the server
- `reconfigure()`: Update configuration
- `handle()`: Main HTTP handler
- Various API handlers for different endpoints

### Integration Points

The main integration points between components are:

1. Server creates and configures the Engine
2. Server calls Engine methods in response to API requests
3. Server reads Engine state to provide updates to the UI
4. Server serves files that the Engine has downloaded

## Extension Points

When extending Cloud Torrent, these are the main areas to consider:

1. **Adding API endpoints**: Extend the server's API handlers in `server_api.go`
2. **Adding UI features**: Modify the static files in the `static/files/` directory
3. **Adding torrent functionality**: Extend the engine in `engine.go`
4. **Adding configuration options**: Update both engine config and server config structures 