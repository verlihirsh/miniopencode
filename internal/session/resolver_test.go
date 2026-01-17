package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"opencode-tty/internal/client"
	"opencode-tty/internal/config"
)

type stubClient struct {
	sessions []client.Session
	messages map[string][]client.Message
	created  []string
}

func (s *stubClient) ListSessions(ctx context.Context) ([]client.Session, error) {
	return s.sessions, nil
}

func (s *stubClient) ListMessages(ctx context.Context, sessionID string) ([]client.Message, error) {
	msgs, ok := s.messages[sessionID]
	if !ok {
		return nil, errors.New("missing messages")
	}
	return msgs, nil
}

func (s *stubClient) CreateSession(ctx context.Context, title string) (string, error) {
	id := "new-" + title
	s.created = append(s.created, title)
	s.sessions = append(s.sessions, client.Session{ID: id, Title: title})
	return id, nil
}

func fixedNow() time.Time { return time.Date(2026, 1, 17, 12, 0, 0, 0, time.UTC) }

func baseConfig() config.Config {
	return config.Config{
		Session: config.SessionConfig{
			DailyTitleFormat: "2006-01-02-daily-%d",
			DailyMaxTokens:   10,
			DailyMaxMessages: 100,
		},
	}
}

func TestResolveSpecificSessionExists(t *testing.T) {
	sc := &stubClient{sessions: []client.Session{{ID: "ses-1", Title: "foo"}}}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "ses-1")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id != "ses-1" {
		t.Fatalf("expected existing id, got %s", id)
	}
}

func TestResolveSpecificSessionMissingCreates(t *testing.T) {
	sc := &stubClient{sessions: []client.Session{}}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "target-session")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id != "new-target-session" {
		t.Fatalf("expected created id, got %s", id)
	}
	if len(sc.created) != 1 || sc.created[0] != "target-session" {
		t.Fatalf("expected create with title target-session, got %v", sc.created)
	}
}

func TestResolveDailyCreatesFirst(t *testing.T) {
	sc := &stubClient{sessions: []client.Session{}}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "daily")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	expectedTitle := "2026-01-17-daily-1"
	if id != "new-"+expectedTitle {
		t.Fatalf("expected new daily id, got %s", id)
	}
}

func TestResolveDailyUsesExistingUnderLimit(t *testing.T) {
	title := "2026-01-17-daily-2"
	sc := &stubClient{
		sessions: []client.Session{{ID: "ses-2", Title: title}},
		messages: map[string][]client.Message{
			"ses-2": {{Tokens: &client.TokenUsage{Input: 3, Output: 4}}},
		},
	}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "daily")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id != "ses-2" {
		t.Fatalf("expected reuse existing, got %s", id)
	}
}

func TestResolveDailyRollsOverWhenTokensExceeded(t *testing.T) {
	title := "2026-01-17-daily-3"
	sc := &stubClient{
		sessions: []client.Session{{ID: "ses-3", Title: title}},
		messages: map[string][]client.Message{
			"ses-3": {{Tokens: &client.TokenUsage{Input: 8, Output: 5}}},
		},
	}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "daily")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	expectedNew := "2026-01-17-daily-4"
	if id != "new-"+expectedNew {
		t.Fatalf("expected rollover to new part, got %s", id)
	}
}

func TestResolveDailyCreatesTodayWhenOnlyOldSessions(t *testing.T) {
	sc := &stubClient{
		sessions: []client.Session{{ID: "old", Title: "2026-01-16-daily-5"}},
		messages: map[string][]client.Message{"old": {}},
	}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "daily")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	expected := "2026-01-17-daily-1"
	if id != "new-"+expected {
		t.Fatalf("expected new today part, got %s", id)
	}
}
