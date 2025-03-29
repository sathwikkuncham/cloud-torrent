# API Reference

Cloud Torrent provides a simple HTTP API for controlling torrents programmatically. This document describes the available endpoints and their usage.

## Authentication

If you've configured Cloud Torrent with `--auth`, you'll need to include Basic Authentication with all API requests:

```
Authorization: Basic <base64 encoded username:password>
```

## API Endpoints

### Torrent Management

#### Add Torrent from Magnet Link

```
POST /api/magnet
```

**Parameters:**
- `magnet` (string, required): The magnet URI to add

**Example:**
```bash
curl -X POST "http://localhost:3000/api/magnet?magnet=magnet:?xt=urn:btih:HASH&dn=Name"
```

#### Add Torrent from File

```
POST /api/torrent
```

**Parameters:**
- The torrent file should be sent as the request body

**Example:**
```bash
curl -X POST -T my-torrent.torrent "http://localhost:3000/api/torrent"
```

#### Start Torrent

```
POST /api/torrent/<infohash>/start
```

**Example:**
```bash
curl -X POST "http://localhost:3000/api/torrent/e39c91edeb3032c828217d46059feb476596eea2/start"
```

#### Stop Torrent

```
POST /api/torrent/<infohash>/stop
```

**Example:**
```bash
curl -X POST "http://localhost:3000/api/torrent/e39c91edeb3032c828217d46059feb476596eea2/stop"
```

#### Delete Torrent

```
DELETE /api/torrent/<infohash>
```

**Example:**
```bash
curl -X DELETE "http://localhost:3000/api/torrent/e39c91edeb3032c828217d46059feb476596eea2"
```

### File Management

#### Start/Stop File

```
POST /api/file/<infohash>/<filepath>/start
POST /api/file/<infohash>/<filepath>/stop
```

These endpoints allow starting or stopping individual files within a torrent.

**Example:**
```bash
curl -X POST "http://localhost:3000/api/file/e39c91edeb3032c828217d46059feb476596eea2/path/to/file.mp4/start"
```

### Search

#### Search for Torrents

```
GET /search/<provider>/<query>/<page>
```

**Parameters:**
- `provider` (string): The search provider to use
- `query` (string): The search query
- `page` (number): The page number of results

**Example:**
```bash
curl "http://localhost:3000/search/thepiratebay/ubuntu/1"
```

### Configuration

#### Get Configuration

```
GET /api/config
```

Returns the current configuration.

**Example:**
```bash
curl "http://localhost:3000/api/config"
```

#### Update Configuration

```
POST /api/config
```

Updates the configuration. The request body should be a JSON object with the configuration options to update.

**Example:**
```bash
curl -X POST -H "Content-Type: application/json" -d '{"EnableSeeding": true}' "http://localhost:3000/api/config"
```

### Stats

#### Get System Stats

```
GET /api/stats
```

Returns system statistics including CPU, memory, and disk usage.

**Example:**
```bash
curl "http://localhost:3000/api/stats"
```

## WebSocket API

Cloud Torrent also provides real-time updates via WebSocket. The main state is available at:

```
GET /sync
```

This endpoint uses the velox library to push state changes to the client. The state includes:

- Torrents information
- Download directory structure
- Configuration
- Statistics

## Response Format

Most API responses are in JSON format. Here's an example of a typical response:

```json
{
  "success": true,
  "data": {
    ...
  }
}
```

In case of errors:

```json
{
  "success": false,
  "error": "Error message"
}
```

## Error Codes

The API uses standard HTTP status codes:

- `200 OK`: Successful operation
- `400 Bad Request`: Invalid parameters
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## API Usage Examples

### Complete Download Workflow

1. Add a torrent:
```bash
curl -X POST "http://localhost:3000/api/magnet?magnet=magnet:?xt=urn:btih:HASH&dn=Name"
```

2. Get torrent status:
```bash
curl "http://localhost:3000/api/torrents"
```

3. Access the downloaded file:
```
http://localhost:3000/download/path/to/file
``` 