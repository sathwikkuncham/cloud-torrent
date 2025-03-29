package server

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jpillora/archive"
)

// Increased file limit to support larger torrent directories
const fileNumberLimit = 10000

// Track currently active file transfers to limit concurrent operations
var activeTransfers struct {
	sync.Mutex
	count int
	limit int
}

func init() {
	activeTransfers.limit = 20 // Default limit of concurrent file operations
}

type fsNode struct {
	Name      string
	Size      int64
	Modified  time.Time
	Children  []*fsNode
	IsDir     bool
	FileCount int // Count of files in directory (only for directories)
}

func (s *Server) listFiles() *fsNode {
	rootDir := s.state.Config.DownloadDirectory
	root := &fsNode{IsDir: true}
	if info, err := os.Stat(rootDir); err == nil {
		if err := list(rootDir, info, root, new(int)); err != nil {
			log.Printf("File listing failed: %s", err)
		}
	}
	return root
}

func (s *Server) serveFiles(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/download/") {
		url := strings.TrimPrefix(r.URL.Path, "/download/")
		//dldir is absolute
		dldir := s.state.Config.DownloadDirectory
		file := filepath.Join(dldir, url)
		//only allow fetches/deletes inside the dl dir
		if !strings.HasPrefix(file, dldir) || dldir == file {
			http.Error(w, "Nice try\n"+dldir+"\n"+file, http.StatusBadRequest)
			return
		}
		info, err := os.Stat(file)
		if err != nil {
			http.Error(w, "File stat error: "+err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			// Check if we're already serving too many files
			activeTransfers.Lock()
			if activeTransfers.count >= activeTransfers.limit {
				activeTransfers.Unlock()
				http.Error(w, "Too many concurrent downloads, try again later",
					http.StatusServiceUnavailable)
				return
			}
			activeTransfers.count++
			activeTransfers.Unlock()

			// Ensure we decrement the counter when done
			defer func() {
				activeTransfers.Lock()
				activeTransfers.count--
				activeTransfers.Unlock()
			}()

			if info.IsDir() {
				// For large directories, we should check size before zipping
				var totalSize int64
				var fileCount int
				err := filepath.Walk(file, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						totalSize += info.Size()
						fileCount++
					}
					// Abort if directory is too large
					if totalSize > 5*1024*1024*1024 { // 5GB limit
						return fmt.Errorf("directory too large")
					}
					if fileCount > 10000 {
						return fmt.Errorf("too many files")
					}
					return nil
				})

				if err != nil {
					if err.Error() == "directory too large" {
						http.Error(w, "Directory too large to zip (>5GB)", http.StatusRequestEntityTooLarge)
						return
					} else if err.Error() == "too many files" {
						http.Error(w, "Directory contains too many files to zip (>10000)", http.StatusRequestEntityTooLarge)
						return
					}
					http.Error(w, "Error accessing directory: "+err.Error(), http.StatusInternalServerError)
					return
				}

				log.Printf("Serving zip archive of %s (%s, %d files)",
					file, humanize.Bytes(uint64(totalSize)), fileCount)

				w.Header().Set("Content-Type", "application/zip")
				w.WriteHeader(200)

				//write .zip archive directly into response with buffering
				bufWriter := bufio.NewWriterSize(w, 4*1024*1024) // 4MB buffer
				a := archive.NewZipWriter(bufWriter)
				a.AddDir(file)
				a.Close()
				bufWriter.Flush()
			} else {
				// Log large file transfers
				if info.Size() > 100*1024*1024 { // 100MB
					log.Printf("Serving large file: %s (%s)",
						info.Name(), humanize.Bytes(uint64(info.Size())))
				}

				f, err := os.Open(file)
				if err != nil {
					http.Error(w, "File open error: "+err.Error(), http.StatusBadRequest)
					return
				}
				defer f.Close()

				// Add caching headers for static content
				w.Header().Set("Cache-Control", "max-age=31536000") // 1 year

				// Use ServeContent for efficient transfer including range requests
				http.ServeContent(w, r, info.Name(), info.ModTime(), f)
			}
		case "DELETE":
			if err := os.RemoveAll(file); err != nil {
				http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("Deleted: %s", file)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	s.static.ServeHTTP(w, r)
}

// Custom directory walk with improvements for large directories
func list(path string, info os.FileInfo, node *fsNode, n *int) error {
	if (!info.IsDir() && !info.Mode().IsRegular()) || strings.HasPrefix(info.Name(), ".") {
		return errors.New("Non-regular file")
	}
	(*n)++
	if (*n) > fileNumberLimit {
		return errors.New("Too many files") // Limit number of files walked
	}
	node.Name = info.Name()
	node.Size = info.Size()
	node.Modified = info.ModTime()
	node.IsDir = info.IsDir()

	if !info.IsDir() {
		return nil
	}

	childEntries, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("Failed to list directory: %s", err)
	}

	// Initialize size for directory
	node.Size = 0
	node.FileCount = 0

	// For very large directories, we might want to limit processing
	if len(childEntries) > 1000 {
		// Just provide a summary for very large directories
		var dirCount, fileCount int
		var totalSize int64

		for _, entry := range childEntries {
			if entry.IsDir() {
				dirCount++
			} else {
				fileCount++
				totalSize += entry.Size()
			}
		}

		// Create a summary entry
		node.Children = []*fsNode{
			{
				Name:      fmt.Sprintf("[%d files and %d directories]", fileCount, dirCount),
				Size:      totalSize,
				Modified:  time.Now(),
				IsDir:     false,
				FileCount: fileCount,
			},
		}
		node.Size = totalSize
		node.FileCount = fileCount

		log.Printf("Large directory %s: %d files, %d subdirs, %s total",
			path, fileCount, dirCount, humanize.Bytes(uint64(totalSize)))

		return nil
	}

	// Process normal sized directories
	for _, childInfo := range childEntries {
		childNode := &fsNode{}
		childPath := filepath.Join(path, childInfo.Name())

		if err := list(childPath, childInfo, childNode, n); err != nil {
			// Log the error but continue processing other files
			log.Printf("Error listing %s: %s", childPath, err)
			continue
		}

		node.Size += childNode.Size
		if childInfo.IsDir() {
			node.FileCount += childNode.FileCount
		} else {
			node.FileCount++
		}
		node.Children = append(node.Children, childNode)
	}

	return nil
}
