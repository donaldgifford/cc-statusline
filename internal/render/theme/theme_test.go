package theme_test

import (
	"testing"

	"github.com/donaldgifford/cc-statusline/internal/color"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

func TestGetKnownTheme(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			th := theme.Get(name)
			if th == nil {
				t.Fatalf("Get(%q) returned nil", name)
			}
			if th.Name != name {
				t.Errorf("Get(%q).Name = %q", name, th.Name)
			}
		})
	}
}

func TestGetUnknownFallsBack(t *testing.T) {
	t.Parallel()

	th := theme.Get("nonexistent-theme")
	if th == nil {
		t.Fatal("Get() for unknown theme returned nil")
	}
	if th.Name != theme.DefaultThemeName {
		t.Errorf("expected fallback to %q, got %q", theme.DefaultThemeName, th.Name)
	}
}

func TestGetDefault(t *testing.T) {
	t.Parallel()

	th := theme.Get(theme.DefaultThemeName)
	if th == nil {
		t.Fatal("default theme not found")
	}
	// The default theme should have styles for the standard segments.
	for _, seg := range []string{"cwd", "git_branch", "model", "context"} {
		s := th.Style(seg)
		if s.Fg == "" {
			t.Errorf("default theme missing style for segment %q", seg)
		}
	}
}

func TestStyleMissingSegment(t *testing.T) {
	t.Parallel()

	th := theme.Get(theme.DefaultThemeName)
	s := th.Style("nonexistent_segment")
	if s.Fg != "" || s.Bold {
		t.Error("Style() for unknown segment should return zero-value SegmentStyle")
	}
}

func TestColorize(t *testing.T) {
	t.Parallel()

	enabled := true
	color.SetEnabled(&enabled)
	t.Cleanup(func() { color.SetEnabled(nil) })

	th := theme.Get("tokyo-night")
	result := th.Colorize("model", "[Opus 4.6]")
	// Model in tokyo-night is bold + cyan.
	if result == "[Opus 4.6]" {
		t.Error("Colorize() should wrap text with ANSI codes when color is enabled")
	}
	// Should contain the reset code at the end.
	if len(result) < len("[Opus 4.6]")+len(color.Reset) {
		t.Error("Colorize() output too short, missing ANSI codes")
	}
}

func TestColorizeNoStyle(t *testing.T) {
	t.Parallel()

	enabled := true
	color.SetEnabled(&enabled)
	t.Cleanup(func() { color.SetEnabled(nil) })

	th := theme.Get("tokyo-night")
	result := th.Colorize("nonexistent_segment", "text")
	if result != "text" {
		t.Errorf("Colorize() for unstyled segment should return text unchanged, got %q", result)
	}
}
