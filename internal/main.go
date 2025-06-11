package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nehonix-nhr/internal/process"
	"nehonix-nhr/internal/types"
	"nehonix-nhr/internal/utils"
	"nehonix-nhr/internal/watcher"
)

const Version = "1.0.0"

var (
	scriptFlag            = flag.String("script", "", "Path to the script to run")
	watchFlag            = flag.String("watch", ".", "Directories to watch (comma-separated)")
	ignoreFlag           = flag.String("ignore", "node_modules,dist,.git", "Directories to ignore (comma-separated)")
	extFlag             = flag.String("ext", ".js,.ts,.jsx,.tsx", "File extensions to watch (comma-separated)")
	debounceFlag        = flag.Int("debounce", 250, "Debounce time in milliseconds")
	restartDelayFlag    = flag.Int("restart-delay", 100, "Delay before restart in milliseconds")
	maxRestartsFlag     = flag.Int("max-restarts", 0, "Maximum number of restarts (0 for unlimited)")
	resetAfterFlag      = flag.Int("reset-after", 60000, "Reset restart count after X milliseconds")
	gracefulFlag        = flag.Bool("graceful", true, "Use graceful shutdown")
	gracefulTimeoutFlag = flag.Int("graceful-timeout", 5, "Graceful shutdown timeout in seconds")
	pollingFlag         = flag.Bool("polling", false, "Use polling instead of filesystem events")
	pollingIntervalFlag = flag.Int("polling-interval", 100, "Polling interval in milliseconds")
	followSymlinksFlag  = flag.Bool("follow-symlinks", false, "Follow symlinks")
	batchChangesFlag    = flag.Bool("batch", true, "Batch file changes")
	batchTimeoutFlag    = flag.Int("batch-timeout", 300, "Batch timeout in milliseconds")
	hashingFlag         = flag.Bool("hash", true, "Enable file hashing")
	clearScreenFlag     = flag.Bool("clear", true, "Clear screen on restart")
	ignoreFileFlag      = flag.String("ignore-file", "", "Custom ignore file")
	watchDotFlag        = flag.Bool("watch-dot", false, "Watch dot files")
	maxFileSizeFlag     = flag.Int("max-size", 10, "Maximum file size in MB")
	excludeEmptyFlag    = flag.Bool("exclude-empty", true, "Exclude empty files")
	parallelFlag        = flag.Bool("parallel", true, "Enable parallel processing")
	healthCheckFlag     = flag.Bool("health", true, "Enable health checking")
	healthIntervalFlag  = flag.Int("health-interval", 30, "Health check interval in seconds")
	memoryLimitFlag     = flag.Int("memory", 500, "Memory limit in MB")
)

