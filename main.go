package main

import (
	"context"
	"github.com/jhonoid/gomon/internal/config"
	"github.com/jhonoid/gomon/internal/logger"
	"github.com/jhonoid/gomon/internal/monitor"
	"log/slog"
)

func main() {
	logger.InitDefault()
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Fail to load configurations", "error", err)
	}

	slog.Warn("NO SUPPORT FOR POLLING YET.")
	if config.Polling {
		slog.Info("No support for polling yet")
		return
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	monitor := monitor.NewMonitor(ctx, config)

	if err := monitor.Watch(); err != nil {
		slog.Error("exit", "error", err)
		cancelFunc()
	}

	<-make(chan struct{})
}
