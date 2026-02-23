# Implementation Plan

Detailed implementation plan for cc-statusline, derived from
[DESIGN.md](DESIGN.md). Each phase is a shippable increment. Tasks within a
phase can be worked in roughly the listed order, though some are parallelizable.

References: [DATA_SOURCES.md](DATA_SOURCES.md),
[EXPERIMENTAL.md](EXPERIMENTAL.md)

### Resolved Decisions

Decisions from the open questions review, referenced throughout the plan:

| #   | Decision                                                                                                    | Rationale                                                                                                                                                                                                                                                                                                                                                                 |
| --- | ----------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D1  | **Drop Viper, keep Cobra.**                                                                                 | Viper pulls 10+ transitive deps for functionality replaceable with `os.ReadFile` + `gopkg.in/yaml.v3` (~30 lines). Cobra adds only 2 deps (`pflag`, `mousetrap`) and earns its keep for subcommand dispatch, persistent flags, and help generation. Net effect: `go.sum` drops from ~54 lines to ~12.                                                                     |
| D2  | **List-based segment ordering with global separator.**                                                      | Simpler to implement and validate than format strings. Fewer imported packages (no template/parsing library). Users control order via the `segments` list and customize the separator globally.                                                                                                                                                                           |
| D3  | **Both single-line and multi-line in Phase 1.**                                                             | Claude Code supports multi-line (each `println` = new status bar row). Config uses a list-of-lists: `lines: [[cwd, git_branch, model, context], [cost, duration]]`. Single list = single line (common case).                                                                                                                                                              |
| D4  | **Read OAuth token from Claude Code's existing credential store.**                                          | Claude Code already stores an OAuth token in the macOS Keychain (`Claude Code-credentials`) or `~/.claude/.credentials.json` on Linux. We read it directly -- no copy-paste needed. If the token is expired, attempt a refresh via `POST https://console.anthropic.com/v1/oauth/token`. Fallback: `cc-statusline auth` for manual paste or `CC_STATUSLINE_TOKEN` env var. |
| D5  | **Follow ccusage's JSONL approach: additive optionals, silent skip.**                                       | ccusage uses no version detection or schema migrations. All new fields are optional. Invalid lines are silently skipped. Only `timestamp`, `message.usage.input_tokens`, and `message.usage.output_tokens` are truly required. Mirror this pattern.                                                                                                                       |
| D6  | **Follow ccusage for pricing data. Bundle trimmed Anthropic-only pricing (~5KB), check for updates daily.** | Fetch the full LiteLLM pricing JSON, extract only Claude/Anthropic models, cache locally with 24-hour TTL. Bundle a trimmed fallback via `//go:embed` (~5KB vs ~500KB full). Daily check avoids stale pricing without hammering the source.                                                                                                                               |
| D7  | **Default segments: `cwd`, `git_branch`, `model`, `context`.**                                              | Matches the current bash statusline: `~/code/cc-statusline (feat/clauderino) [Opus 4.6] ctx:43%`. Users add more via config.                                                                                                                                                                                                                                              |
| D8  | **Ship named color themes.**                                                                                | Include standard themes: Tokyo Night, Rose Pine, Catppuccin. One is the default. Users select via `theme:` in config. Per-segment color override also supported.                                                                                                                                                                                                          |
| D9  | **`install` writes the binary's absolute path.**                                                            | More reliable than relying on PATH. If the binary moves, user re-runs `install`.                                                                                                                                                                                                                                                                                          |
| D10 | **Token storage: `~/.config/cc-statusline/` with `0600` perms.**                                            | XDG spec says `~/.cache/` is disposable (deleting it should not lose data). Auth tokens are not disposable. `~/.config/` is the pragmatic industry default used by `gh`, `gcloud`, `terraform`, etc. Primary source is the OS keyring (reading Claude Code's existing credentials), so the config fallback is rarely needed.                                              |

---

## Phase 0: Project Foundation

Fix structural issues in the scaffold, remove Viper, and establish the package
layout that all subsequent phases build on. No user-facing features yet.

### Tasks

- [x] **0.1: Fix main.go location.** The goreleaser config
      (`./cmd/cc-statusline`) and Makefile (`./cmd/$(PROJECT_NAME)`) expect the
      entry point at `cmd/cc-statusline/main.go`, but it currently lives at the
      repo root. Move `main.go` to `cmd/cc-statusline/main.go` and update
      imports. Verify `make build` and `goreleaser check` still work.
- [x] **0.2: Remove Viper, keep Cobra. (D1)** Remove `github.com/spf13/viper`
      from `go.mod`. Delete the `initConfig()` function and
      `cobra.OnInitialize(initConfig)` call. Replace with a minimal config
      loader using `os.ReadFile` + `gopkg.in/yaml.v3` (see task 1.11). Run
      `go mod tidy` to drop Viper's transitive dependencies. Verify `go.sum`
      shrinks from ~54 lines to ~12.
- [x] **0.3: Clean up Cobra scaffold.** Replace the boilerplate descriptions in
      `cmd/root.go` (the Cobra-generated placeholder text). Remove the unused
      `toggle` flag. Update the `Short` and `Long` descriptions to describe
      cc-statusline. Keep the `--config` persistent flag (re-implement without
      Viper in 0.2).
- [x] **0.4: Establish internal package layout.** Create the directory structure
      for internal packages:
  - `internal/model/` -- Go structs for the stdin JSON payload.
  - `internal/render/` -- Segment rendering and output formatting.
  - `internal/render/segments/` -- Individual segment implementations.
  - `internal/render/theme/` -- Named color themes.
  - `internal/config/` -- Configuration loading (stdlib YAML).
  - `internal/cache/` -- File-based caching (Phase 2+).
  - `internal/jsonl/` -- JSONL transcript parsing (Phase 2).
  - `internal/usageapi/` -- OAuth usage API client (Phase 3).
  - `internal/git/` -- Git subprocess helpers.
  - `internal/color/` -- ANSI escape code constants and helpers.
  - `internal/errlog/` -- Error logging for experimental features (Phase 2+).
- [x] **0.5: Add testable entry point.** Create a
      `Run(in io.Reader, out io.Writer, errOut io.Writer, cfg *config.Config) error`
      function in an internal package that the Cobra root command calls with
      `cmd.InOrStdin()` and `cmd.OutOrStdout()`. This is the core function that
      all tests exercise. `main()` and Cobra are thin wrappers.
- [x] **0.6: Add test fixtures.** Create `testdata/` directory with sample stdin
      JSON payloads:
  - `testdata/basic.json` -- All fields populated.
  - `testdata/minimal.json` -- Only required fields.
  - `testdata/nulls.json` -- Null `current_usage`, missing `vim` and `agent`.
  - `testdata/malformed.json` -- Invalid JSON.
  - `testdata/empty.json` -- Empty object `{}`.
- [x] **0.7: Verify CI.** Confirm `make ci` passes (lint + test + build +
      license check) with the restructured project. Fix any linter issues from
      the scaffold cleanup.

### Success Criteria

- `make build` produces a binary at `build/bin/cc-statusline`.
- `goreleaser check` passes.
- `go.sum` has ~12 lines (down from ~54 after Viper removal).
- `echo '{}' | ./build/bin/cc-statusline` runs without error (outputs nothing or
  a minimal default).
- `make ci` passes.
- Package layout exists and compiles (packages may contain only `doc.go`
  placeholder files).

---

## Phase 1: Core Statusline

Ship a working statusline that reads the stdin JSON payload and outputs
formatted text. All data comes from stable, documented sources. No experimental
features. Default output matches the current bash statusline:

```
~/code/cc-statusline (feat/clauderino) [Opus 4.6] ctx:43%
```

### Tasks

#### 1A: Stdin Parsing

- [x] **1.1: Define stdin JSON types.** Create Go structs in
      `internal/model/status.go` matching the full stdin JSON schema from
      [DATA_SOURCES.md](DATA_SOURCES.md). Use `*int` / `*float64` for nullable
      fields (`used_percentage`, `remaining_percentage`). Use pointer-to-struct
      for conditionally absent objects (`vim`, `agent`, `current_usage`). Add
      `json` struct tags for all fields.
- [x] **1.2: Implement stdin reader.** In `internal/model/reader.go`: read JSON
      from `io.Reader` using `json.NewDecoder`. Handle: valid JSON, empty stdin
      (return zero-value struct, not error), malformed JSON (return error).
      Return typed `*StatusData` or error. Use stdlib `encoding/json` only.
- [x] **1.3: Test stdin parsing.** Table-driven tests using the fixtures from
      0.6. Cover: all fields present, missing optional fields, null values,
      malformed input, empty input, unknown fields silently ignored (forward
      compatibility with future Claude Code versions).

#### 1B: Color & Themes

- [x] **1.4: Implement ANSI color helpers.** In `internal/color/color.go`:
      define constants for ANSI escape codes (reset, bold, dim, 16 standard
      colors for fg/bg). Implement
      `Colorize(text string, codes ...string) string` helper that wraps text in
      escape codes + reset. Implement `Enabled() bool` that checks `--no-color`
      flag and `NO_COLOR` env var. When disabled, `Colorize` returns the text
      unmodified. Zero external dependencies.
- [x] **1.5: Implement named themes. (D8)** In `internal/render/theme/`: define
      a `Theme` struct mapping segment names to fg/bg color codes. Implement
      built-in themes:
  - `tokyo-night` (default) -- Cool blues and purples.
  - `rose-pine` -- Muted pinks and golds.
  - `catppuccin` -- Pastel palette. Config selects theme via
    `theme: tokyo-night`. Per-segment color overrides in config take precedence
    over theme.
- [x] **1.6: Test color and theme logic.** Test: colorize with and without
      `NO_COLOR`, theme loading by name, per-segment color override precedence,
      unknown theme name falls back to default.

#### 1C: Segment Rendering

- [x] **1.7: Define the segment interface.** A segment is a unit of statusline
      output:

  ```go
  type Segment interface {
      Name() string
      Source() string // "stable" or "experimental:jsonl" or "experimental:usage_api"
      Render(data *model.StatusData, theme *theme.Theme) (string, error)
  }
  ```

  Segments with `experimental:*` source are skipped unless the corresponding
  flag is enabled. A segment returning empty string is omitted from output.

- [x] **1.8: Implement stable segments.** One segment per file under
      `internal/render/segments/`:
  - `cwd.go` -- Current working directory, abbreviated with `~` for home dir.
    Colored per theme.
  - `git_branch.go` -- Git branch wrapped in parens: `(feat/clauderino)`.
    Colored per theme (default: magenta/pink).
  - `model.go` -- Model display name in brackets: `[Opus 4.6]`. Colored per
    theme.
  - `context.go` -- Context window usage: `ctx:43%`. Color-coded by threshold
    (green <50%, yellow 50-80%, red >80%). Handles null `used_percentage` (show
    `ctx:--`).
  - `cost.go` -- Session cost: `$0.23`. Colored per theme.
  - `duration.go` -- Session duration as human-readable: `12m`, `1h23m`.
    Computed from `cost.total_duration_ms`.
  - `tokens.go` -- Input/output token counts: `15k/4k tokens`. Off by default.
  - `lines.go` -- Lines added/removed: `+156 -23`. Green for added, red for
    removed.
  - `vim.go` -- Vim mode: `NORMAL` or `INSERT`. Only renders when `vim` is
    present in the JSON.
  - `agent.go` -- Agent name. Only renders when `agent` is present in the JSON.
- [x] **1.9: Implement git segment helper.** In `internal/git/git.go`: run
      `git rev-parse --abbrev-ref HEAD` as a subprocess with the `cwd` from the
      stdin JSON as working directory. Timeout at 500ms via
      `context.WithTimeout`. Return branch name or empty string on
      error/timeout/not-a-repo.
- [x] **1.10: Implement the rendering pipeline.** In
      `internal/render/renderer.go`: take config (segment list per line,
      separator, theme) and `*StatusData`. For each configured line, iterate
      segments, call `Render()`, collect non-empty results, join with separator.
      Write each line as a separate `fmt.Fprintln(out, ...)`. Skip experimental
      segments if their flag is not enabled.
- [x] **1.11: Test segments individually.** Table-driven tests for each segment.
      Edge cases: zero cost, null percentages, missing vim/agent, very long
      paths (>80 chars), zero-length durations, empty branch name,
      `used_percentage` at 0/50/80/100.
- [x] **1.12: Test the full rendering pipeline.** End-to-end tests: JSON in,
      formatted string out via the `Run()` function. Test single-line and
      multi-line configs. Test segment ordering matches config order.

#### 1D: Configuration

- [x] **1.13: Implement config loader. (D1, D2, D3)** In
      `internal/config/config.go`: define the `Config` struct:

  ```yaml
  # ~/.cc-statusline.yaml
  theme: tokyo-night
  separator: " "
  lines:
    - [cwd, git_branch, model, context] # line 1
  # Single list shorthand (equivalent to one line):
  # segments: [cwd, git_branch, model, context]
  color: true
  experimental:
    jsonl: false
    usage_api: false
  ```

  Load with `os.ReadFile` + `yaml.Unmarshal`. Return defaults when file doesn't
  exist. Support both `lines:` (list-of-lists for multi-line) and `segments:`
  (flat list shorthand for single-line). CLI flags `--no-color`,
  `--experimental-jsonl`, `--experimental-usage-api` override config values.
  Explicit env var support: `NO_COLOR=1` disables color.

- [x] **1.14: Test config loading.** Test: defaults when no file, file with all
      fields, file with `segments:` shorthand vs `lines:`, CLI flag overrides,
      `NO_COLOR` env var, missing file (no error), malformed YAML (error),
      unknown segment name in list (warn and skip).

#### 1E: CLI Commands

- [x] **1.15: Implement the default (root) command.** Cobra root command's
      `RunE`: load config, read stdin, parse JSON, build segment list, render,
      print to stdout. This is the primary execution path invoked by Claude
      Code.
- [x] **1.16: Implement `install` subcommand. (D9)** Resolve the binary's
      absolute path via `os.Executable()` + `filepath.EvalSymlinks()`. Read
      `~/.claude/settings.json` (or `$CLAUDE_CONFIG_DIR/settings.json`). Merge
      the `statusLine` entry using `encoding/json` (unmarshal to
      `map[string]any`, set key, marshal back with indent). Handle: file doesn't
      exist (create), file exists with other keys (preserve them), file has a
      different `statusLine` (warn to stderr, overwrite). Write:

  ```json
  {
    "statusLine": {
      "type": "command",
      "command": "/absolute/path/to/cc-statusline",
      "padding": 0
    }
  }
  ```

- [x] **1.17: Implement `uninstall` subcommand.** Read
      `~/.claude/settings.json`, delete the `statusLine` key, write back. If
      file doesn't exist or key is absent, no-op with a message.
- [x] **1.18: Implement `version` subcommand.** Print version and commit hash
      (injected via `-ldflags` at build time). Format:
      `cc-statusline version v0.1.0 (abc1234)`.
- [x] **1.19: Test CLI commands.** Test `install` and `uninstall` against a temp
      directory (inject config dir path, don't touch real `~/.claude/`). Test:
      fresh install, install over existing settings, install over existing
      statusLine, uninstall, uninstall when not installed. Test `version` output
      format.

#### 1F: Documentation & Release

- [x] **1.20: Write README.md.** Installation instructions (goreleaser binaries,
      `go install`), quick start (`cc-statusline install`, what it looks like),
      configuration reference (theme, segments, lines, separator), segment list
      with descriptions, available themes.
- [x] **1.21: Add integration test.** Build the binary with `go build`, pipe
      test JSON via `exec.Command`, assert formatted output on stdout. This
      catches issues that unit tests miss (ldflags, Cobra wiring, stdin/stdout
      plumbing, config file resolution).
- [x] **1.22: Tag and release v0.1.0.** Verify `make release-local` produces
      working binaries for all platforms. Create the release via the existing
      goreleaser pipeline.

### Success Criteria

- Default output (no config file) matches the reference layout:
  `~/code/cc-statusline (feat/clauderino) [Opus 4.6] ctx:43%` with themed
  colors.
- `cc-statusline install` writes the correct entry with the absolute binary path
  to `~/.claude/settings.json`.
- `cc-statusline uninstall` removes the entry cleanly.
- Multi-line config works: two `lines:` entries produce two status bar rows.
- `--no-color` and `NO_COLOR=1` produce output without ANSI escape codes.
- Configuration file controls theme, segment order/lines, and separator.
- `make ci` passes. Test coverage meets the 60% target.
- `make release-local` produces working multi-platform binaries.
- `go.sum` contains ~12 lines (Cobra + yaml.v3 + their minimal transitive deps).

---

## Phase 2: JSONL Transcript Parsing (Experimental)

Add segments that derive data from Claude Code's local JSONL transcript files.
All behind `--experimental-jsonl`. Follow ccusage's approach for schema handling
and deduplication (D5, D6).

### Tasks

#### 2A: Experimental Infrastructure

- [x] **2.1: Implement experimental flag gating.** When building the segment
      list, check each segment's `Source()`. Skip segments tagged
      `experimental:jsonl` unless `config.Experimental.JSONL` is true. The
      rendering pipeline should not call `Render()` on gated segments.
- [x] **2.2: Implement error logging.** In `internal/errlog/`: write to
      `~/.cache/cc-statusline/error.log`. Append-only with ISO 8601 timestamps.
      Rotate when file exceeds 1MB (truncate and start fresh). Create cache dir
      with `0700`, log file with `0600`. Used by experimental features to log
      errors without polluting stdout.
- [x] **2.3: Implement the `err` fallback.** When an experimental segment's
      `Render()` returns an error: substitute `err` (dim red) in the segment's
      position. Log the full error via 2.2. The rest of the statusline renders
      normally.

#### 2B: JSONL Parsing

- [x] **2.4: Implement JSONL reader. (D5)** In `internal/jsonl/reader.go`: open
      a JSONL file, stream-parse lines using `bufio.Scanner` + `json.Unmarshal`.
      Define Go structs for the JSONL entry schema. Required fields: `timestamp`
      (ISO 8601), `message.usage.input_tokens`, `message.usage.output_tokens`.
      All other fields are optional (pointer types or `omitempty`): `sessionId`,
      `version`, `costUSD`, `requestId`, `message.model`, `message.id`,
      `message.usage.cache_creation_input_tokens`,
      `message.usage.cache_read_input_tokens`, `isApiErrorMessage`, `cwd`.
      Silently skip lines that fail JSON parse or are missing required fields
      (no error accumulation, just `continue`). Deduplicate by hashing
      `message.id` + `requestId`; if either is missing, skip dedup for that
      entry.
- [x] **2.5: Implement transcript file discovery.** Primary: use
      `transcript_path` from the stdin JSON (points to the current session's
      file). For daily aggregation, scan directories in order:
      `$CLAUDE_CONFIG_DIR/projects/`, `$XDG_CONFIG_HOME/claude/projects/`,
      `~/.claude/projects/`. Use `filepath.Glob` with `**/*.jsonl` pattern.
- [x] **2.6: Implement file-based caching.** In `internal/cache/cache.go`:
      JSON-file cache at `~/.cache/cc-statusline/`. Each entry stores:
      serialized data, expiry timestamp, source file mtime at cache time.
      `Get(key, sourcePath)` returns data if not expired and source file mtime
      hasn't changed. `Set(key, data, ttl, sourceMtime)` writes atomically
      (`os.CreateTemp` + `os.Rename`). Dir: `0700`, files: `0600`.
- [x] **2.7: Implement pricing data fetcher. (D6)** Fetch the full LiteLLM
      pricing JSON from
      `raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json`.
      Filter to only Claude/Anthropic models and cache the trimmed result at
      `~/.cache/cc-statusline/pricing.json` with 24-hour TTL (daily refresh
      cycle). Bundle a pre-trimmed fallback (~5KB) in the binary via
      `//go:embed internal/pricing/fallback.json`. On fetch failure, use local
      cache if present, else bundled fallback. Log when using stale/bundled
      pricing.
- [x] **2.8: Test JSONL parsing.** Create JSONL test fixtures in `testdata/`.
      Test: valid entries, malformed lines (skipped gracefully), duplicate
      entries (deduplicated), empty file, entries from multiple sessions,
      timestamp ordering.
- [x] **2.9: Test caching.** Test: cache miss (first read), cache hit (within
      TTL), cache expiry (past TTL), cache invalidation (source file mtime
      changed), atomic write (no partial reads), directory creation on first
      use.

#### 2C: Experimental Segments

- [x] **2.10: Implement daily cost segment.** Aggregate `costUSD` across all
      JSONL entries from today (UTC). Display as `$X.XX today`. Cache with
      5-second TTL, invalidated by transcript file mtime change.
- [x] **2.11: Implement burn rate segment.** Calculate tokens/minute and
      cost/hour from the current 5-hour block. A block starts when there's a gap
      of >5 hours between entries. Display as `$X.XX/hr`. Cache with 5-second
      TTL.
- [x] **2.12: Implement per-model breakdown segment.** Group session costs by
      model ID. Display as `opus:$X.XX sonnet:$X.XX`. Shorten model IDs to
      display names.
- [x] **2.13: Test experimental segments.** Table-driven tests with mock JSONL
      data. Test: segments omitted when flag disabled, `err` displayed on parse
      failure, correct aggregation math (including deduplication), cache
      hit/miss behavior, multiple models in breakdown.

#### 2D: Documentation

- [ ] **2.14: Update README with experimental section.** Document how to enable
      `--experimental-jsonl`, what segments it adds, what the risks are (link to
      EXPERIMENTAL.md).

### Success Criteria

- With `--experimental-jsonl` disabled (default): output is identical to
  Phase 1. No JSONL files are read. No cache files created.
- With `--experimental-jsonl` enabled: additional segments appear showing daily
  cost, burn rate, per-model breakdown.
- If JSONL parsing fails: affected segments show `err`, rest of statusline is
  unaffected, error logged to `~/.cache/cc-statusline/error.log`.
- Cache works: consecutive rapid invocations (~300ms apart) reuse cached data
  instead of re-parsing.
- Pricing data fetched once and cached for 24 hours. Bundled fallback used when
  offline.
- `make ci` passes. New code covered by tests.

---

## Phase 3: OAuth Usage API (Experimental)

Add segments that display live usage limits from Anthropic's undocumented OAuth
endpoint. All behind `--experimental-usage-api`.

### Tasks

#### 3A: Prototyping

- [ ] **3.1: Prototype the API call.** Write a standalone Go program in
      `cmd/usage-probe/` (not shipped in release) that reads the OAuth token
      from the macOS Keychain or `~/.claude/.credentials.json`, calls
      `GET https://api.anthropic.com/api/oauth/usage` with
      `Authorization: Bearer <token>` and `anthropic-beta: oauth-2025-04-20`
      headers, prints the raw response, and validates it against the known
      schema. Use this to confirm behavior on a real account. Document any
      discrepancies in [DATA_SOURCES.md](DATA_SOURCES.md).

#### 3B: Auth & Token Management

- [ ] **3.2: Implement credential reader. (D4, D10)** In
      `internal/usageapi/auth.go`: read the OAuth token from Claude Code's
      existing credential store. Priority order:
  1. `CC_STATUSLINE_TOKEN` env var (explicit override).
  2. macOS Keychain: run
     `security find-generic-password -s "Claude Code-credentials" -w`, parse the
     JSON, extract `claudeAiOauth.accessToken`. Check `expiresAt` against
     `time.Now()`.
  3. Linux file: read `~/.claude/.credentials.json`, same JSON structure and
     extraction.
  4. Manual fallback: read from `~/.config/cc-statusline/auth.json` (written by
     `cc-statusline auth`). Return the token string and expiry status, or a
     structured error indicating which sources were tried and why each failed.
- [ ] **3.3: Implement token refresh.** If the token is expired (past
      `expiresAt`): `POST https://console.anthropic.com/v1/oauth/token` with
      `Content-Type: application/json` body
      `{"grant_type": "refresh_token", "refresh_token": "<refreshToken>"}`.
      Parse the new access token from the response. **Do NOT write back to the
      macOS Keychain or `~/.claude/.credentials.json`** -- Claude Code manages
      its own credentials, and writing back causes race conditions. Cache the
      refreshed token at `~/.config/cc-statusline/auth.json` with `0600` perms
      for subsequent cc-statusline invocations. Important: refresh tokens are
      single-use. If refresh fails, log the error and suggest running
      `claude auth` from the terminal to re-authenticate.
- [ ] **3.4: Implement `auth` subcommand.** Add `cc-statusline auth` for cases
      where automatic credential reading fails (no Keychain access, no
      credentials file, broken token). Print which credential sources were
      checked and why they failed. Accept a pasted token via stdin and write to
      `~/.config/cc-statusline/auth.json` with `0600` perms. Also support
      `cc-statusline auth --status` to report: which credential source is
      active, whether the token is expired, and whether the usage API is
      reachable.
- [ ] **3.5: Test credential reader.** Test: env var takes priority, Keychain
      command mocked via interface (accept a `CommandRunner` func so tests don't
      call real `security`), Linux credentials file read from temp dir, expired
      token detected from `expiresAt`, fallback chain order respected. Test
      token refresh with `httptest.NewServer` mocking `console.anthropic.com`.
      Test that refreshed token is NOT written to Keychain.

#### 3C: API Client

- [ ] **3.6: Implement the usage API client.** In `internal/usageapi/client.go`:
      `GET https://api.anthropic.com/api/oauth/usage` with headers
      `Authorization: Bearer <token>` and `anthropic-beta: oauth-2025-04-20`.
      Timeout at 2 seconds via `context.WithTimeout`. Parse response into typed
      structs:

  ```go
  type UsageResponse struct {
      FiveHour       *UsageWindow `json:"five_hour"`
      SevenDay       *UsageWindow `json:"seven_day"`
      SevenDayOpus   *UsageWindow `json:"seven_day_opus"`
      SevenDaySonnet *UsageWindow `json:"seven_day_sonnet"`
      ExtraUsage     *ExtraUsage  `json:"extra_usage"`
  }
  type UsageWindow struct {
      Utilization json.Number `json:"utilization"` // observed as int, float, or string
      ResetsAt    *string     `json:"resets_at"`   // ISO 8601 UTC or null
  }
  type ExtraUsage struct {
      IsEnabled    bool         `json:"is_enabled"`
      MonthlyLimit *int         `json:"monthly_limit"` // cents (divide by 100)
      UsedCredits  *int         `json:"used_credits"`  // cents
      Utilization  *json.Number `json:"utilization"`
  }
  ```

  Parse `Utilization` defensively via `json.Number` -- convert to float64 at
  usage site. On 401/403: log suggestion to run `cc-statusline auth`. On other
  HTTP errors or unexpected response shapes: log raw body for debugging. One
  retry with 500ms backoff on 5xx errors. Use stdlib `net/http` only.

- [ ] **3.7: Implement usage API caching.** Cache API responses at
      `~/.cache/cc-statusline/usage.json` with 30-second TTL. On fetch failure,
      return cached data if within a 5-minute grace period (stale-while-error).
      Track consecutive failures; after 5, log a warning suggesting
      `cc-statusline auth --status`.
- [ ] **3.8: Test the API client.** Use `httptest.NewServer` for canned
      responses. Test: success, 401 (triggers auth suggestion log), 500 (retry
      then fail), timeout, malformed JSON body, `utilization` as int vs float vs
      string, null `resets_at`, null `extra_usage`, `extra_usage.monthly_limit`
      in cents, cache hit within TTL, cache expiry, grace period fallback,
      consecutive failure counter.

#### 3D: Experimental Segments

- [ ] **3.9: Implement 5-hour window segment.** Display time remaining and
      percentage used. Format: `5h: 3h12m left (36%)`. Color-coded by
      utilization: green <50%, yellow 50-80%, red >80%. Compute time remaining
      from `resets_at` minus `time.Now()`. When data unavailable: `5h: err`.
- [ ] **3.10: Implement weekly limits segment.** Display per-model and overall
      percentages with reset times. Format:
      `wk: sonnet 45% (resets Sat 2p) / all 62% (resets Thu 9p)`. Convert
      `resets_at` from UTC to user's local timezone, format as short day + 12h
      time. Use `seven_day_sonnet` and `seven_day` from the response. When data
      unavailable: `wk: err`.
- [ ] **3.11: Implement extra usage segment.** Display spending against limit.
      Format: `extra: $12.50 / $50.00`. Convert `used_credits` and
      `monthly_limit` from cents to dollars. Only renders when
      `extra_usage.is_enabled` is true. When data unavailable: `extra: err`.
- [ ] **3.12: Test usage segments.** Table-driven tests with mock API data.
      Test: segments omitted when flag disabled, `err` on API failure, correct
      percentage formatting, reset time formatting across timezones (use
      `time.FixedZone`), boundary values (0%, 100%, >100%), `extra_usage`
      disabled vs enabled, cents-to-dollars conversion.

#### 3E: Documentation

- [ ] **3.13: Update EXPERIMENTAL.md.** Fill in the auth requirements section
      with the credential reader priority chain. Document the actual response
      schema. Document the token refresh flow and the single-use refresh token
      caveat. Update the failure mode table with real-world observations.
- [ ] **3.14: Update DATA_SOURCES.md.** Add the confirmed response schema.
      Document: the `anthropic-beta` header requirement, `utilization` type
      variance, `extra_usage` cents-vs-dollars, credential storage locations,
      refresh endpoint at `console.anthropic.com` (not `api.anthropic.com`).
- [ ] **3.15: Update README.** Document `--experimental-usage-api`,
      `cc-statusline auth` and `auth --status`, credential auto-detection, what
      segments it enables, what to expect when the API breaks.

### Success Criteria

- With `--experimental-usage-api` disabled (default): no HTTP requests made, no
  credential stores read, output unchanged.
- With `--experimental-usage-api` enabled and valid credentials found: 5-hour
  window, weekly limit, and extra usage segments appear with live data.
- With `--experimental-usage-api` enabled but no credentials: segments show
  `err`, log suggests `cc-statusline auth`.
- Token auto-detection works: reads from Keychain (macOS) or credentials file
  (Linux) without user action.
- Expired tokens are auto-refreshed when refresh token is available.
- On API failure: affected segments show `err`, rest of statusline unaffected,
  error logged.
- API responses cached at 30s TTL. Consecutive invocations within TTL make zero
  HTTP requests.
- `cc-statusline auth --status` reports credential source and token health.
- Auth credentials stored at `~/.config/cc-statusline/auth.json` with `0600`.
  Not logged or printed after setup.
- `make ci` passes. New code covered by tests.

---

## Open Questions

All previously identified questions have been resolved through research and are
captured in the Resolved Decisions table (D1-D10) or incorporated directly into
phase tasks. The following items are observations to monitor during
implementation:

### Monitor During Phase 3

1. **OAuth endpoint rate limiting.** claude-code-meter polls every 30 seconds
   and claude-pulse uses a 60-second cache without hitting rate limits. Our
   30-second cache TTL should be safe, but observe during task 3.1 prototyping.
   If we see 429s, increase the TTL.

2. **Token expiry cadence.** The `expiresAt` field in the credentials JSON tells
   us when the token expires, but the typical lifetime is not documented.
   Observe during prototyping whether it's hours, days, or weeks. This affects
   how often users will encounter refresh flows.

3. **Refresh token single-use behavior.** Refresh tokens are reportedly
   single-use. If our refresh races with Claude Code's own refresh, both could
   fail. The fallback is `claude auth` from the terminal. Monitor whether this
   is a real problem or theoretical.
