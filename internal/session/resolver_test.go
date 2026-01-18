package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"miniopencode/internal/client"
	"miniopencode/internal/config"
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
	require.NoError(t, err)
	assert.Equal(t, "ses-1", id, "should return existing session ID")
	assert.Empty(t, sc.created, "should not create new session")
}

func TestResolveSpecificSessionMissingCreates(t *testing.T) {
	sc := &stubClient{sessions: []client.Session{}}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "target-session")
	require.NoError(t, err)
	assert.Equal(t, "new-target-session", id, "should create session with target title")
	require.Len(t, sc.created, 1)
	assert.Equal(t, "target-session", sc.created[0])
}

func TestResolveDailyCreatesFirst(t *testing.T) {
	sc := &stubClient{sessions: []client.Session{}}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "daily")
	require.NoError(t, err)

	expectedTitle := "2026-01-17-daily-1"
	assert.Equal(t, "new-"+expectedTitle, id, "should create first daily session")
	require.Len(t, sc.created, 1)
	assert.Equal(t, expectedTitle, sc.created[0])
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
	require.NoError(t, err)
	assert.Equal(t, "ses-2", id, "should reuse existing session under limits")
	assert.Empty(t, sc.created, "should not create new session")
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
	require.NoError(t, err)

	expectedNew := "2026-01-17-daily-4"
	assert.Equal(t, "new-"+expectedNew, id, "should create new part when tokens exceeded")
	require.Len(t, sc.created, 1)
	assert.Equal(t, expectedNew, sc.created[0])
}

func TestResolveDailyCreatesTodayWhenOnlyOldSessions(t *testing.T) {
	sc := &stubClient{
		sessions: []client.Session{{ID: "old", Title: "2026-01-16-daily-5"}},
		messages: map[string][]client.Message{"old": {}},
	}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	id, err := r.Resolve(context.Background(), "daily")
	require.NoError(t, err)

	expected := "2026-01-17-daily-1"
	assert.Equal(t, "new-"+expected, id, "should create first session for today")
	require.Len(t, sc.created, 1)
	assert.Equal(t, expected, sc.created[0])
}

func TestResolveDailyRollsOverOnMessageLimit(t *testing.T) {
	const testMessageLimit = 2
	
	title := "2026-01-17-daily-10"
	cfg := baseConfig()
	cfg.Session.DailyMaxMessages = testMessageLimit
	
	sc := &stubClient{
		sessions: []client.Session{{ID: "ses-10", Title: title}},
		messages: map[string][]client.Message{
			"ses-10": {
				{Tokens: &client.TokenUsage{Input: 1, Output: 1}},
				{Tokens: &client.TokenUsage{Input: 1, Output: 1}},
				{Tokens: &client.TokenUsage{Input: 1, Output: 1}}, // 3 messages exceeds limit
			},
		},
	}
	r := Resolver{Client: sc, Config: cfg, Now: fixedNow}

	id, err := r.Resolve(context.Background(), "daily")
	require.NoError(t, err)

	expectedNew := "2026-01-17-daily-11"
	assert.Equal(t, "new-"+expectedNew, id, "should create new part when message limit exceeded")
}

func TestResolveEmptySessionID(t *testing.T) {
	sc := &stubClient{sessions: []client.Session{}}
	r := Resolver{Client: sc, Config: baseConfig(), Now: fixedNow}

	_, err := r.Resolve(context.Background(), "")
	assert.Error(t, err, "should return error when no session ID provided")
	assert.Contains(t, err.Error(), "no default session configured")
}
