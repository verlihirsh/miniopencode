package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"opencode-tty/internal/config"
	"opencode-tty/internal/proxy"
	"opencode-tty/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	headless := flag.Bool("headless", false, "run in headless stdin/stdout mode")
	configPath := flag.String("config", "", "path to config file (default: ~/.config/miniopencode.yaml)")

	// UI flags
	mode := flag.String("mode", "", "UI mode: input|output|full")
	showThinking := flag.Bool("show-thinking", false, "show thinking blocks")
	showTools := flag.Bool("show-tools", false, "show tool calls")
	wrap := flag.Bool("wrap", false, "wrap text in output")
	inputHeight := flag.Int("input-height", 0, "input box height")
	maxOutputLines := flag.Int("max-output-lines", 0, "maximum output lines")
	theme := flag.String("theme", "", "theme name")

	logPath := flag.String("log", "", "write debug logs to file (or set DEBUG=1 for default path)")

	// Server flags
	host := flag.String("host", "", "server host")
	port := flag.Int("port", 0, "server port")

	// Session flags
	defaultSession := flag.String("session", "", "default session ID or 'daily'")
	dailyMaxTokens := flag.Int("daily-max-tokens", 0, "daily session max tokens")
	dailyMaxMessages := flag.Int("daily-max-messages", 0, "daily session max messages")

	// Defaults flags
	agent := flag.String("agent", "", "default agent")
	providerID := flag.String("provider", "", "default provider ID")
	modelID := flag.String("model", "", "default model ID")

	flag.Parse()

	opts := config.Options{}
	if *mode != "" {
		opts.Mode = mode
	}
	if flag.Lookup("show-thinking").Value.String() == "true" {
		opts.ShowThinking = showThinking
	}
	if flag.Lookup("show-tools").Value.String() == "true" {
		opts.ShowTools = showTools
	}
	if flag.Lookup("wrap").Value.String() == "true" {
		opts.Wrap = wrap
	}
	if *inputHeight > 0 {
		opts.InputHeight = inputHeight
	}
	if *maxOutputLines > 0 {
		opts.MaxOutputLines = maxOutputLines
	}
	if *theme != "" {
		opts.Theme = theme
	}
	if *host != "" {
		opts.Host = host
	}
	if *port > 0 {
		opts.Port = port
	}
	if *defaultSession != "" {
		opts.DefaultSession = defaultSession
	}
	if *dailyMaxTokens > 0 {
		opts.DailyMaxTokens = dailyMaxTokens
	}
	if *dailyMaxMessages > 0 {
		opts.DailyMaxMessages = dailyMaxMessages
	}
	if *agent != "" {
		opts.Agent = agent
	}
	if *providerID != "" {
		opts.ProviderID = providerID
	}
	if *modelID != "" {
		opts.ModelID = modelID
	}

	cfg, err := config.Load(*configPath, opts)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// File logging (avoid corrupting the TUI). Enabled via --log or DEBUG=1.
	var logFile *os.File
	if *logPath != "" || os.Getenv("DEBUG") != "" {
		path := *logPath
		if path == "" {
			path = filepath.Join(os.TempDir(), "miniopencode.log")
		}
		f, err := tea.LogToFile(path, "miniopencode")
		if err != nil {
			log.Fatalf("log to file: %v", err)
		}
		logFile = f
		defer logFile.Close()
		log.Printf("logging enabled: %s", path)
	}

	if *headless {
		p := proxy.NewProxy(proxy.Config{Host: cfg.Server.Host, Port: fmt.Sprintf("%d", cfg.Server.Port)})
		p.RunHeadless()
		return
	}

	ctx := context.Background()
	if err := tui.Run(ctx, cfg); err != nil {
		log.Fatalf("tui: %v", err)
	}
}
