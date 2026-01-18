package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"miniopencode/internal/client"
	"miniopencode/internal/config"
	"miniopencode/internal/session"
)

// TestE2ESessionResolutionWithSSE is an integration test that validates the complete workflow:
// 1. Resolve a daily session (create or reuse)
// 2. Send a prompt to the session
// 3. Receive SSE events for the response
func TestE2ESessionResolutionWithSSE(t *testing.T) {
	var (
		mu              sync.Mutex
		sessionCreated  bool
		promptReceived  bool
		sseConnected    bool
		currentSessionID string
	)

	// Create mock opencode server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		// List sessions endpoint
		case r.URL.Path == "/session" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if currentSessionID == "" {
				json.NewEncoder(w).Encode([]map[string]string{})
			} else {
				json.NewEncoder(w).Encode([]map[string]interface{}{
					{"id": currentSessionID, "title": "2026-01-17-daily-1"},
				})
			}

		// Create session endpoint
		case r.URL.Path == "/session" && r.Method == http.MethodPost:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			currentSessionID = "ses-" + body["title"]
			sessionCreated = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"id": currentSessionID})

		// Get session messages (for token counting)
		case r.URL.Path == "/session/"+currentSessionID+"/message" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			// Return empty messages array (session is under limits)
			json.NewEncoder(w).Encode([]map[string]interface{}{})

		// Send prompt endpoint
		case r.URL.Path == "/session/"+currentSessionID+"/prompt_async" && r.Method == http.MethodPost:
			promptReceived = true
			w.WriteHeader(http.StatusAccepted)

		// SSE event stream
		case r.URL.Path == "/event":
			sseConnected = true
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
			
			// Send sample SSE events
			events := []string{
				"event: message.part.updated\ndata: {\"delta\":\"Hello\"}\n\n",
				"event: message.part.updated\ndata: {\"delta\":\" World\"}\n\n",
				"event: message.updated\ndata: {\"status\":\"completed\"}\n\n",
			}
			
			for _, event := range events {
				fmt.Fprint(w, event)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(10 * time.Millisecond)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Setup client and resolver
	cfg := config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 4096,
		},
		Session: config.SessionConfig{
			DailyTitleFormat: "2006-01-02-daily-%d",
			DailyMaxTokens:   250000,
			DailyMaxMessages: 4000,
		},
	}

	c := client.New(client.Config{BaseURL: server.URL})
	resolver := session.Resolver{
		Client: c,
		Config: cfg,
		Now:    func() time.Time { return time.Date(2026, 1, 17, 12, 0, 0, 0, time.UTC) },
	}

	ctx := context.Background()

	// Step 1: Resolve daily session
	t.Run("ResolveSession", func(t *testing.T) {
		sessionID, err := resolver.Resolve(ctx, "daily")
		require.NoError(t, err)
		assert.NotEmpty(t, sessionID)
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return sessionCreated
		}, time.Second, 10*time.Millisecond, "session should be created")
	})

	// Step 2: Send a prompt
	t.Run("SendPrompt", func(t *testing.T) {
		mu.Lock()
		sessionID := currentSessionID
		mu.Unlock()

		err := c.SendPromptAsync(ctx, sessionID, client.PromptInput{
			Parts: []client.InputPart{
				{Type: "text", Text: "Hello from integration test"},
			},
			Agent: "test",
		})
		require.NoError(t, err)
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return promptReceived
		}, time.Second, 10*time.Millisecond, "prompt should be received")
	})

	// Step 3: Consume SSE events
	t.Run("ConsumeSSE", func(t *testing.T) {
		eventChan := make(chan client.SSEEvent, 10)
		errChan := make(chan error, 1)

		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		go c.ConsumeSSE(ctx, eventChan, errChan)

		// Wait for SSE connection
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return sseConnected
		}, time.Second, 10*time.Millisecond, "SSE should connect")

		// Collect events
		var events []client.SSEEvent
		deadline := time.After(1 * time.Second)
		for len(events) < 3 {
			select {
			case evt := <-eventChan:
				events = append(events, evt)
			case err := <-errChan:
				t.Logf("SSE error (may be expected on close): %v", err)
			case <-deadline:
				break
			}
		}

		assert.GreaterOrEqual(t, len(events), 1, "should receive at least one SSE event")
		if len(events) > 0 {
			assert.Equal(t, "message.part.updated", events[0].Event)
		}
	})
}

// TestE2EDailySessionRollover tests that when limits are exceeded, a new daily session part is created.
func TestE2EDailySessionRollover(t *testing.T) {
	const testTokenLimit = 1000 // Set low limit to trigger rollover
	
	var (
		mu           sync.Mutex
		createdParts []string
	)

	// Create mock server that simulates a session with high token usage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.URL.Path == "/session" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			// Return existing session that's over the token limit
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": "ses-part1", "title": "2026-01-17-daily-1"},
			})

		case r.URL.Path == "/session" && r.Method == http.MethodPost:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			createdParts = append(createdParts, body["title"])
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"id": "ses-" + body["title"],
			})

		case r.URL.Path == "/session/ses-part1/message":
			w.Header().Set("Content-Type", "application/json")
			// Return messages that exceed the token limit
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id": "msg1",
					"tokens": map[string]int{
						"input":  600,
						"output": 500,
					},
				},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := config.Config{
		Session: config.SessionConfig{
			DailyTitleFormat: "2006-01-02-daily-%d",
			DailyMaxTokens:   testTokenLimit,
			DailyMaxMessages: 4000,
		},
	}

	c := client.New(client.Config{BaseURL: server.URL})
	resolver := session.Resolver{
		Client: c,
		Config: cfg,
		Now:    func() time.Time { return time.Date(2026, 1, 17, 12, 0, 0, 0, time.UTC) },
	}

	// Resolve daily session - should create part 2 because part 1 exceeds limits
	sessionID, err := resolver.Resolve(context.Background(), "daily")
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, createdParts, 1)
	assert.Equal(t, "2026-01-17-daily-2", createdParts[0], "should create next part when limits exceeded")
}

// TestE2EConfigAndClientIntegration tests that config properly drives client behavior.
func TestE2EConfigAndClientIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]string{})
	}))
	defer server.Close()

	// Create client with server URL
	c := client.New(client.Config{BaseURL: server.URL})

	// Verify client can make requests (even if they're mocked)
	ctx := context.Background()
	sessions, err := c.ListSessions(ctx)
	assert.NoError(t, err, "client should successfully make requests")
	assert.NotNil(t, sessions, "should return sessions list")
}
