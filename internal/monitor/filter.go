package monitor 

import (
	"os"
	"path/filepath"
	"fmt"
	"slices"
	"log/slog"
)


// filter parses the Monitor's configuration to extract absolute directories
// to watch, along with file and pattern targets to completely ignore.
func (m *Monitor) filter() (watchDirs []string, ignoreTargets []string, err error) {
	for _, path := range m.config.Watch {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get absolute path for watch target %q: %w", path, err)
		}

		info, err := os.Lstat(absPath)
		if err != nil {
			return nil, nil, fmt.Errorf("watch path does not exist: %s", absPath)
		}
		if !info.IsDir() {
			dir := filepath.Dir(absPath)
			if !slices.Contains(watchDirs,dir) && !slices.Contains(m.config.Ignore,dir){
				watchDirs = append(watchDirs,dir)
			}
		}else {
			if !slices.Contains(m.config.Ignore,absPath) {
				watchDirs = append(watchDirs, absPath)
			}
		}
	}

	for _, pattern := range m.config.Ignore {
		_, err := filepath.Match(pattern, "syntax-test")
		if err != nil {
			continue
		}

		if absPattern, err := filepath.Abs(pattern); err == nil {
			if _, statErr := os.Stat(absPattern); statErr == nil {
				ignoreTargets = append(ignoreTargets, absPattern)
				continue
			}
		}

		ignoreTargets = append(ignoreTargets, pattern)
	}

	slog.Debug("Filter results","watchDirs", watchDirs,"ignoreTargets",ignoreTargets)
	return watchDirs, ignoreTargets, nil
}
