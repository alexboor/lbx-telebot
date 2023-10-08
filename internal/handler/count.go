package handler

import (
	"context"
	"strings"
	"time"

	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/jdkato/prose/v2"
	tele "gopkg.in/telebot.v3"
)

// Count counts incoming messages and stores sender to database
func (h Handler) Count(c tele.Context) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) {
		return nil
	}

	// Store user profile data
	ctx := context.Background()
	profile := model.NewProfile(msg.Sender)
	err := h.Storage.StoreProfile(ctx, profile)
	if err != nil {
		return err
	}

	// Count message
	dt := time.Unix(msg.Unixtime, 0)

	var count model.Count
	count.Message++
	if msg.IsForwarded() {
		count.Forward++
	} else {
		if msg.Audio != nil || msg.Video != nil || msg.Animation != nil || msg.Document != nil || msg.Photo != nil ||
			msg.Voice != nil {
			count.Media++
		}
		if msg.Sticker != nil {
			count.Sticker++
		}
		if msg.IsReply() {
			count.Reply++
		}

		doc, err := prose.NewDocument(strings.ToLower(msg.Text))
		if err != nil {
			return err
		}
		uniqWords := make(map[string]struct{})
		for _, tok := range doc.Tokens() {
			// filtering words only 2 symbols and longer
			// this is the right place to filtering stopwords
			if len(tok.Text) > 1 {
				uniqWords[tok.Text] = struct{}{}
			}
		}
		count.Word = len(uniqWords)
	}

	if err := h.Storage.Count(ctx, msg.Sender.ID, msg.Chat.ID, dt, count); err != nil {
		return err
	}

	return nil
}
