# Nehonix WatchTower

A professional-grade file watcher and development server for TypeScript/JavaScript applications, built with Go. This tool is designed as a high-performance alternative to nodemon, specifically optimized for modern Node.js/TypeScript applications with advanced features not found in other reloaders.

_Designed for fortify2-js server but can be used in other applications_

## Features

### Core Capabilities

- High-performance file watching and reloading
- Intelligent TypeScript/JavaScript detection and handling
- Advanced project configuration detection
- Sophisticated file system monitoring
- Production-grade process management
- Professional CLI interface with color-coded output

### Advanced Features

**Smart Process Management**

- Graceful shutdown with configurable timeout
- Process restart statistics and history
- Maximum restart limits with auto-reset
- Configurable restart delay
- Environment variable preservation

**Advanced File Watching**

- File hashing for precise change detection
- Batch processing of file changes
- Polling and filesystem event support
- Symlink following capability
- Custom ignore patterns via file

**Performance Optimization**

- Parallel file processing
- Memory usage monitoring and limits
- File size restrictions
- Efficient batch change processing
- Smart event debouncing

**Health Monitoring**

- Process health checks
- Memory usage tracking
- Restart statistics
- Error tracking and history
- Performance metrics

## Installation

1. Make sure you have Go installed (version 1.21 or higher)
2. Install using Go:
   ```bash
   go install github.com/nehonix/watchtower@latest
   ```
   Or build from source:
   ```bash
   git clone https://github.com/nehonix/watchtower.git
   cd watchtower
   go build -o watchtower
   ```

## Configuration

WatchTower supports multiple ways to configure its behavior:

### 1. Configuration File

Create a `watchtower.config.json` (or `.watchtowerrc.json`) in your project root:

```json
{
  "script": "src/server.ts",
  "watch": ["src", "config"],
  "ignore": ["node_modules", "dist", "coverage"],
  "extensions": [".ts", ".js", ".jsx", ".tsx"],

  "gracefulShutdown": true,
  "gracefulShutdownTimeout": 5,
  "maxRestarts": 5,
  "resetRestartsAfter": 60000,
  "restartDelay": 100,

  "batchChanges": true,
  "batchTimeout": 300,
  "enableHashing": true,
  "usePolling": false,
  "pollingInterval": 100,
  "followSymlinks": false,
  "watchDotFiles": false,
  "ignoreFile": ".watchtowerignore",

  "parallelProcessing": true,
  "memoryLimit": 500,
  "maxFileSize": 10,
  "excludeEmptyFiles": true,
  "debounceMs": 250,

  "healthCheck": true,
  "healthCheckInterval": 30,
  "clearScreen": true,

  "typescriptRunner": "tsx",
  "tsNodeFlags": "--esm"
}
```

### 2. Ignore File

Create a `.watchtowerignore` file to specify patterns to ignore:

```text
# Comments are supported
*.log
*.tmp
temp/
**/*.test.js
coverage/
dist/
build/
.git/
```

### 3. Command Line Arguments

All configuration options can be overridden via command line arguments:

```bash
watchtower -script server.js \
  --watch="src,config" \
  --ignore="node_modules,dist" \
  --ext=".ts,.js"
```

Command line arguments take precedence over configuration file settings.

### Configuration Priority

WatchTower uses the following priority order when loading configuration:

1. Command line arguments (highest priority)
2. `watchtower.config.json` or `.watchtowerrc.json`
3. `.watchtowerignore` file
4. Default values (lowest priority)

## Usage Examples

### Basic Usage

```bash
watchtower -script your-script.js
```

### Advanced Usage

```bash
watchtower -script server.js \
  --watch="src,config" \
  --ignore="node_modules,dist,coverage" \
  --ext=".ts,.js,.jsx,.tsx" \
  --batch=true \
  --hash=true \
  --health=true \
  --graceful=true
```

### Command Line Options

#### Core Options

- `-script` - Path to the script to run (required)
- `-watch` - Directories to watch, comma-separated (default: ".")
- `-ignore` - Directories to ignore, comma-separated (default: "node_modules,dist,.git")
- `-ext` - File extensions to watch (default: ".js,.ts,.jsx,.tsx")

#### Process Management

- `-graceful` - Enable graceful shutdown (default: true)
- `-graceful-timeout` - Graceful shutdown timeout in seconds (default: 5)
- `-max-restarts` - Maximum number of restarts (default: 0 = unlimited)
- `-reset-after` - Reset restart count after X milliseconds (default: 60000)
- `-restart-delay` - Delay before restart in milliseconds (default: 100)

#### File Watching

- `-batch` - Enable batch processing of changes (default: true)
- `-batch-timeout` - Batch timeout in milliseconds (default: 300)
- `-hash` - Enable file hashing for precise change detection (default: true)
- `-polling` - Use polling instead of filesystem events (default: false)
- `-polling-interval` - Polling interval in milliseconds (default: 100)
- `-follow-symlinks` - Follow symbolic links (default: false)
- `-watch-dot` - Watch dot files (default: false)
- `-ignore-file` - Path to custom ignore file

#### Performance

- `-parallel` - Enable parallel processing (default: true)
- `-memory` - Memory limit in MB (default: 500)
- `-max-size` - Maximum file size in MB (default: 10)
- `-exclude-empty` - Exclude empty files (default: true)
- `-debounce` - Debounce time in milliseconds (default: 250)

#### Monitoring

- `-health` - Enable health checking (default: true)
- `-health-interval` - Health check interval in seconds (default: 30)
- `-clear` - Clear screen on restart (default: true)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.
