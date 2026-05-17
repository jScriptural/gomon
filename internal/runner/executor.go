package runner

import (
	"context"
	"github.com/jscriptural/gomon/internal/config"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"
	"strings"
)

type Executor struct {
	mu              sync.Mutex
	config          *config.Config
	cmd             *exec.Cmd
	ctx             context.Context
	cancelChild     context.CancelFunc
	canRunPostStart chan struct{}
	signalChan      chan os.Signal
}

func NewExecutor(ctx context.Context, config *config.Config) *Executor {
	e := &Executor{
		mu:              sync.Mutex{},
		config:          config,
		cmd:             nil,
		ctx:             ctx,
		cancelChild:     nil,
		canRunPostStart: make(chan struct{}),
		signalChan:      make(chan os.Signal, 5),
	}

	return e
}

func (e *Executor) start() error {
	e.mu.Lock()
	childctx, cancel := context.WithCancel(e.ctx)
	e.cancelChild = cancel
	cmd := exec.CommandContext(
		childctx,
		e.config.Run,
		e.config.Env...,
	)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(SIGTERM)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e.cmd = cmd
	e.mu.Unlock()
	if err := e.cmd.Start(); err != nil {
		return err
	}
	close(e.canRunPostStart)

	if err := e.cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (e *Executor) Trigger() error {
	slog.Info("TRIGGERED")
	//debounce
	time.Sleep(time.Duration(e.config.Delay))

	//terminateOldProcess:
	if e.cmd != nil {
		slog.Info("Terminating old process", "PID", e.cmd.Process.Pid)
		e.cancelChild()
		time.Sleep(10 * time.Microsecond)
		slog.Info("Old process terminated")
	}

	if c := e.config.Hooks.PreBuild; c != "" {
		slog.Info("Running prebuild hook", "prebuild", c)
		dur, err := e.hooksRunPreBuild()
		if err != nil {
			slog.Error("prebuild failed", "error", err)
			return err
		}
		slog.Info("prebuild successful", "duration", dur)
	}

	if c := e.config.Build; c != "" {
		slog.Info("Building executable", "build", c)
		dur, err := e.runBuild()
		if err != nil {
			slog.Error("Build failed", "error", err, "durarion", dur.String())
			return err
		}
		slog.Info("Build successful", "duration", dur)
	}

	if c := e.config.Hooks.PostBuild; c != "" {
		slog.Info("Running postbuild hook", "postbuild", c)
		dur, err := e.hooksRunPostBuild()
		if err != nil {
			slog.Error("postbuild failed", "error", err)
			return err
		}
		slog.Info("postbuild successful", "duration", dur)
	}

	if c := e.config.Hooks.PreStart; c != "" {
		slog.Info("Running prestart hook", "prestart", c)
		dur, err := e.hooksRunPreStart()
		if err != nil {
			slog.Error("prestart failed", "error", err)
			return err
		}
		slog.Info("prestart successful", "duration", dur)
	}

	if e.config.Hooks.PostStart != "" {
		go e.hooksRunPostStart()
	}
	err := e.start()
	return err
}

func (e *Executor) runBuild() (time.Duration, error) {
	start := time.Now()
	var dur time.Duration

	vec := strings.Fields(e.config.Build);

	n := len(vec)
	var cmd *exec.Cmd

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
