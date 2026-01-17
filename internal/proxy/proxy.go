package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Config holds proxy configuration.
type Config struct {
	Host            string
	Port            string
	BaseURLOverride string
	SessionID       string
}

// BaseURL returns an explicit base URL if provided, otherwise builds from host/port.
func (c Config) BaseURL() string {
	if c.BaseURLOverride != "" {
		return c.BaseURLOverride
	}
	host := c.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := c.Port
	if port == "" {
		port = "4096"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}

// Command types for stdin commands.
type Command struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type PromptPayload struct {
	Text       string `json:"text"`
	ProviderID string `json:"provider_id,omitempty"`
	ModelID    string `json:"model_id,omitempty"`
}

type SessionPayload struct {
	Title string `json:"title,omitempty"`
	ID    string `json:"id,omitempty"`
}

// Proxy handles communication between stdin/stdout and opencode server.
type Proxy struct {
	config    Config
	baseURL   string
	sseClient *http.Client
	mu        sync.Mutex
	sseResp   *http.Response
}

// NewProxy constructs a Proxy with a computed base URL.
func NewProxy(config Config) *Proxy {
	baseURL := config.BaseURL()
	config.BaseURLOverride = baseURL
	return &Proxy{
		config:    config,
		baseURL:   baseURL,
		sseClient: &http.Client{},
	}

}

// BaseURL returns the computed base URL.
func (p *Proxy) BaseURL() string {
	return p.baseURL
}

// CheckHealth checks if server is running.
func (p *Proxy) CheckHealth() bool {
	resp, err := http.Get(p.baseURL + "/global/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// createSession creates a new session.
func (p *Proxy) createSession(title string) (string, error) {
	body := map[string]string{"title": title}
	b, _ := json.Marshal(body)

	resp, err := http.Post(p.baseURL+"/session", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	if id, ok := result["ID"].(string); ok {
		return id, nil
	}
	return "", fmt.Errorf("no session ID in response")
}

// listSessions lists all sessions.
func (p *Proxy) listSessions() ([]map[string]interface{}, error) {
	resp, err := http.Get(p.baseURL + "/session")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// sendPrompt sends a message to the session.
func (p *Proxy) sendPrompt(sessionID string, payload PromptPayload) error {
	body := map[string]interface{}{
		"parts": []map[string]string{
			{"type": "text", "text": payload.Text},
		},
	}

	if payload.ProviderID != "" && payload.ModelID != "" {
		body["model"] = map[string]string{
			"providerID": payload.ProviderID,
			"modelID":    payload.ModelID,
		}
	}

	b, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/session/%s/prompt_async", p.baseURL, sessionID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("prompt failed: %s", string(body))
	}

	return nil
}

// startSSE connects to SSE endpoint and streams events to stdout.
func (p *Proxy) startSSE() error {
	req, err := http.NewRequest(http.MethodGet, p.baseURL+"/event", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := p.sseClient.Do(req)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.sseResp = resp
	p.mu.Unlock()

	go p.readSSE(resp)
	return nil
}

// output sends JSON to stdout.
func (p *Proxy) output(eventType string, data interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := map[string]interface{}{
		"type": eventType,
		"data": data,
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}

// outputError sends error to stdout.
func (p *Proxy) outputError(err error) {
	p.output("error", map[string]string{"message": err.Error()})
}

// outputRaw sends raw SSE event to stdout.
func (p *Proxy) outputRaw(eventType string, raw json.RawMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := map[string]interface{}{
		"type": eventType,
		"data": raw,
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}

// readSSE reads SSE events and outputs to stdout.
func (p *Proxy) readSSE(resp *http.Response) {
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "" {
				continue
			}

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			p.output("sse", event)
		}
	}

	if err := scanner.Err(); err != nil {
		p.outputError(fmt.Errorf("SSE read error: %v", err))
	}
}

// handleCommand processes a command from stdin.
func (p *Proxy) handleCommand(cmd Command) {
	switch cmd.Type {
	case "health":
		healthy := p.CheckHealth()
		p.output("health", map[string]bool{"healthy": healthy})

	case "session.create":
		var payload SessionPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			p.outputError(err)
			return
		}
		id, err := p.createSession(payload.Title)
		if err != nil {
			p.outputError(err)
			return
		}
		p.config.SessionID = id
		p.output("session.created", map[string]string{"id": id})

	case "session.list":
		sessions, err := p.listSessions()
		if err != nil {
			p.outputError(err)
			return
		}
		p.output("session.list", sessions)

	case "session.select":
		var payload SessionPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			p.outputError(err)
			return
		}
		p.config.SessionID = payload.ID
		p.output("session.selected", map[string]string{"id": payload.ID})

	case "prompt":
		if p.config.SessionID == "" {
			p.outputError(fmt.Errorf("no session selected"))
			return
		}
		var payload PromptPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			p.outputError(err)
			return
		}
		if err := p.sendPrompt(p.config.SessionID, payload); err != nil {
			p.outputError(err)
			return
		}
		p.output("prompt.sent", map[string]string{"session_id": p.config.SessionID})

	case "sse.start":
		if err := p.startSSE(); err != nil {
			p.outputError(err)
			return
		}
		p.output("sse.started", nil)

	case "sse.stop":
		p.mu.Lock()
		if p.sseResp != nil {
			p.sseResp.Body.Close()
			p.sseResp = nil
		}
		p.mu.Unlock()
		p.output("sse.stopped", nil)

	default:
		p.outputError(fmt.Errorf("unknown command: %s", cmd.Type))
	}
}

// RunHeadless starts the proxy, reading from stdin and writing to stdout.
func (p *Proxy) RunHeadless() {
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	p.output("ready", map[string]string{
		"host": p.config.Host,
		"port": p.config.Port,
	})

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var cmd Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			p.outputError(fmt.Errorf("invalid JSON: %v", err))
			continue
		}

		p.handleCommand(cmd)
	}

	if err := scanner.Err(); err != nil {
		p.outputError(err)
	}
}
