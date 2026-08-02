# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`lbx-telebot` (module `github.com/alexboor/lbx-telebot`) is a Telegram bot for a group community, written in Go, using `gopkg.in/telebot.v3`. It tracks message activity per user/chat, computes engagement scores, runs prediction "events" (betting), shows weather alerts, and other group utilities. See README.md for the full end-user command reference (`/help`, `/top`, `/bottom`, `/profile`, `/event`, `/score`, `/topic`, `/today`, `/newyear`, `/meteoalarm`).

## Commands

- Run locally: `make run` (or `go run cmd/main.go`). The Makefile auto-loads a `.env` file if present (copy `example.env` to `.env` and fill in `TELEGRAM_BOT_TOKEN`, `DB_*`, `ALLOW_CHATS`, `CHATGPT_TOKEN`).
- Run all tests: `make test` (or `go test -v ./...`).
- Run a single test: `go test -v ./internal/score/... -run TestCalculateScore`.
- Build: `go build -o app cmd/main.go` (Docker build uses `go mod vendor` then builds with `-ldflags` for a version stamp — see Dockerfile).
- Requires a running Postgres instance reachable via `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASSWORD`/`DB_NAME`; the app runs its own migrations on startup (see below), there is no separate migrate step.

## Architecture

**Entry point**: `cmd/main.go` wires everything together: builds the `cfg.Cfg`, connects Postgres (`internal/storage/postgres`), creates the in-memory store (`internal/storage/memory`), constructs the `handler.Handler`, starts the `scheduler.Schedule`, backfills profile data for all `AllowedChats` on boot, then registers all telebot command/event handlers and starts polling.

**Layering**:
- `internal/handler` — one file per command/feature (`event.go`, `score.go`, `profile.go`, `rating.go`, `topic.go`, `today.go`, `meteoalarm.go`, `new-year.go`, `count.go`, `version.go`, `help.go`). All handlers are methods on `handler.Handler`, which holds `Storage` (the `storage.Storage` interface), `Memory` (in-memory KV), and `*cfg.Cfg`. `middleware.go` has `IsAllowedGroup`/`IsAllowedChat` guards — command handlers must not act on chats outside `Config.AllowedChats`, and most commands are group-only while a few (like `/event`) accept both group and private chat.
- `internal/message` — builds the actual reply strings/keyboards/images sent to Telegram (Markdown-formatted text via `internal.MarkdownOpt`, chart images via `go-chart`). Handlers call into `message` rather than formatting strings inline.
- `internal/model` — domain types (`Profile`, `Count`, `Event`, `Participant`, `Option`) plus parsing logic for command payloads. `model.Option` parses free-form `/top`/`/bottom`/`/profile` payloads (a mix of a numeric limit, a duration like `7d`/`72h`, and/or a target profile name, in either order) — see `option.go` for the parsing rules and `option_test.go` for expected behavior.
- `internal/storage` — defines the `Storage` interface (the contract every handler codes against). Two implementations:
  - `internal/storage/postgres` — the real backend (pgx/v4 pool). Schema is defined as raw SQL strings under `internal/storage/postgres/migration/` and run in order from `migration.Migrations` on every connect (`postgres.New` → `migrate`); there is no migration framework or version tracking, so schema changes are additive SQL appended to that list.
  - `internal/storage/memory` — a generic `sync.RWMutex`-guarded `map[string]interface{}` used as a scratch KV store (e.g. caching meteoalarm results between poll cycles), not a Storage implementation.
- `internal/scheduler` — cron-style background jobs via `gocron`. Registered in `scheduler.New` (`schedule.go`): a 10-minute meteoalarm poll, a daily profile cleanup (`00:00`), and a daily score recalculation (`00:15`). Job bodies live in sibling files (`meteoalarm.go`, `score.go`, `examples.go` — the latter documents the pattern for adding new jobs but isn't itself wired in).
- `internal/score` — scoring algorithm (`calculateScore`) that turns per-day `model.DateCount` activity into a point total (message-volume tiers, +1 per forward, +1 per media, -1 for inactive days in the user's active date range). Covered by `internal/score/score_test.go` — check there for the exact tier boundaries before changing scoring rules.
- `internal/meteoalarm` / `internal/wikimedia` — thin API clients for external data sources (weather alerts, "on this day" facts) consumed by the corresponding handlers/scheduler jobs.
- `internal/cfg` — all configuration is env-var driven (no config file); see `cfg.go` for the exact env var names and parsing (comma-separated `ALLOW_CHATS` int64 list, Postgres DSN assembled from `DB_*`, optional `SCORE_TARGET_CHAT`).
- `internal/const.go` — single source of truth for command strings (`/top`, `/event`, ...), timeouts, and misc constants; add new commands here rather than inlining string literals in `main.go`.

**Conventions to follow when extending this code**:
- New commands: add the constant to `internal/const.go`, register the handler in `cmd/main.go`, implement it as a method on `handler.Handler` in a new/existing `internal/handler/*.go` file, and build any user-facing text in `internal/message`, not inline in the handler.
- New Storage methods: add to the `Storage` interface in `internal/storage/storage.go` first, then implement in `internal/storage/postgres`; there's only one production implementation, so interface and implementation tend to change together.
- New scheduled jobs: follow the pattern in `internal/scheduler/schedule.go` (`cronScheduler.NewJob(...)`), with the job body in its own file the same way `meteoalarm.go`/`score.go` do it.
- `docs/msg_obj.json` is a captured sample of a raw Telegram message object, useful as a reference for the shape of `tele.Message`/`msg.Sender` fields.

## Spec-driven process

Non-trivial features are planned as specs before implementation. All specs live under `/specs`:

- `/specs/INDEX.md` — the registry of all features; every feature must be listed here (id, slug, status, one-line summary).
- `/specs/_templates/` — `spec.md`, `plan.md`, `tasks.md` templates that new feature directories are copied from.
- `/specs/progress/` — feature directories currently being worked on.
- `/specs/done/` — feature directories that are completed; move the whole directory here (don't copy) when a feature ships, and update its status in `INDEX.md`.

Each feature gets its own directory named `NNN-feature-slug/` (zero-padded sequential id + kebab-case slug), containing:
- `spec.md` — feature description and acceptance criteria.
- `plan.md` — technical plan: what to build and how.
- `tasks.md` — checklist of tasks as markdown checkboxes, kept up to date as work progresses.
- `notes.md` — optional; decisions made and open questions.

When asked to plan or implement a feature, check `/specs/INDEX.md` first for an existing spec before creating a new one, and keep `tasks.md`/`notes.md` in the feature directory current as work happens rather than tracking that progress only in conversation.
