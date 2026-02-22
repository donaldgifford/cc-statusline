# Experimental Features

cc-statusline includes features that depend on undocumented or unstable data sources. These are disabled by default and must be explicitly opted into. This document explains what experimental features exist, how to enable them, and what to expect when they break.

## How Experimental Features Work

### Enabling

Experimental features are enabled in two ways:

**CLI flags** (per-invocation):

```bash
cc-statusline --experimental-jsonl --experimental-usage-api
```

**Config file** (`~/.cc-statusline.yaml`, persistent):

```yaml
experimental:
  jsonl: true
  usage_api: true
```

CLI flags override config file values for that invocation.

### Behavior When Things Break

Experimental features follow a strict degradation contract:

1. **When disabled (default):** Statusline segments that depend on experimental data are omitted entirely. The statusline renders only stable data. No errors, no placeholders.

2. **When enabled and working:** Segments render normally, indistinguishable from stable data.

3. **When enabled and broken:** Affected segments display `err` in place of the expected value. The rest of the statusline continues to render normally from stable sources.

   Example of normal output:
   ```
   Opus | $0.23 | ctx 12% | 5h: 3h12m left (36%) | wk: sonnet 45%
   ```

   Example with a broken usage API:
   ```
   Opus | $0.23 | ctx 12% | 5h: err | wk: err
   ```

4. **Errors are logged**, not displayed. Details go to `~/.cache/cc-statusline/error.log` for debugging. The statusline itself stays clean.

5. **No retries** within a single invocation. Each statusline render is a short-lived process. If an experimental data source fails, we show `err` and move on. The next Claude Code update cycle (debounced at 300ms) will try again.

### Caching

Experimental data sources use file-based caching to avoid expensive operations on every statusline update:

- **JSONL parsing:** Cached with a short TTL (~1-5 seconds). Invalidated when the transcript file's modification time changes.
- **Usage API:** Cached with a longer TTL (~30-60 seconds). Usage percentages don't change fast enough to justify per-update API calls.
- **Model pricing:** Cached with a 24-hour TTL. Falls back to a bundled copy if the fetch fails.

Cache location: `~/.cache/cc-statusline/`

## Experimental: JSONL Transcript Parsing

**Flag:** `--experimental-jsonl`
**Config:** `experimental.jsonl: true`
**Phase:** 2

### What It Does

Reads Claude Code's local JSONL transcript files to calculate usage data that isn't available in the stdin JSON payload:

- **Daily cost total:** Aggregate cost across all sessions today.
- **5-hour block burn rate:** Tokens per minute and cost per hour based on the current billing window.
- **5-hour block projection:** Estimated total cost if current rate continues for the full 5-hour window.
- **Per-model cost breakdown:** How much of the session cost came from Opus vs Sonnet vs Haiku.

### Why It's Experimental

Claude Code's JSONL transcript format is an internal implementation detail. It is not documented, not versioned, and not guaranteed to be stable. Anthropic could change the schema, move the file location, or stop writing these files entirely. If this happens:

- JSONL-dependent segments will show `err`.
- We will update cc-statusline to handle the new format if possible.
- If the format becomes unreadable, the feature will be removed and the segments will be omitted.

### Data Source Details

See [DATA_SOURCES.md](DATA_SOURCES.md#local-jsonl-transcript-files-phase-2) for the full schema and file location details.

### Known Risks

- **Schema changes:** A new Claude Code version could change field names, add required fields, or restructure the JSONL entries.
- **File location changes:** The transcript directory could move or be made configurable in ways we don't track.
- **Duplicate entries:** JSONL files can contain duplicate entries across files. We deduplicate by `messageId` + `requestId` hash, following ccusage's approach.
- **Large files:** Active users can accumulate large transcript histories. Parsing is bounded to the current session's file (from `transcript_path` in stdin JSON) by default. Full-history analysis (daily totals) scans all files but is cached aggressively.

## Experimental: OAuth Usage API

**Flag:** `--experimental-usage-api`
**Config:** `experimental.usage_api: true`
**Phase:** 3

### What It Does

Queries Anthropic's internal usage API to display real-time limit information:

- **5-hour window:** Percentage used, time remaining until reset.
- **Weekly limits:** Percentage used for all models and per-model breakdown (e.g., Sonnet-only vs overall). Reset timestamps.

### Why It's Experimental

The endpoint at `https://api.anthropic.com/api/oauth/usage` is undocumented. It is used internally by Anthropic's own tools and has been reverse-engineered by third-party projects (notably claude-code-meter). There are no stability guarantees:

- The endpoint could be removed, moved, or restructured at any time.
- Authentication requirements could change.
- Response format could change without notice.
- Rate limiting behavior is unknown.

### What Happens When It Breaks

| Failure Mode | Behavior |
|---|---|
| HTTP error (4xx, 5xx) | Affected segments show `err`. Error details logged. |
| Auth failure (401, 403) | Segments show `err`. Log message suggests re-authenticating. |
| Unexpected response shape | Segments show `err`. Raw response logged for debugging. |
| Timeout (>2 seconds) | Segments show `err`. Cached data used if available and within TTL. |
| Endpoint removed (persistent failures) | After N consecutive failures, feature auto-disables for the session with a log warning. |

### Auth Requirements

The usage API requires a Bearer token with `user:profile` and `user:inference` scopes, plus the `anthropic-beta: oauth-2025-04-20` header.

cc-statusline reads the token automatically from Claude Code's existing credential store. No separate authentication flow is needed in most cases:

1. **`CC_STATUSLINE_TOKEN` env var** -- explicit override, highest priority.
2. **macOS Keychain** -- reads `Claude Code-credentials` entry, extracts `claudeAiOauth.accessToken`.
3. **Linux credentials file** -- reads `~/.claude/.credentials.json`, same extraction.
4. **Manual fallback** -- reads from `~/.config/cc-statusline/auth.json` (written by `cc-statusline auth`).

If the stored token is expired (checked via `expiresAt` field), cc-statusline attempts to refresh it via `POST https://console.anthropic.com/v1/oauth/token`. Refreshed tokens are cached at `~/.config/cc-statusline/auth.json` with `0600` permissions. They are **not** written back to the Keychain or `~/.claude/.credentials.json` to avoid race conditions with Claude Code's own token management.

If automatic detection and refresh both fail, run `cc-statusline auth` to paste a token manually, or `cc-statusline auth --status` to diagnose which credential sources are available.

**Important:** Refresh tokens are single-use. If cc-statusline's refresh races with Claude Code's own refresh, both may fail. In that case, run `claude auth` from the terminal to re-authenticate Claude Code, which will write fresh credentials to the store.

### Data Source Details

See [DATA_SOURCES.md](DATA_SOURCES.md#oauth-usage-api-phase-3) for endpoint details and expected response format.

## Migration to Stable

If Anthropic implements [claude-code#11917](https://github.com/anthropics/claude-code/issues/11917) and adds usage metrics to the stdin JSON payload, the experimental data sources will be superseded:

1. Features backed by the new official fields will be promoted to stable (always-on, no flag needed).
2. The experimental flags will continue to work but become no-ops for data that is now available from stable sources.
3. After a deprecation period, the experimental code paths for superseded data will be removed.

The goal is for users to never notice the transition. Same segments, same output, better data source.
