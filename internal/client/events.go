package client

import (
	"encoding/json"
	"fmt"
)

type PartTime struct {
	Start float64  `json:"start"`
	End   *float64 `json:"end,omitempty"`
}

type MessageTime struct {
	Created   float64  `json:"created"`
	Completed *float64 `json:"completed,omitempty"`
}

type Tokens struct {
	Input     int `json:"input"`
	Output    int `json:"output"`
	Reasoning int `json:"reasoning"`
}

type Part struct {
	ID        string   `json:"id"`
	SessionID string   `json:"sessionID"`
	MessageID string   `json:"messageID"`
	PartType  string   `json:"type"`
	Text      string   `json:"text,omitempty"`
	Tool      string   `json:"tool,omitempty"`
	CallID    string   `json:"callID,omitempty"`
	Time      PartTime `json:"time,omitempty"`
}

type MessageInfo struct {
	ID         string       `json:"id"`
	SessionID  string       `json:"sessionID"`
	Role       string       `json:"role"`
	ModelID    string       `json:"modelID,omitempty"`
	ProviderID string       `json:"providerID,omitempty"`
	Agent      string       `json:"agent,omitempty"`
	Cost       float64      `json:"cost,omitempty"`
	Tokens     *Tokens      `json:"tokens,omitempty"`
	Time       *MessageTime `json:"time,omitempty"`
}

type MessagePartUpdatedEvent struct {
	Type       string `json:"type"`
	Properties struct {
		Part  Part   `json:"part"`
		Delta string `json:"delta,omitempty"`
	} `json:"properties"`
}

type MessageUpdatedEvent struct {
	Type       string `json:"type"`
	Properties struct {
		Info MessageInfo `json:"info"`
	} `json:"properties"`
}

type GenericEvent struct {
	Type string `json:"type"`
}

func ParseEvent(ev SSEEvent) (any, error) {
	if len(ev.Data) == 0 {
		return nil, fmt.Errorf("empty event data")
	}

	var generic GenericEvent
	if err := json.Unmarshal(ev.Data, &generic); err != nil {
		return nil, fmt.Errorf("parse event type: %w", err)
	}

	switch generic.Type {
	case "message.part.updated":
		var parsed MessagePartUpdatedEvent
		if err := json.Unmarshal(ev.Data, &parsed); err != nil {
			return nil, fmt.Errorf("parse message.part.updated: %w", err)
		}
		return &parsed, nil

	case "message.updated":
		var parsed MessageUpdatedEvent
		if err := json.Unmarshal(ev.Data, &parsed); err != nil {
			return nil, fmt.Errorf("parse message.updated: %w", err)
		}
		return &parsed, nil

	default:
		return &generic, nil
	}
}

func (p *Part) IsComplete() bool {
	return p.Time.End != nil
}

func (m *MessageInfo) IsComplete() bool {
	return m.Time != nil && m.Time.Completed != nil
}

type UpdateOp int

const (
	OpAppend UpdateOp = iota
	OpSet
)

type PartKind string

const (
	PartKindText      PartKind = "text"
	PartKindReasoning PartKind = "reasoning"
	PartKindTool      PartKind = "tool"
	PartKindOther     PartKind = "other"
)

type StreamUpdate struct {
	MessageID string
	PartID    string
	Kind      PartKind
	Op        UpdateOp
	Text      string
	Complete  bool
}

func (ev *MessagePartUpdatedEvent) ToStreamUpdate() StreamUpdate {
	part := ev.Properties.Part
	delta := ev.Properties.Delta

	var op UpdateOp
	var text string
	if delta != "" {
		op = OpAppend
		text = delta
	} else {
		op = OpSet
		text = part.Text
	}

	var kind PartKind
	switch part.PartType {
	case "text":
		kind = PartKindText
	case "reasoning":
		kind = PartKindReasoning
	case "tool":
		kind = PartKindTool
	default:
		kind = PartKindOther
	}

	return StreamUpdate{
		MessageID: part.MessageID,
		PartID:    part.ID,
		Kind:      kind,
		Op:        op,
		Text:      text,
		Complete:  part.IsComplete(),
	}
}
