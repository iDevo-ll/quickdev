package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nehonix-nhr/internal/types"
)

// Default config file names
const (
	ConfigFileName = "watchtower.config.json"
	RCFileName    = ".watchtowerrc.json"
	IgnoreFileName = ".watchtowerignore"
)

// LoadConfig loads configuration from various sources in order of precedence:
// 1. Command line arguments
// 2. watchtower.config.json or .watchtowerrc.json in project root
// 3. Default values
func LoadConfig(cliConfig *types.FileWatcherConfig, projectRoot string) (*types.FileWatcherConfig, error) {
	// Try to load config file
	configFile, err := findAndLoadConfigFile(projectRoot)
	if err != nil {
		return nil, err
	}

	// If no config file found, return CLI config
	if configFile == nil {
		return cliConfig, nil
	}

	// Merge configs with CLI taking precedence
	config := mergeConfigs(configFile, cliConfig)

	// Load ignore patterns if specified
	if config.CustomIgnoreFile != "" {
		patterns, err := loadIgnoreFile(filepath.Join(projectRoot, config.CustomIgnoreFile))
		if err != nil {
			return nil, fmt.Errorf("error loading ignore file: %w", err)
		}
		config.IgnorePaths = append(config.IgnorePaths, patterns...)
	}

	return config, nil
}

// findAndLoadConfigFile looks for config files in the project root
func findAndLoadConfigFile(projectRoot string) (*types.ConfigFile, error) {
	// Try watchtower.config.json first
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return loadConfigFile(configPath)
	}

	// Try .watchtowerrc.json next
	rcPath := filepath.Join(projectRoot, RCFileName)
	if _, err := os.Stat(rcPath); err == nil {
		return loadConfigFile(rcPath)
	}

	// No config file found, not an error
	return nil, nil
}

// loadConfigFile loads and parses a config file
func loadConfigFile(path string) (*types.ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config types.ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// loadIgnoreFile loads patterns from a .watchtowerignore file
func loadIgnoreFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
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

// mergeConfigs merges CLI config with file config, CLI takes precedence
func mergeConfigs(fileConfig *types.ConfigFile, cliConfig *types.FileWatcherConfig) *types.FileWatcherConfig {
	config := &types.FileWatcherConfig{
		// Core settings
		Enabled:     true,
		WatchPaths:  fileConfig.Watch,
		IgnorePaths: fileConfig.Ignore,
		Extensions:  fileConfig.Extensions,

		// Process Management
		GracefulShutdown:        fileConfig.GracefulShutdown,
		GracefulShutdownTimeout: fileConfig.GracefulShutdownTimeout,
		MaxRestarts:             fileConfig.MaxRestarts,
		ResetRestartsAfter:      fileConfig.ResetRestartsAfter,
		RestartDelay:            fileConfig.RestartDelay,

		// File Watching
		BatchChanges:    fileConfig.BatchChanges,
		BatchTimeout:    fileConfig.BatchTimeout,
		EnableFileHashing: fileConfig.EnableHashing,
		UsePolling:      fileConfig.UsePolling,
		PollingInterval: fileConfig.PollingInterval,
		FollowSymlinks:  fileConfig.FollowSymlinks,
		WatchDotFiles:   fileConfig.WatchDotFiles,
		CustomIgnoreFile: fileConfig.IgnoreFile,

		// Performance
		ParallelProcessing: fileConfig.ParallelProcessing,
		MemoryLimit:       fileConfig.MemoryLimit,
		MaxFileSize:       fileConfig.MaxFileSize,
		ExcludeEmptyFiles: fileConfig.ExcludeEmptyFiles,
		DebounceMs:        fileConfig.DebounceMs,

		// Monitoring
		HealthCheck:         fileConfig.HealthCheck,
		HealthCheckInterval: fileConfig.HealthCheckInterval,
		ClearScreen:         fileConfig.ClearScreen,
	}

	// Override with CLI values if provided
	if cliConfig.WatchPaths != nil && len(cliConfig.WatchPaths) > 0 {
		config.WatchPaths = cliConfig.WatchPaths
	}
	if cliConfig.IgnorePaths != nil && len(cliConfig.IgnorePaths) > 0 {
		config.IgnorePaths = cliConfig.IgnorePaths
	}
	if cliConfig.Extensions != nil && len(cliConfig.Extensions) > 0 {
		config.Extensions = cliConfig.Extensions
	}
	if cliConfig.MaxRestarts != 0 {
		config.MaxRestarts = cliConfig.MaxRestarts
	}
	if cliConfig.ResetRestartsAfter != 0 {
		config.ResetRestartsAfter = cliConfig.ResetRestartsAfter
	}
	if cliConfig.RestartDelay != 0 {
		config.RestartDelay = cliConfig.RestartDelay
	}
	if cliConfig.BatchTimeout != 0 {
		config.BatchTimeout = cliConfig.BatchTimeout
	}
	if cliConfig.PollingInterval != 0 {
		config.PollingInterval = cliConfig.PollingInterval
	}
	if cliConfig.MemoryLimit != 0 {
		config.MemoryLimit = cliConfig.MemoryLimit
	}
	if cliConfig.MaxFileSize != 0 {
		config.MaxFileSize = cliConfig.MaxFileSize
	}
	if cliConfig.DebounceMs != 0 {
		config.DebounceMs = cliConfig.DebounceMs
	}
	if cliConfig.HealthCheckInterval != 0 {
		config.HealthCheckInterval = cliConfig.HealthCheckInterval
	}

	return config
} 