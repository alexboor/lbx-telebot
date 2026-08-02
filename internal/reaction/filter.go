package reaction

import (
	"context"

	"github.com/alexboor/lbx-telebot/internal/storage"
	"golang.org/x/exp/slices"
	tele "gopkg.in/telebot.v3"
)

// Filter returns a tele.MiddlewarePoller filter that intercepts message_reaction updates
// and hands them off to Handle. Telebot's normal dispatcher has no handler endpoint for
// reactions, so they're processed here, at the poller level, instead. Every update this
// filter sees that isn't a reaction is passed through unchanged; reaction updates are
// always consumed (never forwarded to the normal dispatcher).
func Filter(s storage.Storage, allowedChats []int64) func(*tele.Update) bool {
	return func(u *tele.Update) bool {
		if u.MessageReaction == nil {
			return true
		}

		r := u.MessageReaction
		if r.Chat == nil || !slices.Contains(allowedChats, r.Chat.ID) {
			return false
		}

		go Handle(context.Background(), s, r)
		return false
	}
}
