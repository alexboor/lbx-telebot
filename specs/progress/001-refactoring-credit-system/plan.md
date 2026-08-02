# Refactoring credit system — technical plan

## Context

See spec.md for the feature summary and scope decisions. Two library/protocol gaps drive most of the design complexity:

1. `gopkg.in/telebot.v3` (pinned `v3.3.8`) receives Telegram's `message_reaction` update in its raw `Update` struct, but its dispatcher (`Bot.ProcessUpdate`) never inspects it and there is no `OnReaction` handler endpoint — reactions are silently dropped today. Telebot does ship a purpose-built hook for this: `MiddlewarePoller`, which wraps any `Poller` and runs a `Filter func(*Update) bool` over every raw update before normal dispatch. That's the integration point used here — no forking/patching of the library.
2. Telegram's `message_reaction` event carries the chat, message id, and the *reactor* — but not the original message's author. Since credit goes to the author, the bot needs its own (chat, message) → author lookup, populated from messages it already sees via the existing `Count()` handler.

## Approach

Reuse the existing `counting` table / `internal/score` scoring pipeline rather than building a separate credits ledger, per the user's decision. Add reaction handling as a new `internal/reaction` package (mirroring the `internal/score` logic-package convention), hooked in via `MiddlewarePoller` at the poller level in `cmd/main.go`, since there's no handler-registration path for reactions in telebot.

## Implementation steps

### 1. Opt in to reaction updates + hook them via `MiddlewarePoller`

`cmd/main.go`:
- Set `LongPoller.AllowedUpdates: tele.AllowedUpdates` in the `tele.Settings` literal (currently nil, so Telegram never sends `message_reaction`). `tele.AllowedUpdates` is a package var that already includes `message_reaction` alongside everything currently relied upon — additive, not a behavior change for existing update types.
- After `bot, err := tele.NewBot(opts)` succeeds and before `sch.Start()`/`bot.Start()`:
  ```go
  bot.Poller = tele.NewMiddlewarePoller(bot.Poller, reaction.Filter(pg, config.AllowedChats))
  ```
  (`bot.Poller` is exported and safely reassignable post-construction.)

### 2. New `internal/reaction` package

