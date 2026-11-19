package worker

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPinger_CountsSuccessfulRequests(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "pong")
	}))
	defer peer.Close()

	var logs bytes.Buffer
	p := NewPinger(peer.Client(), peer.URL+"/pong", "Ping", 10*time.Millisecond, slog.New(slog.NewJSONHandler(&logs, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.Run(ctx)
		close(done)
	}()

	require.Eventually(t, func() bool { return p.Count() >= 3 }, 2*time.Second, 5*time.Millisecond)
	cancel()
	<-done

	assert.Contains(t, logs.String(), `"msg":"Ping"`)
}

func TestPinger_DoesNotCountFailures(t *testing.T) {
	var calls atomic.Int64
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer peer.Close()

	p := NewPinger(peer.Client(), peer.URL+"/pong", "Ping", 5*time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	p.Run(ctx)

	assert.Positive(t, calls.Load())
	assert.Zero(t, p.Count())
}
