package statusline_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testTimeout = 30 * time.Second

// TestIntegration builds the binary and pipes test JSON to verify end-to-end behavior.
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildBinary(t, "")
	tests := []struct {
		name    string
		input   string
		contain []string
	}{
		{
			name:    "basic JSON",
			input:   `{"model":{"id":"claude-opus-4-6","display_name":"Opus 4.6"},"cwd":"/tmp","context_window":{"used_percentage":43,"context_window_size":200000}}`,
			contain: []string{"Opus 4.6", "43%"},
		},
		{
			name:    "empty object",
			input:   `{}`,
			contain: []string{"ctx:--"},
		},
		{
			name:  "empty stdin",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, binPath, "--no-color")
			cmd.Stdin = strings.NewReader(tt.input)
			cmd.Env = append(os.Environ(), "NO_COLOR=1")
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("command failed: %v", err)
			}

			output := string(out)
			for _, s := range tt.contain {
				if !strings.Contains(output, s) {
					t.Errorf("output %q does not contain %q", output, s)
				}
			}
		})
	}
}

// TestIntegrationVersion verifies the version subcommand.
func TestIntegrationVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ldflags := "-X main.version=v0.1.0-test -X main.commit=abc1234"
	binPath := buildBinary(t, ldflags)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, binPath, "version").Output()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "v0.1.0-test") {
		t.Errorf("version output %q does not contain version", output)
	}
	if !strings.Contains(output, "abc1234") {
		t.Errorf("version output %q does not contain commit", output)
	}
}

func buildBinary(t *testing.T, ldflags string) string {
	t.Helper()

	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "cc-statusline")

	args := []string{"build"}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	args = append(args, "-o", binPath, "github.com/donaldgifford/cc-statusline/cmd/cc-statusline")

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binPath
}
