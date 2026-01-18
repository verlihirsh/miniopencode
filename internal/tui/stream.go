package tui

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"miniopencode/internal/client"
)

type Streamer struct {
	Client *client.Client
	Events chan Chunk
	Errors chan error

	mu           sync.RWMutex
	messageRoles map[string]string
	partTexts    map[string]string // tracks last known text per partID for delta computation
}

func (s *Streamer) Start(ctx context.Context) {
	s.messageRoles = make(map[string]string)
	s.partTexts = make(map[string]string)
	raw := make(chan client.SSEEvent, 32)
	errs := make(chan error, 1)

	go func() {
		s.Client.ConsumeSSE(ctx, raw, errs)
		close(raw)
	}()

	go func() {
		defer close(s.Events)
		defer close(s.Errors)
		for {
			select {
			case ev, ok := <-raw:
				if !ok {
					return
				}
				if len(ev.Data) == 0 {
					continue
				}
				chunk := s.parseSSEToChunk(ev)
				if chunk.Kind == ChunkSkip {
					continue
				}
				if chunk.Text == "" && chunk.Kind != ChunkMeta {
					continue
				}
				s.Events <- chunk
			case err, ok := <-errs:
				if !ok {
					return
				}
				log.Printf("tui: sse error: %v", err)
				s.Errors <- fmt.Errorf("sse stream error: %w", err)
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Streamer) parseSSEToChunk(ev client.SSEEvent) Chunk {
	parsed, err := client.ParseEvent(ev)
	if err != nil {
		log.Printf("tui: parse event error: %v", err)
		return Chunk{Kind: ChunkSkip}
	}

	switch e := parsed.(type) {
	case *client.MessageUpdatedEvent:
		info := e.Properties.Info
		s.mu.Lock()
		s.messageRoles[info.ID] = info.Role
		s.mu.Unlock()
		log.Printf("tui: message.updated id=%s role=%s", info.ID, info.Role)
		if info.Role == "assistant" {
			return Chunk{
				Kind:      ChunkMeta,
				MessageID: info.ID,
				Complete:  info.IsComplete(),
			}
		}
		return Chunk{Kind: ChunkSkip}

	case *client.MessagePartUpdatedEvent:
		msgID := e.Properties.Part.MessageID
		s.mu.RLock()
		role := s.messageRoles[msgID]
		s.mu.RUnlock()

		if role == "user" {
			log.Printf("tui: skipping user message part msgID=%s", msgID)
			return Chunk{Kind: ChunkSkip}
		}

		update := e.ToStreamUpdate()
		text := strings.Clone(update.Text)

		s.mu.Lock()
		if update.Op == client.OpSet {
			prev := s.partTexts[update.PartID]
			if text == prev {
				s.mu.Unlock()
				return Chunk{Kind: ChunkSkip}
			}
			if len(text) > len(prev) && text[:len(prev)] == prev {
				text = strings.Clone(text[len(prev):])
			}
			s.partTexts[update.PartID] = text
		} else {
			s.partTexts[update.PartID] += text
		}
		s.mu.Unlock()

		return Chunk{
			Kind:      streamUpdateKindToChunkKind(update.Kind),
			Text:      text,
			PartID:    update.PartID,
			MessageID: update.MessageID,
			Complete:  update.Complete,
		}
	default:
		return Chunk{Kind: ChunkSkip}
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
