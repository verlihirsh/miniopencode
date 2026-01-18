plan.md — miniopencode TUI recovery plan (SSE + rendering)

Goals
1. Correct, stable TUI layout using canonical Bubble Tea/Bubbles patterns:
   - viewport for transcript output
   - textinput for input (input exists only in input pane; echo appears in output after send)
   - spinner placeholder inside viewport while waiting for assistant output
   - proper scroll UX (pager-style: header/footer with scroll percent, mouse wheel)
2. Correct SSE handling using a proper SSE client library and schema-driven event decoding (no heuristics).
3. Fix duplication by applying server-defined update semantics (delta vs set/full update).
4. Implement “symbol-by-symbol” display correctly: buffer incoming text and typewriter it via Bubble Tea ticks (no goroutine sleeps, no tearing).
---
Phase 1 — Establish ground truth from server docs [DONE]
1.1 Read opencode server documentation [DONE]
- Fetch and review: GET http://127.0.0.1:4096/doc
- Extract:
  - SSE endpoint(s) (/event vs /global/event) and required headers
  - exact event types and payload schema (message.updated, message.part.updated, etc.)
  - semantics: whether payload contains delta, full content, or both; and how to detect completion
  - message/part identity fields (messageID, partID, timestamps, etc.)
- Deliverable:
  - A small internal spec summary: event types → struct fields → intended rendering behavior.
  - Created: docs/sse-event-spec.md
1.2 Add safe logging for raw SSE payloads (file-only) [DONE]
- Ensure all debug output goes to log file (never stdout/stderr during TUI).
- Add logging that records (bounded):
  - event name, byte size, and first N bytes of JSON
- Purpose: verify runtime behavior matches /doc without corrupting UI.
- Implementation: Added logging in client.go flush() with 512 byte preview limit.
---
Phase 2 — Replace SSE implementation (stop homegrown parser) [DONE]
2.1 Choose SSE client library [DONE]
Pick one:
- Preferred: github.com/r3labs/sse/v2 (widely used, simple API, reconnect support)
- Alternative: github.com/tmaxmax/go-sse (lower-level)
Decision criteria:
- correct SSE framing (multi-line data: handling)
- easy cancellation via context
- optional reconnect semantics (nice-to-have)
Decision: Selected github.com/tmaxmax/go-sse - actively maintained, proper context handling, sse.Read iterator API.
2.2 Implement SSE client wrapper [DONE]
Create internal/client/sse.go:
- Connect to documented SSE endpoint
- Expose:
  - chan Event { Name string; Data []byte }
  - chan error
- Handle:
  - context cancellation
  - server disconnect errors
  - keepalive comments safely
- Remove old bufio.Scanner SSE parsing path from client.go.
Deliverable:
- SSE stream no longer depends on bufio.Scanner.
- Implementation: internal/client/sse.go using tmaxmax/go-sse with sse.Read iterator.
- Tests: internal/client/sse_test.go with multi-line data and context cancellation tests.
---
Phase 3 — Schema-driven event decoding ("real answer") [DONE]
3.1 Define typed event structs based on /doc [DONE]
Create internal/client/events.go with minimal structs needed:
- MessageUpdatedEvent
- MessagePartUpdatedEvent
- any "done/complete" event (if documented)
Avoid map[string]any and heuristic extraction.
Implementation: Created typed structs with ParseEvent dispatch function.
3.2 Normalize to internal UI update messages [DONE]
Define internal update type used by TUI:
type UpdateOp int
const (
  OpAppend UpdateOp = iota
  OpSet
)
type StreamUpdate struct {
  MessageID string
  PartID    string
  Kind      ChunkKind // answer/thinking/tool
  Op        UpdateOp  // Set or Append per docs
  Text      string
  Complete  bool
}
Rules:
- If server delivers delta: OpAppend
- If server delivers full content: OpSet
- Never "TrimPrefix guess" unless explicitly defined by docs.
Deliverable:
- Deterministic updates that eliminate duplicate output.
- Implementation: StreamUpdate type with ToStreamUpdate() method on MessagePartUpdatedEvent.
- Tests: events_test.go with comprehensive coverage.
---
Phase 4 — Transcript model that supports Set/Append by IDs [DONE]
4.1 Replace current transcript append-only behavior [DONE]
Refactor transcript state:
- Messages indexed by messageID (or fallback to last pending assistant if doc says IDs arrive late)
- Parts indexed by partID
Data model:
type TranscriptMessage struct {
  ID string
  Role Role // user/assistant
  Pending bool
  Parts map[string]*TranscriptPart
  PartOrder []string
}
type TranscriptPart struct {
  ID string
  Kind ChunkKind
  Text string
}
Operations:
- AddUserEcho(text)
- EnsurePendingAssistant(messageID)
- ApplyUpdate(StreamUpdate) where:
  - OpSet replaces part text
  - OpAppend appends to part text
  - first non-empty update clears Pending
