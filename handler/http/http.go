package http

import (
	"encoding/json"
	"io"
	"net/http"
)

type Counter interface {
	Count() int64
}

type Handler struct {
	name    string
	counter Counter
}

func New(name string, counter Counter) *Handler {
	return &Handler{name: name, counter: counter}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /"+h.name, h.reply)
	mux.HandleFunc("GET /count", h.count)
	mux.HandleFunc("GET /healthz", h.health)

	return mux
}

func (h *Handler) reply(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, h.name)
}

func (h *Handler) count(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]int64{h.name + "s": h.counter.Count()})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
