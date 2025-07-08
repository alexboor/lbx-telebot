package handler

import (
	"strings"

	"github.com/alexboor/lbx-telebot/internal/score"
	tele "gopkg.in/telebot.v3"
)

func (h Handler) Score(c tele.Context) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) && !h.IsAllowedChat(msg) {
		return nil
	}

	parts := strings.Split(msg.Payload, " ")
	subcmd := parts[0]
	//subPayload := parts[1:]

	if len(parts) == 0 {

	}

	switch subcmd {
	case "recalculate":
		c.Send(score.CalculateAllScore(h.Storage, h.Config.ScoreTargetChat))
	case "cleanup":
		score.CleanupProfile(h.Storage, h.Config.ScoreTargetChat, c.Bot())
	default:
		c.Send(score.ShowScores10(h.Storage))
	}

	//score.CalculateAllScore(h.Storage)

	return nil
}
