# miniopencode 🖥️ - Terminal OpenCode Client

<p align="center">
  <strong>A terminal client for OpenCode with headless proxy + interactive TUI</strong><br>
  Stream AI responses, manage sessions, and stay in your terminal workflow
</p>

<p align="center">
  <a href="#installation"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go" alt="Go 1.25+"></a>
  <a href="#features"><img src="https://img.shields.io/badge/SSE-Streaming-blueviolet" alt="SSE Streaming"></a>
  <a href="#modes"><img src="https://img.shields.io/badge/Modes-TUI%20%7C%20Headless-orange" alt="Modes"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#features">Features</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#usage">Usage</a> •
  <a href="#development">Development</a>
</p>

---

## Why miniopencode?

A **native terminal experience** for OpenCode with real-time streaming and flexible deployment options.

**miniopencode** gives you:
- **🚀 Two modes**: Interactive TUI for humans, headless JSON proxy for scripts
- **⚡ Real-time streaming**: Server-Sent Events (SSE) with proper message chunking
- **🎨 Beautiful output**: Markdown rendering with syntax highlighting in your terminal
- **📊 Session management**: Daily sessions with token/message limits
- **🔧 Single binary**: No external dependencies required

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Modes](#modes)
  - [TUI Mode](#tui-mode-default)
  - [Headless Mode](#headless-mode)
- [Features](#features)
- [Configuration](#configuration)
- [Session Management](#session-management)
- [Usage](#usage)
  - [TUI Keybindings](#tui-keybindings)
  - [Headless Commands](#headless-commands)
- [Development](#development)
- [Architecture](#architecture)

---

## Installation

### Build from Source

**Requirements**: Go 1.25.6 or higher

```bash
# Clone the repository
git clone <repository-url>
cd miniopencode

# Build to bin directory
go build -o bin/miniopencode ./cmd/miniopencode

# Or build to current directory
go build -o miniopencode ./cmd/miniopencode
```

---

## Quick Start

### 1. Start OpenCode Server

Ensure your OpenCode server is running at `http://127.0.0.1:4096` (or configure with flags/config file).

### 2. Launch miniopencode

```bash
# TUI mode (default) - Interactive chat interface
miniopencode

# Headless mode - JSON proxy for scripts
miniopencode --headless

# Custom configuration
miniopencode --config ~/.config/miniopencode.yaml

# Custom server address
miniopencode --host 127.0.0.1 --port 4096
```

### 3. Start Chatting (TUI Mode)

Type your message, press `Enter` to send. Watch AI responses stream in real-time with markdown rendering.

**Press `?` for help, `Ctrl+C` to quit.**

---

## Modes

### TUI Mode (Default)

Interactive terminal interface with **three display modes**:

| Mode | Description | View | Use Case |
|------|-------------|------|----------|
| **`input`** | Only shows input pane | Input only | Dedicated input terminal |
| **`output`** | Only shows output pane | Output only | Monitor AI responses |
| **`full`** | Shows both panes | Output + Input | Standard interactive use |

**Mode selection**: Set at startup via config file or `--mode` flag.  
**Multi-instance support**: Run multiple instances in different modes connected to the same session.

**Example workflow**:
```bash
# Terminal 1: Input only
miniopencode --mode input --session work

# Terminal 2: Output only (monitor responses)
miniopencode --mode output --session work
```

#### TUI Features

- **Real-time SSE streaming** with proper message chunking
- **Markdown rendering** with syntax highlighting
- **Scroll controls**: Arrow keys, `Ctrl+U/D` (half page), `Home/End`
- **Dynamic resizing**: `Ctrl+W` then `+`/`-`/`=` adjusts input height
- **Message categorization**: Thinking, tool calls, answers (color-coded)
- **Output truncation**: Configurable max lines to prevent memory bloat

### Headless Mode

JSON-based stdin/stdout proxy for **scripts, automation, CI/CD**:

```bash
miniopencode --headless
```

Send JSON commands via stdin, receive JSON responses via stdout. Perfect for:
- Shell scripts automating OpenCode interactions
- CI/CD pipelines running AI agents
- Testing frameworks validating AI outputs
- Integration with other tools (VSCode extensions, CLI wrappers)

---

## Features

### 🚀 Core Features

- **Dual-mode operation**: Interactive TUI or headless JSON proxy
- **Real-time SSE streaming**: Proper server-sent events with chunked message delivery
- **Session management**: Create, list, switch sessions; auto-create daily sessions
- **Message history**: Full transcript with role labels and timestamps
- **Token/cost tracking**: View input/output tokens, reasoning tokens, and costs per message

### 🎨 TUI Features

- **Three display modes**: `input`, `output`, `full`
- **Markdown rendering**: Syntax-highlighted code blocks, tables, lists
- **Smart categorization**: Thinking (yellow), Tools (cyan), Answers (white)
- **Smooth scrolling**: Viewport-based pager with scroll indicators
- **Dynamic resizing**: Adjust input pane height on the fly
- **Keyboard-driven**: Navigation without mouse required
- **Themeable**: Customize colors and border styles via config

### 🤖 AI Integration

- **Model selection**: Choose provider (Anthropic, OpenAI) and model per-request
- **Agent support**: Specify agent types (build, explore, oracle, etc.)
- **System prompts**: Custom system messages for specialized tasks
- **Variant support**: Use different model variants (extended thinking, etc.)

### 📊 Session Features

- **Daily sessions**: Auto-create sessions named `YYYY-MM-DD-daily-1`, `YYYY-MM-DD-daily-2`, etc.
- **Token limits**: Rotate to new daily session when limits exceeded (default: 250k tokens, 4000 messages)
- **Session reuse**: Find and resume existing daily sessions under limits
- **Manual sessions**: Create named sessions for long-term projects

---

## Configuration

Configuration is merged from **three sources** (in order of precedence):

1. **Defaults** (hardcoded)
2. **YAML config file** (optional)
3. **CLI flags** (highest priority)

### Config File Location

Default: `~/.config/miniopencode.yaml`

Override with `--config` flag:
```bash
miniopencode --config /path/to/config.yaml
```

### Example Configuration

See [`miniopencode.example.yaml`](miniopencode.example.yaml) for full reference:

```yaml
server:
  host: 127.0.0.1
  port: 4096

session:
  default_session: daily  # or "" (empty) or "ses_xxxxx" (specific ID)
  daily_title_format: "2006-01-02-daily-%d"
  daily_max_tokens: 250000
  daily_max_messages: 4000

defaults:
  agent: build
  provider_id: anthropic
  model_id: claude-3-5-sonnet

ui:
  mode: full  # input | output | full
  show_thinking: true
  show_tools: true
  wrap: true
  input_height: 6
  max_output_lines: 4000
  theme: default

theme:
  border_style: rounded
  output_border_color: "#89b4fa"
  input_border_color: "#a6e3a1"
  status_color: "#6c7086"
  thinking_color: "#f9e2af"
  tool_color: "#94e2d5"
  answer_color: "#cdd6f4"
```

### CLI Flags

```bash
# Server connection
--host STRING         OpenCode server host (default: 127.0.0.1)
--port INT            OpenCode server port (default: 4096)

# Session management
--session STRING      Session ID or "daily" (default: from config)
--daily-max-tokens INT     Daily session max tokens
--daily-max-messages INT   Daily session max messages

# Default model/agent
--agent STRING        Default agent type (default: build)
--provider STRING     Default provider ID (default: anthropic)
--model STRING        Default model ID (default: claude-3-5-sonnet)

# UI/display
--mode STRING              Display mode: input|output|full (default: full)
--show-thinking            Show thinking blocks
--show-tools               Show tool calls
--wrap                     Wrap text in output
--input-height INT         Input box height
--max-output-lines INT     Maximum output lines
--theme STRING             Theme name

# Mode selection
--headless            Run in headless JSON proxy mode

# Debugging
--log PATH            Write debug logs to file (or use DEBUG=1 env var)

# Other
--config PATH         Path to config file (default: ~/.config/miniopencode.yaml)
```

---

## Session Management

### Default Session Resolution

When you start miniopencode, it resolves the default session based on `default_session` config:

| Config Value | Behavior |
|--------------|----------|
| `""` (empty) | Use explicitly provided session ID, or create new session |
| `daily` | Find/create daily session (see below) |
| `ses_xxxxx` | Use specific session ID; create if missing |

### Daily Session Logic

When `default_session: daily`:

1. **Find existing daily session** matching today's date (`YYYY-MM-DD-daily-N`)
2. **Check limits**: If latest daily session is under token/message limits, reuse it
3. **Create new part**: If limits exceeded, create next part (`daily-2`, `daily-3`, etc.)

**Token limit default**: 250,000 tokens  
**Message limit default**: 4,000 messages

**Example**:
```
2026-01-18-daily-1  (150k tokens, 2000 messages) → reuse
2026-01-18-daily-2  (270k tokens, 4500 messages) → exceeded, create daily-3
```

### Manual Session Creation

Create named sessions for long-term projects:

```bash
# TUI mode: Use config or flag
miniopencode --session my-project

# Headless mode: Send JSON command
echo '{"type":"session.create","payload":{"title":"my-project"}}' | miniopencode --headless
```

---

## Usage

### TUI Keybindings

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `?` | Show help |
| `Ctrl+C` | Quit |
| `↑` / `↓` | Scroll output up/down |
| `Ctrl+U` / `Ctrl+D` | Scroll half page up/down |
| `Home` / `End` | Jump to top/bottom of output |
| `Ctrl+W` | Enter resize mode |
| `+` / `-` | Increase/decrease input height (in resize mode) |
| `=` | Reset input height to default (in resize mode) |

### Headless Commands

Send JSON commands via stdin, receive responses via stdout.

#### Available Commands

**Health Check**
```json
{"type":"health"}
```

**Session: Create**
```json
{"type":"session.create","payload":{"title":"my-session"}}
```

**Session: List**
```json
{"type":"session.list"}
```

**Session: Select**
```json
{"type":"session.select","payload":{"id":"ses_xxxxx"}}
```

**Prompt**
```json
{
  "type": "prompt",
  "payload": {
    "text": "Your prompt here",
    "provider_id": "anthropic",
    "model_id": "claude-3-5-sonnet"
  }
}
```

**SSE: Start/Stop**
```json
{"type":"sse.start"}
{"type":"sse.stop"}
```

#### Example: Shell Script

```bash
#!/bin/bash

# Start miniopencode in headless mode
exec 3< <(miniopencode --headless)
exec 4> >(miniopencode --headless)

# Create session
echo '{"type":"session.create","payload":{"title":"script-session"}}' >&4
read -r response <&3
echo "Session created: $response"

# Send prompt
echo '{"type":"prompt","payload":{"text":"Explain Go interfaces"}}' >&4
read -r response <&3
echo "Response: $response"

# Cleanup
exec 3<&-
exec 4>&-
```

---

## Development

### Prerequisites

- **Go 1.25.6+** (required for build)
- **OpenCode server** running at `http://127.0.0.1:4096`

### Project Structure

```
miniopencode/
├── cmd/miniopencode/     # Entrypoint (flags, mode selection)
├── internal/
│   ├── config/           # YAML/CLI config loader, defaults
│   ├── proxy/            # Headless stdin/stdout JSON proxy
│   ├── client/           # HTTP + SSE client (sessions, prompt_async)
│   ├── session/          # Daily session resolver (token/message limits)
│   └── tui/              # Complete TUI implementation
│       ├── app.go        # TUI app structure
│       ├── model.go      # Bubble Tea model
│       ├── events.go     # Bubble Tea events
│       ├── stream.go     # SSE streaming handler
│       ├── send.go       # Prompt sending logic
│       ├── render.go     # Rendering and styling
│       ├── markdown.go   # Markdown rendering with glamour
│       ├── transcript.go # Message history
│       ├── truncate.go   # Output truncation
│       └── keymap.go     # Keyboard bindings
├── docs/
│   ├── plan.md           # Development plan (TUI recovery)
│   └── sse-event-spec.md # SSE event specification
├── bin/                  # Build output directory
├── go.mod                # Go module definition
└── README.md             # This file
```

### Build & Test

```bash
# Build to bin directory
go build -o bin/miniopencode ./cmd/miniopencode

# Or build to current directory
go build -o miniopencode ./cmd/miniopencode

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/tui -run TestModelUpdate
```

### Running Development Build

```bash
# TUI mode
./bin/miniopencode

# Headless mode
./bin/miniopencode --headless

# With debug logging
DEBUG=1 ./bin/miniopencode --log /tmp/debug.log

# Custom config
./bin/miniopencode --config miniopencode.example.yaml

# Multiple modes simultaneously (different terminals)
./bin/miniopencode --mode input --session dev  # Terminal 1
./bin/miniopencode --mode output --session dev # Terminal 2
```

### Debugging

Enable debug logging to troubleshoot issues:

```bash
# Write logs to default temp path
DEBUG=1 ./bin/miniopencode

# Write logs to custom path
./bin/miniopencode --log /tmp/miniopencode.log

# View logs in real-time
tail -f /tmp/miniopencode.log
```

**What gets logged**:
- SSE event payloads (raw JSON, truncated to 512 bytes)
- HTTP requests/responses
- Session creation/selection
- Daily session resolution logic
- Message chunking and categorization
- Errors and warnings

### Key Dependencies

- **[charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)**: TUI framework (Elm architecture for Go)
- **[charmbracelet/bubbles](https://github.com/charmbracelet/bubbles)**: TUI components (viewport, textinput, spinner)
- **[charmbracelet/glamour](https://github.com/charmbracelet/glamour)**: Markdown rendering with syntax highlighting
- **[charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling (colors, borders)
- **[tmaxmax/go-sse](https://github.com/tmaxmax/go-sse)**: SSE client with proper framing
- **[gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)**: YAML config parsing

### Coding Guidelines

- **go fmt** before commit
- **go test ./...** must pass
- Keep imports organized: stdlib → external → internal
- Wrap errors with context; never ignore errors
- Avoid global state; prefer explicit config/structs
- Write tests for new features (see existing `*_test.go` files)

---

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────┐
│                    miniopencode                         │
│                                                         │
│  ┌──────────────┐          ┌──────────────┐           │
│  │  TUI Mode    │          │ Headless Mode│           │
│  │  (Bubble Tea)│          │ (JSON Proxy) │           │
│  └──────┬───────┘          └──────┬───────┘           │
│         │                         │                    │
│         └──────────┬──────────────┘                    │
│                    │                                   │
│         ┌──────────▼──────────┐                        │
│         │   HTTP + SSE Client │                        │
│         │  (internal/client)  │                        │
│         └──────────┬──────────┘                        │
│                    │                                   │
└────────────────────┼───────────────────────────────────┘
                     │
                     │ HTTP/SSE
                     │
          ┌──────────▼──────────┐
          │   OpenCode Server   │
          │  (127.0.0.1:4096)   │
          └─────────────────────┘
```

### SSE Event Handling

1. **Connect**: Establish SSE connection to `/event` (project) or `/global/event`
2. **Listen**: Receive events (`message.updated`, `message.part.updated`, etc.)
3. **Buffer**: Buffer incoming text chunks
4. **Categorize**: Heuristically categorize chunks (thinking/tool/answer) based on patterns
5. **Render**: Display categorized chunks with color-coding and markdown rendering
6. **Truncate**: Limit output to `max_output_lines` to prevent memory bloat

### Daily Session Resolution Algorithm

```
IF default_session == "daily":
  1. Find all sessions matching today's date pattern (YYYY-MM-DD-daily-N)
  2. Sort by part number (N) descending
  3. For latest session:
     - Check total input + output tokens < daily_max_tokens
     - Check message count < daily_max_messages
  4. IF under limits:
       REUSE existing session
     ELSE:
       CREATE new session with incremented part number
ELSE IF default_session == "ses_xxxxx":
  Use specific session ID (create if missing)
ELSE:
  Use explicitly provided session or create new
```

---

## Acknowledgments

Built using:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [go-sse](https://github.com/tmaxmax/go-sse) - SSE client

---

<p align="center">
  <strong>Built for developers who live in the terminal 🚀</strong>
</p>
