package handler

import (
	"fmt"
	"time"

	tele "gopkg.in/telebot.v3"
)

// SetTopic set topic in the chat
func (h Handler) SetTopic(c tele.Context) error {
	msg := c.Message()
	b := c.Bot()

	if !h.IsAllowedGroup(msg) {
		return nil
	}

	if err := b.SetGroupTitle(c.Chat(), msg.Payload); err != nil {
		return fmt.Errorf("failed to set group title: %v", err)
	}

	fmt.Printf("new title: %s dt: %s\n", msg.Payload, time.Now().Format("02-01-2006 15:04:05"))

	return nil
}
