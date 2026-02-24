# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

cc-statusline is a Go CLI tool that provides a configurable statusline for Claude Code. Built with Cobra (CLI framework), no Viper. Licensed under Apache 2.0.

## Build & Development Commands

```bash
make build          # Build binary to build/bin/cc-statusline
make test           # Run all tests with race detector
make test-pkg PKG=./internal/model  # Test a specific package
make test-coverage  # Tests with coverage report (coverage.out)
make lint           # Run golangci-lint
make lint-fix       # Run golangci-lint with auto-fix
make fmt            # Format code (gofmt + goimports)
make check          # Quick pre-commit check (lint + test)
make ci             # Full CI pipeline (lint + test + build + license check)
make clean          # Remove build artifacts and cache
```

## Architecture

- `cmd/cc-statusline/main.go` -- Entry point, injects version/commit ldflags, delegates to `cmd.Execute()`
- `cmd/root.go` -- Root Cobra command: loads config, sets color, calls `statusline.RunWithConfig()`
- `cmd/install.go` / `cmd/uninstall.go` -- Manage `~/.claude/settings.json` statusLine entry
- `cmd/version.go` -- Print version and commit hash
- `internal/statusline/` -- Core `Run()` / `RunWithConfig()` entry points (all tests exercise these)
- `internal/model/` -- Go structs for stdin JSON payload + `ReadStatus()` parser
- `internal/config/` -- YAML config loader (`os.ReadFile` + `yaml.v3`), defaults, CLI flag overrides
- `internal/color/` -- ANSI escape codes, `Colorize()`, `NO_COLOR` support (thread-safe via atomic)
- `internal/render/` -- `Segment` interface, `Renderer` pipeline (iterates lines, joins segments)
- `internal/render/segments/` -- 16 segment implementations: 10 stable (cwd, git_branch, model, context, cost, duration, tokens, lines, vim, agent), 3 experimental:jsonl (daily_cost, burn_rate, model_breakdown), 3 experimental:usage_api (five_hour, weekly_limits, extra_usage)
- `internal/render/theme/` -- Named themes (tokyo-night, rose-pine, catppuccin)
- `internal/git/` -- `Branch()` helper via subprocess with 500ms timeout
- `internal/cache/` -- File-based caching with TTL + mtime invalidation, atomic writes
- `internal/jsonl/` -- JSONL transcript reader with dedup, file discovery
- `internal/usageapi/` -- OAuth credential reader (env → keychain → file → manual), token refresh, HTTP client with retry, cached client (30s TTL, 5min grace)
- `internal/pricing/` -- LiteLLM pricing data fetcher with embedded fallback
- `internal/errlog/` -- Error logging for experimental features (append-only, 1MB rotate)
- `cmd/auth.go` -- `cc-statusline auth` subcommand for manual token management
- Config file: `~/.cc-statusline.yaml` (or `--config` flag)

## Code Style

Follows the **Uber Go Style Guide** strictly, enforced by golangci-lint v2 with 30+ linters. Key constraints:

- Max cyclomatic complexity: 15
- Max cognitive complexity: 30
- Max function length: 100 lines / 50 statements
- Max line length: 150 characters
- Max nesting depth: 4
- Naked returns only in functions ≤5 lines
- All `nolint` directives require an explanation and specific linter name
- Import ordering: stdlib → third-party → `github.com/donaldgifford` (enforced by gci)
- Run `make fmt` before `make lint` -- gci formatting issues are fixed by goimports

## Testing Patterns

- Tests that mutate global color state (`color.SetEnabled`) must NOT use `t.Parallel()`
- Tests that use `t.Setenv()` must NOT use `t.Parallel()` (Go runtime panics)
- Use `disableColor(t)` helper pattern for tests expecting uncolored output
- Integration tests in `internal/statusline/integration_test.go` build the binary and pipe JSON via `exec.CommandContext`
- Test fixtures in `testdata/` (basic.json, minimal.json, nulls.json, malformed.json, empty.json; JSONL fixtures for experimental tests)
- Tests needing cache on macOS must create `$HOME/Library/Caches` in temp dir (for `os.UserCacheDir()`)
- `errcheck` with `check-blank: true` means `_ = fn()` is still flagged; use bare call with `//nolint:errcheck` or handle the error

## Experimental Feature Architecture

- Segments declare their source via `Source()`: `"stable"`, `"experimental:jsonl"`, or `"experimental:usage_api"`
- The renderer skips segments whose source isn't enabled via config/flags
- On render error, experimental segments show `err` (dim red); stable segments are silently skipped
- Usage API segments receive a `UsageFetcher` function via `segments.AllConfig` (dependency injection)
- JSONL segments discover transcript files via `$CLAUDE_CONFIG_DIR` or `~/.claude/projects/`

## Tool Versions (mise)

Go 1.25.7, golangci-lint 2.8.0. Full tool versions managed via `mise.toml`.

## CI/CD

- **CI** (`ci.yml`): lint → test with coverage (codecov) → goreleaser snapshot build → auto-labeling
- **Release** (`release.yml`): semver via PR labels (`major`/`minor`/`patch`/`dont-release`), GPG-signed, multi-platform (Linux/macOS, amd64/arm64)
- **PR labels** (`pr-labels.yml`): requires exactly one semver label
- **License check** (`license-check.yml`): allowed licenses -- Apache-2.0, MIT, BSD-2-Clause, BSD-3-Clause, ISC, MPL-2.0
