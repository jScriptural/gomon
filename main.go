package main

import (
	"context"
	"github.com/jscriptural/gomon/internal/config"
	"github.com/jscriptural/gomon/internal/logger"
	"github.com/jscriptural/gomon/internal/monitor"
	"log/slog"
)

func main() {
	logger.InitDefault()
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Fail to load configurations", "error", err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	monitor := monitor.NewMonitor(ctx, config)

	if err := monitor.Watch(); err != nil {
		slog.Error("exit", "error", err)
		cancelFunc()
	}

	slog.Debug("Main blocks: reading unbuffered chan")

	<-make(chan struct{})
}
