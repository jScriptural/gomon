package config

import (
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
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
		Watch: []string{"."},
		Env:   os.Environ(),
		Ignore: []string{
			"./.git",
			"./.gitignore",
			"./node_modules",
			"*.swp",
			"*.swx",
		},
	}

	f, err := os.Open(configPath)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			defer func() {
				if e := writeConfigFile(config); e != nil {
					slog.Info("Failed to update config", "error", e)
				}
			}()
		default:
			return nil,err;
		}
	} else {
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
		f.Close()
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
	ignore := fs.StringSliceP("ignore", "x", nil, "Comma-separated file extensions and directory to ignore(support simple glob patterns)")
	env := fs.StringSlice("env", nil, "Comma-separated, key=value pair")
	delay := fs.StringP("delay", "d", "500ms", "Debounce delay")

	fs.Parse(args)

	if fs.Changed("build") {
		config.Build = *build
	}

	if fs.Changed("run") {
		config.Run = *run
	}

	if fs.Changed("ignore") {
		config.Ignore = *ignore
	}
	if fs.Changed("ext") {
		config.Ext = *ext
	}
	if fs.Changed("env") {
		for _,v := range *env {
			config.Env = append(config.Env,v)
		}
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

	if fs.Changed("delay") {
		var d Duration
		n, err := strconv.Atoi(*delay)
		if err == nil {
			d = Duration(time.Duration(n) * time.Millisecond)
		} else {
			t, e := time.ParseDuration(*delay)
			if e != nil {
				return fmt.Errorf("Invalid delay duration flag: %w:%w", err, e)
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
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")
	if err := enc.Encode(c); err != nil {
		return err
	}

	return nil
}
