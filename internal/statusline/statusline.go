// Package statusline provides the core entry point for rendering the statusline.
package statusline

import (
	"context"
	"io"
	"os/exec"

	"github.com/donaldgifford/cc-statusline/internal/config"
	"github.com/donaldgifford/cc-statusline/internal/errlog"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render"
	"github.com/donaldgifford/cc-statusline/internal/render/segments"
	"github.com/donaldgifford/cc-statusline/internal/usageapi"
)

// Run reads session JSON from in, renders the statusline, and writes to out.
// This is the core function that all tests exercise.
func Run(in io.Reader, out, _ io.Writer) error {
	return RunWithConfig(in, out, nil)
}

// RunWithConfig reads session JSON and renders the statusline using the
// provided configuration. If cfg is nil, the default config is used.
// Callers should set color.SetEnabled before calling this function.
func RunWithConfig(in io.Reader, out io.Writer, cfg *config.Config) error {
	data, err := model.ReadStatus(in)
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = config.Default()
	}

	rcfg := render.Config{
		Lines:                cfg.ResolvedLines(),
		Separator:            cfg.Separator,
		ThemeName:            cfg.Theme,
		ExperimentalJSONL:    cfg.Experimental.JSONL,
		ExperimentalUsageAPI: cfg.Experimental.UsageAPI,
	}

	segCfg := segments.AllConfig{}
	if cfg.Experimental.UsageAPI {
		segCfg.UsageFetcher = buildUsageFetcher()
	}

	r := render.New(rcfg, segments.All(segCfg))
	return r.Render(out, data)
}

// buildUsageFetcher creates a cached usage API fetcher from available credentials.
func buildUsageFetcher() func() (*usageapi.UsageResponse, error) {
	creds, err := usageapi.ReadCredentials(usageapi.AuthConfig{
		RunCommand: defaultCommandRunner,
	})
	if err != nil {
		errlog.Log("usage API credentials: %v", err)
		return func() (*usageapi.UsageResponse, error) {
			return nil, err
		}
	}

	token := creds.AccessToken
	if creds.Expired && creds.RefreshToken != "" {
		refreshed, refreshErr := usageapi.RefreshToken(creds.RefreshToken)
		if refreshErr != nil {
			errlog.Log("usage API token refresh: %v", refreshErr)
		} else {
			token = refreshed.AccessToken
		}
	}

	client := usageapi.NewCachedClient(usageapi.NewClient(token))
	return client.Fetch
}

func defaultCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}
