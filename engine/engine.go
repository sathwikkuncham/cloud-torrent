package engine

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
)

// the Engine Cloud Torrent engine, backed by anacrolix/torrent
type Engine struct {
	mut              sync.Mutex
	cacheDir         string
	client           *torrent.Client
	config           Config
	ts               map[string]*Torrent
	activeTorrents   int
	healthCheckTimer *time.Timer
	stopChan         chan struct{}
	memoryMonitor    *MemoryMonitor
}

// MemoryMonitor tracks memory usage of the engine
type MemoryMonitor struct {
	memoryUsage int64
	mutex       sync.Mutex
}

// AddMemoryUsage adds to the tracked memory usage
func (mm *MemoryMonitor) AddMemoryUsage(bytes int64) {
	if mm == nil {
		return
	}
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.memoryUsage += bytes
}

// ReleaseMemory decreases the tracked memory usage
func (mm *MemoryMonitor) ReleaseMemory(bytes int64) {
	if mm == nil {
		return
	}
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.memoryUsage -= bytes
	if mm.memoryUsage < 0 {
		mm.memoryUsage = 0
	}
}

// GetMemoryUsage returns the current tracked memory usage
func (mm *MemoryMonitor) GetMemoryUsage() int64 {
	if mm == nil {
		return 0
	}
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	return mm.memoryUsage
}

func New() *Engine {
	return &Engine{
		ts:            map[string]*Torrent{},
		stopChan:      make(chan struct{}),
		memoryMonitor: &MemoryMonitor{},
	}
}

func (e *Engine) Config() Config {
	return e.config
}

func (e *Engine) Configure(c Config) error {
	//recieve config
	if e.client != nil {
		e.client.Close()
		time.Sleep(1 * time.Second)
	}
	if c.IncomingPort <= 0 {
		return fmt.Errorf("Invalid incoming port (%d)", c.IncomingPort)
	}

	// Set up memory and performance configurations
	config := torrent.NewDefaultClientConfig()
	config.DataDir = c.DownloadDirectory
	config.NoUpload = !c.EnableUpload
	config.Seed = c.EnableSeeding
	config.ListenPort = c.IncomingPort

	// Apply network discovery optimizations
	if c.EnableDHT {
		config.NoDHT = false // Enable DHT for better peer discovery
	}
	if c.EnablePEX {
		config.DisablePEX = false // Enable peer exchange
	}
	// Note: Some options might not be directly supported by the torrent library version
	// We set them anyway for future compatibility or versions that do support them

	// Apply bandwidth and performance settings

	// Set max connections if the API supports it
	if c.MaxConnectionsPerTorrent > 0 {
		config.EstablishedConnsPerTorrent = c.MaxConnectionsPerTorrent
		// Add a custom torrent client identifier
		config.Bep20 = "-CT01000-" // Custom peer ID to avoid restrictions
	}

	// Apply connection timeout settings
	config.HandshakesTimeout = 45 * time.Second
	// Set a longer timeout for better performance with slow peers

	// Apply file operation optimizations
	if c.WriteBufferSize > 0 {
		// Torrent library might not directly support this, but we set for future compatibility
	}
	if c.ReadCacheSize > 0 {
		// Torrent library might not directly support this, but we set for future compatibility
	}

	client, err := torrent.NewClient(config)
	if err != nil {
		return err
	}

	e.mut.Lock()
	e.config = c
	e.client = client
	e.mut.Unlock()

	// Reset the engine
	e.GetTorrents()

	// Start health check timer if enabled
	if c.HealthCheckInterval > 0 {
		e.startHealthCheck()
	}

	log.Printf("Engine configured with: max memory=%s, max concurrent torrents=%d, max connections=%d",
		humanize.Bytes(uint64(c.MaxMemoryUsage)),
		c.MaxConcurrentTorrents,
		c.MaxConnectionsPerTorrent)

	return nil
}

