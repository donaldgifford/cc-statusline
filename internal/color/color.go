// Package color provides ANSI escape code constants and helpers.
package color

import (
	"os"
	"strings"
	"sync/atomic"
)

// ANSI escape code constants.
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	// Foreground colors.
	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	// Bright foreground colors.
	FgBrightBlack   = "\033[90m"
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"

	// Background colors.
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// Color override state: 0=check env, 1=force enabled, 2=force disabled.
var colorOverride atomic.Int32

// Colorize wraps text with the given ANSI codes and appends a reset.
// When color is disabled, returns the text unmodified.
func Colorize(text string, codes ...string) string {
	if !Enabled() || len(codes) == 0 {
		return text
	}
	var b strings.Builder
	for _, c := range codes {
		b.WriteString(c)
	}
	b.WriteString(text)
	b.WriteString(Reset)
	return b.String()
}

// Enabled returns true if color output is allowed.
// Color is disabled when NO_COLOR is set (any value) or SetEnabled(false) was called.
func Enabled() bool {
	switch colorOverride.Load() {
	case 1:
		return true
	case 2:
		return false
	default:
		_, set := os.LookupEnv("NO_COLOR")
		return !set
	}
}

// SetEnabled overrides color detection. Pass false to disable, true to enable.
// Passing nil resets to environment-based detection.
func SetEnabled(enabled *bool) {
	if enabled == nil {
		colorOverride.Store(0)
		return
	}
	if *enabled {
		colorOverride.Store(1)
	} else {
		colorOverride.Store(2)
	}
}
