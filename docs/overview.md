# Cloud Torrent Overview

Cloud Torrent is a self-hosted remote torrent client, written in Go (golang). It provides a clean web UI to start torrents remotely, which are downloaded to the server's local disk. Users can then retrieve or stream these files via HTTP.

## Key Features

- **Single Binary**: The entire application is compiled into a single executable file, making deployment simple.
- **Cross Platform**: Can run on Windows, macOS, Linux, and other platforms that Go supports.
- **Embedded Torrent Search**: Built-in search functionality to find torrents without leaving the application.
- **Real-time Updates**: The interface updates in real-time to show download progress and other statistics.
- **Mobile-friendly**: The web interface is designed to work well on mobile devices.
- **Fast Content Server**: Uses Go's efficient `http.ServeContent` to serve files to clients.
- **HTTP Authentication**: Optional basic authentication to protect your instance.
- **HTTPS Support**: Can be configured with TLS certificates for secure access.
- **Memory Management**: Smart memory usage tracking and limits to prevent RAM exhaustion.
- **Download Reliability**: Health monitoring and automatic recovery of stalled downloads.
- **Resource Control**: Configurable limits for bandwidth, connections, and concurrent downloads.

## Use Cases

- Self-hosted alternative to online torrent services
- Remote downloading of large files to a server
- Media server for streaming downloaded content
- Automated downloading via the API

## Project Status

Cloud Torrent is being actively developed with new features planned for future releases. The project is currently in version 0.X.Y and is considered stable for production use, though it continues to evolve.

## Core Components

The application consists of several key components:

1. **Torrent Engine**: The backend engine that handles the actual torrent downloading, powered by the anacrolix/torrent library
2. **Web Server**: Provides the HTTP interface for managing torrents and accessing files
3. **Search API**: Connects to various torrent search providers to enable built-in search
4. **File System API**: Manages the downloaded files and provides access to them

Each of these components is detailed further in the [Architecture](./architecture.md) documentation.

## System Requirements

- Memory: 2GB+ recommended (configurable memory limits)
- Disk Space: Depends on the size of torrents you download
- CPU: Modern multi-core processor recommended for handling multiple concurrent downloads
- Network: Stable internet connection

## Recent Improvements

We've recently enhanced Cloud Torrent with significant reliability improvements:
- Improved memory management with configurable limits
- Detection and recovery of stalled downloads
- Better metadata handling and communication
- Enhanced file serving for large files
- Detailed status reporting

See the [Improvements](./improvements.md) document for details about these enhancements. 