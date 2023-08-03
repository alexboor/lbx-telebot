package handler

import (
	"golang.org/x/exp/slices"
	tele "gopkg.in/telebot.v3"
)

func (h Handler) IsChatAllowed(msg *tele.Message) bool {
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" && !slices.Contains(h.Config.AllowedChats, msg.Chat.ID) {
		return false
	}
	return true
}
