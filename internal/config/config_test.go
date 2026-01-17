package config

import (
	"os"
	"path/filepath"
	"testing"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != 4096 {
		t.Fatalf("unexpected server defaults: %+v", cfg.Server)
	}
	if cfg.Session.DailyTitleFormat != "2006-01-02-daily-%d" {
		t.Fatalf("unexpected daily title format: %s", cfg.Session.DailyTitleFormat)
	}
	if cfg.UI.Mode != "full" || cfg.UI.ShowThinking != true || cfg.UI.ShowTools != true {
		t.Fatalf("unexpected ui defaults: %+v", cfg.UI)
	}
}

func TestLoadMergesYamlAndCLI(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "miniopencode.yaml")
	yamlContent := `server:
  host: yaml-host
  port: 1111
session:
  default_session: yaml-default
ui:
  mode: input
  show_thinking: true
  show_tools: false
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	cli := Options{
		Host:             strPtr("cli-host"),
		Port:             intPtr(2222),
		DefaultSession:   strPtr("cli-default"),
		Mode:             strPtr("full"),
		ShowThinking:     boolPtr(false),
		ShowTools:        boolPtr(true),
		DailyMaxTokens:   intPtr(3333),
		DailyMaxMessages: intPtr(44),
	}

	cfg, err := Load(yamlPath, cli)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Server.Host != "cli-host" || cfg.Server.Port != 2222 {
		t.Fatalf("cli overrides server failed: %+v", cfg.Server)
	}
	if cfg.Session.DefaultSession != "cli-default" {
		t.Fatalf("cli default_session not applied: %s", cfg.Session.DefaultSession)
	}
	if cfg.Session.DailyMaxTokens != 3333 || cfg.Session.DailyMaxMessages != 44 {
		t.Fatalf("cli daily limits not applied: %+v", cfg.Session)
	}
	if cfg.UI.Mode != "full" {
		t.Fatalf("ui mode override failed: %s", cfg.UI.Mode)
	}
	if cfg.UI.ShowThinking != false || cfg.UI.ShowTools != true {
		t.Fatalf("ui toggles override failed: %+v", cfg.UI)
	}
	// untouched fields fall back to YAML/default
	if cfg.UI.Wrap != true {
		t.Fatalf("expected wrap default true, got %v", cfg.UI.Wrap)
	}
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cli := Options{}
	cfg, err := Load(filepath.Join("/nonexistent", "miniopencode.yaml"), cli)
	if err != nil {
		t.Fatalf("load missing should not error: %v", err)
	}
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != 4096 {
		t.Fatalf("defaults not applied: %+v", cfg.Server)
	}
}
