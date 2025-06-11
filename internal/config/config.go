package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"quickdev/internal/types"
)

// Default config file names
const (
	ConfigFileName = "quickdev.config.json"
	RCFileName    = ".quickdevrc.json"
	IgnoreFileName = ".quickdevignore"
)

// LoadConfig loads and merges configuration from various sources
func LoadConfig(cliConfig *types.FileWatcherConfig, projectRoot string) (*types.FileWatcherConfig, error) {
	// Try to load config file
	fileConfig, err := loadConfigFile(projectRoot)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading config file: %v", err)
	}

	// Merge configs with CLI taking precedence
	finalConfig := mergeConfigs(fileConfig, cliConfig)

	// Load ignore patterns from .quickdevignore
	if patterns, err := loadIgnoreFile(finalConfig.CustomIgnoreFile, projectRoot); err == nil {
		finalConfig.IgnorePaths = append(finalConfig.IgnorePaths, patterns...)
	}

	return finalConfig, nil
}

// loadConfigFile attempts to load configuration from quickdev.config.json or .quickdevrc.json
func loadConfigFile(projectRoot string) (*types.FileWatcherConfig, error) {
	configFiles := []string{
		filepath.Join(projectRoot, "quickdev.config.json"),
		filepath.Join(projectRoot, ".quickdevrc.json"),
	}

	var config types.FileWatcherConfig
	var foundConfig bool

	for _, configFile := range configFiles {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("error parsing %s: %v", configFile, err)
		}

		foundConfig = true
		break
	}

	if !foundConfig {
		return &types.FileWatcherConfig{}, nil
	}

	return &config, nil
}

// loadIgnoreFile loads patterns from .quickdevignore file
func loadIgnoreFile(customIgnoreFile string, projectRoot string) ([]string, error) {
	var ignoreFile string
	if customIgnoreFile != "" {
		ignoreFile = customIgnoreFile
	} else {
		ignoreFile = filepath.Join(projectRoot, ".quickdevignore")
	}

	data, err := ioutil.ReadFile(ignoreFile)
	if err != nil {
		return nil, err
	}

	var patterns []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	return patterns, nil
}

// mergeConfigs merges CLI config with file config, with CLI taking precedence
func mergeConfigs(fileConfig, cliConfig *types.FileWatcherConfig) *types.FileWatcherConfig {
	// Start with file config
	result := *fileConfig

	// Override with CLI values if they're non-zero/non-empty
	if len(cliConfig.WatchPaths) > 0 && cliConfig.WatchPaths[0] != "" {
		result.WatchPaths = cliConfig.WatchPaths
	}
	if len(cliConfig.IgnorePaths) > 0 && cliConfig.IgnorePaths[0] != "" {
		result.IgnorePaths = cliConfig.IgnorePaths
	}

	// Special handling for extensions - merge instead of replace if CLI extensions are default
	defaultExts := []string{".js", ".ts", ".jsx", ".tsx"}
	isDefaultExtList := func(exts []string) bool {
		if len(exts) != len(defaultExts) {
			return false
		}
		for i, ext := range exts {
			if ext != defaultExts[i] {
				return false
			}
		}
		return true
	}

	// Only override extensions if CLI provided non-default extensions
	if len(cliConfig.Extensions) > 0 && !isDefaultExtList(cliConfig.Extensions) {
		result.Extensions = cliConfig.Extensions
	} else if len(result.Extensions) == 0 {
		// If no extensions in file config, use defaults
		result.Extensions = defaultExts
	}
	// Otherwise keep the file config extensions

	if cliConfig.GracefulShutdownTimeout > 0 {
		result.GracefulShutdownTimeout = cliConfig.GracefulShutdownTimeout
	}
	if cliConfig.MaxRestarts > 0 {
		result.MaxRestarts = cliConfig.MaxRestarts
	}
	if cliConfig.ResetRestartsAfter > 0 {
		result.ResetRestartsAfter = cliConfig.ResetRestartsAfter
	}
	if cliConfig.RestartDelay > 0 {
		result.RestartDelay = cliConfig.RestartDelay
	}
	if cliConfig.BatchTimeout > 0 {
		result.BatchTimeout = cliConfig.BatchTimeout
	}
	if cliConfig.PollingInterval > 0 {
		result.PollingInterval = cliConfig.PollingInterval
	}
	if cliConfig.DebounceMs > 0 {
		result.DebounceMs = cliConfig.DebounceMs
	}
	if cliConfig.MaxFileSize > 0 {
		result.MaxFileSize = cliConfig.MaxFileSize
	}
	if cliConfig.HealthCheckInterval > 0 {
		result.HealthCheckInterval = cliConfig.HealthCheckInterval
	}
	if cliConfig.MemoryLimit > 0 {
		result.MemoryLimit = cliConfig.MemoryLimit
	}
	if cliConfig.CustomIgnoreFile != "" {
		result.CustomIgnoreFile = cliConfig.CustomIgnoreFile
	}

	// Boolean flags
	result.GracefulShutdown = cliConfig.GracefulShutdown
	result.BatchChanges = cliConfig.BatchChanges
	result.EnableFileHashing = cliConfig.EnableFileHashing
	result.UsePolling = cliConfig.UsePolling
	result.FollowSymlinks = cliConfig.FollowSymlinks
	result.WatchDotFiles = cliConfig.WatchDotFiles
	result.ParallelProcessing = cliConfig.ParallelProcessing
	result.ExcludeEmptyFiles = cliConfig.ExcludeEmptyFiles
	result.HealthCheck = cliConfig.HealthCheck
	result.ClearScreen = cliConfig.ClearScreen

	return &result
}