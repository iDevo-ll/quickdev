package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"quickdev/internal/config"
	"quickdev/internal/process"
	"quickdev/internal/types"
	"quickdev/internal/utils"
	"quickdev/internal/watcher"
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
		fmt.Println(utils.Error("Error: script path is required"))
		flag.Usage()
		os.Exit(1)
	}

	// Get absolute path of script
	scriptPath, err := filepath.Abs(*scriptFlag)
	if err != nil {
		fmt.Printf("%s %v\n", utils.Error("Error resolving script path:"), err)
		os.Exit(1)
	}

	// Find project root (directory containing package.json or parent of script)
	projectRoot := findProjectRoot(scriptPath)

	// Create initial config from CLI args
	cliConfig := &types.FileWatcherConfig{
		Enabled:                true,
		WatchPaths:            strings.Split(*watchFlag, ","),
		IgnorePaths:           strings.Split(*ignoreFlag, ","),
		Extensions:            strings.Split(*extFlag, ","),
		GracefulShutdown:      *gracefulFlag,
		GracefulShutdownTimeout: *gracefulTimeoutFlag,
		MaxRestarts:           *maxRestartsFlag,
		ResetRestartsAfter:    *resetAfterFlag,
		RestartDelay:          *restartDelayFlag,
		BatchChanges:          *batchChangesFlag,
		BatchTimeout:          *batchTimeoutFlag,
		EnableFileHashing:     *hashingFlag,
		UsePolling:            *pollingFlag,
		PollingInterval:       *pollingIntervalFlag,
		FollowSymlinks:        *followSymlinksFlag,
		WatchDotFiles:         *watchDotFlag,
		CustomIgnoreFile:      *ignoreFileFlag,
		ParallelProcessing:    *parallelFlag,
		MemoryLimit:           *memoryLimitFlag,
		MaxFileSize:           *maxFileSizeFlag,
		ExcludeEmptyFiles:     *excludeEmptyFlag,
		DebounceMs:            *debounceFlag,
		HealthCheck:           *healthCheckFlag,
		HealthCheckInterval:   *healthIntervalFlag,
		ClearScreen:           *clearScreenFlag,
	}

	// Load and merge configuration from files
	finalConfig, err := config.LoadConfig(cliConfig, projectRoot)
	if err != nil {
		fmt.Printf("%s %v\n", utils.Error("Error loading configuration:"), err)
		os.Exit(1)
	}

	// Normalize paths to absolute
	for i, path := range finalConfig.WatchPaths {
		// Skip empty paths
		if path == "" {
			continue
		}

		// Join with project root if path is relative
		fullPath := path
		if !filepath.IsAbs(path) {
			fullPath = filepath.Join(projectRoot, path)
		}

		// Convert to absolute path
		absPath, err := filepath.Abs(fullPath)
		if err == nil {
			finalConfig.WatchPaths[i] = absPath
			// fmt.Printf("Normalized watch path: %s -> %s\n", path, absPath)
		} else {
			fmt.Printf("Error normalizing path %s: %v\n", path, err)
		}
	}

	// Print watch configuration
	// fmt.Printf("\nWatch Configuration:\n")
	// fmt.Printf("Project Root: %s\n", projectRoot)
	// fmt.Printf("Watch Paths: %v\n", finalConfig.WatchPaths)
	// fmt.Printf("Extensions: %v\n", finalConfig.Extensions)
	// fmt.Printf("Ignore Paths: %v\n\n", finalConfig.IgnorePaths)

	// Create process manager
	pm := process.NewProcessManager(scriptPath, finalConfig)

	// Create file watcher
	fw := watcher.NewFileWatcher(finalConfig)

	// Start the watcher first
	if err := fw.Start(); err != nil {
		fmt.Printf("%s %v\n", utils.Error("Error starting watcher:"), err)
		os.Exit(1)
	}

	// Print initial status
	printStatus(finalConfig)

	// Start the process
	if err := pm.Start(); err != nil {
		fmt.Printf("%s %v\n", utils.Error("Error starting process:"), err)
		os.Exit(1)
	}

	// Main event loop
	for {
		select {
		case event := <-fw.GetChangeChannel():
			handleFileChange(event, pm)
		case err := <-fw.GetErrorChannel():
			fmt.Printf("%s %v\n", utils.Error("Error:"), err)
		}
	}
}

// findProjectRoot looks for package.json to determine project root
func findProjectRoot(scriptPath string) string {
	dir := filepath.Dir(scriptPath)
	for dir != "" && dir != "." && dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return filepath.Dir(scriptPath)
}

func handleFileChange(event types.FileEvent, pm *process.ProcessManager) {
	// Print change details
	fmt.Printf("\n%s %s\n", utils.Info("File changed:"), utils.Path(event.Path))
	fmt.Printf("%s %s\n", utils.Section("Operation:"), event.Operation)
	fmt.Printf("%s %s\n", utils.Section("Time:"), event.Time.Format("15:04:05"))

	// Restart the process
	if err := pm.Restart(); err != nil {
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
	fmt.Printf("\n%s\n", utils.Header("Nehonix quickdev"))
	fmt.Println(utils.Dimmed("================================"))

	fmt.Printf("%s %s\n", utils.Section("Watching:"), utils.Path(strings.Join(config.WatchPaths, ", ")))
	//print project github link
	fmt.Printf("%s %s\n", utils.Section("Github:"), "https://github.com/nehonix/quickdev")
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
	fmt.Printf("%s v%s\n", utils.Info("Monitoring with quickdev"), Version)
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
