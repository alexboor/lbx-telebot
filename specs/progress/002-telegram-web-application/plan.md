# Telegram Web Application (Mini App) — technical plan

## Context

The bot only talks to users through `gopkg.in/telebot.v3` long-polling commands today; there is no HTTP server anywhere in the process. This plan adds a Telegram Mini App entry point (chat menu button), a REST API, and a Nuxt-built UI, all served by the same Go binary — a monolith, per the requirement. Initial scope is a single profile view (name, avatar, credit score).

`specs/progress/001-refactoring-credit-system/` is the only existing feature and is used as the structural template for this directory.

Decisions confirmed with the user (see `notes.md`):
- Entry point is the bot's global chat **menu button** (`setChatMenuButton`, type `web_app`), not a new slash command or an addition to `/profile`'s output.
- The Nuxt UI is built as a **static SPA** (`nuxi generate`, `ssr: false`) and its output is **embedded into the Go binary** via `go:embed` — no Node process at runtime, keeping the existing `FROM scratch` single-binary Docker model.

## Approach

**Data reuse**: no new storage methods or migrations. `Storage.GetProfileById(ctx, id)` and `Storage.GetScore(ctx, id)` (`internal/storage/storage.go`) already provide everything the profile view needs. `model.Profile` has no avatar field — the avatar is fetched live from Telegram (`bot.ProfilePhotosOf`) and proxied through our own API so the bot token is never exposed to the client.

**Auth**: Telegram Mini Apps identify the user via `initData`, injected into `window.Telegram.WebApp.initData` by the Telegram client, HMAC-signed with the bot token. The frontend sends it as `Authorization: tma <initData>` on every API call; the backend validates the HMAC per Telegram's documented algorithm (`secret_key = HMAC_SHA256(key="WebAppData", data=botToken)`, `hash = HMAC_SHA256(key=secret_key, data=data_check_string)`, where `data_check_string` is every `initData` field except `hash`, sorted and joined with `\n`), checks `auth_date` isn't older than 24h, and extracts the Telegram user id from the `user` JSON field. This is pure/testable logic in a new `internal/webapp` package (stdlib `crypto/hmac`/`crypto/sha256` only) — mirrors the existing "pure calculation function + tests" pattern used by `internal/score`.

**Why `ui/embed.go` and not `internal/webui`**: `go:embed` patterns cannot use `..` or reference paths outside the directory of the declaring `.go` file. Since the Nuxt project lives at top-level `/ui` and its static build output lands at `ui/.output/public`, the embedding shim has to physically live inside `/ui` too. A tiny `ui/embed.go` (`package ui`, `//go:embed all:.output/public`) sits alongside the Nuxt project and is imported by `internal/api` like any other Go package.

**Why `net/http` + `ServeMux`, no router dependency**: matches the codebase's existing minimal-dependency style (no framework anywhere; `cfg` uses plain `os.Getenv`). Three routes don't justify a router dependency.

## Implementation steps

