package runner

import (
	"log/slog"
	"time"
)

func (e *Executor) debouncer() {
	for {
		<-e.isEvent
		dur := time.Until(time.Now().Add(time.Duration(e.config.Delay)))
		if e.timer != nil {
			ok := e.timer.Reset(dur)
			_ = ok
		} else {
			e.timer = time.AfterFunc(dur, func() {
				if err := e.trigger(); err != nil {
					slog.Info("Command fails to complete", "status", err)
				} else {
					slog.Info("Command completes", "status", "success")
				}
			})
		}
	}
}

func (e *Executor) Trigger() {
	e.isEvent <- true
}
