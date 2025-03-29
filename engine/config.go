package engine

type Config struct {
	// Basic configuration
	AutoStart         bool
	DisableEncryption bool
	DownloadDirectory string
	EnableUpload      bool
	EnableSeeding     bool
	IncomingPort      int

	// Memory management
	MaxMemoryUsage        int64 // Maximum memory usage in bytes (0 = unlimited)
	MaxConcurrentTorrents int   // Maximum number of torrents to download simultaneously (0 = unlimited)
	BufferPerTorrent      int64 // Buffer size per torrent in bytes

	// Performance tuning
	MaxConnectionsPerTorrent int   // Maximum number of connections per torrent (0 = use default)
	MaxDownloadRate          int64 // Maximum download rate in bytes/sec (0 = unlimited)
	MaxUploadRate            int64 // Maximum upload rate in bytes/sec (0 = unlimited)
	WriteBufferSize          int   // Write buffer size in KB
	ReadCacheSize            int   // Read cache size in MB

	// Network discovery and optimization
	ListenInterfaces string // Network interfaces to listen on (empty = all)
	EnableDHT        bool   // Enable DHT for peer discovery
	EnablePEX        bool   // Enable Peer Exchange
	EnableLPD        bool   // Enable Local Peer Discovery
	EnableUPnP       bool   // Enable UPnP port mapping
	EnableNATPMP     bool   // Enable NAT-PMP port mapping

	// Reliability settings
	EnableAutoRetry     bool    // Auto retry failed downloads
	MaxRetries          int     // Maximum number of retries per chunk
	RetryBackoffFactor  float32 // Exponential backoff factor for retries
	HealthCheckInterval int     // Health check interval in seconds (0 = disabled)
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		AutoStart:     true,
		EnableUpload:  true,
		EnableSeeding: true,
		IncomingPort:  50007,

		// Memory management defaults
		MaxMemoryUsage:        2 * 1024 * 1024 * 1024, // 2GB default limit
		MaxConcurrentTorrents: 5,
		BufferPerTorrent:      4 * 1024 * 1024, // 4MB per torrent

		// Performance defaults
		MaxConnectionsPerTorrent: 50,
		WriteBufferSize:          256, // 256KB
		ReadCacheSize:            64,  // 64MB

		// Network discovery defaults
		ListenInterfaces: "0.0.0.0",
		EnableDHT:        true,
		EnablePEX:        true,
		EnableLPD:        true,
		EnableUPnP:       true,
		EnableNATPMP:     true,

		// Reliability defaults
		EnableAutoRetry:     true,
		MaxRetries:          3,
		RetryBackoffFactor:  1.5,
		HealthCheckInterval: 30, // 30 seconds
	}
}