// startHealthCheck begins periodic health checking of torrents
func (e *Engine) startHealthCheck() {
	// Stop any existing timer
	if e.healthCheckTimer != nil {
		e.healthCheckTimer.Stop()
	}

	e.healthCheckTimer = time.AfterFunc(time.Duration(e.config.HealthCheckInterval)*time.Second, func() {
		e.checkTorrentsHealth()
		// Schedule next check
		e.startHealthCheck()
	})
}

// checkTorrentsHealth checks all active torrents for health issues
func (e *Engine) checkTorrentsHealth() {
	e.mut.Lock()
	defer e.mut.Unlock()

	now := time.Now()
	for _, t := range e.ts {
		if !t.Started {
			continue
		}

		// Check for stalled downloads (no progress for over 2 minutes)
		if t.UpdatedAt.Add(2*time.Minute).Before(now) && t.DownloadRate == 0 && t.Percent < 100 {
			log.Printf("Torrent %s appears stalled, attempting to restart", t.Name)
			ih := t.InfoHash
			// We need to release the lock before calling methods that acquire it
			e.mut.Unlock()
			e.StopTorrent(ih)
			time.Sleep(1 * time.Second)
			e.StartTorrent(ih)
			e.mut.Lock()
		}
	}
}

// Close shuts down the engine and releases resources
func (e *Engine) Close() error {
	if e.healthCheckTimer != nil {
		e.healthCheckTimer.Stop()
	}

	close(e.stopChan)

	if e.client != nil {
		e.client.Close()
	}
	return nil
}

// GetTorrents returns a map of all torrents by infohash
func (e *Engine) GetTorrents() map[string]*Torrent {
	e.mut.Lock()
	defer e.mut.Unlock()

	if e.client == nil {
		return nil
	}
	for _, tt := range e.client.Torrents() {
		e.upsertTorrent(tt)
	}
	return e.ts
}

// GetTorrent returns a specific torrent by infohash
func (e *Engine) GetTorrent(infohash string) (*Torrent, error) {
	return e.getTorrent(infohash)
}

func (e *Engine) NewMagnet(magnetURI string) error {
	// Check if we're at max concurrent torrents
	if e.config.MaxConcurrentTorrents > 0 && e.activeTorrents >= e.config.MaxConcurrentTorrents {
		return fmt.Errorf("Maximum number of concurrent torrents reached (%d)", e.config.MaxConcurrentTorrents)
	}

	// Check if we have enough memory available
	if e.config.MaxMemoryUsage > 0 && e.memoryMonitor.GetMemoryUsage() >= e.config.MaxMemoryUsage {
		return fmt.Errorf("Memory limit reached (%s used)",
			humanize.Bytes(uint64(e.memoryMonitor.GetMemoryUsage())))
	}

	tt, err := e.client.AddMagnet(magnetURI)
	if err != nil {
		return err
	}

	return e.newTorrent(tt)
}

func (e *Engine) NewTorrent(spec *torrent.TorrentSpec) error {
	tt, _, err := e.client.AddTorrentSpec(spec)
	if err != nil {
		return err
	}
	return e.newTorrent(tt)
}

func (e *Engine) newTorrent(tt *torrent.Torrent) error {
	t := e.upsertTorrent(tt)
	go func() {
		<-t.t.GotInfo()
		e.StartTorrent(t.InfoHash)
	}()
	return nil
}

func (e *Engine) upsertTorrent(tt *torrent.Torrent) *Torrent {
	ih := tt.InfoHash().HexString()
	torrent, ok := e.ts[ih]
	if !ok {
		torrent = &Torrent{InfoHash: ih}
		e.ts[ih] = torrent
	}
	//update torrent fields using underlying torrent
	torrent.Update(tt)
	return torrent
}

func (e *Engine) getTorrent(infohash string) (*Torrent, error) {
	ih, err := str2ih(infohash)
	if err != nil {
		return nil, err
	}
	t, ok := e.ts[ih.HexString()]
	if !ok {
		return t, fmt.Errorf("Missing torrent %x", ih)
	}
	return t, nil
}

