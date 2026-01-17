package tui

import (
	"context"

	"opencode-tty/internal/client"
	"opencode-tty/internal/config"
	"opencode-tty/internal/session"
)

func Run(ctx context.Context, cfg config.Config) error {
	cli := client.New(client.Config{Host: cfg.Server.Host, Port: cfg.Server.Port})
	resolver := session.Resolver{Client: cli, Config: cfg}

	defaultSession := cfg.Session.DefaultSession
	if defaultSession == "" {
		defaultSession = "miniopencode"
	}
	sessionID, err := resolver.Resolve(ctx, defaultSession)
	if err != nil {
		return err
	}

	streamer := &Streamer{Client: cli, Events: make(chan Chunk, 64), Errors: make(chan error, 1)}
	streamer.Start(ctx)

	uiCfg := UIConfig{
		Mode:           cfg.UI.Mode,
		Multiline:      cfg.UI.Multiline,
		InputHeight:    cfg.UI.InputHeight,
		ShowThinking:   cfg.UI.ShowThinking,
		ShowTools:      cfg.UI.ShowTools,
		Wrap:           cfg.UI.Wrap,
		MaxOutputLines: cfg.UI.MaxOutputLines,
	}
	promptCfg := PromptConfig{Agent: cfg.Defaults.Agent, ProviderID: cfg.Defaults.ProviderID, ModelID: cfg.Defaults.ModelID}

	m := NewModel(uiCfg)
	m.streamer = streamer
	m.sessionID = sessionID
	m.promptCfg = promptCfg
	m.chunkCh = streamer.Events
	m.errCh = streamer.Errors
	m.maxOutputLines = cfg.UI.MaxOutputLines

	p := newProgram(m)
	_, err = p.Run()
	return err
}
