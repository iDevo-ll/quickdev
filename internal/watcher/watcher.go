package watcher

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"nehonix-nhr/internal/types"
)

// FileWatcher represents the main file watcher instance
type FileWatcher struct {
	config        *types.FileWatcherConfig
	restartStats  *types.RestartStats
	health        *types.WatcherHealth
	startTime     time.Time
	batchTimer    *time.Timer
	batchChanges  []types.FileChangeEvent
	batchMutex    sync.Mutex
	fileHashes    map[string]string
	hashMutex     sync.RWMutex
	changeChannel chan types.FileChangeEvent
	errorChannel  chan error
	done          chan bool
}
 
// NewFileWatcher creates a new file watcher instance
func NewFileWatcher(config *types.FileWatcherConfig) *FileWatcher {
	return &FileWatcher{
		config:        config,
		startTime:     time.Now(),
		fileHashes:    make(map[string]string),
		changeChannel: make(chan types.FileChangeEvent, 100),
		errorChannel:  make(chan error, 100),
		done:         make(chan bool),
		restartStats: &types.RestartStats{
			RestartHistory: make([]types.RestartHistoryEntry, 0),
		},
		health: &types.WatcherHealth{
			IsHealthy:       true,
			LastHealthCheck: time.Now(),
			Errors:         make([]types.HealthError, 0),
		},
	}
}

// Start begins watching for file changes
func (fw *FileWatcher) Start() error {
	if !fw.config.Enabled {
		return fmt.Errorf("file watcher is disabled")
	}

	// Start health check if enabled
	if fw.config.HealthCheck {
		go fw.runHealthCheck()
	}

	// Initialize batch processing if enabled
	if fw.config.BatchChanges {
		fw.batchTimer = time.NewTimer(time.Duration(fw.config.BatchTimeout) * time.Millisecond)
		go fw.processBatchChanges()
	}

	// Start watching each path
	for _, path := range fw.config.WatchPaths {
		if err := fw.watchPath(path); err != nil {
			return fmt.Errorf("error watching path %s: %v", path, err)
		}
	}

	return nil
}

// Stop gracefully stops the file watcher
func (fw *FileWatcher) Stop() {
	if fw.config.GracefulShutdown {
		fmt.Println("Gracefully shutting down...")
		timeout := time.After(time.Duration(fw.config.GracefulShutdownTimeout) * time.Second)
		select {
		case <-fw.done:
			fmt.Println("Shutdown complete")
		case <-timeout:
			fmt.Println("Shutdown timed out")
		}
	}
	close(fw.done)
}

// watchPath starts watching a specific path
func (fw *FileWatcher) watchPath(path string) error {
	if !fw.config.WatchDotFiles && filepath.Base(path)[0] == '.' {
		return nil
	}

	// Initial file hash calculation if enabled
	if fw.config.EnableFileHashing {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				fw.calculateAndStoreHash(path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Start watching
	go fw.watchPathForChanges(path)
	return nil
}

// watchPathForChanges implements the actual file watching logic
func (fw *FileWatcher) watchPathForChanges(path string) {
	ticker := time.NewTicker(time.Duration(fw.config.PollingInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-fw.done:
			return
		case <-ticker.C:
			fw.checkForChanges(path)
		}
	}
}

// checkForChanges checks for file changes in the given path
func (fw *FileWatcher) checkForChanges(path string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fw.errorChannel <- err
			return nil
		}

		// Skip if path should be ignored
		for _, ignorePath := range fw.config.IgnorePaths {
			if matched, _ := filepath.Match(ignorePath, path); matched {
				return nil
			}
		}

		// Check file size limit
		if !info.IsDir() && fw.config.MaxFileSize > 0 && info.Size() > int64(fw.config.MaxFileSize*1024*1024) {
			return nil
		}

		// Check for changes
		if fw.config.EnableFileHashing {
			if fw.hasFileChanged(path, info) {
				event := fw.createChangeEvent(path, info)
				fw.handleChange(event)
			}
		}

		return nil
	})
}

