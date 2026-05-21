package runner

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
)

func (e *Executor) Signal() {
	sig := []os.Signal{}

	for _, s := range platformSignals {
		sig = append(sig, s)
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
