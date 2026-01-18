package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Config holds client configuration.
type Config struct {
	Host    string
	Port    int
	BaseURL string
	Timeout time.Duration
}

// ModelRef selects provider/model.
type ModelRef struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
}

// InputPart represents a prompt input part.
type InputPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	MIME string `json:"mime,omitempty"`
	URL  string `json:"url,omitempty"`
}

// PromptInput mirrors API body for /prompt_async.
type PromptInput struct {
	Parts   []InputPart `json:"parts"`
	Model   *ModelRef   `json:"model,omitempty"`
	Agent   string      `json:"agent,omitempty"`
	System  string      `json:"system,omitempty"`
	NoReply bool        `json:"noReply,omitempty"`
	Variant string      `json:"variant,omitempty"`
}

// Session represents minimal session info.
type Session struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Part  int    `json:"-"`
}

// TokenUsage captures token counts on a message.
type TokenUsage struct {
	Input     int `json:"input"`
	Output    int `json:"output"`
	Reasoning int `json:"reasoning"`
}

// Message represents a minimal session message with token metadata.
type Message struct {
	ID     string      `json:"id"`
	Tokens *TokenUsage `json:"tokens,omitempty"`
}

// SSEEvent is a parsed SSE event with optional event name and combined data.
type SSEEvent struct {
	Event string
	Data  []byte
}

// Client is a minimal HTTP client for opencode server.
type Client struct {
	baseURL       string
	http          *http.Client
	httpNoTimeout *http.Client
}

// New builds a client from config, defaulting host/port when BaseURL is empty.
func New(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		host := cfg.Host
		if host == "" {
			host = "127.0.0.1"
		}
		port := cfg.Port
		if port == 0 {
			port = 4096
		}
		baseURL = fmt.Sprintf("http://%s:%d", host, port)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL:       baseURL,
		http:          &http.Client{Timeout: timeout},
		httpNoTimeout: &http.Client{Timeout: 0},
	}
}

// ListSessions fetches sessions.
func (c *Client) ListSessions(ctx context.Context) ([]Session, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/session", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list sessions failed: %s", string(body))
	}
	var sessions []Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// CreateSession creates a session with given title.
func (c *Client) CreateSession(ctx context.Context, title string) (string, error) {
	body := map[string]string{"title": title}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/session", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create session failed: %s", string(body))
	}
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	return "", fmt.Errorf("no session id in response")
}

// SendPromptAsync posts to /session/{id}/prompt_async.
func (c *Client) SendPromptAsync(ctx context.Context, sessionID string, input PromptInput) error {
	b, _ := json.Marshal(input)
	url := fmt.Sprintf("%s/session/%s/prompt_async", c.baseURL, sessionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		log.Printf("client: build prompt request failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	log.Printf("client: prompt_async POST start session=%s url=%s", sessionID, url)
	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("client: prompt_async POST error session=%s err=%v", sessionID, err)
		return err
	}
	defer resp.Body.Close()
	log.Printf("client: prompt_async POST done session=%s status=%d", sessionID, resp.StatusCode)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("client: prompt_async POST failed session=%s status=%d body=%s", sessionID, resp.StatusCode, string(body))
		return fmt.Errorf("prompt POST failed: %s", string(body))
	}
	return nil
}

// ConsumeSSE connects to /event and streams events into provided channels.
func (c *Client) ConsumeSSE(ctx context.Context, out chan<- SSEEvent, errs chan<- error) {
	url := c.baseURL + "/event"
	sseClient := NewSSEClient(url)
	sseClient.Connect(ctx, out, errs)
}
