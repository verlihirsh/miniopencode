package client

import (
	"encoding/json"
	"testing"
)

func TestDecodeMessagePartUpdatedEvent_TextPart(t *testing.T) {
	raw := `{
		"type": "message.part.updated",
		"properties": {
			"part": {
				"id": "part-123",
				"sessionID": "session-456",
				"messageID": "msg-789",
				"type": "text",
				"text": "Hello world",
				"time": {"start": 1234567890.123}
			},
			"delta": "world"
		}
	}`

	var ev MessagePartUpdatedEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if ev.Type != "message.part.updated" {
		t.Errorf("expected type 'message.part.updated', got %q", ev.Type)
	}
	if ev.Properties.Delta != "world" {
		t.Errorf("expected delta 'world', got %q", ev.Properties.Delta)
	}
	if ev.Properties.Part.ID != "part-123" {
		t.Errorf("expected part ID 'part-123', got %q", ev.Properties.Part.ID)
	}
	if ev.Properties.Part.MessageID != "msg-789" {
		t.Errorf("expected message ID 'msg-789', got %q", ev.Properties.Part.MessageID)
	}
	if ev.Properties.Part.PartType != "text" {
		t.Errorf("expected part type 'text', got %q", ev.Properties.Part.PartType)
	}
	if ev.Properties.Part.Text != "Hello world" {
		t.Errorf("expected text 'Hello world', got %q", ev.Properties.Part.Text)
	}
}

func TestDecodeMessagePartUpdatedEvent_ReasoningPart(t *testing.T) {
	raw := `{
		"type": "message.part.updated",
		"properties": {
			"part": {
				"id": "part-abc",
				"sessionID": "session-456",
				"messageID": "msg-789",
				"type": "reasoning",
				"text": "thinking...",
				"time": {"start": 1234567890.123, "end": 1234567890.456}
			}
		}
	}`

	var ev MessagePartUpdatedEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if ev.Properties.Part.PartType != "reasoning" {
		t.Errorf("expected part type 'reasoning', got %q", ev.Properties.Part.PartType)
	}
	if ev.Properties.Part.Time.End == nil {
		t.Error("expected time.end to be set")
	}
	if ev.Properties.Delta != "" {
		t.Errorf("expected empty delta, got %q", ev.Properties.Delta)
	}
}

func TestDecodeMessagePartUpdatedEvent_ToolPart(t *testing.T) {
	raw := `{
		"type": "message.part.updated",
		"properties": {
			"part": {
				"id": "part-tool",
				"sessionID": "session-456",
				"messageID": "msg-789",
				"type": "tool",
				"tool": "bash",
				"callID": "call-123",
				"state": {"status": "running"}
			}
		}
	}`

	var ev MessagePartUpdatedEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if ev.Properties.Part.PartType != "tool" {
		t.Errorf("expected part type 'tool', got %q", ev.Properties.Part.PartType)
	}
	if ev.Properties.Part.Tool != "bash" {
		t.Errorf("expected tool 'bash', got %q", ev.Properties.Part.Tool)
	}
}

func TestDecodeMessageUpdatedEvent(t *testing.T) {
	raw := `{
		"type": "message.updated",
		"properties": {
			"info": {
				"id": "msg-123",
				"sessionID": "session-456",
				"role": "assistant",
				"modelID": "claude-3",
				"providerID": "anthropic",
				"agent": "build",
				"cost": 0.001,
				"tokens": {"input": 100, "output": 50, "reasoning": 10},
				"time": {"created": 1234567890.123}
			}
		}
	}`

	var ev MessageUpdatedEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if ev.Type != "message.updated" {
		t.Errorf("expected type 'message.updated', got %q", ev.Type)
	}
	if ev.Properties.Info.ID != "msg-123" {
		t.Errorf("expected message ID 'msg-123', got %q", ev.Properties.Info.ID)
	}
	if ev.Properties.Info.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", ev.Properties.Info.Role)
	}
	if ev.Properties.Info.Tokens.Input != 100 {
		t.Errorf("expected 100 input tokens, got %d", ev.Properties.Info.Tokens.Input)
	}
}

func TestDecodeMessageUpdatedEvent_Completed(t *testing.T) {
	raw := `{
		"type": "message.updated",
		"properties": {
			"info": {
				"id": "msg-123",
				"sessionID": "session-456",
				"role": "assistant",
				"time": {"created": 1234567890.123, "completed": 1234567890.456}
			}
		}
	}`

	var ev MessageUpdatedEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if ev.Properties.Info.Time.Completed == nil {
		t.Error("expected time.completed to be set")
	}
}

