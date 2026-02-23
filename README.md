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
--config string     Config file (default ~/.cc-statusline.yaml)
--no-color          Disable color output
```

Color is also disabled when `NO_COLOR=1` is set.

## License

Apache 2.0
