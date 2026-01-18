# SSE Event Specification

Internal reference documenting opencode server SSE events for miniopencode TUI.

Source: `GET http://127.0.0.1:4096/doc` (OpenAPI 3.1.1)

---

## SSE Endpoints

| Endpoint | Scope | Use Case |
|----------|-------|----------|
| `/event` | Per-project | Session/message events for current project |
| `/global/event` | Global | Cross-project events (installation, server lifecycle) |

**Headers:**
- `Accept: text/event-stream`
- `Cache-Control: no-cache`

---

## Message Streaming Events

### `message.updated`

Fired when a message (user or assistant) is created or updated.

```json
{
  "type": "message.updated",
  "properties": {
    "info": { /* Message object */ }
  }
}
```

**Message (AssistantMessage):**
```json
{
  "id": "string",
  "sessionID": "string",
  "role": "assistant",
  "time": {
    "created": 1234567890.123,
    "completed": 1234567890.456  // optional - present when done
  },
  "modelID": "string",
  "providerID": "string",
  "agent": "string",
  "cost": 0.001,
  "tokens": {
    "input": 100,
    "output": 50,
    "reasoning": 0,
    "cache": { "read": 0, "write": 0 }
  },
  "error": { /* optional error object */ }
}
```

**Use:** Track message lifecycle (created → completed), token costs, errors.

---

### `message.part.updated`

Fired when a message part (text chunk, tool call, reasoning) is created or updated.

```json
{
  "type": "message.part.updated",
  "properties": {
    "part": { /* Part object - full current state */ },
    "delta": "optional incremental text"
  }
}
```

**CRITICAL:** The `delta` field determines update semantics:
- `delta` **present and non-empty**: Use delta for append (OpAppend)
- `delta` **absent or empty**: Use `part.text` as full content (OpSet)

---

## Part Types

All parts share common fields:
```json
{
  "id": "string",
  "sessionID": "string", 
  "messageID": "string",
  "type": "text|reasoning|tool|...",
  "time": {
    "start": 1234567890.123,
    "end": 1234567890.456  // optional - present when part is complete
  }
}
```

### TextPart (`type: "text"`)

Main assistant response text.

```json
{
  "type": "text",
  "text": "full accumulated text content",
  "synthetic": false,
  "ignored": false
}
```

### ReasoningPart (`type: "reasoning"`)

Chain-of-thought / thinking content.

```json
{
  "type": "reasoning", 
  "text": "thinking process text"
}
```

### ToolPart (`type: "tool"`)

Tool/function call invocation.

```json
{
  "type": "tool",
  "callID": "string",
  "tool": "tool_name",
  "state": { /* ToolState */ }
}
```

### Other Part Types

- `subtask` - Subtask delegation
- `file` - File attachment
- `step-start`, `step-finish` - Step boundaries
- `snapshot`, `patch` - Code diff parts
- `agent` - Agent-specific part
- `retry`, `compaction` - Internal lifecycle

---

## Rendering Rules

### Part Kind Mapping

| Server `part.type` | TUI ChunkKind |
|--------------------|---------------|
| `text` | `ChunkAnswer` |
| `reasoning` | `ChunkThinking` |
| `tool` | `ChunkTool` |
| other | `ChunkRaw` (ignore or log) |

### Update Semantics

```
On message.part.updated:
  1. Extract messageID from properties.part.messageID
  2. Extract partID from properties.part.id
  3. Extract partType from properties.part.type
  4. Check for delta:
     - If properties.delta != "": 
         op = OpAppend, text = delta
     - Else:
         op = OpSet, text = properties.part.text
  5. Check completion:
     - complete = (properties.part.time.end != nil)
```

### Completion Detection

| Level | Complete When |
|-------|---------------|
| Part | `time.end` present |
| Message | `time.completed` present |

---

## Example Event Sequence

```
# User sends prompt
→ POST /session/{id}/prompt_async

# Server streams response
← SSE: message.updated (role=assistant, no time.completed)
← SSE: message.part.updated (type=text, delta="Hello")
← SSE: message.part.updated (type=text, delta=" world")
← SSE: message.part.updated (type=text, delta="!", time.end set)
← SSE: message.updated (time.completed set, tokens populated)
```

---

## Anti-Patterns (Current Issues)

1. **Heuristic categorization** - Current code guesses chunk type from field presence
2. **TrimPrefix deduplication** - Assumes delta but receives full content
3. **No ID tracking** - Loses part identity, causes duplicates on OpSet

---

## Implementation Checklist

- [ ] Parse `type` field from event JSON to dispatch
- [ ] Define typed structs for `message.updated`, `message.part.updated`
- [ ] Define typed structs for `TextPart`, `ReasoningPart`, `ToolPart`
- [ ] Use `delta` presence to choose OpAppend vs OpSet
- [ ] Track parts by `partID` for idempotent updates
- [ ] Detect completion via `time.end` / `time.completed`
