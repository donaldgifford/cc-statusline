# cc-statusline Design Document

## What Is This

cc-statusline is a Go-based CLI tool for customizing the Claude Code statusline. It reads session JSON from stdin (piped by Claude Code), combines it with usage data from local sources, and outputs formatted text to stdout for display in the Claude Code status bar.

The goal is to give users visibility into the information that actually matters for managing their Claude Code usage: how much of their 5-hour session window remains, where they stand against weekly limits, and what they're spending on extra usage.

## Why This Exists

### The Problem

Claude Code's built-in statusline shows basic session info, but the data most users care about -- usage limits, spending, and time remaining -- requires checking the web dashboard or running `/usage` manually. You shouldn't have to interrupt your flow to find out if you're about to hit a wall.

Existing tools solve parts of this. But they're all Node/TypeScript-based, which creates a second problem.

### Why Go Over Node

This tool runs locally, reads sensitive session data, and configures LLM behavior. The trust requirements are high. Go provides meaningfully better tooling for this:

- **Static binaries**: Single signed binary with no runtime dependencies. No `node_modules`, no transitive dependency tree to audit. A Go binary is what it is.
- **Security scanning**: `govulncheck` checks against the Go vulnerability database at the call-graph level. `gosec` does static analysis for security anti-patterns. Both are mature and well-integrated into CI.
- **SBOMs and license compliance**: `go-licenses` and `syft` produce accurate SBOMs from Go's explicit dependency graph. Go modules are content-addressed and reproducible.
- **Signed releases**: `goreleaser` produces signed, multi-platform binaries with checksums. Users can verify provenance before running the tool.
- **Minimal attack surface**: No package manager scripts (`postinstall`), no dynamic requires, no eval chains. The dependency tree is small and auditable.

Node has a well-documented history of supply chain attacks (`event-stream`, `ua-parser-js`, `colors`, `faker`, etc.). For a tool that sits between a user and their LLM sessions, "install this npm package globally and let it run on every Claude Code interaction" is not a posture we're comfortable with. Go doesn't eliminate risk, but it dramatically reduces the surface area.

## Existing Landscape

### ccstatusline (sirmalloc/ccstatusline)

The most feature-rich statusline tool available today. TypeScript/React/Ink. ~4,000+ stars.

**What it does well:**
- 25+ widgets (model, git, tokens, cost, context, block timer, custom commands)
- Powerline rendering with themes
- Multi-line status bars
- Interactive TUI for configuration (React/Ink)
- Manages Claude Code's `settings.json` directly (install/uninstall)

**How it works:**
- Claude Code pipes JSON to stdin; ccstatusline parses it and renders formatted output
- Also reads Claude Code's JSONL transcript files directly (for block timer and detailed token data not in the stdin JSON)
- Runs via `bunx ccstatusline@latest` or `npx ccstatusline@latest`
- User settings stored at `~/.config/ccstatusline/settings.json`

**Limitations:**
- Requires Node.js or Bun runtime
- Every status update spawns a new process (`npx`/`bunx` cold start)
- Depends on parsing Claude Code's internal JSONL transcript format (undocumented, could change)
- Node dependency tree inherits the supply chain concerns discussed above

### ccusage (ryoppippi/ccusage)

Usage analysis CLI. TypeScript. ~10,900 stars.

**What it does:**
- Reads local JSONL session files from `~/.config/claude/projects/` (or `~/.claude/projects/`)
- Calculates token usage and costs by day, week, month, session, or 5-hour block
- Pricing data from LiteLLM's open-source model pricing database (fetched once, cached)
- Has a `statusline` subcommand for compact single-line output
- Supports `--json` and `--jq` for machine-readable output
- Also has an MCP server package (`@ccusage/mcp`) for Claude Desktop integration

