# Data Sources

This document describes every data source cc-statusline reads from, what data each provides, and its stability.

## Stable Sources

### Claude Code Stdin JSON

**How it works:** Claude Code pipes a JSON payload to the statusline command's stdin on every update (after each assistant message, permission change, or vim mode toggle). Updates are debounced at 300ms.

**Stability:** Official. This is the documented statusline integration point. The schema has been stable since v1.0.71 with additive changes only.

**Data provided:**

| Field | Path | Type | Notes |
|-------|------|------|-------|
| Model ID | `model.id` | string | e.g., `claude-opus-4-6` |
| Model name | `model.display_name` | string | e.g., `Opus` |
| Session ID | `session_id` | string | UUID |
| Session cost | `cost.total_cost_usd` | float | Cumulative USD for session |
| Session duration | `cost.total_duration_ms` | int | Wall-clock ms since session start |
| API duration | `cost.total_api_duration_ms` | int | Time waiting for API responses |
| Lines added | `cost.total_lines_added` | int | Cumulative |
| Lines removed | `cost.total_lines_removed` | int | Cumulative |
| Context used % | `context_window.used_percentage` | int | Input tokens only; may be null early |
| Context remaining % | `context_window.remaining_percentage` | int | May be null early |
| Context window size | `context_window.context_window_size` | int | 200000 or 1000000 (extended) |
| Total input tokens | `context_window.total_input_tokens` | int | Cumulative across session |
| Total output tokens | `context_window.total_output_tokens` | int | Cumulative across session |
| Current input tokens | `context_window.current_usage.input_tokens` | int | Last API call; null before first call |
| Current output tokens | `context_window.current_usage.output_tokens` | int | Last API call |
| Cache creation tokens | `context_window.current_usage.cache_creation_input_tokens` | int | Last API call |
| Cache read tokens | `context_window.current_usage.cache_read_input_tokens` | int | Last API call |
| Exceeds 200k | `exceeds_200k_tokens` | bool | Whether last response exceeded 200k |
| Working directory | `cwd` | string | Same as `workspace.current_dir` |
| Project directory | `workspace.project_dir` | string | Where Claude Code was launched |
| Transcript path | `transcript_path` | string | Path to session JSONL file |
| Claude Code version | `version` | string | e.g., `1.0.88` |
| Output style | `output_style.name` | string | e.g., `default` |
| Vim mode | `vim.mode` | string | `NORMAL` or `INSERT`; absent when vim mode disabled |
| Agent name | `agent.name` | string | Only present with `--agent` flag |

**Gotchas:**
- `context_window.total_input_tokens` and `total_output_tokens` are cumulative across the session and can exceed the context window size. They are not "current context" values.
- `used_percentage` is calculated from input tokens only (excludes output tokens).
- `current_usage` is null before the first API call.
- `vim` and `agent` are conditionally absent (not null, missing from the JSON entirely).

### Git (Shell)

**How it works:** Standard git commands (`git branch`, `git status`, etc.) executed as subprocesses.

**Stability:** Stable. Git's CLI output formats are well-established.

**Data provided:** Current branch, dirty/clean status, repository root.

## Experimental Sources

### Local JSONL Transcript Files (Phase 2)

**Flag:** `--experimental-jsonl` or `experimental.jsonl: true` in config.

**How it works:** Claude Code writes session transcripts as JSONL files. The path is provided in the stdin JSON (`transcript_path`). Files are also discoverable at:

1. `$CLAUDE_CONFIG_DIR/projects/**/*.jsonl`
2. `$XDG_CONFIG_HOME/claude/projects/**/*.jsonl`
3. `~/.claude/projects/**/*.jsonl`

Each line is a JSON object containing:

```json
{
  "timestamp": "2025-01-01T12:00:00Z",
  "sessionId": "uuid",
  "version": "1.0.88",
  "costUSD": 0.001,
  "requestId": "req-id",
  "message": {
    "model": "claude-sonnet-4-20250514",
    "id": "msg-id",
    "usage": {
      "input_tokens": 50,
      "output_tokens": 10,
      "cache_creation_input_tokens": 25,
      "cache_read_input_tokens": 10
    }
  }
}
```

**Stability:** Undocumented internal format. Could change without notice. Treat defensively.

**Data provided:** Historical token usage, per-request costs, model used per request, session timelines. Enables burn rate calculations and 5-hour block analysis.

**Error behavior:** Parsing failures log a warning and affected statusline segments display `err`.

### Model Pricing Database (Phase 2)

**Flag:** Enabled alongside `--experimental-jsonl` (needed for independent cost calculation).

**How it works:** LiteLLM maintains a public JSON file of model pricing at `https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json`. Fetched on first use, cached locally with a 24-hour TTL. A bundled fallback copy ships with the binary for offline use.

