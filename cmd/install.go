package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Configure Claude Code to use cc-statusline",
	Long:  "Writes the statusLine entry to Claude Code's settings.json with the absolute path to this binary.",
	RunE:  runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(_ *cobra.Command, _ []string) error {
	binPath, err := resolveExecutable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	settingsPath := claudeSettingsPath()
	settings, err := readSettingsFile(settingsPath)
	if err != nil {
		return err
	}

	// Check for existing statusLine.
	if existing, ok := settings["statusLine"]; ok {
		fmt.Fprintf(os.Stderr, "Warning: overwriting existing statusLine entry: %v\n", existing)
	}

	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": binPath,
		"padding": 0,
	}

	return writeSettingsFile(settingsPath, settings)
}

func resolveExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func claudeSettingsPath() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "settings.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".claude", "settings.json")
	}
	return filepath.Join(home, ".claude", "settings.json")
}

func readSettingsFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("read settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}
	return settings, nil
}

func writeSettingsFile(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	fmt.Printf("Installed cc-statusline at %s\n", path)
	return nil
}
