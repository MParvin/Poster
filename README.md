# Social Poster

Social Poster publishes short posts to configured social platforms (Twitter/X, Mastodon, Dev.to, LinkedIn, optional Telegram channel).

It supports two content sources via `SOURCE`:

| Source | How content arrives | Typical deploy |
|---|---|---|
| `telegram` | People message your Telegram bot | **GitHub Actions** (hourly) |
| `github` | New Markdown files in a private git repo | Docker / long-running process |

## Recommended: Telegram bot + GitHub Actions

### Why not “every 4 hours because Telegram only keeps messages for 4 hours”?

Telegram Bot API keeps **unconfirmed** updates for up to **24 hours**, not 4. Confirmed updates are removed from the queue.

That means Telegram itself acts as the message queue / “database”:

1. Job calls `getUpdates` (pending messages)
2. Publishes each message to configured socials
3. Only then confirms the update (`offset = last_update_id + 1`)

If a publish fails, the update stays pending and the next run retries it. No Postgres/SQLite required.

**Schedule choice:** hourly (`0 * * * *` UTC) is a good default — well inside the 24h window, and better than 4h if many messages arrive (API returns at most 100 updates per call). Use workflow_dispatch for manual runs.

### Setup

1. Create a bot with [@BotFather](https://t.me/BotFather) → copy `TELEGRAM_BOT_TOKEN`
2. Message the bot from your account (or add it to a group and message there)
3. Find your numeric chat id (e.g. temporarily log `getUpdates`, or use a helper bot) → set `TELEGRAM_ALLOWED_CHAT_IDS`
4. In the GitHub repo → **Settings → Secrets and variables → Actions**, add:

| Secret | Required | Purpose |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | yes | Bot token |
| `TELEGRAM_ALLOWED_CHAT_IDS` | yes | Comma-separated numeric chat IDs allowed to submit posts |
| `TWITTER_*` / `MASTODON_*` / `DEV_TO_API_KEY` / `LINKEDIN_*` | at least one set | Outbound platforms |
| `TELEGRAM_CHAT_ID` | only if reposting | Outbound Telegram chat/channel |

Optional repository variable: `TELEGRAM_REPOST=true` to also send to `TELEGRAM_CHAT_ID` (off by default to avoid echo loops).

5. Enable Actions. Workflow file: [`.github/workflows/telegram-publish.yml`](.github/workflows/telegram-publish.yml)
6. Message the bot → wait for the hourly run (or **Actions → Telegram publish → Run workflow**)

### Local dry-run

```bash
cp .env.example .env
# set SOURCE=telegram, token, allowlist, and at least one outbound platform
go run . process --once --dry-run
```

## Alternative: GitHub repo source (`SOURCE=github`)

Monitors a private posts repository for newly **added** Markdown files and publishes them. See `.env.example` for `GITHUB_*` / `POSTS_REPO_*` variables. Docker Compose remains available for this mode.

## Configuration reference

```bash
cp .env.example .env
```

| Variable | Mode | Description |
|---|---|---|
| `SOURCE` | both | `telegram` or `github` (default `github`) |
| `TELEGRAM_BOT_TOKEN` | telegram | Bot token |
| `TELEGRAM_ALLOWED_CHAT_IDS` | telegram | Inbound allowlist (numeric IDs) |
| `TELEGRAM_CHAT_ID` | optional | Outbound Telegram destination |
| `TELEGRAM_REPOST` | telegram | Default `false` — post back to Telegram |
| `STATE_FILE_PATH` / `DELIVERIES_FILE_PATH` | both | Checkpoint + per-platform delivery ledger |
| `DRY_RUN` | both | Log only; do not publish or confirm Telegram updates |

Placeholder/default values are rejected at startup.

## Development

```bash
go test ./...
go run . health
go run . process --once --dry-run
```

## Project structure

- `cmd/` — CLI (`process`, `health`)
- `config/` — env loading/validation
- `telegram/` — bot inbox (`getUpdates` / confirm offset)
- `github/` — repo sync + markdown detection
- `posting/` — platform publishers + content adapters
- `state/` — checkpoint, delivery ledger, flock
- `.github/workflows/telegram-publish.yml` — hourly Telegram → socials job

## Security notes

- Never commit `.env` or real tokens
- Always set `TELEGRAM_ALLOWED_CHAT_IDS` so random people cannot feed your social accounts
- Prefer fine-grained credentials for each outbound platform

## License

MIT
