package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBaseURLFromHostPort(t *testing.T) {
	cfg := Config{Host: "example.com", Port: 1234}
	c := New(cfg)
	if got := c.baseURL; got != "http://example.com:1234" {
		t.Fatalf("base url mismatch: %s", got)
	}
}

func TestListSessions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session" || r.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[{"id":"ses1","title":"t1"},{"id":"ses2","title":"t2"}]`)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	sessions, err := c.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sessions) != 2 || sessions[0].ID != "ses1" || sessions[1].Title != "t2" {
		t.Fatalf("unexpected sessions: %+v", sessions)
	}
}

func TestCreateSession(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"ses-new"}`)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	id, err := c.CreateSession(context.Background(), "hello")
	if err != nil || id != "ses-new" {
		t.Fatalf("create: id=%s err=%v", id, err)
	}
	if !strings.Contains(string(body), "hello") {
		t.Fatalf("expected title in body, got %s", string(body))
	}
}

func TestPromptAsyncSendsModelAndAgent(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session/ses123/prompt_async" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
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
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	model := captured["model"].(map[string]any)
	if model["providerID"] != "anthropic" || model["modelID"] != "claude" {
		t.Fatalf("model not captured: %v", model)
	}
	if captured["agent"].(string) != "build" {
		t.Fatalf("agent not captured: %v", captured)
	}
}

func TestSSEReaderCombinesMultiline(t *testing.T) {
	stream := "" +
		"event: message.part.updated\n" +
		"data: {\"delta\":\"hi\"}\n" +
		"data: {\"delta\":\" there\"}\n" +
		"\n" +
		"data: {\"done\":true}\n\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		bw := bufio.NewWriter(w)
		bw.WriteString(stream)
		bw.Flush()
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan SSEEvent, 4)
	errs := make(chan error, 1)
	go c.ConsumeSSE(ctx, events, errs)

	var received []SSEEvent
	deadline := time.After(2 * time.Second)
loop:
	for {
		select {
		case ev := <-events:
			received = append(received, ev)
			if len(received) == 2 {
				break loop
			}
		case err := <-errs:
			t.Fatalf("sse error: %v", err)
		case <-deadline:
			t.Fatalf("timeout waiting for events")
		}
	}

	if received[0].Event != "message.part.updated" || !bytes.Contains(received[0].Data, []byte("hi")) {
		t.Fatalf("unexpected first event: %+v", received[0])
	}
	if !bytes.Contains(received[0].Data, []byte("there")) {
		t.Fatalf("expected combined data: %s", string(received[0].Data))
	}
	if !bytes.Contains(received[1].Data, []byte("done")) {
		t.Fatalf("unexpected second event: %+v", received[1])
	}
}
