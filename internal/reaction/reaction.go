// Package reaction credits a message's author when other members tap a positive
// emoji reaction on it, folding the credit into the existing /score system.
package reaction

import (
	"context"
	"errors"
	"log/slog"

	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/alexboor/lbx-telebot/internal/storage"
	"github.com/jackc/pgx/v4"
	tele "gopkg.in/telebot.v3"
)

// PositiveEmoji is the curated allow-list of tap-to-react emoji that award credit.
// Telegram's wire format for the heart reaction is the bare U+2764 without a variation selector.
var PositiveEmoji = map[string]struct{}{
	"👍":   {},
	"❤":   {},
	"🔥":   {},
	"🥰":   {},
	"👏":   {},
	"🎉":   {},
	"🤩":   {},
	"💯":   {},
	"😁":   {},
	"❤‍🔥": {},
}

const reactionTypeEmoji = "emoji"

// newlyAdded returns the emoji present in newReactions but not in oldReactions,
// ignoring custom_emoji reactions.
func newlyAdded(oldReactions, newReactions []tele.Reaction) []string {
	old := make(map[string]struct{}, len(oldReactions))
	for _, r := range oldReactions {
		if r.Type == reactionTypeEmoji {
			old[r.Emoji] = struct{}{}
		}
	}

	var added []string
	for _, r := range newReactions {
		if r.Type != reactionTypeEmoji {
			continue
		}
		if _, ok := old[r.Emoji]; !ok {
			added = append(added, r.Emoji)
		}
	}

	return added
}

// hasNewPositive reports whether any newly-added emoji is in PositiveEmoji.
func hasNewPositive(oldReactions, newReactions []tele.Reaction) bool {
	for _, emoji := range newlyAdded(oldReactions, newReactions) {
		if _, ok := PositiveEmoji[emoji]; ok {
			return true
		}
	}

	return false
}

// Handle awards +1 score credit to a message's author when the message receives a new
// positive reaction from someone other than the author. A single reaction event carrying
// multiple new positive emoji at once still only awards +1. The credit is dated to the
// original message's date (not this event's), so it lands on the counting row the message
// itself already created and never shifts the author's inactivity-gap accounting.
func Handle(ctx context.Context, s storage.Storage, r *tele.MessageReaction) {
	if !hasNewPositive(r.OldReaction, r.NewReaction) {
		return
	}

	authorId, messageDate, err := s.GetMessageAuthor(ctx, r.Chat.ID, int64(r.MessageID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("no stored author for reacted message",
				slog.Any("chat", r.Chat.ID), slog.Any("message", r.MessageID))
		} else {
			slog.Error("failed to get message author", slog.Any("error", err))
		}
		return
	}

	if r.User != nil && r.User.ID == authorId {
		return
	}

	if err := s.Count(ctx, authorId, r.Chat.ID, messageDate, model.Count{PositiveReaction: 1}); err != nil {
		slog.Error("failed to award reaction credit", slog.Any("error", err))
	}
}
