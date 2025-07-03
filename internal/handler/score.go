package handler

import (
	"github.com/alexboor/lbx-telebot/internal/score"
	tele "gopkg.in/telebot.v3"
)

func (h Handler) Score(c tele.Context) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) && !h.IsAllowedChat(msg) {
		return nil
	}

	score.CalculateAllScore(h.Storage)

	return nil
}