// hasFileChanged checks if a file has changed by comparing hashes
func (fw *FileWatcher) hasFileChanged(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	newHash := fw.calculateHash(path)
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

// calculateHash calculates MD5 hash of a file
func (fw *FileWatcher) calculateHash(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// calculateAndStoreHash calculates and stores the hash of a file
func (fw *FileWatcher) calculateAndStoreHash(path string) {
	hash := fw.calculateHash(path)
	fw.hashMutex.Lock()
	fw.fileHashes[path] = hash
	fw.hashMutex.Unlock()
}

// createChangeEvent creates a FileChangeEvent for a changed file
func (fw *FileWatcher) createChangeEvent(path string, info os.FileInfo) types.FileChangeEvent {
	fw.hashMutex.RLock()
	prevHash := fw.fileHashes[path]
	fw.hashMutex.RUnlock()

	// Get relative path from watch directory
	relPath, err := filepath.Rel(fw.config.WatchPaths[0], path)
	if err != nil {
		relPath = path
	}

	return types.FileChangeEvent{
		Type:         "change",
		Filename:     filepath.Base(path),
		FullPath:     path,
		RelativePath: relPath,
		Timestamp:    time.Now(),
		Size:         info.Size(),
		Hash:         fw.calculateHash(path),
		PreviousHash: prevHash,
		IsDirectory:  info.IsDir(),
		Stats:        info,
	}
}

// handleChange processes a file change event
func (fw *FileWatcher) handleChange(event types.FileChangeEvent) {
	if fw.config.BatchChanges {
		fw.batchMutex.Lock()
		fw.batchChanges = append(fw.batchChanges, event)
		fw.batchMutex.Unlock()
	} else {
		fw.changeChannel <- event
	}
}

// processBatchChanges handles batched file changes
func (fw *FileWatcher) processBatchChanges() {
	for {
		select {
		case <-fw.done:
			return
		case <-fw.batchTimer.C:
			fw.batchMutex.Lock()
			if len(fw.batchChanges) > 0 {
				batchEvent := types.BatchChangeEvent{
					Changes:    fw.batchChanges,
					TotalFiles: len(fw.batchChanges),
					Timestamp:  time.Now(),
				}
				fw.batchChanges = make([]types.FileChangeEvent, 0)
				fw.processBatchEvent(batchEvent)
			}
			fw.batchMutex.Unlock()
			fw.batchTimer.Reset(time.Duration(fw.config.BatchTimeout) * time.Millisecond)
		}
	}
}

// processBatchEvent handles a batch of changes
func (fw *FileWatcher) processBatchEvent(batch types.BatchChangeEvent) {
	// Calculate batch duration
	batch.Duration = time.Since(batch.Timestamp)

	// Filter out any unwanted changes
	filteredChanges := make([]types.FileChangeEvent, 0)
	for _, change := range batch.Changes {
		// Skip empty files if configured
		if fw.config.ExcludeEmptyFiles && change.Size == 0 && !change.IsDirectory {
			continue
		}

		// Skip files that exceed size limit
		if fw.config.MaxFileSize > 0 && change.Size > int64(fw.config.MaxFileSize*1024*1024) {
			continue
		}

		// Check file extensions
		ext := filepath.Ext(change.Filename)
		isValidExt := false
		for _, allowedExt := range fw.config.Extensions {
			if ext == allowedExt {
				isValidExt = true
				break
			}
		}
		if !isValidExt {
			continue
		}

		filteredChanges = append(filteredChanges, change)
	}

	// Update batch with filtered changes
	batch.Changes = filteredChanges
	batch.TotalFiles = len(filteredChanges)

	// Send changes to channel if any remain after filtering
	if batch.TotalFiles > 0 {
		for _, change := range batch.Changes {
			fw.changeChannel <- change
		}
	}
}

// runHealthCheck periodically checks the watcher's health
func (fw *FileWatcher) runHealthCheck() {
	ticker := time.NewTicker(time.Duration(fw.config.HealthCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fw.done:
			return
		case <-ticker.C:
			fw.checkHealth()
		}
	}
}

// checkHealth performs a health check
func (fw *FileWatcher) checkHealth() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fw.health.Uptime = time.Since(fw.startTime)
	fw.health.LastHealthCheck = time.Now()
	fw.health.MemoryUsage = &types.MemoryUsage{
		HeapTotal:     m.HeapSys,
		HeapUsed:      m.HeapAlloc,
		External:      m.HeapReleased,
		ProcessMemory: m.Sys,
	}

	// Check memory limit
	if fw.config.MemoryLimit > 0 && m.Sys > uint64(fw.config.MemoryLimit*1024*1024) {
		fw.health.IsHealthy = false
		fw.health.Errors = append(fw.health.Errors, types.HealthError{
			Timestamp: time.Now(),
			Error:     "Memory limit exceeded",
			Resolved:  false,
		})
	}
}

// GetChangeChannel returns the channel for file change events
func (fw *FileWatcher) GetChangeChannel() chan types.FileChangeEvent {
	return fw.changeChannel
}

// GetErrorChannel returns the channel for errors
func (fw *FileWatcher) GetErrorChannel() chan error {
	return fw.errorChannel
} 