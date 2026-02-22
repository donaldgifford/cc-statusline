# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

cc-statusline is a Go CLI tool that provides a statusline for Claude Code. Built with Cobra (CLI framework) and Viper (configuration management). Licensed under Apache 2.0.

## Build & Development Commands

```bash
make build          # Build binary to build/bin/cc-statusline
make test           # Run all tests with race detector
make test-pkg PKG=./pkg/foo  # Test a specific package
make test-coverage  # Tests with coverage report (coverage.out)
make lint           # Run golangci-lint
make lint-fix       # Run golangci-lint with auto-fix
make fmt            # Format code (gofmt + goimports)
make check          # Quick pre-commit check (lint + test)
make ci             # Full CI pipeline (lint + test + build + license check)
make clean          # Remove build artifacts and cache
```

## Architecture

- `main.go` — Entry point, delegates to `cmd.Execute()`
- `cmd/root.go` — Root Cobra command, config initialization via Viper
- Config file: `~/.cc-statusline.yaml` (or `--config` flag)

The build injects version and commit hash via `-ldflags` at compile time.

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

## Tool Versions (mise)

Go 1.25.7, golangci-lint 2.8.0. Full tool versions managed via `mise.toml`.

## CI/CD

- **CI** (`ci.yml`): lint → test with coverage (codecov) → goreleaser snapshot build → auto-labeling
- **Release** (`release.yml`): semver via PR labels (`major`/`minor`/`patch`/`dont-release`), GPG-signed, multi-platform (Linux/macOS, amd64/arm64)
- **PR labels** (`pr-labels.yml`): requires exactly one semver label
- **License check** (`license-check.yml`): allowed licenses — Apache-2.0, MIT, BSD-2-Clause, BSD-3-Clause, ISC, MPL-2.0
