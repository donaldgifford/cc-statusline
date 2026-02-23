package jsonl

import (
	"os"
	"path/filepath"
)

// DiscoverFiles returns paths to JSONL transcript files.
// It checks directories in priority order:
// 1. $CLAUDE_CONFIG_DIR/projects/
// 2. $XDG_CONFIG_HOME/claude/projects/
// 3. ~/.claude/projects/.
func DiscoverFiles() []string {
	dirs := configDirs()
	var files []string
	for _, dir := range dirs {
		pattern := filepath.Join(dir, "projects", "*", "*.jsonl")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}
	return files
}

func configDirs() []string {
	var dirs []string
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		dirs = append(dirs, d)
	}
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		dirs = append(dirs, filepath.Join(d, "claude"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".claude"))
	}
	return dirs
}
