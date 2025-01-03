package handler

import (
	"fmt"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (h Handler) NewYearCmd(c tele.Context) error {
	msg := c.Message()

	// bot can check private and group messages
	if !h.IsAllowedGroup(msg) && !h.IsAllowedChat(msg) {
		return nil
	}

	if len(msg.Payload) == 0 {
		msg.Payload = "Europe/Podgorica"
	}
	location, err := time.LoadLocation(msg.Payload)
	if err != nil {
		m := fmt.Sprintf("I don't know %s timezone. Try another.\nFor example Europe/Podgorica", msg.Payload)
		return c.Send(m)
	}

	now := time.Now().UTC().In(location)
	newYearTime := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, location)

	m := fmt.Sprintf("New Year in %v will be in %v", location, newYearTime.Sub(now).Round(time.Second))
	return c.Send(m)
}
