package tui

import (
	"context"
	"fmt"
	"log"

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
				chunk := parseSSEToChunk(ev)
				if chunk.Text == "" {
					continue
				}
				s.Events <- chunk
			case err := <-errs:
				log.Printf("tui: sse error: %v", err)
				s.Errors <- fmt.Errorf("sse stream error: %w", err)
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func parseSSEToChunk(ev client.SSEEvent) Chunk {
	parsed, err := client.ParseEvent(ev)
	if err != nil {
		log.Printf("tui: parse event error: %v", err)
		return Chunk{Kind: ChunkRaw, Text: ""}
	}

	switch e := parsed.(type) {
	case *client.MessagePartUpdatedEvent:
		update := e.ToStreamUpdate()
		return Chunk{
			Kind:      streamUpdateKindToChunkKind(update.Kind),
			Text:      update.Text,
			PartID:    update.PartID,
			MessageID: update.MessageID,
			Complete:  update.Complete,
		}
	default:
		return Chunk{Kind: ChunkRaw, Text: ""}
	}
}

func streamUpdateKindToChunkKind(pk client.PartKind) ChunkKind {
	switch pk {
	case client.PartKindText:
		return ChunkAnswer
	case client.PartKindReasoning:
		return ChunkThinking
	case client.PartKindTool:
		return ChunkTool
	default:
		return ChunkRaw
	}
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
