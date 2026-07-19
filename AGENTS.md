# Agent Instructions for Social Poster Project

## Overview

Guidance for AI agents working on the Social Poster project — a Go CLI that publishes content to social platforms from either a Telegram bot inbox (`SOURCE=telegram`, GitHub Actions) or newly added Markdown files in a private Git repo (`SOURCE=github`).

## Development Principles

1. **Modularity**: Keep platforms and core concerns separated (`config`, `github`, `posting`, `state`, `cmd`).
2. **Security**:
   * Credentials load only from environment / `.env` (never hardcode secrets).
   * Do not put tokens on process argv; git auth uses a temporary `.netrc` under a disposable `HOME`.
   * Mask tokens in logs (`truncateToken`).
   * Reject placeholder config values at startup.
3. **Configuration**: All knobs go through `config/config.go` and `.env.example`.
4. **Error Handling**: Fail closed for missing platforms; do not advance the checkpoint when publishing is incomplete. Record successful platform deliveries so retries do not duplicate.
5. **Dependencies**: Prefer the Go standard library; keep module deps minimal.
6. **Git Usage**: Shell out to `git` with validated HTTPS github.com URLs only.

## Code Structure

- `cmd/` — Cobra CLI (`process`, `health`)
- `config/` — configuration loading/validation (`SOURCE=telegram|github`)
- `telegram/` — bot inbox via `getUpdates` (Telegram is the queue; confirm after success)
- `github/` — repository sync and post discovery
- `posting/` — platform APIs and content adapters
- `state/` — checkpoint SHA, delivery ledger, flock
- `Dockerfile` & `docker-compose.yml` — deployment
- `.github/workflows/ci.yml` — CI
- `.github/workflows/telegram-publish.yml` — hourly Telegram → socials

## Adding a platform

1. Add fields to `config/config.go` and `.env.example`
2. Add `HasX()` / include in `ConfiguredPlatforms()`
3. Create `posting/<platform>.go`
4. Wire the job in `posting/publish.go`
5. Add adapter rules in `posting/content.go` if needed
6. Extend tests

## State model

- `STATE_FILE_PATH` — last processed commit SHA
- `DELIVERIES_FILE_PATH` — JSON map of `commit|path|platform -> success`
- Process lock beside the state file prevents concurrent runs

## Testing

```bash
go test ./...
```

CI also runs `go vet`, `govulncheck`, and a Docker image build.

## Deployment

- Compose defaults to continuous mode (`PROCESS_INTERVAL=5m`, `restart: unless-stopped`)
- Taskfile rsync **excludes** `.env`; secrets must already exist on the server

Update this file when architecture or conventions change.
