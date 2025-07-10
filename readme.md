# Nehonix QuickDev

A professional-grade file watcher and development server for TypeScript/JavaScript applications. This tool is designed as a high-performance alternative to nodemon, specifically optimized for modern Node.js/TypeScript applications with advanced features not found in other reloaders.

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

### Option 1: Install via npm (Recommended)
```bash
npm install -g nquickdev
```

### Option 2: Install via Go
1. Make sure you have Go installed (version 1.21 or higher)
2. Install using Go:
   ```bash
   go install github.com/nehonix/quickdev@latest
   ```
   Or build from source:
   ```bash
   git clone https://github.com/nehonix/quickdev.git
   cd quickdev
   go build -o quickdev
   ```

## Configuration

quickdev supports multiple ways to configure its behavior:

### 1. Configuration File

Create a `quickdev.config.json` (or `.quickdevrc.json`) in your project root:

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
    "ignoreFile": ".quickdevignore",

    "parallelProcessing": true,
    "memoryLimit": 500,
    "maxFileSize": 10,
    "excludeEmptyFiles": true,
    "debounceMs": 250,

    "healthCheck": true,
    "healthCheckInterval": 30,
    "clearScreen": true,

    "typescriptRunner": "bun",
    "tsNodeFlags": "--esm"
}

```
Always use bun for better perfomance and fast ops (default: tsx). Install bun in your machine._

### Configuration Options

#### Core Settings

- `script` - Path to the script to run (required)
- `watch` - Directories to watch, array of paths
- `ignore` - Directories to ignore, array of paths
- `extensions` - File extensions to watch

#### Process Management

- `gracefulShutdown` - Enable graceful shutdown (default: true)
- `gracefulShutdownTimeout` - Graceful shutdown timeout in seconds (default: 5)
- `maxRestarts` - Maximum number of restarts (default: 5)
- `resetRestartsAfter` - Reset restart count after X milliseconds (default: 60000)
- `restartDelay` - Delay before restart in milliseconds (default: 100)

#### File Watching

- `batchChanges` - Enable batch processing of changes (default: true)
- `batchTimeout` - Batch timeout in milliseconds (default: 300)
- `enableHashing` - Enable file hashing for precise change detection (default: true)
- `usePolling` - Use polling instead of filesystem events (default: false)
- `pollingInterval` - Polling interval in milliseconds (default: 100)
- `followSymlinks` - Follow symbolic links (default: false)
- `watchDotFiles` - Watch dot files (default: false)
- `ignoreFile` - Path to custom ignore file

#### Performance

- `parallelProcessing` - Enable parallel processing (default: true)
- `memoryLimit` - Memory limit in MB (default: 500)
- `maxFileSize` - Maximum file size in MB (default: 10)
- `excludeEmptyFiles` - Exclude empty files (default: true)
- `debounceMs` - Debounce time in milliseconds (default: 250)

#### Monitoring

- `healthCheck` - Enable health checking (default: true)
- `healthCheckInterval` - Health check interval in seconds (default: 30)
- `clearScreen` - Clear screen on restart (default: true)

#### TypeScript Settings (leave blank to use default runner (recommanded))

- `typescriptRunner` - TypeScript execution engine to use ("tsx" or "ts-node", default: tries "tsx" first, then "ts-node")
- `tsNodeFlags` - Additional flags for the TypeScript runner (default: "--esm" for ts-node)

### 2. Ignore File

Create a `.quickdevignore` file to specify patterns to ignore:

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
quickdev -script server.js \
  --watch="src,config" \
  --ignore="node_modules,dist" \
  --ext=".ts,.js"
```

Command line arguments take precedence over configuration file settings.

### Configuration Priority

quickdev uses the following priority order when loading configuration:

1. Command line arguments (highest priority)
2. `quickdev.config.json` or `.quickdevrc.json`
3. `.quickdevignore` file
4. Default values (lowest priority)

### TypeScript Support

quickdev provides robust TypeScript support with configurable execution options:

#### TypeScript Runner Selection

You can specify your preferred TypeScript runner in the config:

```json
{
  "typescriptRunner": "tsx", // or "ts-node"
  "tsNodeFlags": "--esm" // or any other flags you need
}
```

The runner selection follows this order:

1. Uses the specified `typescriptRunner` if configured
2. Falls back to `tsx` if available
3. Falls back to `ts-node` if available
4. Fails if no TypeScript runner is found

#### TypeScript Flags

- For `tsx`: Pass any additional flags via `tsNodeFlags`
- For `ts-node`: Default flags are `--esm`, can be overridden via `tsNodeFlags`

Example with custom flags:

```json
{
  "typescriptRunner": "ts-node",
  "tsNodeFlags": "--esm --transpileOnly --compilerOptions '{\"module\":\"ESNext\"}'"
}
```

## Usage Examples

### Basic Usage

```bash
quickdev -script your-script.js
```

### Advanced Usage

```bash
quickdev -script server.js \
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
