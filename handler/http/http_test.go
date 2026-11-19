package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeCounter int64

func (c fakeCounter) Count() int64 { return int64(c) }

func TestRoutes(t *testing.T) {
	routes := New("pong", fakeCounter(42)).Routes()

	tests := []struct {
		method, path string
		status       int
		body         string
	}{
		{http.MethodGet, "/pong", http.StatusOK, "pong"},
		{http.MethodGet, "/count", http.StatusOK, `{"pongs":42}` + "\n"},
		{http.MethodGet, "/ping", http.StatusNotFound, ""},
		{http.MethodPost, "/pong", http.StatusMethodNotAllowed, ""},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			routes.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			assert.Equal(t, tt.status, rec.Code)
			if tt.body != "" {
				assert.Equal(t, tt.body, rec.Body.String())
			}
		})
	}
}