func TestParseEvent_DispatchesCorrectly(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		data      string
		wantType  string
	}{
		{
			name:      "message.part.updated",
			eventType: "message.part.updated",
			data:      `{"type":"message.part.updated","properties":{"part":{"id":"p1","type":"text","text":"hi"}}}`,
			wantType:  "message.part.updated",
		},
		{
			name:      "message.updated",
			eventType: "message.updated",
			data:      `{"type":"message.updated","properties":{"info":{"id":"m1","role":"assistant"}}}`,
			wantType:  "message.updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := SSEEvent{Event: tt.eventType, Data: []byte(tt.data)}
			parsed, err := ParseEvent(ev)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			switch p := parsed.(type) {
			case *MessagePartUpdatedEvent:
				if p.Type != tt.wantType {
					t.Errorf("expected type %q, got %q", tt.wantType, p.Type)
				}
			case *MessageUpdatedEvent:
				if p.Type != tt.wantType {
					t.Errorf("expected type %q, got %q", tt.wantType, p.Type)
				}
			default:
				t.Errorf("unexpected parsed type: %T", parsed)
			}
		})
	}
}

func TestToStreamUpdate_WithDelta_ReturnsOpAppend(t *testing.T) {
	ev := MessagePartUpdatedEvent{
		Type: "message.part.updated",
		Properties: struct {
			Part  Part   `json:"part"`
			Delta string `json:"delta,omitempty"`
		}{
			Part: Part{
				ID:        "part-1",
				MessageID: "msg-1",
				PartType:  "text",
				Text:      "Hello world",
			},
			Delta: "world",
		},
	}

	update := ev.ToStreamUpdate()

	if update.Op != OpAppend {
		t.Errorf("expected OpAppend, got %v", update.Op)
	}
	if update.Text != "world" {
		t.Errorf("expected text 'world', got %q", update.Text)
	}
	if update.PartID != "part-1" {
		t.Errorf("expected partID 'part-1', got %q", update.PartID)
	}
	if update.MessageID != "msg-1" {
		t.Errorf("expected messageID 'msg-1', got %q", update.MessageID)
	}
	if update.Kind != PartKindText {
		t.Errorf("expected kind PartKindText, got %v", update.Kind)
	}
}

func TestToStreamUpdate_NoDelta_ReturnsOpSet(t *testing.T) {
	ev := MessagePartUpdatedEvent{
		Type: "message.part.updated",
		Properties: struct {
			Part  Part   `json:"part"`
			Delta string `json:"delta,omitempty"`
		}{
			Part: Part{
				ID:        "part-2",
				MessageID: "msg-2",
				PartType:  "text",
				Text:      "Full content here",
			},
		},
	}

	update := ev.ToStreamUpdate()

	if update.Op != OpSet {
		t.Errorf("expected OpSet, got %v", update.Op)
	}
	if update.Text != "Full content here" {
		t.Errorf("expected text 'Full content here', got %q", update.Text)
	}
}

func TestToStreamUpdate_ReasoningPart(t *testing.T) {
	ev := MessagePartUpdatedEvent{
		Type: "message.part.updated",
		Properties: struct {
			Part  Part   `json:"part"`
			Delta string `json:"delta,omitempty"`
		}{
			Part: Part{
				ID:        "part-3",
				MessageID: "msg-3",
				PartType:  "reasoning",
				Text:      "thinking...",
			},
		},
	}

	update := ev.ToStreamUpdate()

	if update.Kind != PartKindReasoning {
		t.Errorf("expected kind PartKindReasoning, got %v", update.Kind)
	}
}

func TestToStreamUpdate_ToolPart(t *testing.T) {
	ev := MessagePartUpdatedEvent{
		Type: "message.part.updated",
		Properties: struct {
			Part  Part   `json:"part"`
			Delta string `json:"delta,omitempty"`
		}{
			Part: Part{
				ID:        "part-4",
				MessageID: "msg-4",
				PartType:  "tool",
				Tool:      "bash",
			},
		},
	}

	update := ev.ToStreamUpdate()

	if update.Kind != PartKindTool {
		t.Errorf("expected kind PartKindTool, got %v", update.Kind)
	}
}

func TestToStreamUpdate_Complete(t *testing.T) {
	endTime := 1234567890.456
	ev := MessagePartUpdatedEvent{
		Type: "message.part.updated",
		Properties: struct {
			Part  Part   `json:"part"`
			Delta string `json:"delta,omitempty"`
		}{
			Part: Part{
				ID:        "part-5",
				MessageID: "msg-5",
				PartType:  "text",
				Text:      "done",
				Time:      PartTime{Start: 1234567890.123, End: &endTime},
			},
		},
	}

	update := ev.ToStreamUpdate()

	if !update.Complete {
		t.Error("expected Complete=true")
	}
}