**Stability:** Third-party, actively maintained. Updated frequently as Anthropic changes pricing.

**Data provided:** Per-model input/output token prices, context window sizes, tiered pricing thresholds.

**Error behavior:** If the fetch fails and no cache exists, falls back to the bundled copy. Cost calculations note when using potentially stale pricing.

### OAuth Usage API (Phase 3)

**Flag:** `--experimental-usage-api` or `experimental.usage_api: true` in config.

**How it works:** `GET https://api.anthropic.com/api/oauth/usage` with headers `Authorization: Bearer <token>` and `anthropic-beta: oauth-2025-04-20`. Returns the same data Anthropic uses internally for the `/usage` command. Used by claude-code-meter, claude-pulse, and Claude-Usage-Tracker.

**Stability:** Undocumented beta endpoint. No stability guarantees. Could require auth changes, change response format, or be removed entirely.

**Response schema:**

```json
{
  "five_hour": {
    "utilization": 8.0,
    "resets_at": "2026-02-22T14:00:00Z"
  },
  "seven_day": {
    "utilization": 77.0,
    "resets_at": "2026-02-28T19:00:00Z"
  },
  "seven_day_opus": {
    "utilization": 45.0,
    "resets_at": null
  },
  "seven_day_sonnet": {
    "utilization": 12.0,
    "resets_at": "2026-02-28T19:00:00Z"
  },
  "seven_day_oauth_apps": null,
  "extra_usage": {
    "is_enabled": true,
    "monthly_limit": 5000,
    "used_credits": 1250,
    "utilization": 25.0
  },
  "iguana_necktie": null
}
```

**Data provided:**

| Field | Type | Notes |
|-------|------|-------|
| `five_hour.utilization` | number (0-100+) | Can exceed 100. Observed as int, float, or string -- parse defensively. |
| `five_hour.resets_at` | string or null | ISO 8601 UTC. |
| `seven_day.utilization` | number (0-100+) | Overall weekly limit. |
| `seven_day.resets_at` | string or null | ISO 8601 UTC. |
| `seven_day_opus` | object or null | Opus-specific weekly quota. |
| `seven_day_sonnet` | object or null | Sonnet-specific weekly quota. |
| `extra_usage.is_enabled` | bool | Whether overage billing is active. |
| `extra_usage.monthly_limit` | int or null | **In cents.** Divide by 100 for dollars. |
| `extra_usage.used_credits` | int or null | **In cents.** |
| `extra_usage.utilization` | number or null | Percentage of extra budget used. |
| `iguana_necktie` | null | Unknown/canary field. Always null. Ignore. |

**Auth requirements:**

The token is read from Claude Code's existing credential store (no separate OAuth flow needed):

1. **macOS Keychain:** `security find-generic-password -s "Claude Code-credentials" -w` returns JSON with `claudeAiOauth.accessToken` (prefix `sk-ant-oat01-`), `refreshToken`, `expiresAt` (Unix epoch seconds), and `subscriptionType`.
2. **Linux file:** `~/.claude/.credentials.json` with the same JSON structure.
3. **Env var:** `CC_STATUSLINE_TOKEN` for explicit override.

The `anthropic-beta: oauth-2025-04-20` header is **required** -- without it, the endpoint returns an error.

**Token refresh:** `POST https://console.anthropic.com/v1/oauth/token` (note: `console.anthropic.com`, not `api.anthropic.com`) with body `{"grant_type": "refresh_token", "refresh_token": "<refreshToken>"}`. Refresh tokens are **single-use**. Do not write refreshed tokens back to the Keychain -- Claude Code manages its own credentials.

**Known quirks:**
- Claude Code refreshes tokens in memory but does **not** write them back to the Keychain/file. The stored token may be expired while Claude Code itself works fine.
- The `/login` slash command inside Claude Code handles MCP OAuth, not the main OAuth. Using it can destroy the `claudeAiOauth` entry. The correct re-auth command is `claude auth` from the terminal.
- Tokens require `user:profile` and `user:inference` scopes. Tokens from `setup-token` may only have `user:inference` and will fail.

**Error behavior:** API errors, auth failures, unexpected response shapes, or timeouts all result in affected segments displaying `err`. One retry with 500ms backoff on 5xx errors. Errors are logged to `~/.cache/cc-statusline/error.log` for debugging.

## Data Not Currently Available

| Data | Notes |
|------|-------|
| Plan tier (Pro/Max 5x/Max 20x) | Not exposed in any known API or local data. |
| Exact token allocations per plan | Anthropic does not publish these numbers. |

If [anthropics/claude-code#11917](https://github.com/anthropics/claude-code/issues/11917) is implemented, some or all of the experimental data sources may be superseded by official fields in the stdin JSON payload. The architecture should allow transparent migration when this happens.
