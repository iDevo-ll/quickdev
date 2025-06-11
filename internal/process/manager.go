package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"nehonix-nhr/internal/types"
)

// ProcessManager handles the running process
type ProcessManager struct {
	cmd           *exec.Cmd
	mutex         sync.Mutex
	stats         *types.RestartStats
	config        *types.FileWatcherConfig
	lastRestart   time.Time
	restartCount  int
	isRunning     bool
	scriptPath    string
	processEnv    []string
	projectRoot   string
}

// NewProcessManager creates a new process manager
func NewProcessManager(scriptPath string, config *types.FileWatcherConfig) *ProcessManager {
	return &ProcessManager{
		scriptPath:  scriptPath,
		config:     config,
		stats:      &types.RestartStats{},
		processEnv: os.Environ(),
		projectRoot: findProjectRoot(scriptPath),
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

// determineRunner determines which runner to use based on file extension and project setup
func (pm *ProcessManager) determineRunner() (string, []string, error) {
	ext := filepath.Ext(pm.scriptPath)
	
	// For JavaScript files, use Node directly
	if ext == ".js" || ext == ".jsx" {
		return "node", []string{pm.scriptPath}, nil
	}

	// For TypeScript files, we need to determine the appropriate runner
	if ext == ".ts" || ext == ".tsx" {
		// Check for local tsx
		if _, err := os.Stat(filepath.Join(pm.projectRoot, "node_modules", ".bin", "tsx")); err == nil {
			return filepath.Join(pm.projectRoot, "node_modules", ".bin", "tsx"), []string{pm.scriptPath}, nil
		}

		// Check for local ts-node
		if _, err := os.Stat(filepath.Join(pm.projectRoot, "node_modules", ".bin", "ts-node")); err == nil {
			return filepath.Join(pm.projectRoot, "node_modules", ".bin", "ts-node"), []string{"--esm", pm.scriptPath}, nil
		}

		// Check for global tsx
		if tsxPath, err := exec.LookPath("tsx"); err == nil {
			return tsxPath, []string{pm.scriptPath}, nil
		}

		// Check for global ts-node
		if tsNodePath, err := exec.LookPath("ts-node"); err == nil {
			return tsNodePath, []string{"--esm", pm.scriptPath}, nil
		}

		// If no TypeScript runner is found, suggest installation
		return "", nil, fmt.Errorf("no TypeScript runner found. Please install tsx or ts-node:\nnpm install -g tsx\n   or\nnpm install -g ts-node")
	}

	return "", nil, fmt.Errorf("unsupported file extension: %s", ext)
}

// Start starts the process
func (pm *ProcessManager) Start() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.isRunning {
		return fmt.Errorf("process is already running")
	}

	return pm.startProcess()
}

// Restart restarts the process
func (pm *ProcessManager) Restart(reason string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Check if we've exceeded max restarts
	if pm.config.MaxRestarts > 0 {
		now := time.Now()
		if now.Sub(pm.lastRestart) > time.Duration(pm.config.ResetRestartsAfter)*time.Millisecond {
			pm.restartCount = 0
		} else if pm.restartCount >= pm.config.MaxRestarts {
			return fmt.Errorf("exceeded maximum number of restarts (%d)", pm.config.MaxRestarts)
		}
	}

	startTime := time.Now()

	// Stop the current process
	if pm.isRunning {
		if err := pm.stopProcess(); err != nil {
			return fmt.Errorf("failed to stop process: %v", err)
		}
	}

	// Start the new process
	if err := pm.startProcess(); err != nil {
		pm.recordRestartFailure(startTime, reason)
		return fmt.Errorf("failed to start process: %v", err)
	}

	pm.recordRestartSuccess(startTime, reason)
	return nil
}

// Stop stops the process
func (pm *ProcessManager) Stop() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	return pm.stopProcess()
}

