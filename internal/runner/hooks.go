package runner

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (e *Executor) runHooks(which int) (time.Duration, error) {
	start := time.Now()
	var dur time.Duration

	var vec []string
	switch which {
	case PREBUILD:
		vec = strings.Fields(e.config.Hooks.PreBuild)
	case POSTBUILD:
		vec = strings.Fields(e.config.Hooks.PostBuild)
	case POSTSTART:
		vec = strings.Fields(e.config.Hooks.PostStart)
	case PRESTART:
		vec = strings.Fields(e.config.Hooks.PreStart)
	default:
		return time.Since(start), nil
	}

	var cmd *exec.Cmd
	n := len(vec)
	switch {
	case n == 1:
		cmd = exec.CommandContext(
			e.ctx,
			vec[0],
		)
	case n > 1:
		cmd = exec.CommandContext(
			e.ctx,
			vec[0],
			vec[1:]...,
		)
	}

	cmd.Env = e.config.Env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Cancel = func() error {
		return cmd.Process.Signal(SIGTERM)
	}

	if err := cmd.Run(); err != nil {
		dur = time.Since(start)
		return dur, err
	}

	dur = time.Since(start)
	return dur, nil

}

func (e *Executor) runHooksPostStart() {
	<-e.canRunPostStart

	vec := strings.Fields(e.config.Hooks.PostStart)
	var cmd *exec.Cmd
	n := len(vec)
	switch {
	case n == 1:
		cmd = exec.CommandContext(
			e.ctx,
			vec[0],
		)
	case n > 1:
		cmd = exec.CommandContext(
			e.ctx,
			vec[0],
			vec[1:]...,
		)
	default:
		return
	}

	cmd.Env = e.config.Env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Cancel = func() error {
		return cmd.Process.Signal(SIGTERM)
	}
	if err := cmd.Run(); err != nil {
		slog.Error("poststart failed", "error", err)
		return
	}
}
