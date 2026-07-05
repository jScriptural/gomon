package runner

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (e *Executor) runHooks(which int) (time.Duration, error) {
	var vec []string

	start := time.Now()
	switch which {
	case PREBUILD:
		vec = e.config.Hooks.PreBuild
	case POSTBUILD:
		vec = e.config.Hooks.PostBuild
	case POSTSTART:
		vec = e.config.Hooks.PostStart
	case PRESTART:
		vec = e.config.Hooks.PreStart
	default:
		_ = vec
		return time.Since(start), nil
	}

	for _, c := range vec {
		cmdToken := strings.Fields(c)
		if err := e.executeCmd(cmdToken); err != nil {
			return time.Since(start), err
		}
	}

	return time.Since(start), nil
}

func (e *Executor) runHooksPostStart() {
	<-e.canRunPostStart

	vec := e.config.Hooks.PostStart

	for _, c := range vec {
		cmdToken := strings.Fields(c)
		if err := e.executeCmd(cmdToken); err != nil {
			slog.Error("poststart failed", "error", err)
			return
		}
	}
}

func (e *Executor) executeCmd(cmdToken []string) error {
	var cmd *exec.Cmd

	tokenCount := len(cmdToken)
	switch {
	case tokenCount == 1:
		cmd = exec.CommandContext(
			e.ctx,
			cmdToken[0],
		)
	case tokenCount > 1:
		cmd = exec.CommandContext(
			e.ctx,
			cmdToken[0],
			cmdToken[1:]...,
		)
	default:
		return nil
	}

	cmd.Env = e.config.Env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Cancel = func() error {
		return cmd.Process.Signal(SIGTERM)
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
