# Development Guide

This guide provides information for developers who want to contribute to Cloud Torrent or customize it for their own needs.

## Setting Up Development Environment

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.13 or higher recommended)
- [Git](https://git-scm.com/downloads)
- Any modern web browser for testing

### Getting the Source Code

Clone the repository:

```bash
git clone https://github.com/jpillora/cloud-torrent.git
cd cloud-torrent
```

### Building From Source

Build the application:

```bash
go build -o cloud-torrent
```

Run the application:

```bash
./cloud-torrent
```

For development, you may want to enable hot reloading using tools like [air](https://github.com/cosmtrek/air) or [realize](https://github.com/oxequa/realize).

### Running Tests

Run the Go tests:

```bash
go test ./...
```

## Project Architecture

Cloud Torrent follows a modular architecture with clear separation of concerns:

1. **Engine**: Core BitTorrent functionality
2. **Server**: HTTP API and file serving
3. **Frontend**: Web UI for interaction

For more detail on the code structure, see the [Code Structure](./code-structure.md) documentation.

## Adding Features

### Adding a New API Endpoint

1. Identify the appropriate server file (usually `server_api.go`)
2. Add a new handler function
3. Register the endpoint in the `handle()` function in `server.go`

Example of adding a new endpoint:

```go
// In server_api.go
func (s *Server) handleNewFeature(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

// In server.go, add to the handle() function
case "/api/new-feature":
    s.handleNewFeature(w, r)
```

### Adding Engine Functionality

1. Identify the feature you want to add
2. Implement it in the Engine package
3. Expose it through the Server API if needed

### Modifying the UI

The UI files are embedded in the binary but are located in the `static/files/` directory for development.

1. Modify the HTML, CSS, or JavaScript files
2. Test your changes by running the application locally
3. Build the application to embed the updated files

## Coding Guidelines

### Go Code

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Add comments for exported functions and types
- Write tests for new functionality

### JavaScript Code

- Follow a consistent style
- Minimize external dependencies
- Ensure compatibility with recent browsers

## Release Process

1. Update version number in appropriate files
2. Run tests to ensure everything works
3. Create a release tag
4. Build for all platforms:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o cloud-torrent-linux-amd64
   GOOS=darwin GOARCH=amd64 go build -o cloud-torrent-darwin-amd64
   GOOS=windows GOARCH=amd64 go build -o cloud-torrent-windows-amd64.exe
   ```
5. Create a GitHub release with the built binaries

## Debugging

### Common Issues

#### Build Errors

- Make sure you have the latest Go version
- Ensure all dependencies are available
- Check for breaking changes in dependencies

#### Runtime Errors

- Check the logs output by Cloud Torrent
- Verify that your firewall isn't blocking required ports
- Ensure the download directory is writable

### Debugging Tools

- Use `debug` statements in your code
- Set higher verbosity with `-v` flag when running
- Use browser developer tools to debug UI issues

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request

When submitting a pull request:
- Clearly describe the problem and solution
- Include any relevant issue numbers
- Ensure all tests pass
- Keep code quality high

## Community

Join the discussion:
- [GitHub Issues](https://github.com/jpillora/cloud-torrent/issues)
- Ask questions with the tag 'cloud-torrent' on Stack Overflow 