# Nehonix File Reloader (NHR)

A lightning-fast file watcher and reloader for TypeScript/JavaScript applications, built with Go. This tool is designed as a high-performance alternative to nodemon, specifically optimized for fortify2-js server and similar Node.js applications.

## Features

- 🚀 Blazing fast file watching and reloading
- 💡 Smart debouncing to prevent multiple reloads
- 🎯 Configurable file extensions and ignore patterns
- 📁 Recursive directory watching
- 🛡️ Safe process management
- 🎨 Beautiful colored console output

## Installation

1. Make sure you have Go installed (version 1.21 or higher)
2. Clone this repository
3. Build the binary:
   ```bash
   go build -o nhr
   ```

## Usage

```bash
./nhr -script your-script.js [options]
```

### Command Line Options

- `-script`: Path to the script to run (required)
- `-watch`: Directories to watch, comma-separated (default: ".")
- `-ignore`: Directories to ignore, comma-separated (default: "node_modules,dist,.git")
- `-ext`: File extensions to watch, comma-separated (default: ".js,.ts,.jsx,.tsx")

### Examples

Watch current directory and run `server.js`:

```bash
./nhr -script server.js
```

Watch specific directories:

```bash
./nhr -script server.js -watch "src,config"
```

Custom ignore patterns:

```bash
./nhr -script server.js -ignore "node_modules,dist,coverage"
```

Watch specific file extensions:

```bash
./nhr -script server.js -ext ".ts,.js"
```

## How It Works

Nehonix File Reloader uses Go's efficient file system notifications (fsnotify) to watch for file changes. When a change is detected:

1. The change event is debounced to prevent multiple rapid reloads
2. The running process is gracefully terminated
3. The script is restarted automatically
4. All stdout/stderr output is properly forwarded to the console

## Performance

Unlike Node.js-based file watchers, NHR is built in Go, providing:

- Minimal memory footprint
- Fast file system events processing
- Efficient process management
- No dependency overhead

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