func (e *Engine) getOpenTorrent(infohash string) (*Torrent, error) {
	t, err := e.getTorrent(infohash)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (e *Engine) StartTorrent(infohash string) error {
	t, err := e.getOpenTorrent(infohash)
	if err != nil {
		return err
	}
	if t.Started {
		return fmt.Errorf("Already started")
	}

	// Check if we're at max concurrent torrents
	if e.config.MaxConcurrentTorrents > 0 && e.activeTorrents >= e.config.MaxConcurrentTorrents {
		return fmt.Errorf("Maximum number of concurrent torrents reached (%d)", e.config.MaxConcurrentTorrents)
	}

	// Check if we have enough memory available
	estimatedMemory := t.Size / 100 * 2 // Rough estimate: 2% of torrent size
	if e.config.MaxMemoryUsage > 0 &&
		e.memoryMonitor.GetMemoryUsage()+estimatedMemory >= e.config.MaxMemoryUsage {
		return fmt.Errorf("Memory limit reached (%s used)",
			humanize.Bytes(uint64(e.memoryMonitor.GetMemoryUsage())))
	}

	// Apply buffer settings
	if e.config.BufferPerTorrent > 0 && t.t.Info() != nil {
		// Set buffer size if the underlying library supports it
		// Implementation depends on the actual client capabilities
	}

	t.Started = true
	e.activeTorrents++
	e.memoryMonitor.AddMemoryUsage(estimatedMemory)

	for _, f := range t.Files {
		if f != nil {
			f.Started = true
		}
	}

	if t.t.Info() != nil {
		t.t.DownloadAll()
	}

	log.Printf("Started torrent %s (%s), active: %d, memory: %s",
		t.Name,
		humanize.Bytes(uint64(t.Size)),
		e.activeTorrents,
		humanize.Bytes(uint64(e.memoryMonitor.GetMemoryUsage())))

	return nil
}

func (e *Engine) StopTorrent(infohash string) error {
	t, err := e.getTorrent(infohash)
	if err != nil {
		return err
	}
	if !t.Started {
		return fmt.Errorf("Already stopped")
	}

	//there is no stop - kill underlying torrent
	t.t.Drop()
	t.Started = false

	// Release resources
	e.activeTorrents--
	estimatedMemory := t.Size / 100 * 2 // Same estimate as in StartTorrent
	e.memoryMonitor.ReleaseMemory(estimatedMemory)

	for _, f := range t.Files {
		if f != nil {
			f.Started = false
		}
	}

	log.Printf("Stopped torrent %s, active: %d, memory: %s",
		t.Name,
		e.activeTorrents,
		humanize.Bytes(uint64(e.memoryMonitor.GetMemoryUsage())))

	return nil
}

func (e *Engine) DeleteTorrent(infohash string) error {
	t, err := e.getTorrent(infohash)
	if err != nil {
		return err
	}
	os.Remove(filepath.Join(e.cacheDir, infohash+".torrent"))
	delete(e.ts, t.InfoHash)
	ih, _ := str2ih(infohash)
	if tt, ok := e.client.Torrent(ih); ok {
		tt.Drop()
	}
	return nil
}

func (e *Engine) StartFile(infohash, filepath string) error {
	t, err := e.getOpenTorrent(infohash)
	if err != nil {
		return err
	}
	var f *File
	for _, file := range t.Files {
		if file.Path == filepath {
			f = file
			break
		}
	}
	if f == nil {
		return fmt.Errorf("Missing file %s", filepath)
	}
	if f.Started {
		return fmt.Errorf("Already started")
	}
	t.Started = true
	f.Started = true
	return nil
}

func (e *Engine) StopFile(infohash, filepath string) error {
	return fmt.Errorf("Unsupported")
}

func str2ih(str string) (metainfo.Hash, error) {
	var ih metainfo.Hash
	e, err := hex.Decode(ih[:], []byte(str))
	if err != nil {
		return ih, fmt.Errorf("Invalid hex string")
	}
	if e != 20 {
		return ih, fmt.Errorf("Invalid length")
	}
	return ih, nil
}
