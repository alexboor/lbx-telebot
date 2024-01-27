package handler

import (
	"golang.org/x/exp/slices"
	tele "gopkg.in/telebot.v3"
)

// IsAllowedGroup checks that chat is group and id of given group is allowed
func (h Handler) IsAllowedGroup(msg *tele.Message) bool {
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" || !slices.Contains(h.Config.AllowedChats, msg.Chat.ID) {
		return false
	}
	return true
}

// IsAllowedChat checks that message is private chat with bot
func (h Handler) IsAllowedChat(msg *tele.Message) bool {
	if msg.Chat.Type == "private" {
		return true
	}
	return false
}
