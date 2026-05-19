package runner

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	SIGTERM   = syscall.SIGTERM
	SIGABRT   = syscall.SIGABRT
	SIGALRM   = syscall.SIGALRM
	SIGBUS    = syscall.SIGBUS
	SIGCHLD   = syscall.SIGCHLD
	SIGCLD    = syscall.SIGCLD
	SIGCONT   = syscall.SIGCONT
	SIGFPE    = syscall.SIGFPE
	SIGHUP    = syscall.SIGHUP
	SIGINT    = syscall.SIGINT
	SIGIO     = syscall.SIGIO
	SIGIOT    = syscall.SIGIOT
	SIGKILL   = syscall.SIGKILL
	SIGPIPE   = syscall.SIGPIPE
	SIGPOLL   = syscall.SIGPOLL
	SIGPROF   = syscall.SIGPROF
	SIGPWR    = syscall.SIGPWR
	SIGQUIT   = syscall.SIGQUIT
	SIGSEGV   = syscall.SIGSEGV
	SIGSTKFLT = syscall.SIGSTKFLT
	SIGSTOP   = syscall.SIGSTOP
	SIGSYS    = syscall.SIGSYS
	SIGTRAP   = syscall.SIGTRAP
	SIGTSTP   = syscall.SIGTSTP
	SIGTTIN   = syscall.SIGTTIN
	SIGTTOU   = syscall.SIGTTOU
	SIGUNUSED = syscall.SIGUNUSED
	SIGURG    = syscall.SIGURG
	SIGUSR1   = syscall.SIGUSR1
	SIGUSR2   = syscall.SIGUSR2
	SIGVTALRM = syscall.SIGVTALRM
	SIGWINCH  = syscall.SIGWINCH
	SIGXCPU   = syscall.SIGXCPU
	SIGXFSZ   = syscall.SIGXFSZ
)

func (e *Executor) Signal() {
	sig := []os.Signal{
		SIGTERM, SIGINT, SIGABRT, SIGALRM,
		SIGBUS, SIGCHLD, SIGCLD, SIGCONT, SIGFPE,
		SIGHUP, SIGIO, SIGIOT, SIGKILL, SIGPOLL,
		SIGPIPE, SIGPROF, SIGPWR, SIGQUIT, SIGSEGV,
		SIGSTKFLT, SIGSTOP, SIGSYS, SIGTRAP,
		SIGTSTP, SIGTTIN, SIGTTOU, SIGUNUSED,
		SIGURG, SIGUSR1, SIGUSR2, SIGVTALRM,
		SIGWINCH, SIGXCPU, SIGXFSZ,
	}

	signal.Notify(e.signalChan, sig...)

	go e.forwardSignal()
}

func (e *Executor) forwardSignal() {
	for {
		for sig := range e.signalChan {
			if e.isCmdActive {
				err := e.cmd.Process.Signal(sig)
				if err != nil {
					if errors.Is(err, os.ErrProcessDone) {
						continue
					}
					slog.Error("Fail to forward signal", "signal", sig, "error", err)
				}
			}
			if sig == SIGQUIT {
				slog.Info("gomon quits")
				os.Exit(0)
				return
			}
		}
	}
}
