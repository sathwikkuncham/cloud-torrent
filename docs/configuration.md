# Configuration

Cloud Torrent can be configured in multiple ways. This document covers the available configuration options and methods.

## Configuration Methods

There are three ways to configure Cloud Torrent:

1. **Command-line arguments**: Passed when starting the application
2. **Environment variables**: Set in the shell before starting
3. **Configuration file**: Loaded from the specified JSON file

## Command-line Options

These options can be specified when running Cloud Torrent:

| Option | Short | Description | Default | Env Variable |
|--------|-------|-------------|---------|--------------|
| `--title` | `-t` | Title of this instance | `Cloud Torrent` | `TITLE` |
| `--port` | `-p` | Listening port | `3000` | `PORT` |
| `--host` | `-h` | Listening interface | `0.0.0.0` (all) | - |
| `--auth` | `-a` | Optional basic auth (user:password) | - | `AUTH` |
| `--config-path` | `-c` | Configuration file path | `cloud-torrent.json` | - |
| `--key-path` | `-k` | TLS Key file path | - | - |
| `--cert-path` | `-r` | TLS Certificate file path | - | - |
| `--log` | `-l` | Enable request logging | `false` | - |
| `--open` | `-o` | Open now with your default browser | `false` | - |

## Configuration File

Cloud Torrent uses a JSON configuration file for torrent engine settings. By default, it looks for `cloud-torrent.json` in the current directory, but you can specify a different path with the `--config-path` option.

Example `cloud-torrent.json`:

```json
{
  "DownloadDirectory": "/path/to/downloads",
  "IncomingPort": 50007,
  "EnableUpload": true,
  "EnableSeeding": false,
  "AutoStart": true
}
```

### Configuration Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `DownloadDirectory` | String | Directory to store downloaded files | `./downloads` |
| `IncomingPort` | Integer | Port for BitTorrent connections | `50007` |
| `EnableUpload` | Boolean | Allow uploading to peers | `true` |
| `EnableSeeding` | Boolean | Keep uploading after download completes | `false` |
| `AutoStart` | Boolean | Automatically start torrents when added | `true` |

## Environment Variables

Some options can be configured using environment variables:

| Variable | Description | Equivalent Flag |
|----------|-------------|----------------|
| `PORT` | Listening port | `--port` |
| `TITLE` | Title of this instance | `--title` |
| `AUTH` | Basic auth credentials | `--auth` |

## Security Recommendations

When deploying Cloud Torrent, consider these security recommendations:

1. **Always use authentication**: Set a username and password with `--auth` option
2. **Use HTTPS**: Configure with `--key-path` and `--cert-path` for TLS
3. **Run as non-root user**: If using a system service, configure it to run as a limited user
4. **Firewall access**: Restrict access to the server port

## Advanced Configuration

### Using Behind a Reverse Proxy

If you're running Cloud Torrent behind a reverse proxy like Nginx:

1. Set `--host` to `127.0.0.1` to restrict direct access
2. Configure your reverse proxy to handle TLS termination
3. Set appropriate headers for WebSocket support

Example Nginx configuration:

```nginx
server {
    listen 443 ssl;
    server_name torrent.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Search Providers Configuration

Cloud Torrent includes a scraper for torrent search. The search providers are configured internally and automatically updated from the project repository. 