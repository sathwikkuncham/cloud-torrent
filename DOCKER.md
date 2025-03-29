# Running Cloud Torrent in Docker

This document explains how to run the improved Cloud Torrent application in Docker.

## Quick Start

1. Make sure you have Docker and Docker Compose installed on your system.

2. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/cloud-torrent.git
   cd cloud-torrent
   ```

3. Create required directories:
   ```bash
   mkdir -p downloads config
   ```

4. Copy the sample configuration:
   ```bash
   cp config-sample.json config/cloud-torrent.json
   ```

5. Build and start the Docker container:
   ```bash
   docker-compose up -d
   ```

6. Access the Cloud Torrent web interface at http://localhost:3000

## Configuration

The Docker container is configured to use:
- Port 3000 for the web interface
- Port 50007 for incoming torrent connections
- `./downloads` directory for storing downloaded files
- `./config` directory for storing configuration

### Memory and Resource Limits

By default, the Docker container is limited to:
- 512MB of RAM
- 1 CPU core

You can adjust these limits in the `docker-compose.yml` file.

The sample configuration file includes:
- Maximum memory usage of 400MB
- Maximum of 3 concurrent torrents
- 4MB buffer per torrent
- Health checks every 30 seconds
- Automatic retry for failed downloads

### Testing Memory Management

To test the memory management features:
1. Try adding several large torrents simultaneously
2. Observe the memory usage with `docker stats cloud-torrent`
3. Check the logs with `docker logs cloud-torrent`

You should see log messages about memory limits and throttling being applied.

## API Endpoints

The improved version includes new API endpoints:

### Detailed Torrent Status
```
POST /api/status
Body: [torrent infohash]
```

### System Health Status
```
POST /api/health
```

## Troubleshooting

If you encounter issues:

1. Check the logs:
   ```bash
   docker logs cloud-torrent
   ```

2. Increase log verbosity by setting the `--log` flag in the CMD in Dockerfile

3. Restart the container:
   ```bash
   docker-compose restart
   ```

4. If you need to reset everything:
   ```bash
   docker-compose down
   rm -rf downloads/* config/*
   cp config-sample.json config/cloud-torrent.json
   docker-compose up -d
   ``` 