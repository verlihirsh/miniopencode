package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig
	Session  SessionConfig
	Defaults DefaultsConfig
	UI       UIConfig
	Theme    ThemeConfig
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type SessionConfig struct {
	DefaultSession   string `yaml:"default_session"`
	DailyTitleFormat string `yaml:"daily_title_format"`
	DailyMaxTokens   int    `yaml:"daily_max_tokens"`
	DailyMaxMessages int    `yaml:"daily_max_messages"`
}

type DefaultsConfig struct {
	Agent      string `yaml:"agent"`
	ProviderID string `yaml:"provider_id"`
	ModelID    string `yaml:"model_id"`
}

type UIConfig struct {
	Mode           string `yaml:"mode"`
	ShowThinking   bool   `yaml:"show_thinking"`
	ShowTools      bool   `yaml:"show_tools"`
	Wrap           bool   `yaml:"wrap"`
	InputHeight    int    `yaml:"input_height"`
	MaxOutputLines int    `yaml:"max_output_lines"`
	Theme          string `yaml:"theme"`
}

type ThemeConfig struct {
	BorderStyle       string `yaml:"border_style"`
	OutputBorderColor string `yaml:"output_border_color"`
	InputBorderColor  string `yaml:"input_border_color"`
	StatusColor       string `yaml:"status_color"`
	ThinkingColor     string `yaml:"thinking_color"`
	ToolColor         string `yaml:"tool_color"`
	AnswerColor       string `yaml:"answer_color"`
}

type Options struct {
	Host             *string
	Port             *int
	DefaultSession   *string
	DailyMaxTokens   *int
	DailyMaxMessages *int
	Mode             *string
	ShowThinking     *bool
	ShowTools        *bool
	Wrap             *bool
	InputHeight      *int
	MaxOutputLines   *int
	Theme            *string
	Agent            *string
	ProviderID       *string
	ModelID          *string
}

func Default() Config {
	return Config{
		Server: ServerConfig{Host: "127.0.0.1", Port: 4096},
		Session: SessionConfig{
			DefaultSession:   "",
			DailyTitleFormat: "2006-01-02-daily-%d",
			DailyMaxTokens:   250000,
			DailyMaxMessages: 4000,
		},
		Defaults: DefaultsConfig{},
		UI: UIConfig{
			Mode:           "full",
			ShowThinking:   true,
			ShowTools:      true,
			Wrap:           true,
			InputHeight:    6,
			MaxOutputLines: 4000,
			Theme:          "default",
		},
		Theme: ThemeConfig{
			BorderStyle:       "rounded",
			OutputBorderColor: "#89b4fa",
			InputBorderColor:  "#a6e3a1",
			StatusColor:       "#6c7086",
			ThinkingColor:     "#f9e2af",
			ToolColor:         "#94e2d5",
			AnswerColor:       "#cdd6f4",
		},
	}
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "miniopencode.yaml")
}

func Load(path string, opts Options) (Config, error) {
	cfg := Default()
	if path == "" {
		path = DefaultConfigPath()
	}
	data, err := os.ReadFile(path)
	if err == nil {
		cfgFromFile, err := parseYAML(data)
		if err != nil {
			return cfg, err
		}
		cfg = cfgFromFile
	} else if !errors.Is(err, os.ErrNotExist) {
		return cfg, err
	}
	cfg = applyOptions(cfg, opts)
	return cfg, nil
}

type yamlConfig struct {
	Server *struct {
		Host *string `yaml:"host"`
		Port *int    `yaml:"port"`
	} `yaml:"server"`
	Session *struct {
		DefaultSession   *string `yaml:"default_session"`
		DailyTitleFormat *string `yaml:"daily_title_format"`
		DailyMaxTokens   *int    `yaml:"daily_max_tokens"`
		DailyMaxMessages *int    `yaml:"daily_max_messages"`
	} `yaml:"session"`
	Defaults *struct {
		Agent      *string `yaml:"agent"`
		ProviderID *string `yaml:"provider_id"`
		ModelID    *string `yaml:"model_id"`
	} `yaml:"defaults"`
	UI *struct {
		Mode           *string `yaml:"mode"`
		ShowThinking   *bool   `yaml:"show_thinking"`
		ShowTools      *bool   `yaml:"show_tools"`
		Wrap           *bool   `yaml:"wrap"`
		InputHeight    *int    `yaml:"input_height"`
		MaxOutputLines *int    `yaml:"max_output_lines"`
		Theme          *string `yaml:"theme"`
	} `yaml:"ui"`
	Theme *struct {
		BorderStyle       *string `yaml:"border_style"`
		OutputBorderColor *string `yaml:"output_border_color"`
		InputBorderColor  *string `yaml:"input_border_color"`
		StatusColor       *string `yaml:"status_color"`
		ThinkingColor     *string `yaml:"thinking_color"`
		ToolColor         *string `yaml:"tool_color"`
		AnswerColor       *string `yaml:"answer_color"`
	} `yaml:"theme"`
}

