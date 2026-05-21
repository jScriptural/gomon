package config

import (
	"encoding/json"
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"time"
)

const (
	configPath = "gomon.config.json"
)

var (
	usage = func() {
	}
)

func LoadConfig() (*Config, error) {
	config := &Config{
		Delay: Duration(time.Duration(500 * time.Millisecond)),
		Watch: List{Dir: []string{"./"}},
		Env:   os.Environ(),
		Ignore: List{
			Dir: []string{
				"./node_modules",
				"./.git",
				"./github",
			},
			File: []string{"./.gitignore"},
			Glob: []string{
				"*.swp",
				"*.swx",
				"*.tmp*",
			},
		},
	}

	f, err := os.Open(configPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		defer func() {
			if e := writeConfigFile(config); e != nil {
				slog.Info("Failed to update config", "error", e)
			}
		}()

	} else {
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
		_ = f.Close()
	}

	err = parseFlags(os.Args[1:], config, usage)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func parseFlags(args []string, config *Config, usage func()) error {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.SetInterspersed(true)
	fs.Usage = usage

	build := fs.StringP("build", "b", "", "primary build command")
	run := fs.StringP("run", "r", "", "Primary binary execution command")
	preBuild := fs.String("prebuild", "", "Command to run before building binary")
	postStart := fs.String("poststart", "", "Command to run after starting application")
	preStart := fs.String("prestart", "", "Command to run before starting application")
	postBuild := fs.String("postbuild", "", "Command to run after building binary")
	ext := fs.StringSliceP("ext", "e", nil, "Comma-separated file extensions and directory to watch(support simple glob patterns)")
	ignoreDir := fs.StringSlice("ignoredir", nil, "Comma-separated directory to ignore")
	ignoreFile := fs.StringSlice("ignorefile", nil, "Comma-separated list of file to ignore")
	ignoreGlob := fs.StringSlice("ignoreglob", nil, "Comma-separated glob-pattern to ignore")
	watchDir := fs.StringSlice("watchdir", nil, "Comma-separated directory to watch. If `ext` is empty, watch all files not ignored else filter by `ext`")
	watchFile := fs.StringSlice("watchfile", nil, "Comma-separated file list to watch. If `ext` is not empty, file will be ignored if its extension is not in `ext`")
	watchGlob := fs.StringSlice("watchglob", nil, "Comma-separated glob-pattern to watch. If `ext` is not empty, file that match glob will be ignored if its extension is not in `ext`")
	env := fs.StringSlice("env", nil, "Comma-separated, key=value pair")
	delay := fs.StringP("delay", "d", "500ms", "Debounce delay")
	polling := fs.BoolP("polling", "p", false, "Prefer polling")

	_ = fs.Parse(args)

	if fs.Changed("build") {
		config.Build = *build
	}

	if fs.Changed("run") {
		config.Run = *run
	}

	if fs.Changed("watchdir") {
		config.Watch.Dir = *watchDir
	}

	if fs.Changed("watchfile") {
		config.Watch.Dir = *watchFile
	}

	if fs.Changed("watchglob") {
		config.Watch.Dir = *watchGlob
	}

	if fs.Changed("ignoredir") {
		config.Ignore.Dir = *ignoreDir
	}

	if fs.Changed("ignorefile") {
		config.Ignore.File = *ignoreFile
	}

	if fs.Changed("ignoreglob") {
		config.Ignore.Glob = *ignoreGlob
	}

	if fs.Changed("ext") {
		config.Ext = *ext
	}
	if fs.Changed("env") {
		config.Env = append(config.Env, *env...)
	}
	if fs.Changed("postbuild") {
		config.Hooks.PostBuild = *postBuild
	}
	if fs.Changed("poststart") {
		config.Hooks.PostStart = *postStart
	}
	if fs.Changed("prestart") {
		config.Hooks.PreStart = *preStart
	}
	if fs.Changed("prebuild") {
		config.Hooks.PreBuild = *preBuild
	}

	if fs.Changed("polling") {
		config.Polling = *polling
	}

	if fs.Changed("delay") {
		var d Duration
		n, err := strconv.Atoi(*delay)
		if err == nil {
			d = Duration(time.Duration(n) * time.Millisecond)
		} else {
			t, e := time.ParseDuration(*delay)
			if e != nil {
				return fmt.Errorf("invalid delay duration flag: %w:%w", err, e)
			}
			d = Duration(t)
		}
		config.Delay = d
	}

	return nil
}

func writeConfigFile(c *Config) error {
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")
	if err := enc.Encode(c); err != nil {
		return err
	}

	return nil
}
