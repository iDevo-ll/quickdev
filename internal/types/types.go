package types

import (
	"os"
	"regexp"
	"time"
)

// FileWatcherConfig represents the configuration for the file watcher
type FileWatcherConfig struct {
	Enabled                bool          `json:"enabled"`
	WatchPaths            []string      `json:"watch"`
	IgnorePaths           []string      `json:"ignore"`
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
	EnableFileHashing     bool          `json:"enableHashing"`
	ClearScreen           bool          `json:"clearScreen"`
	CustomIgnoreFile      string        `json:"ignoreFile"`
	WatchDotFiles         bool          `json:"watchDotFiles"`
	MaxFileSize           int           `json:"maxFileSize"`
	ExcludeEmptyFiles     bool          `json:"excludeEmptyFiles"`
	ParallelProcessing    bool          `json:"parallelProcessing"`
	HealthCheck           bool          `json:"healthCheck"`
	HealthCheckInterval   int           `json:"healthCheckInterval"`
	MemoryLimit           int           `json:"memoryLimit"`
	TypeScriptRunner      string        `json:"typescriptRunner"` // "tsx" or "ts-node"
	TSNodeFlags           string        `json:"tsNodeFlags"`      // Additional flags for ts-node/tsx
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

// FileEvent represents a file change event
type FileEvent struct {
	Path      string    `json:"path"`
	Operation string    `json:"operation"`
	Time      time.Time `json:"time"`
}

// RestartHistoryEntry represents a single restart event
type RestartHistoryEntry struct {
	Time      time.Time     `json:"time"`
	ExitCode  int          `json:"exitCode"`
	Error     string       `json:"error"`
	Duration  time.Duration `json:"duration"`
}

// RestartStats tracks process restart statistics
type RestartStats struct {
	TotalRestarts    int                   `json:"totalRestarts"`
	LastRestart      time.Time             `json:"lastRestart"`
	RestartHistory   []RestartHistoryEntry `json:"restartHistory"`
	AverageUptime    time.Duration         `json:"averageUptime"`
	LongestUptime    time.Duration         `json:"longestUptime"`
	ShortestUptime   time.Duration         `json:"shortestUptime"`
	LastExitCode     int                   `json:"lastExitCode"`
	LastErrorMessage string                `json:"lastErrorMessage"`
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

// WatcherHealth represents health check information
type WatcherHealth struct {
	LastCheck      time.Time `json:"lastCheck"`
	Status         string    `json:"status"`
	MemoryUsage    uint64    `json:"memoryUsage"`
	CPUUsage       float64   `json:"cpuUsage"`
	FileCount      int       `json:"fileCount"`
	WatchedDirs    int       `json:"watchedDirs"`
	ErrorCount     int       `json:"errorCount"`
	LastError      string    `json:"lastError"`
	LastErrorTime  time.Time `json:"lastErrorTime"`
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