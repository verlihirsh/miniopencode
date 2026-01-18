package tui

import (
	"strings"
	"time"

	"opencode-tty/internal/client"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type TranscriptPart struct {
	ID   string
	Kind ChunkKind
	Text strings.Builder
}

type TranscriptMessage struct {
	ID      string
	Role    Role
	Created time.Time
	Pending bool
	Parts   []TranscriptPart
}

type Transcript struct {
	messages []TranscriptMessage
}

func (t *Transcript) AddUserMessage(text string) {
	var b strings.Builder
	b.WriteString(text)
	t.messages = append(t.messages, TranscriptMessage{
		Role:    RoleUser,
		Created: time.Now(),
		Parts:   []TranscriptPart{{Kind: ChunkAnswer, Text: b}},
	})
}

func (t *Transcript) EnsurePendingAssistant(messageID string) {
	if len(t.messages) == 0 || t.messages[len(t.messages)-1].Role != RoleAssistant {
		t.messages = append(t.messages, TranscriptMessage{
			ID:      messageID,
			Role:    RoleAssistant,
			Created: time.Now(),
			Pending: true,
		})
		return
	}
	if t.messages[len(t.messages)-1].ID == "" && messageID != "" {
		t.messages[len(t.messages)-1].ID = messageID
	}
}

func (t *Transcript) EnsureAssistantMessage(messageID string) {
	t.EnsurePendingAssistant(messageID)
}

func partKindToChunkKind(pk client.PartKind) ChunkKind {
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

func (t *Transcript) ApplyUpdate(update client.StreamUpdate) {
	t.EnsurePendingAssistant(update.MessageID)
	msg := &t.messages[len(t.messages)-1]

	if msg.ID == "" && update.MessageID != "" {
		msg.ID = update.MessageID
	}

	chunkKind := partKindToChunkKind(update.Kind)

	var part *TranscriptPart
	for i := range msg.Parts {
		if msg.Parts[i].ID == update.PartID {
			part = &msg.Parts[i]
			break
		}
	}

	if part == nil {
		msg.Parts = append(msg.Parts, TranscriptPart{
			ID:   update.PartID,
			Kind: chunkKind,
		})
		part = &msg.Parts[len(msg.Parts)-1]
	}

	switch update.Op {
	case client.OpAppend:
		part.Text.WriteString(update.Text)
	case client.OpSet:
		part.Text.Reset()
		part.Text.WriteString(update.Text)
	}

	if update.Text != "" && msg.Pending {
		msg.Pending = false
	}
}

func (t *Transcript) AppendAssistantChunk(messageID, partID string, kind ChunkKind, text string) {
	update := client.StreamUpdate{
		MessageID: messageID,
		PartID:    partID,
		Kind:      chunkKindToPartKind(kind),
		Op:        client.OpAppend,
		Text:      text,
	}
	t.ApplyUpdate(update)
}

func chunkKindToPartKind(ck ChunkKind) client.PartKind {
	switch ck {
	case ChunkAnswer:
		return client.PartKindText
	case ChunkThinking:
		return client.PartKindReasoning
	case ChunkTool:
		return client.PartKindTool
	default:
		return client.PartKindOther
	}
}

func (t *Transcript) AddAssistantSystemLine(text string) {
	t.messages = append(t.messages, TranscriptMessage{Role: RoleAssistant, Created: time.Now()})
	msg := &t.messages[len(t.messages)-1]
	msg.Parts = append(msg.Parts, TranscriptPart{Kind: ChunkAnswer})
	msg.Parts[len(msg.Parts)-1].Text.WriteString(text)
}

func (t Transcript) Render(showThinking, showTools bool, spinnerFrame string, showSpinner bool) string {
	var b strings.Builder
	for i, m := range t.messages {
		if i > 0 {
			b.WriteString("\n\n")
		}

		if m.Role == RoleUser {
			b.WriteString(answerStyle.Render("You:"))
			b.WriteString(" ")
			if len(m.Parts) > 0 {
				b.WriteString(m.Parts[0].Text.String())
			}
			continue
		}

		b.WriteString(answerStyle.Render("Assistant:"))
		if showSpinner && m.Pending {
			b.WriteString("\n")
			b.WriteString(answerStyle.Render(spinnerFrame))
			continue
		}
		for _, p := range m.Parts {
			if p.Kind == ChunkThinking && !showThinking {
				continue
			}
			if p.Kind == ChunkTool && !showTools {
				continue
			}
			b.WriteString("\n")
			text := p.Text.String()
			switch p.Kind {
			case ChunkThinking:
				b.WriteString(thinkingStyle.Render(text))
			case ChunkTool:
				b.WriteString(toolStyle.Render(text))
			default:
				b.WriteString(renderMarkdown(80, text))
			}
		}
	}
	return b.String()
}
