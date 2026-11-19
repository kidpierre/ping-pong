package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"ping-pong/config"
	handlerhttp "ping-pong/handler/http"
	"ping-pong/worker"
)

func Run(cfg *config.Config, name, peer string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})).With("service", name)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := &http.Client{Timeout: cfg.RequestTimeout}
	pinger := worker.NewPinger(client, cfg.PeerURL+"/"+peer, strings.ToUpper(name[:1])+name[1:], cfg.Interval, logger)
	handler := handlerhttp.New(name, pinger)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: cfg.RequestTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		pinger.Run(ctx)
	})

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	var runErr error
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		runErr = fmt.Errorf("http server: %w", err)
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("shutdown http server: %w", err))
	}
	wg.Wait()

	logger.Info("service stopped")
	return runErr
}
