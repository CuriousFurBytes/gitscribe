# AGENTS.md

## Project Overview

GitScribe is a keyboard-first Git TUI written in Go (`github.com/CuriousFurBytes/gitscribe`). Built on Bubble Tea / Charm. Provides status, diffs, history, commit authoring, PR creation, and optional AI-assisted message generation. Single binary from `main.go` at the module root. Requires Go 1.21+.

## Setup

```bash
go mod download
```

Runtime requirements: `git`, `gh` (for PR creation), an AI CLI command (e.g. `claude`) if AI generation is enabled.

## Build / Test / Lint

```bash
# Build
go build -o gitscribe .

# Test (default — run before every commit)
go test ./...

# Test with race detector (required before PRs)
go test -race ./...

# Vet
go vet ./...
```

**Default verification:** `go build ./... && go test ./... && go vet ./...`

## Code Style

- Standard `gofmt` formatting. No custom linter config.
- Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:` prefixes.
- One sub-package per concern under `internal/`: `app`, `ai`, `config`, `git`, `github`, `execx`, `logger`, `repo`, `templates`, `theme`.

## Testing

- All tests are pure unit tests; no live git repo or network required.
- Test helpers `newReadyTestModel` / `newReadyTestModelWithFiles` live in `internal/app/`.
- Tests live in the same package as the code (`package app`, `package config`, etc.).
- Table-driven tests preferred.
- Run `go test -race ./...` before opening a PR.

## Commit / PR Rules

- Conventional Commits, imperative mood, ≤72 char subject.
- PRs target `main`.
- All tests must pass locally before merging.

## Security & Secrets

- Never commit config files containing API keys.
- Config loaded from `$XDG_CONFIG_HOME/gitscribe/config.toml` and `<repo>/.gitscribe.toml`.
- `ai.output_format` must be `"json"` — the parser hard-rejects other values.
- Do not read or modify `.env*`, `*.pem`, or files under `secrets/`.

## Architecture Notes

- Entry: `main.go` (repo root) → detects repo via `internal/repo`, loads config via `internal/config`, starts Bubble Tea with `internal/app.Model`.
- TUI: `internal/app/model.go` holds `*Model`; `update.go` dispatches; `*_screen.go` / `*_modal.go` render. Screen constants in `screens.go`.
- Git ops: `internal/git/service.go` shells out; `internal/execx/command.go` is the subprocess helper.
- AI: `internal/ai/service.go` builds prompt, shells out to configured command, parses JSON response.
- Config: `internal/config/config.go` — `Defaults()` for all keys, `Validate()` for constraints.
- No CGO, no generated code, no build tags. Self-contained binary.
- Do not add direct imports of `internal/app` from `internal/git`, `internal/ai`, or `internal/config`. Dependencies flow inward only.
