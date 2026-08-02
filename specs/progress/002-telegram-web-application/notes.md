# Notes

Decisions confirmed with the user on 2026-08-02:

- **Entry point**: the bot's global Telegram chat menu button (`setChatMenuButton`, type `web_app`), not a new slash command and not an addition to `/profile`'s existing output. Rejected alternatives: a new `/app` command, and a button appended to `/profile`'s reply.
- **UI build/serve strategy**: Nuxt built as a static SPA (`nuxi generate`, `ssr: false`) and embedded into the Go binary via `go:embed`, served with `net/http`. Rejected alternatives: copying static files into the Docker image without embedding (loses the single self-contained binary), and running Nuxt SSR as a subprocess (breaks the single-binary/`scratch` model, adds a Node runtime dependency).

Key technical facts discovered during research:

- `gopkg.in/telebot.v3 v3.3.8` (already the repo's pinned version) fully supports Web Apps: `tele.WebApp{URL}`, `tele.MenuButton{Type, Text, WebApp}` with `tele.MenuButtonWebApp`, and `bot.SetMenuButton(chat *User, mb interface{})` / `bot.MenuButton(chat *User)`. Passing `nil` as the chat sets the *default* menu button for all private chats.
- The codebase has zero `net/http` usage today — the new HTTP server is fully additive, started as a goroutine alongside the existing `bot.Start()` poller in `cmd/main.go`.
- `model.Profile` has no avatar field and there's no existing precedent for fetching profile photos anywhere in the codebase — `bot.ProfilePhotosOf` is unused today.
- `go:embed` cannot reference paths outside the declaring file's own directory (no `..`), which is why the embedding shim must live inside `/ui` itself rather than under `internal/`.

Build & deploy is now specified concretely in `plan.md`'s "Build & deploy" section: `Dockerfile` gets a `node:20-alpine` build stage feeding `go:embed`, `deploy.yml` gets a `ports` mapping, and the GitHub Actions workflow passes `http_port` through to Ansible. The only pieces that stay outside this repo's control (host-side, one-time, manual) are: adding `HTTP_PORT`/`PUBLIC_BASE_URL` to the deploy host's `/home/sta/envs/telebot.env`, and pointing an existing reverse proxy at the container's published port with TLS — Telegram requires HTTPS for `WebApp.URL` and the Go process only ever serves plain HTTP itself.
