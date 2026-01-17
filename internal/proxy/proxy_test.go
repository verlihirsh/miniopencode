package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProxyBuildsBaseURL(t *testing.T) {
	cfg := Config{Host: "example.com", Port: "1234"}
	p := NewProxy(cfg)
	if p.BaseURL() != "http://example.com:1234" {
		t.Fatalf("baseURL mismatch: got %q", p.BaseURL())
	}
}

func TestCheckHealth(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/global/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unhealthy.Close()

	t.Run("healthy", func(t *testing.T) {
		cfg := Config{BaseURLOverride: healthy.URL}
		p := NewProxy(cfg)
		if ok := p.CheckHealth(); !ok {
			t.Fatalf("expected healthy server")
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		cfg := Config{BaseURLOverride: unhealthy.URL}
		p := NewProxy(cfg)
		if ok := p.CheckHealth(); ok {
			t.Fatalf("expected unhealthy server")
		}
	})
}