// startProcess starts the managed process
func (pm *ProcessManager) startProcess() error {
	// Add delay before restart if configured
	if pm.config.RestartDelay > 0 {
		time.Sleep(time.Duration(pm.config.RestartDelay) * time.Millisecond)
	}

	// Clear screen if configured
	if pm.config.ClearScreen {
		fmt.Print("\033[H\033[2J")
	}

	// Determine the appropriate runner
	runner, args, err := pm.determineRunner()
	if err != nil {
		return err
	}

	// Create the command
	pm.cmd = exec.Command(runner, args...)
	pm.cmd.Dir = pm.projectRoot // Set working directory to project root
	pm.cmd.Env = pm.processEnv
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr

	// Start the process
	if err := pm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}

	pm.isRunning = true
	pm.lastRestart = time.Now()
	pm.restartCount++

	// Monitor the process
	go func() {
		pm.cmd.Wait()
		pm.mutex.Lock()
		pm.isRunning = false
		pm.mutex.Unlock()
	}()

	return nil
}

// stopProcess stops the managed process
func (pm *ProcessManager) stopProcess() error {
	if !pm.isRunning {
		return nil
	}

	if pm.config.GracefulShutdown {
		// Send SIGTERM first
		if err := pm.cmd.Process.Signal(os.Interrupt); err != nil {
			return err
		}

		// Wait for graceful shutdown
		done := make(chan error)
		go func() {
			done <- pm.cmd.Wait()
		}()

		select {
		case <-time.After(time.Duration(pm.config.GracefulShutdownTimeout) * time.Second):
			// Force kill if timeout
			if err := pm.cmd.Process.Kill(); err != nil {
				return err
			}
		case err := <-done:
			if err != nil {
				return err
			}
		}
	} else {
		// Force kill
		if err := pm.cmd.Process.Kill(); err != nil {
			return err
		}
	}

	pm.isRunning = false
	return nil
}

// recordRestartSuccess records a successful restart
func (pm *ProcessManager) recordRestartSuccess(startTime time.Time, reason string) {
	duration := time.Since(startTime)
	
	// Update stats
	pm.stats.TotalRestarts++
	pm.stats.SuccessfulRestarts++
	pm.stats.LastRestart = &startTime

	// Update timing stats
	if duration < pm.stats.FastestRestart || pm.stats.FastestRestart == 0 {
		pm.stats.FastestRestart = duration
	}
	if duration > pm.stats.SlowestRestart {
		pm.stats.SlowestRestart = duration
	}

	// Calculate average
	totalTime := pm.stats.AverageRestartTime * time.Duration(pm.stats.TotalRestarts-1)
	pm.stats.AverageRestartTime = (totalTime + duration) / time.Duration(pm.stats.TotalRestarts)

	// Add to history
	pm.addToHistory(startTime, reason, duration, true)
}

// recordRestartFailure records a failed restart
func (pm *ProcessManager) recordRestartFailure(startTime time.Time, reason string) {
	pm.stats.TotalRestarts++
	pm.stats.FailedRestarts++
	pm.addToHistory(startTime, reason, time.Since(startTime), false)
}

// addToHistory adds a restart event to the history
func (pm *ProcessManager) addToHistory(timestamp time.Time, reason string, duration time.Duration, success bool) {
	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memUsage := &types.MemoryUsage{
		HeapTotal:     memStats.HeapSys,
		HeapUsed:      memStats.HeapAlloc,
		External:      memStats.HeapReleased,
		ProcessMemory: memStats.Sys,
	}

	// Parse file count from reason if it's a batch change
	fileCount := 1
	if strings.Contains(reason, "files changed") {
		parts := strings.Split(reason, " ")
		for i, part := range parts {
			if part == "files" && i > 0 {
				if count, err := strconv.Atoi(parts[i-1]); err == nil {
					fileCount = count
				}
				break
			}
		}
	}

	entry := types.RestartHistoryEntry{
		Timestamp:   timestamp,
		Reason:      reason,
		Duration:    duration,
		Success:     success,
		FileCount:   fileCount,
		MemoryUsage: memUsage,
	}

	// Limit history size
	if len(pm.stats.RestartHistory) >= 100 {
		pm.stats.RestartHistory = pm.stats.RestartHistory[1:]
	}
	pm.stats.RestartHistory = append(pm.stats.RestartHistory, entry)
}

// GetStats returns the current restart statistics
func (pm *ProcessManager) GetStats() *types.RestartStats {
	return pm.stats
}

// IsRunning returns whether the process is currently running
func (pm *ProcessManager) IsRunning() bool {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.isRunning
}

// SetEnvironment sets environment variables for the process
func (pm *ProcessManager) SetEnvironment(env []string) {
	pm.processEnv = env
} 