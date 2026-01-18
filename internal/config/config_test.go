package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	t.Run("ServerDefaults", func(t *testing.T) {
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 4096, cfg.Server.Port)
	})

	t.Run("SessionDefaults", func(t *testing.T) {
		assert.Equal(t, "2006-01-02-daily-%d", cfg.Session.DailyTitleFormat)
		assert.Equal(t, 250000, cfg.Session.DailyMaxTokens)
		assert.Equal(t, 4000, cfg.Session.DailyMaxMessages)
		assert.Empty(t, cfg.Session.DefaultSession)
	})

	t.Run("UIDefaults", func(t *testing.T) {
		assert.Equal(t, "full", cfg.UI.Mode)
		assert.True(t, cfg.UI.ShowThinking)
		assert.True(t, cfg.UI.ShowTools)
		assert.True(t, cfg.UI.Wrap)
		assert.Equal(t, 6, cfg.UI.InputHeight)
		assert.Equal(t, 4000, cfg.UI.MaxOutputLines)
		assert.Equal(t, "default", cfg.UI.Theme)
	})

	t.Run("DefaultsDefaults", func(t *testing.T) {
		assert.Empty(t, cfg.Defaults.Agent)
		assert.Empty(t, cfg.Defaults.ProviderID)
		assert.Empty(t, cfg.Defaults.ModelID)
	})
}

func TestLoadMergesYamlAndCLI(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "miniopencode.yaml")
	yamlContent := `server:
  host: yaml-host
  port: 1111
session:
  default_session: yaml-default
  daily_max_tokens: 100000
ui:
  mode: input
  show_thinking: true
  show_tools: false
defaults:
  agent: yaml-agent
`
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cli := Options{
		Host:             strPtr("cli-host"),
		Port:             intPtr(2222),
		DefaultSession:   strPtr("cli-default"),
		Mode:             strPtr("full"),
		ShowThinking:     boolPtr(false),
		ShowTools:        boolPtr(true),
		DailyMaxTokens:   intPtr(3333),
		DailyMaxMessages: intPtr(44),
		Agent:            strPtr("cli-agent"),
	}

	cfg, err := Load(yamlPath, cli)
	require.NoError(t, err)

	t.Run("CLIOverridesServer", func(t *testing.T) {
		assert.Equal(t, "cli-host", cfg.Server.Host, "CLI should override YAML host")
		assert.Equal(t, 2222, cfg.Server.Port, "CLI should override YAML port")
	})

	t.Run("CLIOverridesSession", func(t *testing.T) {
		assert.Equal(t, "cli-default", cfg.Session.DefaultSession)
		assert.Equal(t, 3333, cfg.Session.DailyMaxTokens)
		assert.Equal(t, 44, cfg.Session.DailyMaxMessages)
	})

	t.Run("CLIOverridesUI", func(t *testing.T) {
		assert.Equal(t, "full", cfg.UI.Mode)
		assert.False(t, cfg.UI.ShowThinking)
		assert.True(t, cfg.UI.ShowTools)
	})

	t.Run("CLIOverridesDefaults", func(t *testing.T) {
		assert.Equal(t, "cli-agent", cfg.Defaults.Agent)
	})

	t.Run("UntouchedFieldsFromYAML", func(t *testing.T) {
		// Wrap is not overridden by CLI, should come from default (true)
		assert.True(t, cfg.UI.Wrap)
	})
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cli := Options{}
	cfg, err := Load(filepath.Join("/nonexistent", "miniopencode.yaml"), cli)
	require.NoError(t, err, "missing config file should not error")

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 4096, cfg.Server.Port)
	assert.Equal(t, "full", cfg.UI.Mode)
}

func TestLoadWithEmptyPath(t *testing.T) {
	cli := Options{
		Host: strPtr("custom-host"),
		Port: intPtr(9999),
	}
	cfg, err := Load("", cli)
	require.NoError(t, err)

	assert.Equal(t, "custom-host", cfg.Server.Host)
	assert.Equal(t, 9999, cfg.Server.Port)
}

func TestLoadYAMLOnly(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "test.yaml")
	yamlContent := `server:
  host: yaml-only-host
  port: 5555
ui:
  mode: output
  show_thinking: false
`
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	cfg, err := Load(yamlPath, Options{})
	require.NoError(t, err)

	assert.Equal(t, "yaml-only-host", cfg.Server.Host)
	assert.Equal(t, 5555, cfg.Server.Port)
	assert.Equal(t, "output", cfg.UI.Mode)
	assert.False(t, cfg.UI.ShowThinking)
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "invalid.yaml")
	invalidContent := `this is not: valid: yaml::`
	err := os.WriteFile(yamlPath, []byte(invalidContent), 0o644)
	require.NoError(t, err)

	_, err = Load(yamlPath, Options{})
	assert.Error(t, err, "invalid YAML should return error")
}
