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
	Kind ChunkKind
	Text string
}

func categorize(raw map[string]any) ChunkKind {
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

func makeChunk(data []byte) Chunk {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return Chunk{Kind: ChunkRaw, Text: string(data)}
	}
	kind := categorize(m)
	return Chunk{Kind: kind, Text: string(data)}
}
