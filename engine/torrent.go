package engine

import (
	"log"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/dustin/go-humanize"
)

// TorrentStatus represents the health status of a torrent
type TorrentStatus int

const (
	TorrentStatusUnknown TorrentStatus = iota
	TorrentStatusHealthy
	TorrentStatusSlow
	TorrentStatusStalled
	TorrentStatusError
)

// TorrentError tracks errors encountered during torrent operations
type TorrentError struct {
	Time    time.Time
	Message string
}

type Torrent struct {
	//anacrolix/torrent
	InfoHash   string
	Name       string
	Loaded     bool
	Downloaded int64
	Size       int64
	Files      []*File
	//cloud torrent
	Started      bool
	Dropped      bool
	Percent      float32
	DownloadRate float32
	t            *torrent.Torrent
	UpdatedAt    time.Time

	// Enhanced tracking
	Status          TorrentStatus
	Errors          []TorrentError
	LastProgress    time.Time
	RetryCount      int
	MetadataLoaded  bool
	MetadataPercent float32
	HealthChecks    int
	BytesLastCheck  int64
	PeersConnected  int
	PeersTotal      int

	// Mutex for updates to this torrent
	Mu sync.Mutex
}

type File struct {
	//anacrolix/torrent
	Path      string
	Size      int64
	Chunks    int
	Completed int
	//cloud torrent
	Started bool
	Percent float32
	f       *torrent.File

	// Enhanced tracking
	Priority    int   // Download priority (higher = more important)
	RetryCount  int   // Number of retry attempts for this file
	LastError   error // Last error encountered while downloading this file
	BytesPerSec int64 // Current download rate for this specific file
}

func (torrent *Torrent) Update(t *torrent.Torrent) {
	torrent.Mu.Lock()
	defer torrent.Mu.Unlock()

	torrent.Name = t.Name()
	torrent.Loaded = t.Info() != nil

	// Update peer information
	torrent.PeersConnected = t.Stats().ActivePeers
	torrent.PeersTotal = t.Stats().TotalPeers

	// Update metadata status if torrent is not fully loaded
	if !torrent.Loaded && !torrent.MetadataLoaded {
		if t.Info() != nil {
			torrent.MetadataLoaded = true
			torrent.MetadataPercent = 100
			log.Printf("Metadata fully loaded for torrent: %s", torrent.Name)
		} else {
			// Set a simple metadata progress indicator
			stats := t.Stats()
			if stats.ActivePeers > 0 {
				torrent.MetadataPercent = 50
			} else {
				torrent.MetadataPercent = 1
			}
		}
	}

	if torrent.Loaded {
		torrent.updateLoaded(t)
	}
	torrent.t = t
}

func (torrent *Torrent) updateLoaded(t *torrent.Torrent) {
	prevBytes := torrent.Downloaded
	torrent.Size = t.Length()
	totalChunks := 0
	totalCompleted := 0

	tfiles := t.Files()
	if len(tfiles) > 0 && torrent.Files == nil {
		torrent.Files = make([]*File, len(tfiles))
	}
	//merge in files
	for i, f := range tfiles {
		path := f.Path()
		file := torrent.Files[i]
		if file == nil {
			file = &File{
				Path:     path,
				Priority: 1, // Default priority
			}
			torrent.Files[i] = file
		}
		chunks := f.State()

		file.Size = f.Length()
		file.Chunks = len(chunks)
		completed := 0
		for _, p := range chunks {
			if p.Complete {
				completed++
			}
		}
		file.Completed = completed
		file.Percent = percent(int64(file.Completed), int64(file.Chunks))
		file.f = f

		// Calculate file-specific download rate
		if torrent.UpdatedAt.IsZero() {
			file.BytesPerSec = 0
		} else {
			dt := time.Since(torrent.UpdatedAt).Seconds()
			bytesProgress := int64(float64(file.Size) * (float64(file.Percent) / 100.0))
			prevBytesProgress := int64(float64(file.Size) * (float64(file.Percent-1) / 100.0))
			if dt > 0 && bytesProgress > prevBytesProgress {
				file.BytesPerSec = int64(float64(bytesProgress-prevBytesProgress) / dt)
			}
		}

		totalChunks += file.Chunks
		totalCompleted += file.Completed
	}

	//calculate rate
	now := time.Now()
	bytes := t.BytesCompleted()
	torrent.Percent = percent(bytes, torrent.Size)

	// Update status based on download progress
	prevDownloadRate := torrent.DownloadRate

	if !torrent.UpdatedAt.IsZero() {
		dt := float32(now.Sub(torrent.UpdatedAt))
		db := float32(bytes - torrent.Downloaded)
		rate := db * (float32(time.Second) / dt)
		if rate >= 0 {
			torrent.DownloadRate = rate
		}

		// Check if we're making progress
		if bytes > prevBytes {
			torrent.LastProgress = now
			// If download speed has increased, update status
			if torrent.DownloadRate > prevDownloadRate {
				torrent.Status = TorrentStatusHealthy
			} else if torrent.DownloadRate < prevDownloadRate*0.7 {
				// Download has slowed down significantly
				torrent.Status = TorrentStatusSlow
			}
		} else if now.Sub(torrent.LastProgress) > 1*time.Minute && torrent.Percent < 100 {
			// No progress for a minute
			torrent.Status = TorrentStatusStalled
			log.Printf("Torrent %s appears stalled (no progress for %s)",
				torrent.Name,
				time.Since(torrent.LastProgress).Round(time.Second))

			// Record error
			torrent.addError("Download stalled - no progress for " +
				time.Since(torrent.LastProgress).Round(time.Second).String())
		}
	} else {
		// First update
		torrent.LastProgress = now
		torrent.Status = TorrentStatusHealthy
	}

	// Log meaningful changes in download status
	if torrent.Size > 0 && torrent.Downloaded > 0 && torrent.Percent < 100 {
		downloadedStr := humanize.Bytes(uint64(bytes))
		totalStr := humanize.Bytes(uint64(torrent.Size))
		rateStr := humanize.Bytes(uint64(torrent.DownloadRate)) + "/s"

		// Only log when there's meaningful change
		if bytes-prevBytes > 1024*1024 { // At least 1MB progress
			log.Printf("Torrent %s progress: %s/%s (%.1f%%) at %s",
				torrent.Name, downloadedStr, totalStr,
				torrent.Percent, rateStr)
		}
	}

	torrent.Downloaded = bytes
	torrent.UpdatedAt = now
}

// addError adds a new error to the torrent's error log
func (torrent *Torrent) addError(msg string) {
	// Cap errors at 10 to avoid unbounded growth
	if len(torrent.Errors) >= 10 {
		// Remove oldest error
		torrent.Errors = torrent.Errors[1:]
	}

	torrent.Errors = append(torrent.Errors, TorrentError{
		Time:    time.Now(),
		Message: msg,
	})
}

// HasRecentError checks if the torrent has encountered errors in the last few minutes
func (torrent *Torrent) HasRecentError() bool {
	if len(torrent.Errors) == 0 {
		return false
	}

	// Check if the most recent error was within the last 5 minutes
	mostRecent := torrent.Errors[len(torrent.Errors)-1]
	return time.Since(mostRecent.Time) < 5*time.Minute
}

func percent(n, total int64) float32 {
	if total == 0 {
		return float32(0)
	}
	return float32(int(float64(10000)*(float64(n)/float64(total)))) / 100
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
