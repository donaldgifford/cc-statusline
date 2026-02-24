// Package theme provides named color themes for statusline rendering.
package theme

import "github.com/donaldgifford/cc-statusline/internal/color"

// SegmentStyle defines colors for a single segment.
type SegmentStyle struct {
	Fg   string
	Bold bool
}

// Theme maps segment names to their color styles.
type Theme struct {
	Name   string
	Styles map[string]SegmentStyle
}

// DefaultThemeName is the theme used when no config is specified.
const DefaultThemeName = "tokyo-night"

// Get returns the theme with the given name. Falls back to the default
// theme if the name is unknown.
func Get(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes[DefaultThemeName]
}

// Names returns all available theme names.
func Names() []string {
	return []string{"tokyo-night", "rose-pine", "catppuccin"}
}

// Style returns the SegmentStyle for the named segment in this theme.
// Returns a zero-value SegmentStyle if the segment has no theme entry.
func (t *Theme) Style(segment string) SegmentStyle {
	if s, ok := t.Styles[segment]; ok {
		return s
	}
	return SegmentStyle{}
}

// Colorize applies this theme's style for the given segment to text.
func (t *Theme) Colorize(segment, text string) string {
	s := t.Style(segment)
	var codes []string
	if s.Bold {
		codes = append(codes, color.Bold)
	}
	if s.Fg != "" {
		codes = append(codes, s.Fg)
	}
	if len(codes) == 0 {
		return text
	}
	return color.Colorize(text, codes...)
}

var themes = map[string]*Theme{
	"tokyo-night": {
		Name: "tokyo-night",
		Styles: map[string]SegmentStyle{
			"cwd":        {Fg: color.FgBlue},
			"git_branch": {Fg: color.FgMagenta},
			"model":      {Fg: color.FgCyan, Bold: true},
			"context":    {Fg: color.FgGreen},
			"cost":       {Fg: color.FgYellow},
			"duration":   {Fg: color.FgBrightBlack},
			"tokens":     {Fg: color.FgBrightBlack},
			"lines":      {},
			"vim":        {Fg: color.FgBrightYellow, Bold: true},
			"agent":      {Fg: color.FgBrightCyan},
		},
	},
	"rose-pine": {
		Name: "rose-pine",
		Styles: map[string]SegmentStyle{
			"cwd":        {Fg: color.FgBrightBlue},
			"git_branch": {Fg: color.FgBrightMagenta},
			"model":      {Fg: color.FgBrightCyan, Bold: true},
			"context":    {Fg: color.FgGreen},
			"cost":       {Fg: color.FgBrightYellow},
			"duration":   {Fg: color.FgBrightBlack},
			"tokens":     {Fg: color.FgBrightBlack},
			"lines":      {},
			"vim":        {Fg: color.FgYellow, Bold: true},
			"agent":      {Fg: color.FgCyan},
		},
	},
	"catppuccin": {
		Name: "catppuccin",
		Styles: map[string]SegmentStyle{
			"cwd":        {Fg: color.FgBrightBlue},
			"git_branch": {Fg: color.FgMagenta},
			"model":      {Fg: color.FgBrightGreen, Bold: true},
			"context":    {Fg: color.FgGreen},
			"cost":       {Fg: color.FgYellow},
			"duration":   {Fg: color.FgBrightBlack},
			"tokens":     {Fg: color.FgBrightBlack},
			"lines":      {},
			"vim":        {Fg: color.FgBrightRed, Bold: true},
			"agent":      {Fg: color.FgBrightCyan},
		},
	},
}
