# Telegram Web Application (Mini App)

## Summary

Add a Telegram Mini App, launched from the bot's chat menu button, that shows the requesting user their name, avatar, and credit score. Served as a monolith: one Go process polls the Telegram Bot API, serves the Nuxt-built UI at `/`, and serves a REST API at `/api/v1/`.

## Description

The bot currently only interacts with users through chat commands. This feature adds a web surface: a Telegram Mini App (Web App) that opens inside Telegram when the user taps the bot's chat menu button. The app is a Nuxt frontend (built as a static SPA, embedded into the Go binary) that calls a REST API under `/api/v1/` to fetch the current user's profile data. The HTTP port the app listens on is configured via an environment variable.

Initial version scope: a single profile view showing the user's name, avatar, and credit score (the same score computed by the existing `/score` system). The user is identified via Telegram's `initData` mechanism (HMAC-signed by the bot token), not a separate login flow.

## Acceptance criteria

- [ ] The bot's chat menu button (private chats only, per Telegram platform behavior) launches the Mini App via a `web_app` button pointing at the configured public HTTPS URL.
- [ ] The HTTP server listens on a port read from an environment variable; no hardcoded port.
- [ ] `GET /api/v1/profile`, authenticated via a valid `Authorization: tma <initData>` header, returns the caller's own name, username, and credit score as JSON.
- [ ] `GET /api/v1/avatar`, authenticated the same way, returns the caller's own Telegram profile photo bytes (proxied server-side; the bot token is never exposed to the client).
- [ ] Requests with missing or invalid `initData` (bad HMAC, or `auth_date` older than 24h) are rejected with `401`.
- [ ] A user with no stored profile (never tracked in an `AllowedChats` group) gets a `404` from `/api/v1/profile`, not a crash or 500.
- [ ] A user with no Telegram profile photo gets a graceful empty response from `/api/v1/avatar`, and the UI shows a placeholder instead of a broken image.
- [ ] `/` serves the Nuxt SPA (static, embedded in the Go binary — no Node process at runtime).
- [ ] Existing bot behavior (all chat commands, scheduler jobs) is unaffected by the new HTTP server running alongside the poller.

## Out of scope

- Viewing another user's profile through the Mini App (only the authenticated caller's own data).
- Any Mini App view beyond the single profile screen (no `/top`, `/event`, etc. in the web UI yet).
- TLS termination / reverse proxy setup — the Go process serves plain HTTP; HTTPS in front of it is a deployment concern.
- Full `initData` replay protection (only staleness via `auth_date` is checked).
- Server-side caching of avatar/profile responses.
- Ansible/`deploy.yml` port exposure — the container currently publishes no ports; wiring that up is a manual deployment follow-up, not part of this code change.
