package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Find the project root (where package.json is)
func findProjectRoot(startPath string) string {
	currentPath := startPath
	for {
		if fileExists(filepath.Join(currentPath, "package.json")) {
			return currentPath
		}
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			return startPath // If we can't find project root, return original path
		}
		currentPath = parent
	}
}

// Check if package.json has type: module
func isESMProject(projectRoot string) bool {
	data, err := os.ReadFile(filepath.Join(projectRoot, "package.json"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), `"type": "module"`)
}

// Check for TypeScript configuration
func getTypeScriptConfig(projectRoot string) map[string]string {
	config := make(map[string]string)

	// Check for tsconfig.json
	tsconfigPath := filepath.Join(projectRoot, "tsconfig.json")
	if fileExists(tsconfigPath) {
		data, err := os.ReadFile(tsconfigPath)
		if err == nil {
			// Check for module type
			if strings.Contains(string(data), `"module": "ES"`) || 
			   strings.Contains(string(data), `"module": "ESNext"`) ||
			   strings.Contains(string(data), `"module": "esnext"`) {
				config["moduleType"] = "esm"
			}
		}
	}

	// Check for local installations
	nodeModulesBin := filepath.Join(projectRoot, "node_modules", ".bin")
	
	if fileExists(filepath.Join(nodeModulesBin, "tsx")) {
		config["runner"] = filepath.Join(nodeModulesBin, "tsx")
	} else if fileExists(filepath.Join(nodeModulesBin, "ts-node")) {
		config["runner"] = filepath.Join(nodeModulesBin, "ts-node")
	}

	return config
}

func main() {
	// Parse command line arguments
	scriptPath := flag.String("script", "", "Path to the script to run")
	flag.Parse()

	if *scriptPath == "" {
		fmt.Println("Error: Please provide a script to run using -script flag")
		os.Exit(1)
	}

	// Convert to absolute path
	absScriptPath, err := filepath.Abs(*scriptPath)
	if err != nil {
		fmt.Printf("Error resolving path: %s\n", err)
		os.Exit(1)
	}

	// Find project root
	projectRoot := findProjectRoot(filepath.Dir(absScriptPath))
	fmt.Printf("📂 Project root: %s\n", projectRoot)

	fmt.Println("🚀 Nehonix File Reloader (NHR) started")
	fmt.Printf("👀 Watching: %s\n", *scriptPath)

	// Start the initial process
	cmd := startProcess(absScriptPath, projectRoot)
	
	// Simple file watching using polling (for testing)
	lastModTime := getFileModTime(*scriptPath)
	
	for {
		time.Sleep(1 * time.Second)
		currentModTime := getFileModTime(*scriptPath)
		
		if currentModTime != lastModTime {
			fmt.Printf("📝 File changed: %s\n", *scriptPath)
			
			// Kill the existing process
			if cmd != nil && cmd.Process != nil {
				fmt.Println("🔄 Restarting process...")
				cmd.Process.Kill()
				cmd.Wait()
			}
			
			// Start the new process
			cmd = startProcess(absScriptPath, projectRoot)
			lastModTime = currentModTime
		}
	}
}

func startProcess(scriptPath, projectRoot string) *exec.Cmd {
	var cmd *exec.Cmd

	// Check file extension
	ext := strings.ToLower(filepath.Ext(scriptPath))
	switch ext {
	case ".ts", ".tsx":
		// Get TypeScript configuration
		tsConfig := getTypeScriptConfig(projectRoot)
		isESM := isESMProject(projectRoot) || tsConfig["moduleType"] == "esm"

		// Determine runner and args based on configuration
		if runner, exists := tsConfig["runner"]; exists {
			// Use local installation
			var args []string
			if strings.HasSuffix(runner, "ts-node") {
				args = []string{
					"--transpile-only",
				}
				if isESM {
					args = append(args, "--esm")
				}
				args = append(args, scriptPath)
			} else if strings.HasSuffix(runner, "tsx") {
				args = []string{scriptPath}
			}
			cmd = exec.Command(runner, args...)
		} else {
			// Use global installation or npx
			if isESM {
				cmd = exec.Command("npx", "--yes", "tsx", scriptPath)
			} else {
				cmd = exec.Command("npx", "--yes", "ts-node", "--transpile-only", scriptPath)
			}
		}

		// Set working directory to project root for proper module resolution
		cmd.Dir = projectRoot
		
		// Set NODE_ENV if not set
		env := os.Environ()
		hasNodeEnv := false
		for _, e := range env {
			if strings.HasPrefix(e, "NODE_ENV=") {
				hasNodeEnv = true
				break
			}
		}
		if !hasNodeEnv {
			env = append(env, "NODE_ENV=development")
		}
		cmd.Env = env

	default:
		// For JavaScript files, use node directly
		cmd = exec.Command("node", scriptPath)
		cmd.Dir = projectRoot
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting process: %s\n", err)
		return nil
	}
	
	fmt.Printf("✨ Started process: %s\n", scriptPath)
	return cmd
}

func getFileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
} 