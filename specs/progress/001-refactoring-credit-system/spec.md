# Refactoring credit system

## Summary

Refactor the credit/scoring pipeline (`internal/score`, the `counting`/`score` tables, `/score`): add a positive-engagement signal — reward group members with score credit when other members tap a positive emoji reaction (👍, ❤️, 🔥, etc.) on their message — and fix two pre-existing scoring defects surfaced while reviewing the pipeline end to end.

## Description

The bot already tracks message-activity engagement (message volume, forwards, media, stickers) per user/day and turns it into a `/score` leaderboard (`internal/score`). Originally scoped as a single feature ("emoji reaction credits") adding a positive-engagement signal on top: Telegram's tap-to-react emoji reactions on messages. Renamed/broadened 2026-08-02 after a full pipeline review (see notes.md "Review findings") surfaced two independent, pre-existing defects in the scoring algorithm itself — both pulled into this spec's scope rather than tracked separately, since they're defects in the same credit-rating process this spec is already touching.

Scope decisions (confirmed with the user):
- **Source**: Telegram tap-to-react reactions only — not emoji characters typed as text in a message.
- **Credit target**: the message **author** (the person whose message got reacted to), not the reactor. Rewards people for content the group responds well to.
- **Positive emoji**: a curated allow-list, not every possible reaction emoji (see plan.md for the proposed list).
- **Integration**: folds into the *existing* `/score` system and its underlying `counting` table / daily recalculation job — not a new, separate credits ledger or command.

## Acceptance criteria

### Emoji reaction credits

- [ ] When a user (not the message's own author) adds a reaction from the positive emoji allow-list to a message, the message author's score increases by 1, reflected after the next `/score recalculate` (or the daily cron).
- [ ] The +1 is applied dated to the *original message's* date, not the date/time the reaction itself arrives — see notes.md "Review findings" #1: awarding it against "now" would corrupt `calculateScore`'s inactivity-gap accounting for a dormant author.
- [ ] Reacting to one's own message never awards credit.
- [ ] A single reaction event carrying multiple new positive emoji at once still only awards +1 (not one point per emoji).
- [ ] Removing a reaction never decrements score; re-adding an already-present reaction (no net-new emoji) never re-awards.
- [ ] Reactions in chats outside `Config.AllowedChats` are ignored.
- [ ] Reactions on messages the bot has no stored author for (i.e., the message predates this feature being deployed — `message_author` is kept forever once populated, no retention/pruning) are silently ignored — not treated as an error.
- [ ] Existing bot behavior (message counting, `/top`, `/bottom`, `/profile`, all other commands) is unaffected.

### Score algorithm fixes (pre-existing defects, pulled into scope 2026-08-02)

- [x] `/score` no longer panics in a chat with fewer than 10 scored users (`ShowScores10`, `internal/score/score.go:24`).
- [x] `calculateScore` continues penalizing a dormant user's inactivity up through the present, instead of freezing their score once their last `counting` row passes (`internal/score/score.go:71-72,83`).

## Out of scope

- Emoji typed as text in message content (not reactions) — not detected or scored.
- Reflecting reaction credit in `/top`, `/bottom`, or `/profile` (these read a separate "engagement volume" query path that this feature deliberately does not touch) — only `/score` reflects it.
- Anti-gaming/dedup for repeated add-remove-add reaction cycling by the same reactor — explicitly deferred; judged consistent with the existing lack of anti-gaming on `Forward`/`Media` counts for a small trusted group.
- Telegram Premium custom emoji reactions (`type == "custom_emoji"`) — not classified as positive/negative, ignored.
