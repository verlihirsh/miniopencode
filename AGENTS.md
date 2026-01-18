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
# Build to bin directory
go build -o bin/miniopencode ./cmd/miniopencode

# Or build to current directory
go build -o miniopencode ./cmd/miniopencode
```

### Run
```bash
# TUI (default)
./bin/miniopencode

# Headless legacy proxy
./bin/miniopencode --headless

# Custom config
./bin/miniopencode --config ~/.config/miniopencode.yaml

# With mode flag
./bin/miniopencode --mode input
./bin/miniopencode --mode output
./bin/miniopencode --mode full

# With server flags
./bin/miniopencode --host 127.0.0.1 --port 4096

# With session flags
./bin/miniopencode --session daily
./bin/miniopencode --session ses_xxxxx

# With defaults flags
./bin/miniopencode --agent build --provider anthropic --model claude-3-5-sonnet

# With debug logging
./bin/miniopencode --log /tmp/debug.log
# Or use DEBUG env var for default path
DEBUG=1 ./bin/miniopencode
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
  - Mode is set at startup via config/flag, NOT switchable at runtime
  - Can run multiple instances in different modes connected to same session
  - Example: Terminal 1 with `--mode input`, Terminal 2 with `--mode output`
- Input: single-line only; `Enter` sends.
- Keybindings:
  - Quit: `Ctrl+C`, Help: `?`
  - Resize: `Ctrl+W` then `+`/`-`/`=` adjusts input height
  - Scroll: Arrow keys, `Ctrl+U/D` (half page), `Home/End` (top/bottom)
  - Send: `Enter`
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

## Project Structure
- Root: contains `main.go` (old legacy proxy code, kept for reference)
- `cmd/miniopencode/main.go`: actual entrypoint with flags
- `bin/`: build output directory
- `docs/`: design docs (plan.md, sse-event-spec.md)
- `internal/config`: YAML/CLI config loader, defaults
- `internal/proxy`: headless stdin/stdout proxy
- `internal/client`: HTTP + SSE client (sessions, prompt_async, SSE reader)
- `internal/session`: daily resolver (token/message limits)
- `internal/tui`: complete TUI implementation
  - `app.go`, `program.go`: TUI app structure
  - `model.go`, `events.go`, `event.go`: bubbletea model/events
  - `stream.go`, `send.go`: SSE streaming and prompt sending
  - `render.go`, `styles.go`, `markdown.go`: rendering and theming
  - `transcript.go`: message history
  - `truncate.go`: output truncation
  - `keymap.go`: keyboard bindings
  - `ttycheck.go`: TTY detection
  - `*_test.go`: comprehensive test coverage

---

## Dependencies
Key external libraries (see go.mod):
- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/bubbles`: TUI components
- `github.com/charmbracelet/glamour`: markdown rendering
- `github.com/charmbracelet/lipgloss`: styling
- `github.com/tmaxmax/go-sse`: SSE client
- `gopkg.in/yaml.v3`: config parsing

---

## Coding Guidelines
- go fmt, go test ./... before commit
- Imports: stdlib → external → internal; keep clean
- Errors: wrap with context; never ignore
- Avoid global state; prefer explicit config/structs
- Write tests for new features (see existing *_test.go files)

---

## Build/Release Checklist
- go test ./...
- go mod tidy
- go build -o bin/miniopencode ./cmd/miniopencode
- Update `miniopencode.example.yaml` if config changes
- Keep AGENTS.md in sync (commands, config, keybindings)
- Test both TUI and headless modes