func parseYAML(data []byte) (Config, error) {
	var y yamlConfig
	if err := yaml.Unmarshal(data, &y); err != nil {
		return Config{}, err
	}
	cfg := Default()
	applyYAML(&cfg, y)
	return cfg, nil
}

func applyYAML(cfg *Config, y yamlConfig) {
	if y.Server != nil {
		if y.Server.Host != nil {
			cfg.Server.Host = *y.Server.Host
		}
		if y.Server.Port != nil {
			cfg.Server.Port = *y.Server.Port
		}
	}
	if y.Session != nil {
		if y.Session.DefaultSession != nil {
			cfg.Session.DefaultSession = *y.Session.DefaultSession
		}
		if y.Session.DailyTitleFormat != nil {
			cfg.Session.DailyTitleFormat = *y.Session.DailyTitleFormat
		}
		if y.Session.DailyMaxTokens != nil {
			cfg.Session.DailyMaxTokens = *y.Session.DailyMaxTokens
		}
		if y.Session.DailyMaxMessages != nil {
			cfg.Session.DailyMaxMessages = *y.Session.DailyMaxMessages
		}
	}
	if y.Defaults != nil {
		if y.Defaults.Agent != nil {
			cfg.Defaults.Agent = *y.Defaults.Agent
		}
		if y.Defaults.ProviderID != nil {
			cfg.Defaults.ProviderID = *y.Defaults.ProviderID
		}
		if y.Defaults.ModelID != nil {
			cfg.Defaults.ModelID = *y.Defaults.ModelID
		}
	}
	if y.UI != nil {
		if y.UI.Mode != nil {
			cfg.UI.Mode = *y.UI.Mode
		}
		if y.UI.ShowThinking != nil {
			cfg.UI.ShowThinking = *y.UI.ShowThinking
		}
		if y.UI.ShowTools != nil {
			cfg.UI.ShowTools = *y.UI.ShowTools
		}
		if y.UI.Wrap != nil {
			cfg.UI.Wrap = *y.UI.Wrap
		}
		if y.UI.InputHeight != nil {
			cfg.UI.InputHeight = *y.UI.InputHeight
		}
		if y.UI.MaxOutputLines != nil {
			cfg.UI.MaxOutputLines = *y.UI.MaxOutputLines
		}
		if y.UI.Theme != nil {
			cfg.UI.Theme = *y.UI.Theme
		}
	}
	if y.Theme != nil {
		if y.Theme.BorderStyle != nil {
			cfg.Theme.BorderStyle = *y.Theme.BorderStyle
		}
		if y.Theme.OutputBorderColor != nil {
			cfg.Theme.OutputBorderColor = *y.Theme.OutputBorderColor
		}
		if y.Theme.InputBorderColor != nil {
			cfg.Theme.InputBorderColor = *y.Theme.InputBorderColor
		}
		if y.Theme.StatusColor != nil {
			cfg.Theme.StatusColor = *y.Theme.StatusColor
		}
		if y.Theme.ThinkingColor != nil {
			cfg.Theme.ThinkingColor = *y.Theme.ThinkingColor
		}
		if y.Theme.ToolColor != nil {
			cfg.Theme.ToolColor = *y.Theme.ToolColor
		}
		if y.Theme.AnswerColor != nil {
			cfg.Theme.AnswerColor = *y.Theme.AnswerColor
		}
	}
}

func applyOptions(cfg Config, opts Options) Config {
	if opts.Host != nil {
		cfg.Server.Host = *opts.Host
	}
	if opts.Port != nil {
		cfg.Server.Port = *opts.Port
	}
	if opts.DefaultSession != nil {
		cfg.Session.DefaultSession = *opts.DefaultSession
	}
	if opts.DailyMaxTokens != nil {
		cfg.Session.DailyMaxTokens = *opts.DailyMaxTokens
	}
	if opts.DailyMaxMessages != nil {
		cfg.Session.DailyMaxMessages = *opts.DailyMaxMessages
	}
	if opts.Mode != nil {
		cfg.UI.Mode = *opts.Mode
	}
	if opts.ShowThinking != nil {
		cfg.UI.ShowThinking = *opts.ShowThinking
	}
	if opts.ShowTools != nil {
		cfg.UI.ShowTools = *opts.ShowTools
	}
	if opts.Wrap != nil {
		cfg.UI.Wrap = *opts.Wrap
	}
	if opts.InputHeight != nil {
		cfg.UI.InputHeight = *opts.InputHeight
	}
	if opts.MaxOutputLines != nil {
		cfg.UI.MaxOutputLines = *opts.MaxOutputLines
	}
	if opts.Theme != nil {
		cfg.UI.Theme = *opts.Theme
	}
	if opts.Agent != nil {
		cfg.Defaults.Agent = *opts.Agent
	}
	if opts.ProviderID != nil {
		cfg.Defaults.ProviderID = *opts.ProviderID
	}
	if opts.ModelID != nil {
		cfg.Defaults.ModelID = *opts.ModelID
	}
	return cfg
}