- **`internal/reaction/reaction.go`**:
  - `PositiveEmoji map[string]struct{}` — curated allow-list: `👍 ❤ 🔥 🥰 👏 🎉 🤩 💯 😁 ❤‍🔥` (Telegram's payload uses the bare heart `❤` U+2764 without variation selector — confirm empirically, see verification step 3).
  - `newlyAdded(old, new []tele.Reaction) []string` — pure diff: emoji present in `new` (`Type == "emoji"`) but not in `old`. Ignores `custom_emoji` reactions.
  - `hasNewPositive(old, new []tele.Reaction) bool` — true if any newly-added emoji is in `PositiveEmoji`.
  - `Handle(ctx context.Context, s storage.Storage, r *tele.MessageReaction)`:
    1. Return early if `!hasNewPositive(r.OldReaction, r.NewReaction)`.
    2. `s.GetMessageAuthor(ctx, r.Chat.ID, int64(r.MessageID))` — now returns `(authorId int64, messageDate time.Time, err error)`, see step 3; on `pgx.ErrNoRows` log `Debug` and return (expected condition, not an error).
    3. Skip if `r.User != nil && r.User.ID == authorId` (no self-credit).
    4. Award via existing `s.Count(ctx, authorId, r.Chat.ID, messageDate, model.Count{PositiveReaction: 1})` — reuses the additive upsert already used for every other counter.

  **Fix step (Review findings #1)**: `dt` passed to `s.Count(...)` must be `messageDate` — the original message's date/time as recorded when the message was first counted — never `time.Now()`/the reaction event's own timestamp. Concretely: `messageDate` comes straight from `GetMessageAuthor`'s second return value (step 3 below), which is populated from the same `dt := time.Unix(msg.Unixtime, 0)` `internal/handler/count.go`'s `Count()` already computes for the `counting` upsert — so `reaction.Handle` never re-derives or approximates it. Why this matters: `calculateScore` only penalizes inactivity gaps between a user's `firstDate`/`lastDate` in `counting`, so dating the row "today" for a reaction on an old message from a currently-inactive user would drag `lastDate` forward and retroactively apply `-1` for every previously-frozen gap day. Using the message's own date/time keeps the `PositiveReaction` increment on the same `counting` row the original message already created, so it can never shift `lastDate`.
  - Award rule: cap at **+1 per reaction event**, regardless of how many new positive emoji appear in one `MessageReaction` update.
  - Non-goal (explicit, not an oversight): add/remove/re-add cycling by the same reactor to farm points is not prevented in v1 — consistent with the existing lack of anti-gaming on `Forward`/`Media`. Deferred mitigation: a `(chat, message, reactor)` dedup ledger with a daily cap, if it becomes a real problem.

- **`internal/reaction/filter.go`**:
  ```go
  func Filter(s storage.Storage, allowedChats []int64) func(*tele.Update) bool {
      return func(u *tele.Update) bool {
          if u.MessageReaction == nil {
              return true
          }
          r := u.MessageReaction
          if r.Chat == nil || !slices.Contains(allowedChats, r.Chat.ID) {
              return false
          }
          go Handle(context.Background(), s, r) // fire-and-forget, see risk below
          return false
      }
  }
  ```

### 3. Message → author mapping (new table)

New migration `internal/storage/postgres/migration/reaction.go`:
```sql
create table if not exists message_author
(
    chat_id      numeric not null,
    message_id   numeric not null,
    user_id      numeric not null,
    message_date date    not null,
    primary key (chat_id, message_id)
);
```
`message_date` is the denormalized calendar date the message was counted under (i.e. the same `dt` already computed in `internal/handler/count.go`'s `Count()` from `msg.Unixtime`) — added per "Review findings" #1 in notes.md, so `reaction.Handle` can award credit on the message's own date instead of the reaction's, without re-deriving or approximating it. Appended to `migration.Migrations` (append-only, `if not exists`, matches existing no-version-tracking convention).

**No retention/cleanup**: per explicit user decision (2026-08-02), `message_author` is kept forever, like every other table in the schema — no TTL, no cleanup job, no `created_at`/retention column. A reaction on a message from years ago must still resolve to its author.

`Storage` interface (`internal/storage/storage.go`) gains:
```go
StoreMessageAuthor(ctx context.Context, chatId, messageId, userId int64, messageDate time.Time) error
GetMessageAuthor(ctx context.Context, chatId, messageId int64) (userId int64, messageDate time.Time, err error)
```
Implemented in new `internal/storage/postgres/reaction.go`:
- `StoreMessageAuthor`: `insert ... on conflict (chat_id, message_id) do nothing` (one row per message, never moves once written — unlike additive `counting` upserts).
- `GetMessageAuthor`: `select user_id, message_date from message_author where chat_id = $1 and message_id = $2` (caller handles `pgx.ErrNoRows`).

**Population point**: `internal/handler/count.go`'s `Count()` — the single choke point already gated by `IsAllowedGroup`, covering every message type registered in `main.go`. It already computes `dt := time.Unix(msg.Unixtime, 0)` for the `counting` upsert; reuse the same `dt` here so both tables agree on the message's date:
```go
if err := h.Storage.StoreMessageAuthor(ctx, msg.Chat.ID, int64(msg.ID), msg.Sender.ID, dt); err != nil {
    return err
}
```

### 4. Thread the credit into the existing score pipeline

- `internal/model/count.go`: add `PositiveReaction int` to `Count`.
- New migration: `alter table counting add if not exists positive_reaction numeric default 0`.
- `internal/storage/postgres/count.go`'s `Count()` upsert: add `positive_reaction` to insert columns and `on conflict ... positive_reaction = counting.positive_reaction + excluded.positive_reaction`.
- `internal/storage/postgres/count.go`'s `GetAllCountsByUser` (the only SELECT feeding scoring): add `positive_reaction` to select + `Scan`.
- `internal/score/score.go`'s `calculateScore`: add `dayScore += count.PositiveReaction` next to the existing `Forward`/`Media` lines.

**Explicitly out of scope**: `GetTop`/`GetBottom`/`GetProfileStatisticByName`/`GetProfileStatisticById` (`/top`, `/bottom`, `/profile`) and `GetAvgStatistic()` are a separate "engagement volume" metric independent of `score`/`/score`, and are intentionally left untouched.

**`ScoreTargetChat` scoping note** (see "Review findings" #2 in notes.md): `/score` recalculation (`CalculateAllScore`, both the `/score recalculate` command and the nightly cron) only ever reads `counting` rows for the single `Config.ScoreTargetChat`, independently configured from `Config.AllowedChats`. This is pre-existing behavior for every counting column (message/forward/media), not something this feature introduces: reaction credit earned in an `AllowedChats` chat other than `ScoreTargetChat` will be correctly tracked (per this feature's own scoping) but will never move `/score` for that chat, same as ordinary message activity there already doesn't. Not a defect to fix here — noted so it isn't mistaken for one.

### 5. Fix: `ShowScores10` panic on small chats (Review findings #3)

`internal/score/score.go:24` — `message.GetScores(scores[:10])` unconditionally slices to 10. Change to `message.GetScores(scores[:min(10, len(scores))])` (Go 1.21+ builtin `min`, already available per the module's Go version). No other call site depends on the slice always being length 10.

### 6. Fix: `calculateScore` freezes a dormant user's score (Review findings #4)

`internal/score/score.go:71-72,83` — the inactivity-penalty loop currently walks `firstDate..lastDate` where `lastDate = counts[len(counts)-1].Date`, i.e. the user's own last `counting` row. Change the walk's upper bound to `min(lastDate, today)` (using `time.Now()` truncated to the same day-granularity the rest of the function already uses), so a dormant user keeps accumulating `-1`/day up to the present instead of freezing the moment their last row passes. `today` for the nightly cron (`RecalculateScore`, runs once at `00:15`) is unambiguous; for an ad hoc `/score recalculate` run mid-day, `today` is simply the moment the recalculation runs — consistent with the cron, no separate handling needed.

## Risks / edge cases

- **Poll-loop blocking**: `MiddlewarePoller.Poll` calls `Filter` synchronously in its single receive loop. If `Filter` did blocking DB I/O inline, it would stall delivery of *every* subsequent update (including ordinary messages) for that duration. Mitigated by dispatching `Handle` via `go` so `Filter` returns immediately — reaction processing becomes best-effort/unordered, acceptable for a scoring nice-to-have.
- **Emoji byte-equality**: Telegram's wire format for the heart reaction may not match a naively-typed `❤️` literal (variation selector mismatch) — must be confirmed against a live payload (see verification step 3), not assumed.
- **Missing author row**: reactions on messages that predate this feature being deployed (i.e. sent before `message_author` existed) resolve to `pgx.ErrNoRows` — treated as a silent no-op, not an error to alert on. Once deployed, `message_author` is kept forever (no retention), so this only affects the pre-deployment backlog, never a gap that reopens later.
- **Anonymous reactors** (`User == nil`, `ActorChat != nil`): self-reaction guard is skipped only when `User != nil`; still requires the diff + author lookup to succeed.
- **Reaction farming via repeated add/remove**: explicitly deferred, see step 2 above.

## Files touched

- `cmd/main.go` — poller wiring + `AllowedUpdates`
- `internal/reaction/reaction.go`, `internal/reaction/filter.go`, `internal/reaction/reaction_test.go` (new)
- `internal/handler/count.go` — populate `message_author`
- `internal/storage/storage.go` — 3 new interface methods
- `internal/storage/postgres/reaction.go` (new)
- `internal/storage/postgres/migration/reaction.go` (new), `migration.go` (register), count migration (new `positive_reaction` column)
- `internal/model/count.go` — new field
- `internal/storage/postgres/count.go` — upsert + `GetAllCountsByUser`
- `internal/score/score.go`, `internal/score/score_test.go` — reaction credit tier, `ShowScores10` bound fix, `calculateScore` dormant-decay fix

## Verification plan

Cannot be fully unit tested — needs live Telegram reaction events against a real bot in a real group:

1. Log `LongPoller.AllowedUpdates` at startup to confirm it's non-nil and includes `message_reaction`.
2. Temporarily log every non-nil `u.MessageReaction` in `Filter` before wiring credit logic, to isolate wiring bugs from scoring-logic bugs.
3. Confirm the exact byte sequence of `❤` as delivered by a live reaction, diffed against the `PositiveEmoji` map literal.
4. End-to-end in a real allowed test group: A posts → confirm `message_author` row appears; B taps 👍 → confirm `positive_reaction` increments in `counting` for A/today; run `/score recalculate` → confirm `/score` reflects +1.
5. Self-reaction: A reacts to A's own message → confirm no change.
6. Tap + untap: confirm untap never decrements; re-tapping an already-present emoji doesn't re-award.
7. Reaction in a non-allowed chat → confirm `Filter` drops it before any DB call.
8. `ShowScores10` fix: unit test with `<10` scores confirms no panic (covered in `score_test.go`, no live bot needed).
9. Dormant-decay fix: unit test seeding a user with a `counting` row several days in the past and no recent activity confirms `calculateScore` now penalizes gap days up to "today", not just up to the user's last row (covered in `score_test.go`, no live bot needed).
