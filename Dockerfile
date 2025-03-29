FROM golang:1.23-alpine as builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloud-torrent .

# Create final image
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1000 cloudtorrent && \
    adduser -u 1000 -G cloudtorrent -s /bin/sh -D cloudtorrent

# Create required directories with appropriate permissions
RUN mkdir -p /downloads /config && \
    chown -R cloudtorrent:cloudtorrent /downloads /config

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/cloud-torrent .

# Switch to non-root user
USER cloudtorrent

# Set volumes
VOLUME ["/downloads", "/config"]

# Expose port
EXPOSE 3000

# Run the application with proper defaults
ENTRYPOINT ["/app/cloud-torrent"]
CMD ["--port", "3000", "--config-path", "/config/cloud-torrent.json"]
