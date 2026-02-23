// Package config handles loading and merging configuration from files and flags.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Config holds all user configuration.
type Config struct {
	Theme        string       `yaml:"theme"`
	Separator    string       `yaml:"separator"`
	Color        *bool        `yaml:"color,omitempty"`
	Lines        [][]string   `yaml:"lines,omitempty"`
	Segments     []string     `yaml:"segments,omitempty"`
	Experimental Experimental `yaml:"experimental"`
}

// Experimental holds flags for experimental features.
type Experimental struct {
	JSONL    bool `yaml:"jsonl"`
	UsageAPI bool `yaml:"usage_api"`
}

// DefaultSegments is the default segment list when no config is provided.
var DefaultSegments = []string{"cwd", "git_branch", "model", "context"}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Theme:     theme.DefaultThemeName,
		Separator: " ",
	}
}

// ResolvedLines returns the effective line layout. If Lines is set, it is
// returned directly. If Segments is set (flat list shorthand), it is
// wrapped as a single line. If neither is set, DefaultSegments is used.
func (c *Config) ResolvedLines() [][]string {
	if len(c.Lines) > 0 {
		return c.Lines
	}
	if len(c.Segments) > 0 {
		return [][]string{c.Segments}
	}
	return [][]string{DefaultSegments}
}

// ColorEnabled returns whether color output should be used. Checks (in
// order): explicit config, NO_COLOR env var. Defaults to true.
func (c *Config) ColorEnabled() bool {
	if c.Color != nil {
		return *c.Color
	}
	_, noColor := os.LookupEnv("NO_COLOR")
	return !noColor
}

// Load reads configuration from the given file path. Returns the default
// config when the file does not exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return nil, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// DefaultPath returns the default config file path (~/.cc-statusline.yaml).
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cc-statusline.yaml"
	}
	return filepath.Join(home, ".cc-statusline.yaml")
}
