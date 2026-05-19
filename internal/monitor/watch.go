package monitor


import (
	"github.com/fsnotify/fsnotify"
	"github.com/jscriptural/gomon/internal/config"
	"github.com/jscriptural/gomon/internal/runner"
	"context"
	"os"
	"log/slog"
	"errors"
	"fmt"
	"strings"
	"slices"
	"io/fs"
	"path/filepath"
)

type Monitor struct {
	executor *runner.Executor
	config *config.Config
	watcher *fsnotify.Watcher
	cancelExec context.CancelFunc
	ctx context.Context
}

func NewMonitor(ctx context.Context, config *config.Config) *Monitor {
	ctx, cancel := context.WithCancel(ctx)

	executor := runner.NewExecutor(ctx,config)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Creating a new watcher","error",err)
		os.Exit(1)
	}

	executor.Signal()
	return &Monitor {
		executor: executor,
		config: config,
		watcher: w,
		cancelExec: cancel,
		ctx: ctx,
	}
}

func (m *Monitor) Watch() error {
	wl, excl, err := m.filter()
	if err != nil {
		slog.Error("Watch fails", "error", err)
		return err
	}


	go func() {
		for {
			m.watchLoop(excl);

			if m.ctx.Err() != nil {
				os.Exit(2)
			}

			slog.Warn("Attempting to re-initialize file watcher tree...")

			// Close old instance safely
			m.watcher.Close()

			// Rebuild a fresh watcher
			w, err := fsnotify.NewWatcher()
			if err != nil {
				slog.Error("Failed to recreate watcher engine", "error", err)
				os.Exit(1)
			}
			m.watcher = w

			// Re-register directories
			if err := m.watchDirs(wl,excl); err != nil {
				slog.Error("Failed to re-crawl directories on recovery", "error", err)
				continue
			}

			slog.Info("File watcher self-healed successfully.")
		}
	}()



	if err := m.watchDirs(wl,excl); err != nil {
		return err
	}

	go m.executor.Trigger();
	return nil
}



func (m *Monitor) watchDirs(watchList, cutList []string) error {
	if len(watchList) == 0 {
		return errors.New("empty watch list")
	}

	for _, rootDir := range watchList {
		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrPermission) {
					return fs.SkipDir
				}
				return err
			}

			if d.IsDir() {
				if slices.Contains(cutList,path) {
					return fs.SkipDir
				}
				slog.Debug("Registering directory target to watcher", "path", path)
				if err := m.watcher.Add(path); err != nil {
					return fmt.Errorf("failed adding path %s to watcher: %w", path, err)
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





func (m *Monitor) watchLoop(ignore []string) {
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
			absEventPath, err := filepath.Abs(evt.Name)
			if err != nil {
				continue
			}

			ignored := false
			for _, p := range ignore {
				if absEventPath == p {
					ignored = true
					break
				}

				if match, _ := filepath.Match(p, filepath.Base(absEventPath)); match {
					ignored = true
					break
				}
				if strings.Contains(absEventPath, p) {
					ignored = true
					break
				}
			}

			if !ignored && len(m.config.Ext) > 0 {
				hasTargetExt := false
				for _, ext := range m.config.Ext {
					if strings.HasSuffix(absEventPath, ext) {
						hasTargetExt = true
						break
					}
				}
				if !hasTargetExt {
					ignored = true 
				}
			}

			if !ignored && (evt.Has(fsnotify.Write) || evt.Has(fsnotify.Create) || evt.Has(fsnotify.Remove)) {
				slog.Info("Event captured", "file", evt.Name, "op", evt.Op.String())
				go m.executor.Trigger()
			}
		}
	}
}


