# AGENTS.md - AI Coding Agent Guidelines

## Project Overview

**miniopencode** is a terminal opencode client with headless proxy + interactive TUI.
- Headless: stdin JSON → stdout JSON (legacy proxy)
- TUI: chat interface with modes (input/output/full), SSE streaming, markdown output
- Communicates with opencode server via HTTP/SSE

---

## Build, Test, Run

### Build
```bash
go build -o miniopencode ./cmd/miniopencode
```

### Run
```bash
# TUI (default)
./miniopencode

# Headless legacy proxy
./miniopencode --headless

# Custom config
./miniopencode --config ~/.config/miniopencode.yaml
```

### Test
```bash
go test ./...
```

---

## Config
- Default path: `~/.config/miniopencode.yaml`
- Example: `miniopencode.example.yaml`
- Merge: defaults → YAML → CLI overrides

Key fields:
- `server.host`, `server.port`
- `session.default_session`: "" | session ID | `daily`
- `session.daily_title_format`: default `2006-01-02-daily-%d`
- `session.daily_max_tokens`, `session.daily_max_messages`
- `defaults.agent`, `defaults.provider_id`, `defaults.model_id`
- `ui.mode`: `input|output|full`
- `ui.multiline`: true enables Ctrl+Enter send; Enter inserts newline
- `ui.show_thinking`, `ui.show_tools`, `ui.wrap`, `ui.input_height`, `ui.max_output_lines`, `ui.theme`
- `theme`: border style/colors for output/input/status/thinking/tool/answer

---

## Code Structure
- `cmd/miniopencode`: entrypoint, flags `--headless`, `--config`
- `internal/config`: YAML/CLI loader, defaults
- `internal/proxy`: headless stdin/stdout proxy
- `internal/client`: HTTP + SSE client (sessions, prompt_async, SSE reader)
- `internal/session`: daily resolver (token/message limits)
- `internal/tui`: keymap, model, SSE chunking, markdown render, truncation

---

## TUI Behavior
- Modes: `input` (only input), `output` (only output), `full` (output+input)
- Multiline toggle: `Ctrl+M`. Single line: Enter sends. Multiline: Enter newline; Ctrl+Enter/Ctrl+J sends.
- Keybindings (vim-like):
  - Quit `q`/`ctrl+c`, Help `?`
  - Modes: `g i`, `g o`, `g f`
  - Multiline: `ctrl+m`
  - Resize: `ctrl+w` then `+`/`-`/`=` adjusts input height
  - Scroll: `j/k`, `u/d` half page, `f/b` page, `g/G` top/bottom
  - Toggles: `t t` thinking, `t o` tools
  - Send: Enter (single), Ctrl+Enter/Ctrl+J (multiline)
- Output: SSE chunks categorized (heuristic) into thinking/tool/answer; rendered as markdown; truncated to max lines

---

## Session Handling
- Default session resolution:
  - If `default_session` is an ID: use it or create if missing
  - If `daily`: find `YYYY-MM-DD-daily-<part>`; reuse latest if under limits; else create next part
  - Limits default: tokens 250000, messages 4000

---

## API Notes (opencode server)
- Health: `GET /global/health`
- Sessions: `GET/POST /session`, `GET /session/{id}`, `GET /session/{id}/message`
- Prompt: `POST /session/{id}/prompt_async` with `parts`, optional `model {providerID, modelID}`, `agent`, `system`, `variant`
- SSE: `GET /event` (project) or `/global/event`; events include `message.updated`, `message.part.updated`, `session.*`
- Messages: tokens metadata `tokens.input|output|reasoning`, `cost`, `modelID`, `providerID`, `agent`

---

## Headless Commands (legacy)
- `{"type":"health"}`
- `{"type":"session.create","payload":{"title":"..."}}`
- `{"type":"session.list"}`
- `{"type":"session.select","payload":{"id":"..."}}`
- `{"type":"prompt","payload":{"text":"...","provider_id":"...","model_id":"..."}}`
- `{"type":"sse.start"}` / `{"type":"sse.stop"}`

---

## Coding Guidelines
- go fmt, go test ./... before commit
- Imports: stdlib → external → internal; keep clean
- Errors: wrap with context; never ignore
- Avoid global state; prefer explicit config/structs

---

## Build/Release Checklist
- go test ./...
- go mod tidy
- Update `miniopencode.example.yaml` if config changes
- Keep AGENTS.md in sync (commands, config, keybindings)
