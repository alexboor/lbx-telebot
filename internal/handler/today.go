package handler

import (
	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/message"
	"github.com/alexboor/lbx-telebot/internal/wikimedia"
	tele "gopkg.in/telebot.v3"
)

func (h Handler) TodayCmd(c tele.Context) error {
	event, err := wikimedia.GetOnThisDay()
	if err != nil {
		return err
	}

	resp := message.GetTodayMessage(event)
	return c.Send(resp, internal.MarkdownOpt)
}
