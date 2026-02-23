package color_test

import (
	"testing"

	"github.com/donaldgifford/cc-statusline/internal/color"
)

func TestColorize(t *testing.T) {
	t.Parallel()

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

	// Force color enabled for these tests.
	enabled := true
	color.SetEnabled(&enabled)
	t.Cleanup(func() { color.SetEnabled(nil) })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := color.Colorize(tt.text, tt.codes...)
			if got != tt.want {
				t.Errorf("Colorize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorizeDisabled(t *testing.T) {
	t.Parallel()

	disabled := false
	color.SetEnabled(&disabled)
	t.Cleanup(func() { color.SetEnabled(nil) })

	got := color.Colorize("hello", color.FgRed, color.Bold)
	if got != "hello" {
		t.Errorf("Colorize() with color disabled = %q, want %q", got, "hello")
	}
}

func TestEnabledDefault(t *testing.T) {
	t.Parallel()

	// Reset to default behavior.
	color.SetEnabled(nil)
	// Without NO_COLOR set, Enabled() should return true.
	// We can't test NO_COLOR env var reliably in parallel tests
	// since env vars are process-global, but we can test SetEnabled.
	enabled := true
	color.SetEnabled(&enabled)
	t.Cleanup(func() { color.SetEnabled(nil) })

	if !color.Enabled() {
		t.Error("Enabled() should return true when SetEnabled(true)")
	}
}

func TestSetEnabledToggle(t *testing.T) {
	t.Parallel()

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
}
