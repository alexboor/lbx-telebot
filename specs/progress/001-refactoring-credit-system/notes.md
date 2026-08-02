# Refactoring credit system — notes

## Decisions made

- **Renamed from "emoji reaction credits" to "refactoring credit system", scope broadened.** Confirmed with user 2026-08-02. The pipeline review (below) surfaced two pre-existing scoring defects unrelated to reactions; rather than track them separately, they were pulled into this spec's acceptance criteria/tasks since they're defects in the same credit-rating process this spec already touches. See spec.md's "Score algorithm fixes" acceptance criteria and tasks.md's matching section.
- **No retention/cleanup for `message_author` — kept forever.** Confirmed with user 2026-08-02, overriding the originally-proposed 14-day retention window. `message_author` (and every table in this schema) holds all historical data indefinitely: no TTL, no cleanup job, no `created_at`/retention column. A reaction on a years-old message must still resolve to its author. Removes `CleanupMessageAuthors` from the `Storage` interface, the `internal/scheduler/reaction.go` cleanup job, and the retention-supporting index — see plan.md step 3.

- **Credit target = message author, not reactor.** Confirmed with user 2026-08-02. Rewards content that gets a positive response, not the act of reacting. Requires the new `message_author` mapping table since Telegram's `message_reaction` event doesn't include the original author.
- **Scope = tap-to-react reactions only, not emoji typed in text.** Confirmed with user 2026-08-02. Typed-emoji detection was considered easy (message text is already available) but is a distinct feature — not built here.
- **Integrate into the existing `/score` system**, not a new ledger/command. Confirmed with user 2026-08-02. Reuses `counting` table, `internal/score`, daily recalculation cron, and the existing `/score` display — no new user-facing command.
- **+1 per reaction event, not per emoji.** A single `MessageReaction` update can carry multiple newly-added emoji at once (multi-emoji react); capping at +1 avoids one tap awarding multiple points.
- **No anti-gaming for repeated add/remove/re-add cycling.** Deliberately deferred — consistent with the existing lack of anti-gaming on `Forward`/`Media` counts, judged acceptable for a small trusted group. If it becomes a problem later, mitigation is a `(chat, message, reactor)` dedup ledger with a daily cap.
- **`/top`, `/bottom`, `/profile` are NOT updated to reflect reaction credit.** They read a separate "engagement volume" query path (`GetTop`/`GetBottom`/`GetProfileStatisticByName`/`GetProfileStatisticById`) independent of the `score` table. Only `/score` reflects reaction credit. Revisit as a follow-up if the user wants parity.

## Open questions

- Exact positive-emoji allow-list is a proposal (`👍 ❤ 🔥 🥰 👏 🎉 🤩 💯 😁 ❤‍🔥`) — needs user review/edit before/at implementation time.
- Telegram's wire encoding of `❤` (with or without variation selector) is assumed, not yet empirically verified against a live payload — first implementation step should confirm this before relying on it.
- ~~14-day retention window for `message_author`~~ — resolved 2026-08-02: no retention, kept forever, per explicit user decision (see "Decisions made").
- ~~`dt` for the reaction-credit `Count()` call is unspecified in plan.md~~ — resolved, see "Review findings" below: it must be the original message's date, not the reaction's.

## Review findings (pipeline audit, 2026-08-02)

Before implementation, the full credit-rating pipeline was read end to end (message ingestion → `counting` storage → `calculateScore` → `/score` display, plus the sibling `/top`/`/bottom`/`/profile` "engagement volume" path) to check the original "emoji reaction credits" feature's assumptions against the actual code. Two findings affect that feature's design directly; two more are pre-existing scoring defects, independent of reactions, which prompted the 2026-08-02 rename/scope-broadening to "refactoring credit system" and are now in scope (see spec.md's "Score algorithm fixes" acceptance criteria and tasks.md's matching section).

### 1. `dt` for reaction credit must be the original message's date, not "now" — design change required

