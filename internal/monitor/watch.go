package monitor

import (
	"context"
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/jhonoid/gomon/internal/config"
	"github.com/jhonoid/gomon/internal/runner"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

type Monitor struct {
	executor   *runner.Executor
	config     *config.Config
	watcher    *fsnotify.Watcher
	cancelExec context.CancelFunc
	ctx        context.Context
}

func NewMonitor(ctx context.Context, config *config.Config) *Monitor {
	ctx, cancel := context.WithCancel(ctx)

	executor := runner.NewExecutor(ctx, config)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Creating a new watcher", "error", err)
		os.Exit(1)
	}

	executor.Signal()
	return &Monitor{
		executor:   executor,
		config:     config,
		watcher:    w,
		cancelExec: cancel,
		ctx:        ctx,
	}
}

func (m *Monitor) Watch() error {
	go func() {
		for {
			m.watchLoop()

			if m.ctx.Err() != nil {
				os.Exit(2)
			}

			slog.Warn("Attempting to re-initialize file watcher tree...")

			// Close old instance safely
			_ = m.watcher.Close()

			// Rebuild a fresh watcher
			w, err := fsnotify.NewWatcher()
			if err != nil {
				slog.Error("Failed to recreate watcher engine", "error", err)
				m.cancelExec()
				os.Exit(1)
			}
			m.watcher = w

			// Re-register directories
			if err := m.watchDirs(); err != nil {
				slog.Error("Failed to re-crawl directories on recovery", "error", err)
				os.Exit(5)
			}

			slog.Info("File watcher self-healed successfully.")
		}
	}()

	if err := m.watchDirs(); err != nil {
		return err
	}

	go m.executor.Trigger()
	return nil
}

func (m *Monitor) watchDirs() error {
	defer func() {
		watcherList := m.watcher.WatchList()
		slog.Info("", "Watched Directories", watcherList, "Number of directories", len(watcherList))
	}()

	watchList := m.config.Watch
	ignoreList := m.config.Ignore

	for _, rootDir := range watchList.Dir {
		absRootDir, err := filepath.Abs(rootDir)
		if err != nil {
			slog.Warn("Fail to get absolute path", "error", err)
			continue
		}
		err = filepath.WalkDir(absRootDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrPermission) {
					slog.Warn("Path error", "error", err)
					return fs.SkipDir
				}
				return err
			}
			if d.IsDir() {
				for _, dir := range ignoreList.Dir {
					absDir, err := filepath.Abs(dir)
					if err != nil {
						slog.Warn("Fail to get Absolute path", "error", err)
						continue
					}
					if m, _ := filepath.Match(absDir, path); m {
						return fs.SkipDir
					}
				}
				for _, p := range ignoreList.Glob {
					ok, err := filepath.Match(p, path)
					if err != nil {
						slog.Warn("Bad glob pattern", "glob", p, "error", err)
					}
					if ok {
						return fs.SkipDir
					}
				}

				if err := m.watcher.Add(path); err != nil {
					slog.Warn("Fail to add path to watcher", "path", path, "error", err)
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Monitor) watchLoop() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case err, ok := <-m.watcher.Errors:
			if !ok {
				slog.Error("Watcher stopped")
				return
			}
			slog.Error("Watcher tracking error", "error", err)

		case evt, ok := <-m.watcher.Events:
			if !ok {
				slog.Error("Event channel closed")
				return
			}

			if !m.isIgnoredEvent(evt) {
				slog.Info("Captured Event", "file", evt.Name, "Op", evt.Op.String())
				go m.executor.Trigger()
			}
		}
	}
}

func (m *Monitor) isIgnoredEvent(evt fsnotify.Event) bool {
	ignoreList := m.config.Ignore
	absName, err := filepath.Abs(evt.Name)
	if err != nil {
		absName = evt.Name
	}

	for _, dir := range ignoreList.Dir {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}
		if ok, _ := filepath.Match(absDir, absName); ok {
			return true
		}
	}

	for _, f := range ignoreList.File {
		absF, err := filepath.Abs(f)
		if err != nil {
			absF = f
		}
		if ok, _ := filepath.Match(absF, absName); ok {
			return true
		}
	}

	for _, p := range ignoreList.Glob {
		ok, err := filepath.Match(p, absName)
		if err != nil {
			slog.Warn("Bad glob pattern", "glob", p, "error", err)
			continue
		}
		if ok {
			return true
		}
	}

	for _, ext := range m.config.Ext {
		e := filepath.Ext(absName)
		if e == "" {
			continue
		}
		if ext != e {
			return true
		}
	}

	return !evt.Has(fsnotify.Write) && !evt.Has(fsnotify.Create) && !evt.Has(fsnotify.Remove)
}