**How it gets pricing data:**
- Fetches `model_prices_and_context_window.json` from BerriAI/litellm on GitHub
- Supports tiered pricing (different rates above 200k tokens)
- Three cost modes: `auto` (use JSONL's `costUSD` if available, else calculate), `calculate` (always from tokens), `display` (always from JSONL)

**What it cannot do:**
- Does not know about 5-hour window remaining time or weekly limit percentages (these aren't in the JSONL files)
- Burns through the JSONL files on every invocation (has a semaphore/cache system to mitigate)
- No access to subscription-level data (plan tier, spending cap, balance)

### claude-code-meter (gxjansen/claude-code-meter)

macOS desktop widget. Shows 5-hour and 7-day usage in real time.

**Notable detail:** Uses an undocumented OAuth endpoint (`https://api.anthropic.com/api/oauth/usage`) that returns utilization percentages and reset timestamps. This is the same data Anthropic uses internally for the `/usage` command. Undocumented and could break, but it's the only known source for real-time limit data.

### claude_monitor_statusline (gabriel-dehan/claude_monitor_statusline)

Usage monitoring statusline with plan-aware rate limit tracking.

### starship-claude (martinemde/starship-claude)

Pure Bash statusline inspired by Starship prompt. No runtime dependencies beyond Bash and standard Unix tools.

**What it does well:**
- Zero dependencies beyond Bash, `jq`, `git`, and `bc`
- Tested with [bats](https://github.com/bats-core/bats-core) (Bash Automated Testing System) -- a good model for shell testing
- Lightweight and fast -- no interpreter cold start
- Shows model, cost, context percentage, git info, session duration

**Limitations:**
- Pure Bash imposes a ceiling on complexity. Adding structured caching, validated API responses, or configurable output formats means fighting the language rather than working with it.
- Error handling in Bash is opt-in and fragile (`set -e` has well-known edge cases). Graceful degradation for experimental data sources would be difficult to implement reliably.
- No type safety for JSON parsing -- relies on `jq` pipelines that silently produce empty strings on schema changes.
- Testing with bats is better than no tests, but Bash's string-oriented nature makes it hard to test edge cases (malformed JSON, partial API responses, cache races).

Of the existing tools, starship-claude is the closest in philosophy to what we want: useful information, not ricing. But Bash is the wrong language for the direction we're heading (validated API responses, caching with TTLs, experimental feature flags, structured error handling).

### Other Tools

- **claude-powerline** (Owloops): Vim-style powerline statusline
- **ClaudeUsageTracker** (masorange): macOS menu bar app
- **Claude-Code-Usage-Monitor** (Maciek-roboblog): Terminal monitor with ML-based predictions

## What Claude Code Gives Us

### Stdin JSON Payload

Claude Code pipes this JSON to the statusline command on every update (after each assistant message, permission change, or vim mode toggle). Updates are debounced at 300ms.

```json
{
  "session_id": "abc123...",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/working/directory",
  "model": {
    "id": "claude-opus-4-6",
    "display_name": "Opus"
  },
  "workspace": {
    "current_dir": "/current/working/directory",
    "project_dir": "/original/project/directory"
  },
  "version": "1.0.88",
  "output_style": { "name": "default" },
  "cost": {
    "total_cost_usd": 0.01234,
    "total_duration_ms": 45000,
    "total_api_duration_ms": 2300,
    "total_lines_added": 156,
    "total_lines_removed": 23
  },
  "context_window": {
    "total_input_tokens": 15234,
    "total_output_tokens": 4521,
    "context_window_size": 200000,
    "used_percentage": 8,
    "remaining_percentage": 92,
    "current_usage": {
      "input_tokens": 8500,
      "output_tokens": 1200,
      "cache_creation_input_tokens": 5000,
      "cache_read_input_tokens": 2000
    }
  },
  "exceeds_200k_tokens": false,
  "vim": { "mode": "NORMAL" },
  "agent": { "name": "security-reviewer" }
}
```

**What's in the payload:** model info, session cost, context window usage, token counts, git/workspace info, vim mode, session duration, lines changed.

**What's NOT in the payload:** 5-hour window remaining, weekly limit percentages, extra usage spending/balance, plan tier, reset timestamps.

### Local JSONL Transcript Files

Claude Code writes session transcripts as JSONL files to `~/.config/claude/projects/` (or `~/.claude/projects/`). Each line contains a timestamped entry with token counts, model, cost, and request ID. This is what ccusage parses. The format is undocumented and internal.

### The Data Gap

The information users most want to see -- session window remaining, weekly limit status, spending -- is not available through any official, documented API for individual users. The known sources:

| Data | Source | Status |
|------|--------|--------|
| Session cost, tokens, context | Stdin JSON payload | Official, stable |
| Historical token usage, costs | Local JSONL transcripts | Undocumented internal format |
| Model pricing | LiteLLM pricing DB | Third-party, maintained |
| 5-hour window %, reset time | OAuth usage API | Undocumented, could break |
| Weekly limit %, reset time | OAuth usage API | Undocumented, could break |
| Extra usage balance, spending cap | Unknown | No known programmatic source |
| Per-model limit breakdown | OAuth usage API (partial) | Undocumented |

There is an open feature request ([anthropics/claude-code#11917](https://github.com/anthropics/claude-code/issues/11917)) to expose usage metrics directly in the session JSON. If implemented, this would close the gap significantly.

## What We Want to Display

### Phase 1: Core (Stdin JSON Data)

Data available directly from the stdin JSON, no external sources needed:

- Model name and ID
- Session cost (USD)
- Session duration
- Context window usage (percentage, tokens)
- Lines added/removed
- Current working directory
- Git branch (from workspace or shell)
- Vim mode (when active)

This gets us to parity with a well-written bash script but in a structured, configurable Go binary.

### Phase 2: Historical Usage (Experimental -- JSONL Parsing)

Parse the local JSONL transcript files (same approach as ccusage):

- Daily/session cost totals
- Token usage trends
- 5-hour block burn rate and projections
- Per-model cost breakdown

**Experimental.** Depends on an undocumented internal file format. Enabled via `--experimental-jsonl` flag or `experimental.jsonl: true` in config. When enabled, segments sourced from JSONL data render normally. When disabled (default), those segments are simply omitted. If parsing fails at runtime, affected segments display `err` instead of silently disappearing. See [docs/EXPERIMENTAL.md](EXPERIMENTAL.md).

### Phase 3: Live Limits (Experimental -- OAuth Usage API)

The data users actually want most.

**Primary approach: OAuth usage API.** The undocumented endpoint at `api.anthropic.com/api/oauth/usage` returns real-time utilization percentages and reset timestamps. claude-code-meter uses this successfully. Enabled via `--experimental-usage-api` flag or `experimental.usage_api: true` in config.

**Fallback: official support.** The feature request in [#11917](https://github.com/anthropics/claude-code/issues/11917) asks for usage metrics in the session JSON. If Anthropic adds this, we migrate the data source transparently -- same output, stable source. The architecture should make this swap trivial.

**Experimental.** Undocumented API that could break without notice. Auth requirements not fully understood. When enabled, segments sourced from the usage API render normally when data is available. On API errors, auth failures, or unexpected response shapes, affected segments display `err` rather than crashing or showing stale data. See [docs/EXPERIMENTAL.md](EXPERIMENTAL.md).

### Desired Statusline Output (Conceptual)

```
Opus | $0.23 | ctx 12% | 5h: 3h12m left (36%) | wk: sonnet 45% (resets Sat 2p) / all 62% (resets Thu 9p)
```

Or multi-line:

```
Opus | $0.23 session | ctx 12% | feat/my-branch
5h: 3h12m left (36%) | wk: 45% sonnet / 62% all | extra: $12.34 / $50.00
```

The exact format should be user-configurable.

## Architecture Considerations

### Process Lifetime

Every statusline update spawns a new process. This is Claude Code's design -- there is no long-running daemon mode for statusline commands. This means:

- **Startup time matters.** Go's fast cold start (~5-10ms) is a significant advantage over Node/Bun. No runtime initialization, no module resolution.
- **State between invocations** must be persisted to disk or a cache file. We can't hold state in memory across updates.
- **Expensive operations** (JSONL parsing, API calls) should be cached with TTLs and invalidated based on file modification times, similar to ccusage's semaphore approach.

### Configuration

- Primary config: `~/.cc-statusline.yaml` (already wired via Viper)
- Claude Code integration: write the `statusLine` entry to `~/.claude/settings.json`
- Should support both a `setup` command (interactive) and direct config file editing

### Output Formatting

- Support ANSI colors (16, 256, truecolor)
- Support multi-line output
- Consider Powerline-style rendering as an optional feature
- Keep the default output simple and fast

### Caching Strategy

For data that doesn't come from stdin (JSONL parsing, API calls):

- File-based cache in a temp directory or `~/.cache/cc-statusline/`
- TTL-based expiry (configurable, default ~1-5 seconds for JSONL, longer for API data)
- File modification time checks to invalidate when transcripts update
- Graceful degradation: if cache is stale and refresh fails, show stale data with an indicator

## Is This Worth Doing

**Yes, with caveats.**

The core value proposition is clear: a single signed Go binary with no runtime dependencies that replaces a Node.js tool for a security-sensitive use case. The Go ecosystem's security tooling (govulncheck, gosec, go-licenses, signed goreleaser builds) is genuinely superior for this class of tool.

The main risk is the data gap. The most valuable features (live limit tracking) depend on either an undocumented API or a not-yet-implemented feature request. Phase 1 is viable today with stable data sources. Phases 2 and 3 use undocumented sources but are gated behind explicit experimental flags -- users opt in with full knowledge that these features may break. When they do break, the tool degrades to showing `err` for affected segments rather than crashing. If Anthropic officially exposes usage data in the session JSON (#11917), the experimental flags become unnecessary and we promote those features to stable. See [docs/EXPERIMENTAL.md](EXPERIMENTAL.md) and [docs/DATA_SOURCES.md](DATA_SOURCES.md).

The ccstatusline project proves there's demand (~4,000+ stars). Building a Go alternative that prioritizes security, auditability, and minimal dependencies fills a real gap -- especially for users and organizations that care about supply chain security.

**Start with Phase 1, ship it, and iterate.**