func main() {
	flag.Parse()

	if *scriptFlag == "" {
		fmt.Println("Error: script path is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create configuration
	config := &types.FileWatcherConfig{
		Enabled:                true,
		WatchPaths:            strings.Split(*watchFlag, ","),
		IgnorePaths:           strings.Split(*ignoreFlag, ","),
		Extensions:            strings.Split(*extFlag, ","),
		DebounceMs:            *debounceFlag,
		RestartDelay:          *restartDelayFlag,
		MaxRestarts:           *maxRestartsFlag,
		ResetRestartsAfter:    *resetAfterFlag,
		GracefulShutdown:      *gracefulFlag,
		GracefulShutdownTimeout: *gracefulTimeoutFlag,
		UsePolling:            *pollingFlag,
		PollingInterval:       *pollingIntervalFlag,
		FollowSymlinks:        *followSymlinksFlag,
		BatchChanges:          *batchChangesFlag,
		BatchTimeout:          *batchTimeoutFlag,
		EnableFileHashing:     *hashingFlag,
		ClearScreen:           *clearScreenFlag,
		CustomIgnoreFile:      *ignoreFileFlag,
		WatchDotFiles:         *watchDotFlag,
		MaxFileSize:           *maxFileSizeFlag,
		ExcludeEmptyFiles:     *excludeEmptyFlag,
		ParallelProcessing:    *parallelFlag,
		HealthCheck:           *healthCheckFlag,
		HealthCheckInterval:   *healthIntervalFlag,
		MemoryLimit:           *memoryLimitFlag,
	}

	// Load custom ignore patterns if specified
	if config.CustomIgnoreFile != "" {
		if patterns, err := loadIgnoreFile(config.CustomIgnoreFile); err == nil {
			config.IgnorePaths = append(config.IgnorePaths, patterns...)
		}
	}

	// Create process manager
	scriptPath, err := filepath.Abs(*scriptFlag)
	if err != nil {
		fmt.Printf("Error resolving script path: %v\n", err)
		os.Exit(1)
	}

	pm := process.NewProcessManager(scriptPath, config)

	// Create file watcher
	fw := watcher.NewFileWatcher(config)

	// Start the process
	if err := pm.Start(); err != nil {
		fmt.Printf("Error starting process: %v\n", err)
		os.Exit(1)
	}

	// Start the watcher
	if err := fw.Start(); err != nil {
		fmt.Printf("Error starting watcher: %v\n", err)
		os.Exit(1)
	}

	// Print initial status
	printStatus(config)

	// Main event loop
	for {
		select {
		case event := <-fw.GetChangeChannel():
			handleFileChange(event, pm)
		case err := <-fw.GetErrorChannel():
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func handleFileChange(event types.FileChangeEvent, pm *process.ProcessManager) {
	var reason string
	if event.IsDirectory {
		reason = fmt.Sprintf("Directory changed: %s", utils.Path(event.RelativePath))
	} else {
		reason = fmt.Sprintf("File changed: %s", utils.Path(event.RelativePath))
		if event.PreviousHash != "" {
			reason += fmt.Sprintf(" (hash: %s -> %s)", 
				utils.Dimmed(event.PreviousHash[:8]), 
				utils.Info(event.Hash[:8]))
		}
	}

	// Print change details
	fmt.Printf("\n%s\n", utils.Info(reason))
	if !event.IsDirectory {
		fmt.Printf("%s %.2f KB\n", utils.Section("Size:"), float64(event.Size)/1024)
		fmt.Printf("%s %s\n", utils.Section("Time:"), event.Timestamp.Format("15:04:05"))
	}

	// Restart the process
	if err := pm.Restart(reason); err != nil {
		fmt.Printf("%s %v\n", utils.Error("Error restarting process:"), err)
		return
	}

	// Print restart success
	fmt.Printf("%s\n", utils.Success("Process restarted successfully"))
}

func loadIgnoreFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ignore file: %w", err)
	}

	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			// Convert glob patterns to proper format
			line = strings.ReplaceAll(line, "\\", "/")
			line = strings.TrimPrefix(line, "./")
			patterns = append(patterns, line)
		}
	}

	return patterns, nil
}

func printStatus(config *types.FileWatcherConfig) {
	fmt.Printf("\n%s\n", utils.Header("Nehonix WatchTower"))
	fmt.Println(utils.Dimmed("================================"))
	
	fmt.Printf("%s %s\n", utils.Section("Watching:"), utils.Path(strings.Join(config.WatchPaths, ", ")))
	fmt.Printf("%s %s\n", utils.Section("Ignoring:"), utils.Path(strings.Join(config.IgnorePaths, ", ")))
	fmt.Printf("%s %s\n", utils.Section("Extensions:"), utils.Path(strings.Join(config.Extensions, ", ")))
	
	features := getEnabledFeatures(config)
	fmt.Printf("%s %s\n", utils.Section("Features:"), features)
	
	if config.MaxRestarts > 0 {
		fmt.Printf("%s %d (reset after %ds)\n", 
			utils.Section("Max Restarts:"),
			config.MaxRestarts, 
			config.ResetRestartsAfter/1000)
	}
	
	if config.MemoryLimit > 0 {
		fmt.Printf("%s %d MB\n", utils.Section("Memory Limit:"), config.MemoryLimit)
	}
	
	fmt.Println(utils.Dimmed("================================"))
	fmt.Printf("%s v%s\n", utils.Info("Monitoring with WatchTower"), Version)
	fmt.Printf("%s\n\n", utils.Dimmed("Press Ctrl+C to exit"))
}

func getEnabledFeatures(config *types.FileWatcherConfig) string {
	var features []string
	
	if config.BatchChanges {
		features = append(features, "batching")
	}
	if config.EnableFileHashing {
		features = append(features, "hashing")
	}
	if config.GracefulShutdown {
		features = append(features, "graceful-shutdown")
	}
	if config.HealthCheck {
		features = append(features, "health-monitoring")
	}
	if config.ParallelProcessing {
		features = append(features, "parallel")
	}
	if config.UsePolling {
		features = append(features, "polling")
	}

	return utils.Status(strings.Join(features, ", "))
} 