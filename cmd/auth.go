package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/donaldgifford/cc-statusline/internal/usageapi"
)

var authStatus bool

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage OAuth credentials for the usage API",
	Long: `Manage OAuth credentials for the usage API.

Without flags, accepts a token pasted via stdin and saves it.
With --status, reports the current credential state.`,
	RunE: runAuth,
}

func init() {
	authCmd.Flags().BoolVar(&authStatus, "status", false, "report credential status")
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, _ []string) error {
	if authStatus {
		return reportAuthStatus(cmd.OutOrStdout())
	}
	return acceptToken(cmd.InOrStdin(), cmd.OutOrStdout())
}

func reportAuthStatus(w io.Writer) error {
	cfg := usageapi.AuthConfig{
		RunCommand: defaultCommandRunner,
	}

	result, err := usageapi.ReadCredentials(cfg)
	if err != nil {
		return writeStatus(w, "No credentials found.\nDetails: %v\n\nTo authenticate, run: cc-statusline auth\n", err)
	}

	if err := writeStatus(w, "Source: %s\n", result.Source); err != nil {
		return err
	}

	if result.ExpiresAt.IsZero() {
		return writeStatus(w, "Status: valid (no expiry info)\n")
	}

	if err := writeStatus(w, "Expires: %s\n", result.ExpiresAt.Format("2006-01-02 15:04:05")); err != nil {
		return err
	}

	if !result.Expired {
		return writeStatus(w, "Status: valid\n")
	}

	if err := writeStatus(w, "Status: EXPIRED\n"); err != nil {
		return err
	}
	if result.RefreshToken != "" {
		return writeStatus(w, "A refresh token is available. The token will be refreshed on next use.\n")
	}
	return writeStatus(w, "No refresh token available. Run 'claude auth' to re-authenticate.\n")
}

func acceptToken(r io.Reader, w io.Writer) error {
	if err := writeStatus(w, "Paste your OAuth access token (press Enter when done):\n"); err != nil {
		return err
	}

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return fmt.Errorf("no input received")
	}
	token := strings.TrimSpace(scanner.Text())
	if token == "" {
		return fmt.Errorf("empty token")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("determine config directory: %w", err)
	}

	dir := filepath.Join(configDir, "cc-statusline")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	type storedCreds struct {
		ClaudeAIOAuth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	creds := storedCreds{}
	creds.ClaudeAIOAuth.AccessToken = token

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	authPath := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(authPath, data, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return writeStatus(w, "Token saved to %s\n", authPath)
}

func writeStatus(w io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(w, format, args...)
	return err
}

func defaultCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}
