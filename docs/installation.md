# Installation Guide

Cloud Torrent is designed to be easy to install and run on multiple platforms. This guide covers several installation methods.

## System Requirements

- Any operating system that supports Go (Windows, macOS, Linux, etc.)
- Sufficient disk space for downloads
- Internet connection

## Installation Methods

### Pre-built Binaries

The easiest way to install Cloud Torrent is by downloading a pre-built binary:

1. Visit the [releases page](https://github.com/jpillora/cloud-torrent/releases/latest)
2. Download the appropriate binary for your platform
3. Extract the archive if necessary
4. Run the executable

Alternatively, you can use the installation script:

```bash
curl https://i.jpillora.com/cloud-torrent! | bash
```

### Docker

To run Cloud Torrent using Docker:

```bash
docker run -d -p 3000:3000 -v /path/to/downloads:/downloads jpillora/cloud-torrent
```

Replace `/path/to/downloads` with the directory where you want to store downloaded files.

### Building from Source

If you prefer to build from source:

1. Make sure [Go](https://golang.org/dl/) is installed
2. Run:

```bash
go get -v github.com/jpillora/cloud-torrent
```

The binary will be installed in your `$GOPATH/bin` directory.

## Running Cloud Torrent

### Basic Usage

Run the executable:

```bash
cloud-torrent
```

This will start the server on the default port (3000) and open a browser window.

### Command Line Options

Cloud Torrent accepts various command line options:

```
--title, -t        Title of this instance (default Cloud Torrent, env TITLE)
--port, -p         Listening port (default 3000, env PORT)
--host, -h         Listening interface (default all)
--auth, -a         Optional basic auth in form 'user:password' (env AUTH)
--config-path, -c  Configuration file path (default cloud-torrent.json)
--key-path, -k     TLS Key file path
--cert-path, -r    TLS Certificate file path
--log, -l          Enable request logging
--open, -o         Open now with your default browser
```

### Example Configurations

#### Changing Port

```bash
cloud-torrent --port 8080
```

#### Adding Authentication

```bash
cloud-torrent --auth user:password
```

#### Using HTTPS

```bash
cloud-torrent --cert-path /path/to/cert.pem --key-path /path/to/key.pem
```

## Running as a Service

### Linux (systemd)

Create a file at `/etc/systemd/system/cloud-torrent.service`:

```ini
[Unit]
Description=Cloud Torrent
After=network.target

[Service]
ExecStart=/usr/local/bin/cloud-torrent --port 3000
Restart=always
User=nobody
Group=nogroup
WorkingDirectory=/var/lib/cloud-torrent
Environment=PORT=3000

[Install]
WantedBy=multi-user.target
```

Then:
```bash
mkdir -p /var/lib/cloud-torrent
systemctl enable cloud-torrent
systemctl start cloud-torrent
```

### Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3'
services:
  cloud-torrent:
    image: jpillora/cloud-torrent
    restart: always
    ports:
      - 3000:3000
    volumes:
      - ./downloads:/downloads
```

Run with:

```bash
docker-compose up -d
```

## Verification

After installation, you can verify that Cloud Torrent is running by:

1. Opening a web browser
2. Navigating to `http://localhost:3000` (or your configured port)
3. Seeing the Cloud Torrent interface 