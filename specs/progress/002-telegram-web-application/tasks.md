# Telegram Web Application (Mini App) — tasks

## Config

- [ ] Add `HTTPPort` and `PublicBaseURL` fields + parsing helpers to `internal/cfg/cfg.go`
- [ ] Add `HTTP_PORT` / `PUBLIC_BASE_URL` to `example.env`
- [ ] Add menu button text constant to `internal/const.go`

## `internal/webapp` — initData auth

- [ ] Implement `ValidateInitData(botToken, initData string) (*User, error)` (HMAC validation per Telegram's algorithm)
- [ ] Check `auth_date` staleness (reject if older than 24h)
- [ ] Unit tests: valid, invalid-hash, tampered, and stale `initData` cases

## `internal/api` — REST API + SPA serving

- [ ] `server.go`: `New(cfg, storage, bot) *http.Server`, mux wiring
- [ ] `middleware.go`: auth middleware, `401` on invalid/missing `initData`
- [ ] `profile.go`: `GET /api/v1/profile` (own profile + score, `404` if untracked)
- [ ] `avatar.go`: `GET /api/v1/avatar` (proxy Telegram profile photo, empty-friendly if none)
- [ ] `static.go`: serve embedded Nuxt SPA at `/`, `index.html` fallback for client-side routes

## `ui/` — Nuxt frontend

- [ ] Scaffold Nuxt 3 project (`package.json`, `nuxt.config.ts` with `ssr: false`, static generate)
- [ ] Add Telegram Web App script tag, call `Telegram.WebApp.ready()`
- [ ] `pages/index.vue`: fetch `/api/v1/profile` with `Authorization: tma <initData>`, render avatar/name/score
- [ ] Handle 404 ("no profile yet") and missing-avatar placeholder states
- [ ] `ui/embed.go`: `go:embed all:.output/public` shim

## Wiring

- [ ] `cmd/main.go`: register default menu button via `bot.SetMenuButton` (only if `PublicBaseURL` set)
- [ ] `cmd/main.go`: start `internal/api` HTTP server in a goroutine alongside `bot.Start()`

## Build / deploy

- [ ] `Dockerfile`: add `node:20-alpine AS ui-build` stage (`npm ci`, `npm run generate` in `/ui`)
- [ ] `Dockerfile`: `COPY --from=ui-build /ui/.output/public ./ui/.output/public` into the `build` stage, before `go mod vendor`/`go build`
- [ ] `Makefile`: add `ui-build` target (`cd ui && npm ci && npm run generate`)
- [ ] Document `make ui-build` as a required step before first local `go build`/`go run`
- [ ] `deploy.yml`: add `ports: ["{{http_port}}:{{http_port}}"]` to the `docker_container` task
- [ ] `.github/workflows/docker-image.yaml`: add `-e http_port=8080` to the Ansible deploy step's `options:`
- [ ] Manual (deploy host, outside repo): add `HTTP_PORT` and `PUBLIC_BASE_URL` to `/home/sta/envs/telebot.env`
- [ ] Manual (deploy host, outside repo): configure reverse proxy + TLS for the Mini App domain, proxying to `127.0.0.1:8080`

## Verification

- [ ] `go test -v ./...` passes (covers `internal/webapp` HMAC tests)
- [ ] `make ui-build && go build -o app cmd/main.go` succeeds
- [ ] Manual: open bot's private chat, tap menu button, confirm Mini App loads profile for a tracked user
- [ ] Manual: `curl` `/api/v1/profile` with missing/invalid `Authorization` header returns `401`
