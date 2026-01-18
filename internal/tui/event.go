package tui

import (
	"encoding/json"
	"strings"
)

type ChunkKind string

const (
	ChunkThinking ChunkKind = "thinking"
	ChunkTool     ChunkKind = "tool"
	ChunkAnswer   ChunkKind = "answer"
	ChunkRaw      ChunkKind = "raw"
)

type Chunk struct {
	Kind      ChunkKind
	Text      string
	PartID    string
	MessageID string
	Complete  bool
}

type MessagePartEvent struct {
	Type    string `json:"type"`
	Delta   string `json:"delta"`
	Content string `json:"content"`
	Text    string `json:"text"`
}

func categorize(raw map[string]any) ChunkKind {
	if props, ok := raw["properties"].(map[string]any); ok {
		if part, ok := props["part"].(map[string]any); ok {
			if partType, ok := part["type"].(string); ok {
				switch partType {
				case "tool", "tool-use", "function":
					return ChunkTool
				case "thinking", "reasoning":
					return ChunkThinking
				case "text":
					return ChunkAnswer
				}
			}
		}
	}

	if t, ok := raw["type"].(string); ok {
		lt := strings.ToLower(t)
		switch {
		case strings.Contains(lt, "tool") || strings.Contains(lt, "function"):
			return ChunkTool
		case strings.Contains(lt, "think") || strings.Contains(lt, "reason"):
			return ChunkThinking
		case strings.Contains(lt, "message") || strings.Contains(lt, "content") || strings.Contains(lt, "assistant"):
			return ChunkAnswer
		}
	}
	if _, ok := raw["tool"].(string); ok {
		return ChunkTool
	}
	if _, ok := raw["tool_calls"].([]any); ok {
		return ChunkTool
	}
	if _, ok := raw["thinking"].(string); ok {
		return ChunkThinking
	}
	if _, ok := raw["content"].(string); ok {
		return ChunkAnswer
	}
	return ChunkAnswer
}

func extractText(m map[string]any) string {
	if delta, ok := m["delta"].(string); ok && delta != "" {
		return delta
	}
	if content, ok := m["content"].(string); ok && content != "" {
		return content
	}
	if text, ok := m["text"].(string); ok && text != "" {
		return text
	}
	if msg, ok := m["message"].(string); ok && msg != "" {
		return msg
	}

	if props, ok := m["properties"].(map[string]any); ok {
		if delta, ok := props["delta"].(string); ok && delta != "" {
			return delta
		}
		if part, ok := props["part"].(map[string]any); ok {
			if text, ok := part["text"].(string); ok && text != "" {
				return text
			}
			if content, ok := part["content"].(string); ok && content != "" {
				return content
			}
		}
	}

	return ""
}

func extractPartID(m map[string]any) string {
	if props, ok := m["properties"].(map[string]any); ok {
		if part, ok := props["part"].(map[string]any); ok {
			if id, ok := part["id"].(string); ok {
				return id
			}
		}
	}
	return ""
}

func extractMessageID(m map[string]any) string {
	if props, ok := m["properties"].(map[string]any); ok {
		if part, ok := props["part"].(map[string]any); ok {
			if id, ok := part["messageID"].(string); ok {
				return id
			}
		}
	}
	return ""
}

func isComplete(m map[string]any) bool {
	if props, ok := m["properties"].(map[string]any); ok {
		if part, ok := props["part"].(map[string]any); ok {
			if timeMap, ok := part["time"].(map[string]any); ok {
				_, hasEnd := timeMap["end"]
				return hasEnd
			}
		}
	}
	return false
}

func makeChunk(data []byte) Chunk {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return Chunk{Kind: ChunkRaw, Text: string(data)}
	}

	text := extractText(m)
	if text == "" {
		return Chunk{Kind: ChunkRaw, Text: ""}
	}

	kind := categorize(m)
	partID := extractPartID(m)
	messageID := extractMessageID(m)
	complete := isComplete(m)
	return Chunk{Kind: kind, Text: text, PartID: partID, MessageID: messageID, Complete: complete}
}
