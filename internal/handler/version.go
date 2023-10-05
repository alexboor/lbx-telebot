package handler

import (
	"github.com/alexboor/lbx-telebot/internal"
	tele "gopkg.in/telebot.v3"
)

// Ver is handler for command internal.VerCmd
//
//	it returns version to chat
func (h Handler) Ver(c tele.Context) error {
	return c.Send(internal.Version)
}
