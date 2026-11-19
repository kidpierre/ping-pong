package main

import (
	"log/slog"
	"os"

	"ping-pong/app"
	"ping-pong/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	if err := app.Run(cfg, "ping", "pong"); err != nil {
		slog.Error("service stopped with error", "err", err)
		os.Exit(1)
	}
}
