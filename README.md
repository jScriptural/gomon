# gomon
**gomon** is a lightweight, efficient CLI utility written in Go that automatically restarts your applications (or rebuilds and restarts) upon detecting file changes. It is designed for rapid development workflows, supporting custom build/run commands, hooks, debounced restarts, signal forwarding, and flexible file watching with ignore patterns.

## Features

- **fsnotify-based file watching** with recursive directory traversal and recovery mechanisms.
- **Configurable debouncing** to handle rapid successive file changes efficiently.
- **Lifecycle hooks**: `prebuild`, `postbuild`, `prestart`, `poststart`.
- **Build + Run workflow** or direct execution.
- **Signal forwarding** (SIGTERM, SIGINT, etc.) to child processes.
- **CLI flags** and `gomon.config.json` support for easy configuration.
- **Colorized, structured logging** with slog.
- **Graceful process management** with context cancellation and cleanup.

## Quick Start

1. **Install**:
   ```bash
   go install github.com/jscriptural/gomon@latest
   ```

2. **Run** (example for a Go project):
   ```bash
   gomon -b "go build -o ./app" -r "./app"
   ```

   Or with a config file (auto-generated on first run).

3. **Common flags**:
   - `-b, --build`: Build command
   - `-r, --run`: Run command
   - `-d, --delay`: Debounce delay (e.g., `500ms`, `1s`)
   - `-e, --ext`: File extensions to watch
   - `-x, --ignore`: Patterns/directories to ignore
   - `--prebuild`, `--postbuild`, etc.

## System Architecture and Design

gomon follows a clean, modular architecture with separation of concerns:

```
main.go
├── internal/config     (configuration loading, flags, schema)
├── internal/logger     (colorized slog handler)
├── internal/monitor    (file system watching + filtering)
└── internal/runner     (debouncer, executor, hooks, signals)
```

### Core Components

1. **Config (`internal/config`)**
   - Loads from `gomon.config.json` (creates default if missing).
   - Parses pflag CLI arguments with precedence over file.
   - Custom `Duration` type for JSON marshaling/unmarshaling of durations.
   - Supports `watch`, `ignore`, `ext`, `env`, hooks, `build`, `run`, etc.

2. **Logger (`internal/logger`)**
   - Custom `slog.Handler` with ANSI colors and prefixed output (`gomon`).
   - Timestamps, leveled output, and attribute formatting.

3. **Monitor (`internal/monitor`)**
   - Uses `fsnotify` for event-driven watching.
   - `filter()`: Resolves watch paths, handles files vs directories, builds ignore list.
   - Recursive `filepath.WalkDir` to register directories.
   - Self-healing watcher: Recreates `fsnotify.Watcher` on errors and re-registers paths.
   - Event filtering by extension, ignore patterns (exact match, glob, substring).
   - On relevant events (`Write`, `Create`, `Remove`), triggers the runner.

4. **Runner / Executor (`internal/runner`)**
   - Manages the full build/run lifecycle.
   - Handles process lifecycle with `context.Context` for cancellation.
   - Signal forwarding to child processes.
   - Hook execution (pre/post stages).
   - Graceful termination of old processes before starting new ones.

### Debouncer Implementation (`debouncer.go`)

The debouncer prevents excessive restarts during bursts of file changes (e.g., IDE auto-save or large refactors).

```go
// Simplified core logic
func (e *Executor) debouncer() {
    for {
        <-e.isEvent          // Triggered by file events
        dur := time.Until(time.Now().Add(time.Duration(e.config.Delay)))
        
        if e.timer != nil {
            e.timer.Reset(dur)
        } else {
            e.timer = time.AfterFunc(dur, func() {
                e.trigger()  // Full build/run cycle
            })
        }
    }
}

func (e *Executor) Trigger() {
    e.isEvent <- true
}
```

- **How it works**: Each file event sends to an unbuffered channel. The debouncer resets a `time.Timer` to fire after the configured delay. Only after the delay elapses (with no new events) does the actual `trigger()` execute.
- **Benefits**: Reduces CPU/load, avoids partial builds, improves developer experience.
- Integrated in `NewExecutor` as a background goroutine.

### Program Flow

1. **Startup** (`main.go`):
   - Initialize logger.
   - Load/merge config (file + flags).
   - Create root context and `Monitor`.
   - Start watching.

2. **Watching** (`monitor.Watch()`):
   - Filter watch/ignore paths.
   - Register directories with fsnotify (background recovery loop).
   - Launch event loop (`watchLoop`).
   - Initial executor trigger.

3. **Event Handling**:
   - Filter events → debounce channel.
   - Debouncer timer fires → `executor.trigger()`.

4. **Execution Cycle** (`executor.trigger()`):
   - Lock mutex.
   - Kill previous child if running.
   - Run `prebuild` (if set).
   - Run `build` command (if set).
   - Run `postbuild`.
   - Run `prestart`.
   - Start main process (`start()`).
   - Schedule `poststart` (non-blocking).
   - Unlock; child runs asynchronously with `cmd.Wait()`.

5. **Shutdown**:
   - Context cancellation propagates.
   - Signals forwarded or cause clean exit.

## Configuration Example (`gomon.config.json`)

```json
{
  "watch": [".", "src"],
  "ext": [".go", ".js", ".ts"],
  "ignore": [".git", "node_modules", "*.tmp"],
  "build": "go build -o ./bin/app",
  "run": "./bin/app",
  "delay": "300ms",
  "hooks": {
    "prebuild": "echo 'Preparing build...'",
    "poststart": "echo 'App started!'"
  },
  "env": ["DEBUG=true"]
}
```

## Project Structure

```
.
├── main.go
├── go.mod
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── schema.go
│   ├── logger/
│   │   ├── colors.go
│   │   └── writer.go
│   ├── monitor/
│   │   ├── filter.go
│   │   └── watch.go
│   └── runner/
│       ├── debouncer.go
│       ├── executor.go
│       ├── hooks.go
│       └── signal.go
├── gomon.config.json (generated)
└── LICENSE
```

## Dependencies

- `github.com/fsnotify/fsnotify` – File system notifications.
- `github.com/spf13/pflag` – POSIX/GNU-style flags.
- Standard library (`context`, `os/exec`, `sync`, `slog`, etc.).

## Development & Testing

- The project includes basic tests (e.g., config).
- Self-healing watcher ensures robustness.
- Extensive logging for observability.

## License

MIT License. See [LICENSE](https://github.com/jscriptural/gomon/blob/main/LICENSE) for details.

---

**gomon** prioritizes simplicity, reliability, and performance for modern development hot-reloading needs. Contributions, issues, and feedback are welcome!