Deliverable:
- No more duplication when server sends "updated answer".
- Implementation: transcript.go with ApplyUpdate, EnsurePendingAssistant, Pending field.
- Tests: transcript_test.go with OpAppend, OpSet, Pending, and no-duplication tests.
---
Phase 5 — Pure Bubbles UI layout (pager-style)
5.1 Adopt pager example layout math
Structure View like Bubble Tea pager example:
- Header: status line (session/mode/server + sending indicator)
- Output: viewport (with borders if desired, but consistently applied)
- Footer: scroll percent viewport.ScrollPercent()
- Input box: textinput in full mode only
WindowSize handling:
- compute headerHeight and footerHeight via lipgloss.Height(view)
- set viewport.Width = msg.Width (minus borders/padding if used)
- set viewport.Height = msg.Height - headerHeight - footerHeight - inputHeight (in full mode)
- enable:
  - tea.WithAltScreen()
  - tea.WithMouseCellMotion()
Deliverable:
- No overlap. No output outside panes. Stable resizing.
5.2 Make viewport “follow” behavior correct
- Follow mode enabled when at bottom.
- User scroll disables follow.
- End key restores follow + goto bottom.
Deliverable:
- Smooth UX with long outputs.
---
Phase 6 — Spinner placeholder inside viewport (Option 2)
6.1 Pending assistant message renders spinner inside viewport content
When user sends:
- echo user message into transcript output
- create a pending assistant message with no parts, Pending=true
Render:
- if latest assistant message Pending, show:
  - Assistant:
  - spinner.View() on next line (inside viewport content)
6.2 Ensure spinner animates without SSE activity
Wire spinner.TickMsg to:
- update spinner model
- trigger viewport content re-render (only when there is a pending assistant)
Deliverable:
- spinner visibly animates in output viewport while waiting.
---
Phase 7 — Typewriter (“symbol-by-symbol”) without hacks
7.1 Buffer incoming updates
Add model fields:
- pendingTextBuf []rune (or per-part buffers)
- typingActive bool
On receiving StreamUpdate:
- for OpAppend: enqueue runes in buffer rather than applying immediately
- for OpSet: replace visible text and reset buffer (or compute needed animation per docs)
7.2 Typewriter tick command
Use tea.Tick at ~16ms–33ms:
- each tick, if buffer non-empty:
  - pop 1 rune
  - apply append to transcript part text
  - re-render viewport content
- stop ticking when buffer empty
Deliverable:
- true symbol-by-symbol output, controlled and smooth, not dependent on server chunking.
---
Phase 8 — Tests and regression harness
8.1 Fixture-based decoding tests
- Store a small captured SSE payload sequence (from logs) as testdata JSON lines.
- Unit test: decode → StreamUpdate → ApplyUpdate and assert final transcript text.
8.2 UI behavior sanity checks
- Ensure no panics on resize
- Ensure follow-mode toggles
- Ensure pending spinner appears/disappears correctly
Deliverable:
- Prevent future breakage.
---
Acceptance criteria
- No visible overlap; output never renders outside its border region.
- Input exists only in input pane; after send, user text is echoed in output once.
- While waiting, assistant placeholder shows animated spinner inside viewport.
- SSE decoding is schema-driven, no heuristic map parsing.
- No duplicate assistant text when server sends “full updated content”.
- Scroll percent footer works like pager example; mouse wheel scroll works.
- Symbol-by-symbol rendering works via buffer + ticks, without goroutine sleeps.
