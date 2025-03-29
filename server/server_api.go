package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"github.com/jpillora/cloud-torrent/engine"
)

// TorrentDetailedStatus provides detailed information about a torrent's status
type TorrentDetailedStatus struct {
	InfoHash        string               `json:"infoHash"`
	Name            string               `json:"name"`
	Status          string               `json:"status"`           // Health status as string
	Size            int64                `json:"size"`             // Total size in bytes
	Downloaded      int64                `json:"downloaded"`       // Downloaded bytes
	DownloadRate    float32              `json:"downloadRate"`     // Current download rate in bytes/sec
	Percent         float32              `json:"percent"`          // Percentage complete
	Files           []FileDetailedStatus `json:"files,omitempty"`  // Optional file details
	Errors          []ErrorInfo          `json:"errors,omitempty"` // Recent errors
	PeersConnected  int                  `json:"peersConnected"`   // Number of connected peers
	PeersTotal      int                  `json:"peersTotal"`       // Total peers available
	MetadataPercent float32              `json:"metadataPercent"`  // Metadata download percentage
	TimeAdded       time.Time            `json:"timeAdded"`        // When the torrent was added
	TimeUpdated     time.Time            `json:"timeUpdated"`      // Last update time
	LastProgress    time.Time            `json:"lastProgress"`     // Time of last download progress
}

// FileDetailedStatus provides detailed information about a file's status
type FileDetailedStatus struct {
	Path        string  `json:"path"`
	Size        int64   `json:"size"`
	Downloaded  int64   `json:"downloaded"`
	Percent     float32 `json:"percent"`
	Priority    int     `json:"priority"`
	BytesPerSec int64   `json:"bytesPerSec"`
}

// ErrorInfo contains information about an error
type ErrorInfo struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

