# Cloud Torrent Architecture

This document describes the high-level architecture of Cloud Torrent, explaining how the different components interact.

## System Overview

![Cloud Torrent Architecture](https://docs.google.com/drawings/d/1ekyeGiehwQRyi6YfFA4_tQaaEpUaS8qihwJ-s3FT_VU/pub?w=606&h=305)

Cloud Torrent consists of the following main components:

1. **Engine**: The core torrent downloading functionality (`engine/` directory)
2. **Server**: The HTTP server and API endpoints (`server/` directory)
3. **Static Files**: The web UI (`static/` directory)
4. **Main Application**: Entry point and configuration (`main.go`)

## Component Interactions

### Main Application Flow

1. The user starts the application with configuration options
2. The application initializes the server component with these options
3. The server component initializes the engine component
4. The server starts an HTTP server that serves:
   - The static web UI
   - API endpoints for controlling torrents
   - File access endpoints for downloaded content
   - Search functionality

### Torrent Engine

The engine component (`engine/` directory) is responsible for:

- Managing torrent downloads
- Handling torrent metadata
- Tracking download progress
- Managing files on disk

Key files:
- `engine.go`: Main engine implementation
- `torrent.go`: Torrent type representation
- `config.go`: Engine configuration

The engine uses the [anacrolix/torrent](https://github.com/anacrolix/torrent) library to handle the BitTorrent protocol.

### Server Component

The server component (`server/` directory) provides:

- HTTP endpoints for the UI and API
- Authentication
- File serving
- Real-time updates via WebSockets (using velox)
- Torrent searching via scraper

Key files:
- `server.go`: Main server implementation and HTTP handlers
- `server_api.go`: API endpoint implementations
- `server_files.go`: File serving functionality
- `server_search.go`: Torrent search implementation
- `server_stats.go`: System statistics tracking

### Frontend UI

The web UI is served from the `static/` directory and provides:

- Torrent management interface
- File browser
- Search interface
- Stats display

## Data Flow

1. **Adding a Torrent**:
   - User submits a magnet link or torrent file via the UI
   - Server API receives the request
   - API calls the engine to add the torrent
   - Engine starts downloading the torrent
   - Server provides real-time updates to the UI

2. **Downloading Files**:
   - Engine downloads torrent data to the configured directory
   - Server monitors download progress
   - UI displays progress in real-time

3. **Accessing Files**:
   - User browses files via the UI
   - Server lists files from the download directory
   - User requests a file
   - Server streams the file via HTTP

## Communication Patterns

Cloud Torrent uses several communication patterns:

1. **HTTP REST API**: For torrent management commands
2. **Real-time Updates**: Using velox for pushing updates to the UI
3. **File Streaming**: For serving downloaded content
4. **External APIs**: For torrent searching via the scraper component 