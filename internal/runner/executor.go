package runner

import (
	"context"
	"errors"
	"github.com/jscriptural/gomon/internal/config"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Executor struct {
	mu              sync.Mutex
	config          *config.Config
	cmd             *exec.Cmd
	ctx             context.Context
	cancelChild     context.CancelFunc
	canRunPostStart chan struct{}
	signalChan      chan os.Signal
	timer           *time.Timer
	isEvent         chan bool
	isCmdActive     bool
}

const (
	PREBUILD = iota
	POSTBUILD
	PRESTART
	POSTSTART
)

func NewExecutor(ctx context.Context, config *config.Config) *Executor {
	e := &Executor{
		config:          config,
		cmd:             nil,
		ctx:             ctx,
		cancelChild:     nil,
		canRunPostStart: nil,
		signalChan:      make(chan os.Signal, 5),
		isEvent:         make(chan bool),
		timer:           nil,
		isCmdActive:     false,
	}

	go e.debouncer()
	return e
}

func (e *Executor) trigger() error {
	slog.Info("Starting new Iteration")

	e.mu.Lock()
	if e.isCmdActive {
		slog.Info("Terminating old process", "PID", e.cmd.Process.Pid)
		e.cancelChild()
		slog.Info("waiting for child process to clean up")
		//time.Sleep(100 * time.Millisecond)
		for e.isCmdActive {
			err := e.cmd.Process.Signal(SIGKILL)
			if errors.Is(err, os.ErrProcessDone) {
				break
			}
		}
		slog.Info("Old process terminated")
	}

	if c := e.config.Hooks.PreBuild; c != "" {
		slog.Info("Running prebuild hook", "prebuild", c)
		dur, err := e.runHooks(PREBUILD)
		if err != nil {
			slog.Error("prebuild failed", "error", err)
			e.mu.Unlock()
			return err
		}
		slog.Info("prebuild successful", "duration", dur)
	}

	if c := e.config.Build; c != "" {
		slog.Info("Building executable", "build", c)
		dur, err := e.runBuild()
		if err != nil {
			slog.Error("Build failed", "error", err, "durarion", dur.String())
			e.mu.Unlock()
			return err
		}
		slog.Info("Build successful", "duration", dur)
	}

	if c := e.config.Hooks.PostBuild; c != "" {
		slog.Info("Running postbuild hook", "postbuild", c)
		dur, err := e.runHooks(POSTBUILD)
		if err != nil {
			slog.Error("postbuild failed", "error", err)
			e.mu.Unlock()
			return err
		}
		slog.Info("postbuild successful", "duration", dur)
	}

	if c := e.config.Hooks.PreStart; c != "" {
		slog.Info("Running prestart hook", "prestart", c)
		dur, err := e.runHooks(PRESTART)
		if err != nil {
			slog.Error("prestart failed", "error", err)
			e.mu.Unlock()
			return err
		}
		slog.Info("prestart successful", "duration", dur)
	}

	if e.config.Hooks.PostStart != "" {
		e.canRunPostStart = make(chan struct{})
		go e.runHooksPostStart()
	}

	err := e.start()

	return err
}

func (e *Executor) runBuild() (time.Duration, error) {
	start := time.Now()
	var dur time.Duration

	vec := strings.Fields(e.config.Build)

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

func (e *Executor) start() error {
	childctx, cancel := context.WithCancel(e.ctx)
	e.cancelChild = cancel

	args := strings.Fields(e.config.Run)
	var cmd *exec.Cmd
	n := len(args)
	switch {
	case n == 1:
		cmd = exec.CommandContext(
			childctx,
			args[0],
		)
	case n > 1:
		cmd = exec.CommandContext(
			childctx,
			args[0],
			args[1:]...,
		)
	default:
		return errors.New("no command to run")
	}

	cmd.Cancel = func() error {
		return cmd.Process.Signal(SIGTERM)
	}

	cmd.Env = e.config.Env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e.cmd = cmd

	if err := e.cmd.Start(); err != nil {
		return err
	}

	e.isCmdActive = true
	slog.Info("Spawned Child", "pid", e.cmd.Process.Pid)

	if e.config.Hooks.PostStart != "" {
		close(e.canRunPostStart)
	}

	e.mu.Unlock()

	if err := e.cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Info("Process exited", "Pid", exitErr.Pid(), "ExitCode", exitErr.ExitCode())
			//Well, it works on my machine
			if w, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				switch {
				case w.Signaled():
					slog.Info("Process terminated by signal", "signal", w.Signal())
				case w.Stopped():
					slog.Info("Process stopped by signal", "signal", w.StopSignal())
				}
				e.isCmdActive = false
				return nil
			}
		}
		e.isCmdActive = false
		slog.Debug("Wait unblocks", "error", err)
		return err
	}

	e.isCmdActive = false
	return nil
}
