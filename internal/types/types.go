package types

import (
	"os"
	"regexp"
	"time"
)

// FileWatcherConfig represents the configuration for the file watcher
type FileWatcherConfig struct {
	Enabled                bool          `json:"enabled"`
	WatchPaths            []string      `json:"watchPaths"`
	IgnorePaths           []string      `json:"ignorePaths"`
	IgnorePatterns        []*regexp.Regexp `json:"ignorePatterns"`
	Extensions            []string      `json:"extensions"`
	DebounceMs            int           `json:"debounceMs"`
	RestartDelay          int           `json:"restartDelay"`
	MaxRestarts           int           `json:"maxRestarts"`
	ResetRestartsAfter    int           `json:"resetRestartsAfter"`
	GracefulShutdown      bool          `json:"gracefulShutdown"`
	GracefulShutdownTimeout int         `json:"gracefulShutdownTimeout"`
	UsePolling            bool          `json:"usePolling"`
	PollingInterval       int           `json:"pollingInterval"`
	FollowSymlinks        bool          `json:"followSymlinks"`
	PersistentWatching    bool          `json:"persistentWatching"`
	BatchChanges          bool          `json:"batchChanges"`
	BatchTimeout          int           `json:"batchTimeout"`
	EnableFileHashing     bool          `json:"enableFileHashing"`
	ClearScreen           bool          `json:"clearScreen"`
	CustomIgnoreFile      string        `json:"customIgnoreFile"`
	WatchDotFiles         bool          `json:"watchDotFiles"`
	MaxFileSize           int           `json:"maxFileSize"`
	ExcludeEmptyFiles     bool          `json:"excludeEmptyFiles"`
	ParallelProcessing    bool          `json:"parallelProcessing"`
	HealthCheck           bool          `json:"healthCheck"`
	HealthCheckInterval   int           `json:"healthCheckInterval"`
	MemoryLimit           int           `json:"memoryLimit"`
}

// FileChangeEvent represents a single file change event
type FileChangeEvent struct {
	Type         string    `json:"type"`
	Filename     string    `json:"filename"`
	FullPath     string    `json:"fullPath"`
	RelativePath string    `json:"relativePath"`
	Timestamp    time.Time `json:"timestamp"`
	Size         int64     `json:"size,omitempty"`
	Hash         string    `json:"hash,omitempty"`
	PreviousHash string    `json:"previousHash,omitempty"`
	IsDirectory  bool      `json:"isDirectory"`
	Stats        os.FileInfo `json:"-"`
}

// BatchChangeEvent represents multiple file changes grouped together
type BatchChangeEvent struct {
	Changes     []FileChangeEvent `json:"changes"`
	TotalFiles  int              `json:"totalFiles"`
	Timestamp   time.Time        `json:"timestamp"`
	Duration    time.Duration    `json:"duration"`
}

// RestartHistoryEntry represents a single restart event
type RestartHistoryEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Reason      string    `json:"reason"`
	Duration    time.Duration `json:"duration"`
	Success     bool      `json:"success"`
	FileCount   int       `json:"fileCount"`
	MemoryUsage *MemoryUsage `json:"memoryUsage"`
}

// RestartStats tracks statistics about process restarts
type RestartStats struct {
	TotalRestarts      int                  `json:"totalRestarts"`
	LastRestart        *time.Time           `json:"lastRestart"`
	AverageRestartTime time.Duration        `json:"averageRestartTime"`
	FastestRestart     time.Duration        `json:"fastestRestart"`
	SlowestRestart     time.Duration        `json:"slowestRestart"`
	SuccessfulRestarts int                  `json:"successfulRestarts"`
	FailedRestarts     int                  `json:"failedRestarts"`
	RestartHistory     []RestartHistoryEntry `json:"restartHistory"`
}

// HealthError represents an error in the watcher's health monitoring
type HealthError struct {
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
	Resolved  bool      `json:"resolved"`
}

// MemoryUsage represents memory usage statistics
type MemoryUsage struct {
	HeapTotal     uint64 `json:"heapTotal"`
	HeapUsed      uint64 `json:"heapUsed"`
	External      uint64 `json:"external"`
	ProcessMemory uint64 `json:"processMemory"`
}

// WatcherHealth represents the health status of the file watcher
type WatcherHealth struct {
	IsHealthy         bool          `json:"isHealthy"`
	Uptime           time.Duration  `json:"uptime"`
	MemoryUsage      *MemoryUsage  `json:"memoryUsage"`
	ActiveConnections int           `json:"activeConnections"`
	LastHealthCheck   time.Time     `json:"lastHealthCheck"`
	Errors           []HealthError `json:"errors"`
}

// ConfigFile represents the watchtower.config.json structure
type ConfigFile struct {
	// Core settings
	Script     string   `json:"script"`
	Watch      []string `json:"watch"`
	Ignore     []string `json:"ignore"`
	Extensions []string `json:"extensions"`

	// Process Management
	GracefulShutdown      bool `json:"gracefulShutdown"`
	GracefulShutdownTimeout int `json:"gracefulShutdownTimeout"`
	MaxRestarts           int  `json:"maxRestarts"`
	ResetRestartsAfter    int  `json:"resetRestartsAfter"`
	RestartDelay          int  `json:"restartDelay"`

	// File Watching
	BatchChanges    bool   `json:"batchChanges"`
	BatchTimeout    int    `json:"batchTimeout"`
	EnableHashing   bool   `json:"enableHashing"`
	UsePolling      bool   `json:"usePolling"`
	PollingInterval int    `json:"pollingInterval"`
	FollowSymlinks  bool   `json:"followSymlinks"`
	WatchDotFiles   bool   `json:"watchDotFiles"`
	IgnoreFile      string `json:"ignoreFile"`

	// Performance
	ParallelProcessing bool `json:"parallelProcessing"`
	MemoryLimit        int  `json:"memoryLimit"`
	MaxFileSize        int  `json:"maxFileSize"`
	ExcludeEmptyFiles  bool `json:"excludeEmptyFiles"`
	DebounceMs         int  `json:"debounceMs"`

	// Monitoring
	HealthCheck       bool `json:"healthCheck"`
	HealthCheckInterval int `json:"healthCheckInterval"`
	ClearScreen       bool `json:"clearScreen"`

	// TypeScript specific
	TypeScriptRunner string `json:"typescriptRunner"` // "tsx" or "ts-node"
	TSNodeFlags      string `json:"tsNodeFlags"`      // Additional flags for ts-node
} 