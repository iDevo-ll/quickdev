package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"quickdev/internal/types"
)

// ProcessManager handles the running process
type ProcessManager struct {
	config       *types.FileWatcherConfig
	scriptPath   string
	cmd          *exec.Cmd
	mutex        sync.Mutex
	restartStats *types.RestartStats
	startTime    time.Time
}

// NewProcessManager creates a new process manager
func NewProcessManager(scriptPath string, config *types.FileWatcherConfig) *ProcessManager {
	return &ProcessManager{
		config:     config,
		scriptPath: scriptPath,
		restartStats: &types.RestartStats{
			RestartHistory: make([]types.RestartHistoryEntry, 0),
		},
		startTime: time.Now(),
	}
}

// Start starts the process
func (pm *ProcessManager) Start() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if err := pm.startProcess(); err != nil {
		return fmt.Errorf("error starting process: %v", err)
	}

	return nil
}

// startProcess starts the managed process
func (pm *ProcessManager) startProcess() error {
	// Determine the runner based on file extension
	runner := pm.determineRunner()
	if runner == "" {
		return fmt.Errorf("unsupported script type: %s", filepath.Ext(pm.scriptPath))
	}

	var cmd *exec.Cmd

	// Build command based on runner type
	if runner == "ts-node" || runner == "tsx" {
		// For TypeScript files, use npx to ensure we use the local installation
		args := []string{"-y", runner}

		// Add configured flags if available
		if pm.config.TSNodeFlags != "" {
			flags := strings.Split(pm.config.TSNodeFlags, " ")
			args = append(args, flags...)
		} else if runner == "ts-node" {
			// Default ts-node flags if none configured
			args = append(args, "--esm")
		}

		// Add the script path
		args = append(args, pm.scriptPath)
		cmd = exec.Command("npx", args...)
	} else {
		// For JavaScript files, use node directly
		cmd = exec.Command(runner, pm.scriptPath)
	}

	// Set up command environment
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	pm.cmd = cmd

	// Start process
	if err := pm.cmd.Start(); err != nil {
		return err
	}

	// Monitor process in background
	go func() {
		err := pm.cmd.Wait()
		pm.handleProcessExit(err)
	}()

	return nil
}

// determineRunner determines which runner to use based on file extension and project setup
func (pm *ProcessManager) determineRunner() string {
	ext := strings.ToLower(filepath.Ext(pm.scriptPath))
	fmt.Println("Running...")
	switch ext {
	case ".ts", ".tsx":
		// Use configured TypeScript runner if specified
		if pm.config.TypeScriptRunner != "" {
			return pm.config.TypeScriptRunner
		}

		// Check for local installations first using npx
		if err := exec.Command("npx", "-y", "tsx", "--version").Run(); err == nil {
			return "tsx"
		}
		if err := exec.Command("npx", "-y", "ts-node", "--version").Run(); err == nil {
			return "ts-node"
		}

		// Fallback to checking global installations
		if _, err := exec.LookPath("tsx"); err == nil {
			return "tsx"
		}
		if _, err := exec.LookPath("ts-node"); err == nil {
			return "ts-node"
		}
		return ""
	case ".js", ".jsx":
		return "node"
	default:
		return ""
	}
}

// handleProcessExit handles the process exit
func (pm *ProcessManager) handleProcessExit(err error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Update restart stats
	exitTime := time.Now()
	uptime := exitTime.Sub(pm.startTime)
	exitCode := 0
	errorMsg := ""

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		errorMsg = err.Error()
	}

	// Add to history
	pm.restartStats.RestartHistory = append(pm.restartStats.RestartHistory, types.RestartHistoryEntry{
		Time:     exitTime,
		ExitCode: exitCode,
		Error:    errorMsg,
		Duration: uptime,
	})

	// Update stats
	pm.restartStats.TotalRestarts++
	pm.restartStats.LastRestart = exitTime
	pm.restartStats.LastExitCode = exitCode
	pm.restartStats.LastErrorMessage = errorMsg

	// Update uptime stats
	if len(pm.restartStats.RestartHistory) == 1 {
		pm.restartStats.ShortestUptime = uptime
		pm.restartStats.LongestUptime = uptime
		pm.restartStats.AverageUptime = uptime
	} else {
		if uptime < pm.restartStats.ShortestUptime {
			pm.restartStats.ShortestUptime = uptime
		}
		if uptime > pm.restartStats.LongestUptime {
			pm.restartStats.LongestUptime = uptime
		}

		// Calculate new average
		total := time.Duration(0)
		for _, entry := range pm.restartStats.RestartHistory {
			total += entry.Duration
		}
		pm.restartStats.AverageUptime = total / time.Duration(len(pm.restartStats.RestartHistory))
	}
}

// Restart restarts the process
func (pm *ProcessManager) Restart() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Stop current process
	if pm.cmd != nil && pm.cmd.Process != nil {
		if pm.config.GracefulShutdown {
			// Send SIGTERM and wait for graceful shutdown
			if err := pm.cmd.Process.Signal(os.Interrupt); err != nil {
				pm.cmd.Process.Kill()
			} else {
				// Wait for process to exit or timeout
				done := make(chan error)
				go func() {
					done <- pm.cmd.Wait()
				}()

				select {
				case <-done:
					// Process exited gracefully
				case <-time.After(time.Duration(pm.config.GracefulShutdownTimeout) * time.Second):
					// Timeout, force kill
					pm.cmd.Process.Kill()
				}
			}
		} else {
			// Force kill
			pm.cmd.Process.Kill()
		}
	}

	// Delay before restart if configured
	if pm.config.RestartDelay > 0 {
		time.Sleep(time.Duration(pm.config.RestartDelay) * time.Millisecond)
	}

	// Start new process
	return pm.startProcess()
}

// Stop stops the process
func (pm *ProcessManager) Stop() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.cmd != nil && pm.cmd.Process != nil {
		if pm.config.GracefulShutdown {
			if err := pm.cmd.Process.Signal(os.Interrupt); err != nil {
				return pm.cmd.Process.Kill()
			}
			return pm.cmd.Wait()
		}
		return pm.cmd.Process.Kill()
	}
	return nil
}

// GetRestartStats returns the current restart statistics
func (pm *ProcessManager) GetRestartStats() *types.RestartStats {
	return pm.restartStats
}