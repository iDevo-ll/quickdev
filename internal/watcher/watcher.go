package watcher

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"quickdev/internal/types"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher represents the main file watcher instance
type FileWatcher struct {
	config         *types.FileWatcherConfig
	watcher        *fsnotify.Watcher
	fileHashes     map[string]string
	hashMutex      sync.RWMutex
	changes        chan types.FileEvent
	errors         chan error
	batchTimer     *time.Timer
	batchedChanges map[string]types.FileEvent
	batchMutex     sync.Mutex
	health         *types.WatcherHealth
	startTime      time.Time
}

// NewFileWatcher creates a new file watcher instance
func NewFileWatcher(config *types.FileWatcherConfig) *FileWatcher {
	return &FileWatcher{
		config:         config,
		fileHashes:     make(map[string]string),
		changes:        make(chan types.FileEvent, 100),
		errors:         make(chan error, 100),
		batchedChanges: make(map[string]types.FileEvent),
		health: &types.WatcherHealth{
			Status:    "starting",
			LastCheck: time.Now(),
		},
		startTime: time.Now(),
	}
}

// Start begins watching for file changes
func (fw *FileWatcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %v", err)
	}
	fw.watcher = watcher

	// Add watch paths
	for _, path := range fw.config.WatchPaths {
		if err := fw.addWatchPath(path); err != nil {
			fw.errors <- fmt.Errorf("error adding watch path %s: %v", path, err)
		}
	}

	// Start health monitoring if enabled
	if fw.config.HealthCheck {
		go fw.monitorHealth()
	}

	// Start watching for events
	go fw.watchEvents()

	return nil
}

// Stop gracefully stops the file watcher
func (fw *FileWatcher) Stop() error {
	if fw.watcher != nil {
		return fw.watcher.Close()
	}
	return nil
}

// addWatchPath starts watching a specific path
func (fw *FileWatcher) addWatchPath(path string) error {
	// Convert to absolute path if not already
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error getting absolute path for %s: %v", path, err)
	}
	path = absPath

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error stating path %s: %v", path, err)
	}

	// If it's a directory, walk it and add all subdirectories
	if info.IsDir() {
		// fmt.Printf("Adding directory to watch: %s\n", path)
		return filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip if path should be ignored
			if fw.shouldIgnore(subpath) {
				if info.IsDir() {
					// fmt.Printf("Ignoring directory: %s\n", subpath)
					return filepath.SkipDir
				}
				return nil
			}

			// Add directory to watcher
			if info.IsDir() {
				if err := fw.watcher.Add(subpath); err != nil {
					return fmt.Errorf("error watching directory %s: %v", subpath, err)
				}
				// fmt.Printf("Watching directory: %s\n", subpath)
			} else if fw.hasValidExtension(subpath) {
				// fmt.Printf("Found watchable file: %s\n", subpath)
			}

			// Calculate initial hash for file if enabled
			if !info.IsDir() && fw.config.EnableFileHashing {
				if hash, err := fw.calculateFileHash(subpath); err == nil {
					fw.hashMutex.Lock()
					fw.fileHashes[subpath] = hash
					fw.hashMutex.Unlock()
				}
			}

			return nil
		})
	}

	// If it's a file, just watch its directory
	dir := filepath.Dir(path)
	// fmt.Printf("Adding parent directory to watch: %s\n", dir)
	return fw.watcher.Add(dir)
}

// watchEvents implements the actual file watching logic
func (fw *FileWatcher) watchEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.errors <- err
		}
	}
}

// handleEvent processes a file change event
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// Skip if path should be ignored
	if fw.shouldIgnore(event.Name) {
		return
	}

	// Skip if file extension doesn't match
	if !fw.hasValidExtension(event.Name) {
		return
	}

	// Handle directory events
	info, err := os.Stat(event.Name)
	if err == nil && info.IsDir() {
		if event.Op&fsnotify.Create == fsnotify.Create {
			fw.addWatchPath(event.Name)
		}
		return
	}

	// Check if file content actually changed
	if fw.config.EnableFileHashing && event.Op&fsnotify.Write == fsnotify.Write {
		if !fw.hasFileChanged(event.Name) {
			return
		}
	}

	// Create file event
	fileEvent := types.FileEvent{
		Path:      event.Name,
		Operation: event.Op.String(),
		Time:      time.Now(),
	}

	// Handle batching
	if fw.config.BatchChanges {
		fw.batchEvent(fileEvent)
	} else {
		fw.changes <- fileEvent
	}
}

