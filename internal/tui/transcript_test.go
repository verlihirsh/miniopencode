package tui

import (
	"testing"

	"miniopencode/internal/client"
)

func TestTranscript_ApplyUpdate_OpAppend(t *testing.T) {
	tr := &Transcript{}
	tr.AddUserMessage("Hello")

	update := client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpAppend,
		Text:      "Hello",
	}
	tr.ApplyUpdate(update)

	update2 := client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpAppend,
		Text:      " world",
	}
	tr.ApplyUpdate(update2)

	if len(tr.messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(tr.messages))
	}

	assistantMsg := tr.messages[1]
	if assistantMsg.Role != RoleAssistant {
		t.Errorf("expected assistant role, got %v", assistantMsg.Role)
	}
	if len(assistantMsg.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(assistantMsg.Parts))
	}

	text := assistantMsg.Parts[0].Text.String()
	if text != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", text)
	}
}

func TestTranscript_ApplyUpdate_OpSet(t *testing.T) {
	tr := &Transcript{}
	tr.AddUserMessage("Hello")

	update := client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpSet,
		Text:      "First content",
	}
	tr.ApplyUpdate(update)

	update2 := client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpSet,
		Text:      "Replaced content",
	}
	tr.ApplyUpdate(update2)

	assistantMsg := tr.messages[1]
	text := assistantMsg.Parts[0].Text.String()
	if text != "Replaced content" {
		t.Errorf("expected 'Replaced content', got %q", text)
	}
}

func TestTranscript_ApplyUpdate_MultipleParts(t *testing.T) {
	tr := &Transcript{}
	tr.AddUserMessage("Hello")

	tr.ApplyUpdate(client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpAppend,
		Text:      "Answer text",
	})

	tr.ApplyUpdate(client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-2",
		Kind:      client.PartKindReasoning,
		Op:        client.OpAppend,
		Text:      "Thinking...",
	})

	tr.ApplyUpdate(client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-3",
		Kind:      client.PartKindTool,
		Op:        client.OpAppend,
		Text:      "Running tool",
	})

	assistantMsg := tr.messages[1]
	if len(assistantMsg.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(assistantMsg.Parts))
	}

	if assistantMsg.Parts[0].Kind != ChunkAnswer {
		t.Errorf("expected ChunkAnswer for part 0")
	}
	if assistantMsg.Parts[1].Kind != ChunkThinking {
		t.Errorf("expected ChunkThinking for part 1")
	}
	if assistantMsg.Parts[2].Kind != ChunkTool {
		t.Errorf("expected ChunkTool for part 2")
	}
}

func TestTranscript_ApplyUpdate_Pending(t *testing.T) {
	tr := &Transcript{}
	tr.AddUserMessage("Hello")
	tr.EnsurePendingAssistant("msg-1")

	if len(tr.messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(tr.messages))
	}

	assistantMsg := tr.messages[1]
	if !assistantMsg.Pending {
		t.Error("expected Pending=true before update")
	}

	tr.ApplyUpdate(client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpAppend,
		Text:      "First content",
	})

	assistantMsg = tr.messages[1]
	if assistantMsg.Pending {
		t.Error("expected Pending=false after non-empty update")
	}
}

func TestTranscript_ApplyUpdate_NoDuplication(t *testing.T) {
	tr := &Transcript{}
	tr.AddUserMessage("Hello")

	tr.ApplyUpdate(client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpSet,
		Text:      "Hello",
	})

	tr.ApplyUpdate(client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpSet,
		Text:      "Hello world",
	})

	tr.ApplyUpdate(client.StreamUpdate{
		MessageID: "msg-1",
		PartID:    "part-1",
		Kind:      client.PartKindText,
		Op:        client.OpSet,
		Text:      "Hello world!",
	})

	assistantMsg := tr.messages[1]
	if len(assistantMsg.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(assistantMsg.Parts))
	}

	text := assistantMsg.Parts[0].Text.String()
	if text != "Hello world!" {
		t.Errorf("expected 'Hello world!', got %q", text)
	}
}
