package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

type Pinger struct {
	client   *http.Client
	url      string
	message  string
	interval time.Duration
	logger   *slog.Logger
	count    atomic.Int64
}

func NewPinger(client *http.Client, url, message string, interval time.Duration, logger *slog.Logger) *Pinger {
	return &Pinger{
		client:   client,
		url:      url,
		message:  message,
		interval: interval,
		logger:   logger,
	}
}

func (p *Pinger) Count() int64 {
	return p.count.Load()
}

func (p *Pinger) Run(ctx context.Context) {
	p.logger.Info("pinger started", "url", p.url, "interval", p.interval.String())
	defer func() { p.logger.Info("pinger stopped", "count", p.Count()) }()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.send(ctx); err != nil {
				if ctx.Err() == nil {
					p.logger.Warn("request failed", "url", p.url, "err", err)
				}
				continue
			}
			p.logger.Info(p.message, "count", p.count.Add(1))
		}
	}
}

func (p *Pinger) send(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
