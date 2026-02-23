# cc-statusline

A Go-based statusline for [Claude Code](https://claude.ai/code). Single binary, no npm, no node, minimal dependencies.

## Installation

### From release binaries

Download the latest release from the [releases page](https://github.com/donaldgifford/cc-statusline/releases) and place the binary in your PATH.

### From source

```bash
go install github.com/donaldgifford/cc-statusline/cmd/cc-statusline@latest
```

### Configure Claude Code

```bash
cc-statusline install
```

This writes the binary's absolute path to `~/.claude/settings.json`. To remove:

```bash
cc-statusline uninstall
```

## What it looks like

```
~/code/cc-statusline (feat/clauderino) [Opus 4.6] ctx:43%
```

With color theming applied (tokyo-night default).

## Configuration

Create `~/.cc-statusline.yaml`:

```yaml
# Theme: tokyo-night (default), rose-pine, catppuccin
theme: tokyo-night

# Separator between segments
separator: " "

# Single line (shorthand)
segments: [cwd, git_branch, model, context]

# Multi-line (each list is one status bar row)
# lines:
#   - [cwd, git_branch, model, context]
#   - [cost, duration, lines]

# Disable color
# color: false
```

### Available segments

| Segment | Description | Default |
|---------|-------------|---------|
| `cwd` | Current working directory (~ abbreviated) | yes |
| `git_branch` | Git branch in parentheses | yes |
| `model` | Model name in brackets | yes |
| `context` | Context window usage, color-coded | yes |
| `cost` | Session cost in USD | no |
| `duration` | Session duration (human-readable) | no |
| `tokens` | Input/output token counts | no |
| `lines` | Lines added/removed (green/red) | no |
| `vim` | Vim mode (NORMAL/INSERT) | no |
| `agent` | Agent name | no |

### Themes

- **tokyo-night** (default) - Cool blues and purples
- **rose-pine** - Muted pinks and golds
- **catppuccin** - Pastel palette

### CLI flags

```
--config string              Config file (default ~/.cc-statusline.yaml)
--no-color                   Disable color output
--experimental-jsonl         Enable JSONL transcript parsing segments
--experimental-usage-api     Enable OAuth usage API segments
```

Color is also disabled when `NO_COLOR=1` is set.

## Experimental: JSONL Transcript Parsing

Enable with `--experimental-jsonl` or in config:

```yaml
experimental:
  jsonl: true
```

This adds segments that parse Claude Code's local JSONL transcript files:

| Segment | Description |
|---------|-------------|
| `daily_cost` | Total cost across all sessions today (UTC) |
| `burn_rate` | Cost per hour in the current activity block |
| `model_breakdown` | Per-model cost breakdown (e.g., `opus4.6:$0.50 sonnet4.6:$0.25`) |

These segments read files from `~/.claude/projects/` and cache results with a 5-second TTL. If parsing fails, affected segments show `err` (dim red) while the rest of the statusline renders normally. Errors are logged to `~/.cache/cc-statusline/error.log`.

See [EXPERIMENTAL.md](docs/EXPERIMENTAL.md) for details on risks and stability expectations.

## License

Apache 2.0
