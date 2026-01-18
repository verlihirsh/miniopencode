package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseURLFromHostPort(t *testing.T) {
	cfg := Config{Host: "example.com", Port: 1234}
	c := New(cfg)
	assert.Equal(t, "http://example.com:1234", c.baseURL, "base URL should be constructed from host and port")
}

func TestBaseURLFromExplicitBaseURL(t *testing.T) {
	cfg := Config{BaseURL: "http://custom.server:9999"}
	c := New(cfg)
	assert.Equal(t, "http://custom.server:9999", c.baseURL, "should use explicit BaseURL when provided")
}

func TestListSessions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/session", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[{"id":"ses1","title":"t1"},{"id":"ses2","title":"t2"}]`)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	sessions, err := c.ListSessions(context.Background())
	require.NoError(t, err)
	require.Len(t, sessions, 2)
	assert.Equal(t, "ses1", sessions[0].ID)
	assert.Equal(t, "t1", sessions[0].Title)
	assert.Equal(t, "ses2", sessions[1].ID)
	assert.Equal(t, "t2", sessions[1].Title)
}

func TestListSessionsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := c.ListSessions(ctx)
	assert.Error(t, err, "should return error on context cancellation")
}

func TestCreateSession(t *testing.T) {
	var receivedTitle string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/session", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		receivedTitle = body["title"]

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"ses-new"}`)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	id, err := c.CreateSession(context.Background(), "hello")
	require.NoError(t, err)
	assert.Equal(t, "ses-new", id)
	assert.Equal(t, "hello", receivedTitle)
}

func TestPromptAsyncSendsModelAndAgent(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/session/ses123/prompt_async", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &captured)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	err := c.SendPromptAsync(context.Background(), "ses123", PromptInput{
		Parts: []InputPart{{Type: "text", Text: "hi"}},
		Model: &ModelRef{ProviderID: "anthropic", ModelID: "claude"},
		Agent: "build",
	})
	require.NoError(t, err)

	model := captured["model"].(map[string]any)
	assert.Equal(t, "anthropic", model["providerID"])
	assert.Equal(t, "claude", model["modelID"])
	assert.Equal(t, "build", captured["agent"].(string))
}

func TestPromptAsyncWithoutModel(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	err := c.SendPromptAsync(context.Background(), "ses123", PromptInput{
		Parts: []InputPart{{Type: "text", Text: "test"}},
	})
	require.NoError(t, err)

	_, hasModel := captured["model"]
	assert.False(t, hasModel, "should not include model when not provided")
}
