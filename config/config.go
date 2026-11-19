package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPPort        string
	PeerURL         string
	Interval        time.Duration
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
	LogLevel        slog.Level
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		PeerURL:  strings.TrimRight(os.Getenv("PEER_URL"), "/"),
	}

	var err error
	if cfg.Interval, err = getDuration("INTERVAL", 2*time.Second); err != nil {
		return nil, err
	}
	if cfg.RequestTimeout, err = getDuration("REQUEST_TIMEOUT", 2*time.Second); err != nil {
		return nil, err
	}
	if cfg.ShutdownTimeout, err = getDuration("SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return nil, err
	}
	if err = cfg.LogLevel.UnmarshalText([]byte(getEnv("LOG_LEVEL", "info"))); err != nil {
		return nil, fmt.Errorf("LOG_LEVEL: %w", err)
	}

	if cfg.PeerURL == "" {
		return nil, errors.New("PEER_URL is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return d, nil
}
