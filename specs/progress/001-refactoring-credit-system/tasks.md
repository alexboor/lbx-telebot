# Refactoring credit system — tasks

## Wiring
- [x] Set `LongPoller.AllowedUpdates: tele.AllowedUpdates` in `cmd/main.go`
- [x] Wrap `bot.Poller` with `tele.NewMiddlewarePoller(bot.Poller, reaction.Filter(pg, config.AllowedChats))` in `cmd/main.go`

## `internal/reaction` package
- [x] `reaction.go`: `PositiveEmoji` allow-list constant
- [x] `reaction.go`: `newlyAdded(old, new []tele.Reaction) []string`
- [x] `reaction.go`: `hasNewPositive(old, new []tele.Reaction) bool`
- [x] `reaction.go`: `Handle(ctx, storage.Storage, *tele.MessageReaction)` (author lookup, self-reaction guard, +1 award)
- [x] `filter.go`: `Filter(storage.Storage, allowedChats []int64) func(*tele.Update) bool`
- [x] `reaction_test.go`: table-driven tests for `newlyAdded` and `hasNewPositive`

## Message → author mapping
- [x] New migration `internal/storage/postgres/migration/reaction.go` (`message_author` table + `message_date` column, no retention/TTL — kept forever)
- [x] Register new migration in `migration.Migrations`
- [x] Add `StoreMessageAuthor`/`GetMessageAuthor` to `Storage` interface (`internal/storage/storage.go`) — `StoreMessageAuthor` takes `messageDate time.Time`, `GetMessageAuthor` returns `(userId int64, messageDate time.Time, err error)`
- [x] Implement both in new `internal/storage/postgres/reaction.go`
- [x] Populate `message_author` from `internal/handler/count.go`'s `Count()`, reusing the `dt` already computed there for the `counting` upsert
- [x] `reaction.Handle` uses `message_author`'s `message_date` (not the reaction's own timestamp) as the `dt` passed to `s.Count(...)` — see notes.md "Review findings" #1

## Score integration
- [x] Add `PositiveReaction int` to `internal/model/count.go`'s `Count`
- [x] New migration: `alter table counting add if not exists positive_reaction numeric default 0`
- [x] Update `internal/storage/postgres/count.go`'s `Count()` upsert (insert columns + `on conflict` clause)
- [x] Update `internal/storage/postgres/count.go`'s `GetAllCountsByUser` select + scan
- [x] Add `dayScore += count.PositiveReaction` in `internal/score/score.go`'s `calculateScore`
- [x] Add test cases to `internal/score/score_test.go`

## Score algorithm fixes (pre-existing defects, pulled into spec 001's scope 2026-08-02)

Found during the pipeline review; see notes.md "Review findings" #3–4 for full detail.

- [x] `internal/score/score.go:24` `ShowScores10`: fix `scores[:10]` panic on chats with <10 scored users → `scores[:min(10, len(scores))]`
- [x] `internal/score/score.go:71-72,83` `calculateScore`: bound the inactivity-penalty walk by `min(lastDate, today)` instead of `lastDate` alone, so a dormant user's score keeps decaying instead of freezing. Implemented by threading a `now time.Time` parameter through `calculateScore` (defaulting to `time.Now()` at the `CalculateAllScore` call site) so the behavior stays unit-testable.
- [x] Add test cases to `internal/score/score_test.go` for both fixes (small-chat `ShowScores10` needs no dedicated unit test — `GetScores` already ranges over any slice length; dormant-user decay-to-today covered by `TestCalculateScoreDormantDecay`)

## Verification (manual, live bot)

Everything above is implemented, builds clean (`go build ./...`), passes `go vet ./...`, and all unit tests pass (`go test ./...`). The items below need a live Telegram bot + real group and were **not** run in this environment (no reachable Telegram/Postgres in this sandbox) — still open before calling the feature done.

- [ ] Confirm `AllowedUpdates` includes `message_reaction` at startup
- [ ] Confirm raw `MessageReaction` updates are received (temporary logging)
- [ ] Confirm exact byte sequence of `❤` reaction matches `PositiveEmoji` map key
- [ ] End-to-end: post → react → `/score recalculate` → `/score` reflects +1
- [ ] Self-reaction produces no credit
- [ ] Untap never decrements; re-tap of existing emoji doesn't re-award
- [ ] Reaction in non-allowed chat is dropped before any DB call
- [ ] Reaction on an old message from a currently-inactive user does not shift their `lastDate`/retroactively penalize gap days (see notes.md "Review findings" #1)
- [ ] Reaction on a message from well before the retention window that used to be proposed (e.g. months old) still resolves to its author — confirms no cleanup job is silently dropping rows
- [ ] Migrations (`message_author` table, `counting.positive_reaction` column) apply cleanly against a real Postgres instance
