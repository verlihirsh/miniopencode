package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHeadlessProxyE2E tests the complete headless proxy workflow as a black-box test.
// It simulates a complete session: health check → create session → send prompt → SSE streaming.
func TestHeadlessProxyE2E(t *testing.T) {
	// Setup mock opencode server
	var (
		mu             sync.Mutex
		createdSession string
		receivedPrompt map[string]any
		sseConnected   bool
		healthChecked  bool
		sessionsListed bool
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.URL.Path == "/global/health":
			healthChecked = true
			w.WriteHeader(http.StatusOK)

		case r.URL.Path == "/session" && r.Method == http.MethodPost:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			createdSession = "ses-" + body["title"]
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"id": createdSession})

		case r.URL.Path == "/session" && r.Method == http.MethodGet:
			sessionsListed = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]string{
				{"id": "ses-1", "title": "Session 1"},
				{"id": "ses-2", "title": "Session 2"},
			})

		case strings.HasPrefix(r.URL.Path, "/session/") && strings.HasSuffix(r.URL.Path, "/prompt_async"):
			json.NewDecoder(r.Body).Decode(&receivedPrompt)
			w.WriteHeader(http.StatusAccepted)

		case r.URL.Path == "/event":
			sseConnected = true
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
			// Send a sample SSE event
			fmt.Fprint(w, "data: {\"type\":\"test\",\"message\":\"hello\"}\n\n")
			w.(http.Flusher).Flush()

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create proxy with mock server
	cfg := Config{BaseURLOverride: server.URL}
	proxy := NewProxy(cfg)

	// Test individual command handling
	t.Run("HealthCommand", func(t *testing.T) {
		proxy.handleCommand(Command{Type: "health"})
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return healthChecked
		}, time.Second, 10*time.Millisecond, "health check should be called")
	})

	t.Run("CreateSessionCommand", func(t *testing.T) {
		cmd := Command{
			Type:    "session.create",
			Payload: json.RawMessage(`{"title":"my-session"}`),
		}
		proxy.handleCommand(cmd)
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return createdSession == "ses-my-session"
		}, time.Second, 10*time.Millisecond, "session should be created")
	})

	t.Run("ListSessionsCommand", func(t *testing.T) {
		proxy.handleCommand(Command{Type: "session.list"})
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return sessionsListed
		}, time.Second, 10*time.Millisecond, "sessions should be listed")
	})

	t.Run("SelectSessionCommand", func(t *testing.T) {
		cmd := Command{
			Type:    "session.select",
			Payload: json.RawMessage(`{"id":"ses-selected"}`),
		}
		proxy.handleCommand(cmd)
		assert.Equal(t, "ses-selected", proxy.config.SessionID, "session should be selected")
	})

	t.Run("PromptCommand", func(t *testing.T) {
		proxy.config.SessionID = "ses-test"
		cmd := Command{
			Type:    "prompt",
			Payload: json.RawMessage(`{"text":"test prompt","provider_id":"anthropic","model_id":"claude"}`),
		}
		proxy.handleCommand(cmd)
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return receivedPrompt != nil
		}, time.Second, 10*time.Millisecond, "prompt should be sent")

		mu.Lock()
		parts := receivedPrompt["parts"].([]any)
		firstPart := parts[0].(map[string]any)
		mu.Unlock()
		assert.Equal(t, "test prompt", firstPart["text"], "prompt text should match")
	})

	t.Run("SSEStartCommand", func(t *testing.T) {
		proxy.handleCommand(Command{Type: "sse.start"})
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return sseConnected
		}, time.Second, 10*time.Millisecond, "SSE should connect")
	})
}

// TestHeadlessProxyErrorHandling tests error scenarios in black-box manner.
func TestHeadlessProxyErrorHandling(t *testing.T) {
	// Mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/global/health" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/session") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"server error"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{BaseURLOverride: server.URL}
	proxy := NewProxy(cfg)

	t.Run("UnhealthyServer", func(t *testing.T) {
		healthy := proxy.CheckHealth()
		assert.False(t, healthy, "should detect unhealthy server")
	})

	t.Run("PromptWithoutSession", func(t *testing.T) {
		// Ensure no session is set
		proxy.config.SessionID = ""
		cmd := Command{
			Type:    "prompt",
			Payload: json.RawMessage(`{"text":"test"}`),
		}
		// handleCommand would output error, we just verify it doesn't panic
		require.NotPanics(t, func() {
			proxy.handleCommand(cmd)
		})
	})

	t.Run("InvalidCommand", func(t *testing.T) {
		cmd := Command{Type: "invalid.command"}
		require.NotPanics(t, func() {
			proxy.handleCommand(cmd)
		})
	})
}

// TestHeadlessProxySSELifecycle tests SSE connection lifecycle.
func TestHeadlessProxySSELifecycle(t *testing.T) {
	eventsSent := make(chan string, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/event" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send multiple events
		events := []string{
			`data: {"type":"message.part.updated","delta":"Hello"}` + "\n\n",
			`data: {"type":"message.part.updated","delta":" World"}` + "\n\n",
			`data: {"type":"message.updated","status":"completed"}` + "\n\n",
		}

		for _, event := range events {
			fmt.Fprint(w, event)
			w.(http.Flusher).Flush()
			eventsSent <- event
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	cfg := Config{BaseURLOverride: server.URL}
	proxy := NewProxy(cfg)

	t.Run("StartAndStopSSE", func(t *testing.T) {
		// Start SSE
		err := proxy.startSSE()
		require.NoError(t, err, "SSE should start without error")

		// Wait for some events
		time.Sleep(100 * time.Millisecond)

		// Stop SSE
		proxy.mu.Lock()
		if proxy.sseResp != nil {
			proxy.sseResp.Body.Close()
			proxy.sseResp = nil
		}
		proxy.mu.Unlock()

		// Verify at least some events were sent
		assert.Greater(t, len(eventsSent), 0, "should have received some events")
	})
}

// TestHeadlessProxyModelAndAgentPassing tests that model and agent parameters are correctly passed.
func TestHeadlessProxyModelAndAgentPassing(t *testing.T) {
	var capturedRequest map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/prompt_async") {
			json.NewDecoder(r.Body).Decode(&capturedRequest)
			w.WriteHeader(http.StatusAccepted)
		}
	}))
	defer server.Close()

	cfg := Config{BaseURLOverride: server.URL}
	proxy := NewProxy(cfg)
	proxy.config.SessionID = "ses-test"

	t.Run("WithModelAndProviderID", func(t *testing.T) {
		payload := PromptPayload{
			Text:       "test message",
			ProviderID: "openai",
			ModelID:    "gpt-4",
		}

		err := proxy.sendPrompt("ses-test", payload)
		require.NoError(t, err)

		require.NotNil(t, capturedRequest)
		model := capturedRequest["model"].(map[string]any)
		assert.Equal(t, "openai", model["providerID"])
		assert.Equal(t, "gpt-4", model["modelID"])

		parts := capturedRequest["parts"].([]any)
		firstPart := parts[0].(map[string]any)
		assert.Equal(t, "test message", firstPart["text"])
		assert.Equal(t, "text", firstPart["type"])
	})

	t.Run("WithoutModel", func(t *testing.T) {
		capturedRequest = nil
		payload := PromptPayload{
			Text: "test without model",
		}

		err := proxy.sendPrompt("ses-test", payload)
		require.NoError(t, err)

		require.NotNil(t, capturedRequest)
		_, hasModel := capturedRequest["model"]
		assert.False(t, hasModel, "should not include model when not specified")
	})
}
