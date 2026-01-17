package tui

import (
	"context"
	"fmt"
	"strings"

	"opencode-tty/internal/client"
)

type Streamer struct {
	Client *client.Client
	Events chan Chunk
	Errors chan error
}

func (s *Streamer) Start(ctx context.Context) {
	raw := make(chan client.SSEEvent, 32)
	errs := make(chan error, 1)
	go s.Client.ConsumeSSE(ctx, raw, errs)
	go func() {
		for {
			select {
			case ev := <-raw:
				if len(ev.Data) == 0 {
					continue
				}
				// split if multiple JSON objects in combined data
				parts := splitJSONLines(ev.Data)
				for _, p := range parts {
					s.Events <- makeChunk(p)
				}
			case err := <-errs:
				s.Errors <- err
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func splitJSONLines(b []byte) [][]byte {
	lines := strings.Split(string(b), "\n")
	var out [][]byte
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		out = append(out, []byte(l))
	}
	return out
}

func (s *Streamer) SendPrompt(ctx context.Context, sessionID string, text string, cfg PromptConfig) error {
	input := client.PromptInput{
		Parts: []client.InputPart{{Type: "text", Text: text}},
	}
	if cfg.ModelID != "" || cfg.ProviderID != "" {
		input.Model = &client.ModelRef{ProviderID: cfg.ProviderID, ModelID: cfg.ModelID}
	}
	if cfg.Agent != "" {
		input.Agent = cfg.Agent
	}
	return s.Client.SendPromptAsync(ctx, sessionID, input)
}

type PromptConfig struct {
	Agent      string
	ProviderID string
	ModelID    string
}

type SessionResolver interface {
	Resolve(ctx context.Context, defaultSession string) (string, error)
}

func (s *Streamer) SelectDefaultSession(ctx context.Context, res SessionResolver, def string) (string, error) {
	id, err := res.Resolve(ctx, def)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Streamer) EnsureSession(ctx context.Context, res SessionResolver, def string) (string, error) {
	id, err := s.SelectDefaultSession(ctx, res, def)
	if err != nil {
		return "", fmt.Errorf("session: %w", err)
	}
	return id, nil
}
