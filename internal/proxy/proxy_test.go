package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProxyBuildsBaseURL(t *testing.T) {
	cfg := Config{Host: "example.com", Port: "1234"}
	p := NewProxy(cfg)
	assert.Equal(t, "http://example.com:1234", p.BaseURL(), "baseURL should be constructed from host and port")
}

func TestNewProxyDefaultsHostPort(t *testing.T) {
	cfg := Config{}
	p := NewProxy(cfg)
	assert.Equal(t, "http://127.0.0.1:4096", p.BaseURL(), "should use default host and port")
}

func TestNewProxyUsesBaseURLOverride(t *testing.T) {
	cfg := Config{
		Host:            "ignored.com",
		Port:            "9999",
		BaseURLOverride: "http://override.com:8888",
	}
	p := NewProxy(cfg)
	assert.Equal(t, "http://override.com:8888", p.BaseURL(), "should use BaseURLOverride when provided")
}

func TestCheckHealth(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/global/health", r.URL.Path, "should request health endpoint")
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
		assert.True(t, p.CheckHealth(), "should detect healthy server")
	})

	t.Run("unhealthy", func(t *testing.T) {
		cfg := Config{BaseURLOverride: unhealthy.URL}
		p := NewProxy(cfg)
		assert.False(t, p.CheckHealth(), "should detect unhealthy server")
	})

	t.Run("unreachable", func(t *testing.T) {
		cfg := Config{BaseURLOverride: "http://localhost:1"}
		p := NewProxy(cfg)
		assert.False(t, p.CheckHealth(), "should handle unreachable server")
	})
}

func TestCreateSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/session", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id": "ses-" + body["title"],
		})
	}))
	defer server.Close()

	cfg := Config{BaseURLOverride: server.URL}
	p := NewProxy(cfg)

	t.Run("successful", func(t *testing.T) {
		id, err := p.createSession("test-session")
		require.NoError(t, err)
		assert.Equal(t, "ses-test-session", id)
	})
}

func TestListSessions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/session", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "ses-1", "title": "Session 1"},
			{"id": "ses-2", "title": "Session 2"},
		})
	}))
	defer server.Close()

	cfg := Config{BaseURLOverride: server.URL}
	p := NewProxy(cfg)

	sessions, err := p.listSessions()
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "ses-1", sessions[0]["id"])
	assert.Equal(t, "Session 2", sessions[1]["title"])
}

func TestSendPrompt(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/session/")
		assert.Contains(t, r.URL.Path, "/prompt_async")
		assert.Equal(t, http.MethodPost, r.Method)

		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	cfg := Config{BaseURLOverride: server.URL}
	p := NewProxy(cfg)

	t.Run("with_model", func(t *testing.T) {
		receivedBody = nil
		payload := PromptPayload{
			Text:       "hello world",
			ProviderID: "anthropic",
			ModelID:    "claude",
		}

		err := p.sendPrompt("ses-123", payload)
		require.NoError(t, err)

		require.NotNil(t, receivedBody)
		parts := receivedBody["parts"].([]interface{})
		firstPart := parts[0].(map[string]interface{})
		assert.Equal(t, "hello world", firstPart["text"])

		model := receivedBody["model"].(map[string]interface{})
		assert.Equal(t, "anthropic", model["providerID"])
		assert.Equal(t, "claude", model["modelID"])
	})

	t.Run("without_model", func(t *testing.T) {
		receivedBody = nil
		payload := PromptPayload{
			Text: "test message",
		}

		err := p.sendPrompt("ses-456", payload)
		require.NoError(t, err)

		require.NotNil(t, receivedBody)
		_, hasModel := receivedBody["model"]
		assert.False(t, hasModel, "should not include model when not provided")
	})
}
