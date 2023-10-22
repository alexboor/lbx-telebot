package handler

import (
	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/message"
	tele "gopkg.in/telebot.v3"
)

// Help handler print help text to private of requester
func (h Handler) Help(c tele.Context) error {
	_, err := c.Bot().Send(c.Sender(), message.GetHelp(), internal.MarkdownOpt)
	return err
}