`calculateScore` (`internal/score/score.go:71-72,83-118`) only penalizes inactivity gaps *strictly between* a user's `firstDate` and `lastDate` in `counting` — days before/after that window are never penalized, so a dormant user's score is effectively frozen once their last counting row passes. `plan.md`'s original `Handle()` pseudocode called `s.Count(ctx, authorId, chatId, dt, model.Count{PositiveReaction: 1})` without specifying where `dt` comes from; read naturally that's "now" (the reaction event's timestamp). If a reaction lands on an old message from a currently-inactive user, dating the row "today" would create a brand-new `counting` row that drags `lastDate` forward to today — and the next recalculation would retroactively apply `-1` for every previously-frozen gap day. A well-intentioned +1 could swing a dormant user sharply negative.

**Resolution**: `dt` must be the *original message's* date, so the `PositiveReaction` increment lands on the same `counting` row the original message already created (additive upsert, no new row, no `lastDate` shift). See plan.md's revised step 3 — `message_author` now carries the message's date so `reaction.Handle` doesn't need to re-derive or approximate it from `created_at`.

**Fix step**: `reaction.Handle` must set `dt` = the original message's date/time, taken from `message_author.message_date` (itself populated from the same `dt := time.Unix(msg.Unixtime, 0)` `internal/handler/count.go`'s `Count()` already computes when the message first arrives) — never `time.Now()`/the reaction event's own arrival time. Reflected in plan.md step 2 (`Handle`, "Fix step" callout) and step 3 (`message_author` gains a `message_date` column + `StoreMessageAuthor`/`GetMessageAuthor` signatures updated to carry it); tracked in tasks.md under "Message → author mapping".

### 2. `/score` scoping (`ScoreTargetChat`) vs. reaction-credit scoping (`AllowedChats`) — documentation gap, not a defect

`CalculateAllScore`/the daily cron (`internal/score/score.go:28-60`) only ever compute scores from `GetAllCountsByUser(ctx, Config.ScoreTargetChat, user)` — a single chat id, configured independently from `Config.AllowedChats` (`internal/cfg/cfg.go` — two separate env vars, `ALLOW_CHATS` vs `SCORE_TARGET_CHAT`, no validation that one is a subset of the other). Any counting activity (message/forward/media — today, before this feature) in an `AllowedChats` chat that isn't `ScoreTargetChat` already never affects `/score`. This is pre-existing behavior, not something spec 001 introduces: reaction credit, scoped to `AllowedChats` per spec.md's acceptance criteria, correctly gets tracked in a non-target allowed chat but will silently never move `/score` there — same as every other counting column already behaves. Not a bug; called out here because spec.md/plan.md didn't previously mention `ScoreTargetChat` at all, and this could otherwise be mistaken for a defect introduced by this feature.

### 3. `ShowScores10` panics on a chat with fewer than 10 scored users — pulled into scope

`message.GetScores(scores[:10])` (`internal/score/score.go:24`) unconditionally slices to 10 with no length check. `GetAllScores` can return fewer rows (new/small group), which panics (`slice bounds out of range`) — no `recover()` middleware exists anywhere in the codebase, so a bare `/score` call in a small group propagates an unhandled panic in the update-handling goroutine. Pre-existing, unrelated to reaction credits, but pulled into this spec's scope on rename (2026-08-02) as part of "refactoring credit system" — see spec.md acceptance criteria, plan.md step 5.
**Fix**: `scores[:min(10, len(scores))]`.

### 4. `calculateScore` freezes a dormant user's score instead of continuing to decay it — pulled into scope

Same mechanism as finding 1: because the inactivity-penalty loop (`internal/score/score.go:71-72,83`) only walks `firstDate..lastDate` from the user's own `counting` rows, a user who stops posting entirely stops accumulating `-1`/day the moment their last row passes — the score freezes rather than keeps decaying. Pre-existing, unrelated to reaction credits, but pulled into this spec's scope on rename (2026-08-02) — see spec.md acceptance criteria, plan.md step 6.
**Fix**: bound the penalized window by `min(lastDate, today)` instead of `lastDate` alone, so gaps up to the present are actually penalized. "Today" for the nightly cron (runs once at `00:15`) is unambiguous; for an ad hoc `/score recalculate` run mid-day, "today" is simply the moment the recalculation runs — consistent with the cron, no separate handling needed.

## Key library facts discovered (telebot v3.3.8)

- `Bot.ProcessUpdate` never inspects `Update.MessageReaction`; no `OnReaction` handler endpoint exists. Reactions are silently dropped through normal `bot.Handle` registration.
- `MiddlewarePoller` (built into telebot) wraps a `Poller` and runs a `Filter func(*Update) bool` before normal dispatch — the clean hook point for this feature, no fork/patch needed.
- `LongPoller.AllowedUpdates` defaults to nil in this codebase's current config, which means Telegram does not send `message_reaction` at all today; must explicitly opt in via `tele.AllowedUpdates` (a package var that already includes it).
- `bot.Poller` is exported and safely reassignable after `tele.NewBot(...)`.