func (s *Server) api(r *http.Request) error {
	defer r.Body.Close()
	if r.Method != "POST" {
		return fmt.Errorf("Invalid request method (expecting POST)")
	}

	action := strings.TrimPrefix(r.URL.Path, "/api/")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Failed to download request body")
	}

	//convert url into torrent bytes
	if action == "url" {
		url := string(data)
		remote, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("Invalid remote torrent URL: %s (%s)", err, url)
		}
		defer remote.Body.Close() // Ensure body is closed

		// Enforce max body size (32MB)
		if remote.ContentLength > 32*1024*1024 {
			return fmt.Errorf("Remote torrent file too large: %d bytes", remote.ContentLength)
		}

		data, err = ioutil.ReadAll(remote.Body)
		if err != nil {
			return fmt.Errorf("Failed to download remote torrent: %s", err)
		}
		action = "torrentfile"
	}

	//convert torrent bytes into magnet
	if action == "torrentfile" {
		reader := bytes.NewBuffer(data)
		info, err := metainfo.Load(reader)
		if err != nil {
			return fmt.Errorf("Invalid torrent file: %s", err)
		}
		spec := torrent.TorrentSpecFromMetaInfo(info)
		if err := s.engine.NewTorrent(spec); err != nil {
			return fmt.Errorf("Torrent error: %s", err)
		}
		return nil
	}

	//update after action completes
	defer s.state.Push()

	//interface with engine
	switch action {
	case "configure":
		c := engine.Config{}
		if err := json.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("Invalid configuration format: %s", err)
		}
		if err := s.reconfigure(c); err != nil {
			return fmt.Errorf("Failed to reconfigure: %s", err)
		}

	case "magnet":
		uri := string(data)
		if err := s.engine.NewMagnet(uri); err != nil {
			return fmt.Errorf("Magnet error: %s", err)
		}

	case "torrent":
		cmd := strings.SplitN(string(data), ":", 2)
		if len(cmd) != 2 {
			return fmt.Errorf("Invalid request format")
		}
		state := cmd[0]
		infohash := cmd[1]
		if state == "start" {
			if err := s.engine.StartTorrent(infohash); err != nil {
				return fmt.Errorf("Failed to start torrent: %s", err)
			}
		} else if state == "stop" {
			if err := s.engine.StopTorrent(infohash); err != nil {
				return fmt.Errorf("Failed to stop torrent: %s", err)
			}
		} else if state == "delete" {
			if err := s.engine.DeleteTorrent(infohash); err != nil {
				return fmt.Errorf("Failed to delete torrent: %s", err)
			}
		} else {
			return fmt.Errorf("Invalid state: %s", state)
		}

	case "file":
		cmd := strings.SplitN(string(data), ":", 3)
		if len(cmd) != 3 {
			return fmt.Errorf("Invalid file command format")
		}
		state := cmd[0]
		infohash := cmd[1]
		filepath := cmd[2]
		if state == "start" {
			if err := s.engine.StartFile(infohash, filepath); err != nil {
				return fmt.Errorf("Failed to start file: %s", err)
			}
		} else if state == "stop" {
			if err := s.engine.StopFile(infohash, filepath); err != nil {
				return fmt.Errorf("Failed to stop file: %s", err)
			}
		} else {
			return fmt.Errorf("Invalid file state: %s", state)
		}

	case "status":
		// Detailed status endpoint for a specific torrent
		infohash := string(data)
		if infohash == "" {
			return fmt.Errorf("Infohash required")
		}

		s.state.Lock()
		defer s.state.Unlock()

		torrent, err := s.engine.GetTorrent(infohash)
		if err != nil {
			return fmt.Errorf("Torrent not found: %s", err)
		}

		// Lock the torrent to get consistent data
		torrent.Mu.Lock()
		defer torrent.Mu.Unlock()

		// Convert status to string
		statusString := "unknown"
		switch torrent.Status {
		case engine.TorrentStatusHealthy:
			statusString = "healthy"
		case engine.TorrentStatusSlow:
			statusString = "slow"
		case engine.TorrentStatusStalled:
			statusString = "stalled"
		case engine.TorrentStatusError:
			statusString = "error"
		}

		// Calculate downloaded bytes for each file
		files := make([]FileDetailedStatus, 0, len(torrent.Files))
		for _, f := range torrent.Files {
			if f == nil {
				continue
			}
			downloadedBytes := int64(float64(f.Size) * float64(f.Percent) / 100.0)
			files = append(files, FileDetailedStatus{
				Path:        f.Path,
				Size:        f.Size,
				Downloaded:  downloadedBytes,
				Percent:     f.Percent,
				Priority:    f.Priority,
				BytesPerSec: f.BytesPerSec,
			})
		}

		// Convert errors
		errors := make([]ErrorInfo, 0, len(torrent.Errors))
		for _, e := range torrent.Errors {
			errors = append(errors, ErrorInfo{
				Time:    e.Time,
				Message: e.Message,
			})
		}

		// Create detailed status
		status := TorrentDetailedStatus{
			InfoHash:        torrent.InfoHash,
			Name:            torrent.Name,
			Status:          statusString,
			Size:            torrent.Size,
			Downloaded:      torrent.Downloaded,
			DownloadRate:    torrent.DownloadRate,
			Percent:         torrent.Percent,
			Files:           files,
			Errors:          errors,
			PeersConnected:  torrent.PeersConnected,
			PeersTotal:      torrent.PeersTotal,
			MetadataPercent: torrent.MetadataPercent,
			TimeUpdated:     torrent.UpdatedAt,
			LastProgress:    torrent.LastProgress,
		}

		// Convert to JSON and write response
		b, err := json.Marshal(status)
		if err != nil {
			return fmt.Errorf("Failed to serialize status: %s", err)
		}

		// This isn't typical for the API, but this is a special case to return detailed status
		w := r.Context().Value("http.ResponseWriter").(http.ResponseWriter)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return nil

	case "health":
		// Return overall health status of the engine
		torrents := s.engine.GetTorrents()

		// Get engine stats through reflection or simply report what we can
		stats := struct {
			Torrents       int   `json:"torrents"`
			ActiveTorrents int   `json:"activeTorrents"`
			MemoryUsage    int64 `json:"memoryUsage"`
			Uptime         int64 `json:"uptime"`
		}{
			Torrents:       len(torrents),
			ActiveTorrents: 0, // We can't access private field directly
			MemoryUsage:    0, // We can't access private field directly
			Uptime:         int64(time.Since(s.startTime).Seconds()),
		}

		// Count active torrents manually
		for _, t := range torrents {
			if t.Started {
				stats.ActiveTorrents++
			}
		}

		b, err := json.Marshal(stats)
		if err != nil {
			return fmt.Errorf("Failed to serialize health stats: %s", err)
		}

		w := r.Context().Value("http.ResponseWriter").(http.ResponseWriter)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return nil

	default:
		return fmt.Errorf("Invalid action: %s", action)
	}

	return nil
}