1. **Config** (`internal/cfg/cfg.go`): add `HTTPPort string` (env `HTTP_PORT`, default `8080` via an `initHTTPPort()` helper matching the `initScoreTargetChat` pattern) and `PublicBaseURL string` (env `PUBLIC_BASE_URL`, the HTTPS URL Telegram will load — used to build the menu button's `WebApp.URL`; if unset, skip registering the menu button and log a warning). Add both to `example.env`.

2. **`internal/const.go`**: add a menu button text constant (e.g. `WebAppBtnText = "Open"`).

3. **`internal/webapp` (new package)** — `auth.go`: `ValidateInitData(botToken, initData string) (*User, error)` implementing the HMAC validation above, returning the parsed Telegram user (id, first/last name, username). `auth_test.go`: unit tests with a known bot token + hand-computed valid/invalid/stale/tampered `initData` strings, following `internal/score/score_test.go`'s style.

4. **`internal/api` (new package)**:
   - `server.go` — `New(cfg *cfg.Cfg, storage storage.Storage, bot *tele.Bot) *http.Server` wiring an `http.ServeMux`.
   - `middleware.go` — auth middleware parsing `Authorization: tma <initData>` via `internal/webapp.ValidateInitData`, storing the validated user id in request context; `401` on failure.
   - `profile.go` — `GET /api/v1/profile` (authenticated) → `storage.GetProfileById` + `storage.GetScore` for the caller's own id → JSON `{id, firstName, lastName, username, score, avatarUrl: "/api/v1/avatar"}`. `404` if no stored profile.
   - `avatar.go` — `GET /api/v1/avatar` (authenticated) → `bot.ProfilePhotosOf` for the caller's own id → `bot.File(...)` to resolve the file path → server-side fetch from `api.telegram.org/file/bot<token>/<path>` → stream bytes back with `Content-Type`/`Cache-Control`. Empty-friendly response if no photo.
   - `static.go` — serves the embedded Nuxt SPA (`ui.Dist`) at `/` via `fs.Sub(ui.Dist, ".output/public")` + `http.FileServer`, falling back to `index.html` for unmatched non-`/api/` paths.

5. **`ui/` (new Nuxt 3 project)**: `package.json`, `nuxt.config.ts` (`ssr: false`, static `generate`, Telegram Web App script tag in `app.head`), `pages/index.vue` — on mount calls `Telegram.WebApp.ready()`, reads `Telegram.WebApp.initData`, fetches `/api/v1/profile` with the `Authorization: tma <initData>` header, renders avatar/name/score; handles 404 ("no profile yet") and missing-avatar states. `ui/embed.go` (`package ui`, `//go:embed all:.output/public`, `var Dist embed.FS`).

6. **`cmd/main.go`**: after `bot, err := tele.NewBot(opts)`, if `config.PublicBaseURL != ""` call `bot.SetMenuButton(nil, &tele.MenuButton{Type: tele.MenuButtonWebApp, Text: internal.WebAppBtnText, WebApp: &tele.WebApp{URL: config.PublicBaseURL}})`. Build the API server via `api.New(config, pg, bot)` and start it in a goroutine before the blocking `bot.Start()`; fatal on bind failure at startup, `slog.Error` on later serve errors.

7. **Dockerfile**: add a `node:20-alpine AS ui-build` stage (`npm ci`, `npm run generate` in `/ui`), then `COPY --from=ui-build /ui/.output/public /build/ui/.output/public` in the existing `build` stage *before* `go build` runs (so `go:embed` finds the files). No change to the final `scratch` stage.

8. **Makefile**: add `ui-build: cd ui && npm ci && npm run generate`. Document that `make ui-build` must run once before `make run`/`go build` locally.

9. **Spec bookkeeping**: this directory + `specs/INDEX.md` row (done as part of this change).

## Build & deploy

Current state: `Dockerfile` is a two-stage `golang:1.21-alpine` → `scratch` build with no `EXPOSE`; there's no `docker-compose.yml`; deployment is a single Ansible task in `deploy.yml` (`docker_container`, no `ports:` mapping) run by `.github/workflows/docker-image.yaml` on semver tag push, using an `env_file` at `/home/sta/envs/telebot.env` on the target host (not in this repo).

1. **`Dockerfile`** — insert a Node build stage before the existing Go `build` stage, and feed its output into the Go build so `go:embed` sees it:

   ```dockerfile
   FROM node:20-alpine AS ui-build
   WORKDIR /ui
   COPY ui/package.json ui/package-lock.json ./
   RUN npm ci
   COPY ui/ .
   RUN npm run generate

   FROM golang:1.21-alpine as build
   ARG VERSION=0.0.0

   WORKDIR /build
   RUN apk add --no-cache ca-certificates git tzdata
   COPY . .
   COPY --from=ui-build /ui/.output/public ./ui/.output/public
   RUN go mod vendor
   RUN go build -mod vendor -ldflags "$LD_FLAGS" -o app cmd/main.go;
   RUN echo "$VERSION" >> app.version
   ```

   Final `FROM scratch` stage is unchanged (still copies only the compiled binary + certs + tzdata + version file — the Nuxt output is baked into the binary via `go:embed`, not shipped as separate files).

2. **`deploy.yml`** — add a `ports` mapping so the container's `HTTP_PORT` is reachable from the host (currently the task publishes no ports at all):

   ```yaml
   - name: Deploy playbook
     hosts: all
     become: yes

     tasks:
       - name: Create/Update a container
         docker_container:
           name: lbx-telebot
           image: '{{img}}:{{tag}}'
           env_file: /home/sta/envs/telebot.env
           recreate: yes
           pull: true
           restart_policy: unless-stopped
           ports:
             - "{{http_port}}:{{http_port}}"
   ```

3. **`.github/workflows/docker-image.yaml`** — pass `http_port` through to the Ansible run so it matches the `HTTP_PORT` value baked into the server's env file, in the existing `options:` block of the "Deploy to the server" step:

   ```yaml
           options: |
             -u ${{ secrets.SSH_USER }}
             -e img=boorick/lbx-telebot
             -e tag=${{github.ref_name}}
             -e http_port=8080
   ```

   No other CI changes needed — the existing `docker/build-push-action@v4` step already builds the full multi-stage `Dockerfile` (now including the Node stage) with no workflow-level change required; only the build machine's disk/time budget grows slightly.

4. **Manual, one-time host-side steps (not in this repo, must be done by whoever operates the deploy target)**:
   - Add `HTTP_PORT=8080` and `PUBLIC_BASE_URL=https://<mini-app-domain>` to `/home/sta/envs/telebot.env` on the deploy host.
   - Point a reverse proxy (nginx/Caddy, whatever already fronts other services on that host) at `127.0.0.1:8080` for `<mini-app-domain>`, terminating TLS (e.g. via Let's Encrypt/certbot or Caddy's automatic HTTPS) — Telegram rejects non-HTTPS `WebApp.URL`s, and the Go process itself only ever serves plain HTTP.
   - Ensure the host firewall allows inbound 443 (or whatever the proxy listens on) for `<mini-app-domain>`; `8080` itself should stay proxy-only, not exposed publicly.

## Risks / edge cases

- `go:embed` requires `ui/.output/public` to exist and be non-empty at Go compile time — first-time/local/CI builds must run the Nuxt build first (`make ui-build`). Documented, not auto-detected.
- Telegram requires an HTTPS URL for the menu button's `WebApp.URL` — the Go process itself only listens on plain `HTTP_PORT`; a reverse proxy/TLS terminator in front is assumed and out of scope here. The Ansible `deploy.yml` `docker_container` task currently has no `ports:` mapping at all — needs a manual follow-up outside this change.
- Menu-button-type `web_app` only renders in private chats with the bot (Telegram platform limitation) — acceptable since the feature is "open my own profile."
- `initData` isn't single-use; only staleness (`auth_date` > 24h) is checked for the initial version.
- Users with no stored `Profile` row get a `404` from `/api/v1/profile` — UI must show a friendly empty state.
- Users with no Telegram profile photo get an empty avatar response — UI needs a placeholder.

## Files touched

- `internal/cfg/cfg.go` — add `HTTPPort`, `PublicBaseURL` config fields + parsing
- `example.env` — add `HTTP_PORT`, `PUBLIC_BASE_URL`
- `internal/const.go` — add menu button text constant
- `internal/webapp/auth.go`, `internal/webapp/auth_test.go` (new) — `initData` HMAC validation
- `internal/api/server.go`, `middleware.go`, `profile.go`, `avatar.go`, `static.go` (new) — REST API + SPA serving
- `cmd/main.go` — register menu button, start HTTP server goroutine
- `ui/` (new) — Nuxt 3 SPA project + `embed.go` embedding shim
- `Dockerfile` — add `node:20-alpine` Nuxt build stage, copy output into Go build stage
- `Makefile` — add `ui-build` target
- `deploy.yml` — add `ports` mapping to the `docker_container` task
- `.github/workflows/docker-image.yaml` — pass `-e http_port=8080` to the Ansible deploy step
- *(outside repo)* `/home/sta/envs/telebot.env` on the deploy host — add `HTTP_PORT`, `PUBLIC_BASE_URL`; reverse proxy config for TLS — manual, not tracked here
