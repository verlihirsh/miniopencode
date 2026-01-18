package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSSEClient_Connect_ReceivesEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
		}

		w.Write([]byte("event: message.part.updated\n"))
		w.Write([]byte("data: {\"type\":\"message.part.updated\"}\n\n"))
		flusher.Flush()

		w.Write([]byte("event: message.updated\n"))
		w.Write([]byte("data: {\"type\":\"message.updated\"}\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events := make(chan SSEEvent, 10)
	errs := make(chan error, 1)

	sseClient := NewSSEClient(server.URL)
	go sseClient.Connect(ctx, events, errs)

	var received []SSEEvent
	timeout := time.After(1 * time.Second)

loop:
	for {
		select {
		case ev := <-events:
			received = append(received, ev)
			if len(received) >= 2 {
				break loop
			}
		case err := <-errs:
			t.Fatalf("unexpected error: %v", err)
		case <-timeout:
			break loop
		}
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}

	if received[0].Event != "message.part.updated" {
		t.Errorf("expected event 'message.part.updated', got %q", received[0].Event)
	}
	if received[1].Event != "message.updated" {
		t.Errorf("expected event 'message.updated', got %q", received[1].Event)
	}
}

func TestSSEClient_Connect_MultiLineData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)

		w.Write([]byte("event: test\n"))
		w.Write([]byte("data: line1\n"))
		w.Write([]byte("data: line2\n"))
		w.Write([]byte("data: line3\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events := make(chan SSEEvent, 10)
	errs := make(chan error, 1)

	sseClient := NewSSEClient(server.URL)
	go sseClient.Connect(ctx, events, errs)

	select {
	case ev := <-events:
		expected := "line1\nline2\nline3"
		if string(ev.Data) != expected {
			t.Errorf("expected data %q, got %q", expected, string(ev.Data))
		}
	case err := <-errs:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestSSEClient_Connect_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	events := make(chan SSEEvent, 10)
	errs := make(chan error, 1)

	sseClient := NewSSEClient(server.URL)
	done := make(chan struct{})
	go func() {
		sseClient.Connect(ctx, events, errs)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Connect did not exit after context cancellation")
	}
}
