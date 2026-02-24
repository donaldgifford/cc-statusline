package color_test

import (
	"testing"

	"github.com/donaldgifford/cc-statusline/internal/color"
)

func TestColorize(t *testing.T) {
	// Not parallel: mutates global color state.
	enabled := true
	color.SetEnabled(&enabled)
	t.Cleanup(func() { color.SetEnabled(nil) })

	tests := []struct {
		name  string
		text  string
		codes []string
		want  string
	}{
		{
			name:  "single code",
			text:  "hello",
			codes: []string{color.FgRed},
			want:  "\033[31mhello\033[0m",
		},
		{
			name:  "multiple codes",
			text:  "hello",
			codes: []string{color.Bold, color.FgBlue},
			want:  "\033[1m\033[34mhello\033[0m",
		},
		{
			name:  "no codes returns text unchanged",
			text:  "hello",
			codes: nil,
			want:  "hello",
		},
		{
			name:  "empty text",
			text:  "",
			codes: []string{color.FgRed},
			want:  "\033[31m\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := color.Colorize(tt.text, tt.codes...)
			if got != tt.want {
				t.Errorf("Colorize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorizeDisabled(t *testing.T) {
	// Not parallel: mutates global color state.
	disabled := false
	color.SetEnabled(&disabled)
	t.Cleanup(func() { color.SetEnabled(nil) })

	got := color.Colorize("hello", color.FgRed, color.Bold)
	if got != "hello" {
		t.Errorf("Colorize() with color disabled = %q, want %q", got, "hello")
	}
}

func TestSetEnabledToggle(t *testing.T) {
	// Not parallel: mutates global color state.
	t.Cleanup(func() { color.SetEnabled(nil) })

	disabled := false
	color.SetEnabled(&disabled)
	if color.Enabled() {
		t.Error("Enabled() should return false after SetEnabled(false)")
	}

	enabled := true
	color.SetEnabled(&enabled)
	if !color.Enabled() {
		t.Error("Enabled() should return true after SetEnabled(true)")
	}

	color.SetEnabled(nil)
	// After reset, depends on NO_COLOR env var. Don't assert value.
}
