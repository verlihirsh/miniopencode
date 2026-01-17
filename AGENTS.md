# AGENTS.md - AI Coding Agent Guidelines

## Project Overview

**opencode-tty** is a Go proxy that bridges stdin/stdout with an opencode server via HTTP and SSE.
- Single-file Go application (`main.go`)
- Reads JSON commands from stdin, writes JSON responses to stdout
- Proxies to opencode server for session management and prompt handling

---

## Build, Test, and Run Commands

### Build
```bash
go build -o opencode-tty .
```

### Run
```bash
# With defaults (127.0.0.1:4096)
./opencode-tty

# With custom host/port
OPENCODE_HOST=localhost OPENCODE_PORT=8080 ./opencode-tty
```

### Test
```bash
# Run all tests
go test -v ./...

# Run single test
go test -v -run TestFunctionName ./...

# Run tests with coverage
go test -cover ./...
```

**Note**: No tests exist yet. When adding tests, follow Go conventions:
- Test files: `*_test.go`
- Test functions: `func TestXxx(t *testing.T)`

### Lint and Format
```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...

# If golangci-lint installed:
golangci-lint run
```

---

## Code Style Guidelines

### File Organization
- Keep related types and functions together
- Order: types → constructors → methods → helpers → main

### Imports
```go
import (
    // Standard library (sorted alphabetically)
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"

    // External packages (if any, blank line between groups)
    // "github.com/some/package"

    // Internal packages (if any)
    // "opencode-tty/internal/..."
)
```

### Naming Conventions
| Element | Style | Example |
|---------|-------|---------|
| Packages | lowercase, single word | `main`, `proxy` |
| Exported types/funcs | PascalCase | `Config`, `NewProxy` |
| Unexported | camelCase | `baseURL`, `outputError` |
| Constants | PascalCase or ALL_CAPS | `MaxBufferSize` |
| Acronyms | Consistent case | `HTTP`, `SSE`, `ID` |

### Struct Tags
- Use `json:"fieldName"` for JSON marshaling
- Use `json:"fieldName,omitempty"` for optional fields
- Keep tags on same line when short

```go
type Config struct {
    Host      string `json:"host"`
    Port      string `json:"port"`
    SessionID string `json:"session_id,omitempty"`
}
```

### Error Handling
- Always check errors immediately after function calls
- Return errors to caller, don't swallow them
- Use `fmt.Errorf("context: %v", err)` for wrapping

```go
// GOOD
resp, err := http.Get(url)
if err != nil {
    return fmt.Errorf("failed to fetch %s: %w", url, err)
}
defer resp.Body.Close()

// BAD - Don't ignore errors
resp, _ := http.Get(url)
```

### Defer for Cleanup
- Use `defer` for closing resources immediately after opening
- Defer after nil check, not before

```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()  // Correct placement
```

### Mutex Usage
- Lock/unlock in same function when possible
- Use `defer p.mu.Unlock()` for safety
- Keep critical sections small

```go
func (p *Proxy) output(eventType string, data interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()
    // ... protected operations
}
```

### JSON Handling
- Use `json.RawMessage` for delayed parsing
- Use `map[string]interface{}` for dynamic JSON (sparingly)
- Prefer typed structs when structure is known

---

## API Communication Patterns

### Command Structure (stdin → proxy)
```json
{"type": "command.name", "payload": {...}}
```

### Response Structure (proxy → stdout)
```json
{"type": "event.name", "data": {...}}
```

### Supported Commands
- `health` - Check server health
- `session.create` - Create new session
- `session.list` - List all sessions
- `session.select` - Select active session
- `prompt` - Send prompt to active session
- `sse.start` / `sse.stop` - Control SSE stream

---

## Testing Guidelines (When Adding Tests)

### Table-Driven Tests
```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "foo", "bar", false},
        {"empty input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Something(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.expected {
                t.Errorf("got = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

---

## Do's and Don'ts

### DO
- Run `go fmt` before committing
- Use meaningful variable names
- Keep functions focused and small
- Handle all errors explicitly
- Close resources with defer

### DON'T
- Ignore errors with `_`
- Use global variables
- Leave commented-out code
- Use `panic` for recoverable errors
- Commit without `go vet` passing

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENCODE_HOST` | `127.0.0.1` | Server host |
| `OPENCODE_PORT` | `4096` | Server port |