// batchEvent handles batched file changes
func (fw *FileWatcher) batchEvent(event types.FileEvent) {
	fw.batchMutex.Lock()
	defer fw.batchMutex.Unlock()

	// Add event to batch
	fw.batchedChanges[event.Path] = event

	// Reset or start timer
	if fw.batchTimer != nil {
		fw.batchTimer.Reset(time.Duration(fw.config.BatchTimeout) * time.Millisecond)
	} else {
		fw.batchTimer = time.AfterFunc(time.Duration(fw.config.BatchTimeout)*time.Millisecond, func() {
			fw.flushBatchedChanges()
		})
	}
}

// flushBatchedChanges handles a batch of changes
func (fw *FileWatcher) flushBatchedChanges() {
	fw.batchMutex.Lock()
	defer fw.batchMutex.Unlock()

	// Send all batched changes
	for _, event := range fw.batchedChanges {
		fw.changes <- event
	}

	// Clear batch
	fw.batchedChanges = make(map[string]types.FileEvent)
	fw.batchTimer = nil
}

// hasFileChanged checks if a file has changed by comparing hashes
func (fw *FileWatcher) hasFileChanged(path string) bool {
	newHash, err := fw.calculateFileHash(path)
	if err != nil {
		return true // If we can't calculate hash, assume file changed
	}

	fw.hashMutex.RLock()
	oldHash := fw.fileHashes[path]
	fw.hashMutex.RUnlock()

	if newHash != oldHash {
		fw.hashMutex.Lock()
		fw.fileHashes[path] = newHash
		fw.hashMutex.Unlock()
		return true
	}

	return false
}

// calculateFileHash calculates the hash of a file
func (fw *FileWatcher) calculateFileHash(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// shouldIgnore checks if a path should be ignored
func (fw *FileWatcher) shouldIgnore(path string) bool {
	// Convert path to forward slashes for consistency
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")

	// Check against ignore patterns
	for _, pattern := range fw.config.IgnorePaths {
		// Convert pattern to forward slashes and trim ./ prefix
		pattern = filepath.ToSlash(pattern)
		pattern = strings.TrimPrefix(pattern, "./")

		// Try exact match first
		if path == pattern {
			return true
		}

		// Try glob match
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}

		// Try contains match (for node_modules etc)
		if strings.Contains(path, "/"+pattern+"/") || strings.HasSuffix(path, "/"+pattern) {
			return true
		}
	}

	// Check dot files
	if !fw.config.WatchDotFiles && strings.Contains(filepath.Base(path), ".") {
		// Allow specific extensions even if they start with dot
		if fw.hasValidExtension(path) {
			return false
		}
		return true
	}

	return false
}

// hasValidExtension checks if a file has a valid extension
func (fw *FileWatcher) hasValidExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false
	}

	for _, validExt := range fw.config.Extensions {
		if strings.ToLower(validExt) == ext {
			return true
		}
	}

	return false
}

// monitorHealth periodically checks the watcher's health
func (fw *FileWatcher) monitorHealth() {
	ticker := time.NewTicker(time.Duration(fw.config.HealthCheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fw.updateHealth()
	}
}

// updateHealth performs a health check
func (fw *FileWatcher) updateHealth() {
	fw.health.LastCheck = time.Now()
	fw.health.Status = "healthy"

	// Count watched directories
	watchedDirs := 0
	if fw.watcher != nil {
		watchedDirs = len(fw.watcher.WatchList())
	}
	fw.health.WatchedDirs = watchedDirs

	// Count files being watched
	fileCount := len(fw.fileHashes)
	fw.health.FileCount = fileCount

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fw.health.MemoryUsage = m.Alloc

	// Update error count
	select {
	case err := <-fw.errors:
		fw.health.ErrorCount++
		fw.health.LastError = err.Error()
		fw.health.LastErrorTime = time.Now()
		fw.health.Status = "degraded"
	default:
	}
}

// GetChangeChannel returns the channel for file change events
func (fw *FileWatcher) GetChangeChannel() <-chan types.FileEvent {
	return fw.changes
}

// GetErrorChannel returns the channel for errors
func (fw *FileWatcher) GetErrorChannel() <-chan error {
	return fw.errors
}