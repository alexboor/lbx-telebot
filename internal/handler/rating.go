package handler

import (
	"context"
	"fmt"

	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/message"
	"github.com/alexboor/lbx-telebot/internal/model"
	tele "gopkg.in/telebot.v3"
)

const (
	top = iota
	bottom
)

// GetTop is handler for internal.TopCmd command, it returns top profiles by count
func (h Handler) GetTop(c tele.Context) error {
	return h.getRating(c, top)
}

// GetBottom is handler for internal.BottomCmd command, it returns bottom profiles by count
func (h Handler) GetBottom(c tele.Context) error {
	return h.getRating(c, bottom)
}

func (h Handler) getRating(c tele.Context, rating int) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) {
		return nil
	}

	opt, ok := model.NewRatingOption(msg.Payload)
	if !ok || opt.Limit <= 0 {
		opt.Limit = internal.RatingLimit
	}

	var (
		profiles []model.Profile
		err      error
		ctx      = context.Background()
	)
	if rating == bottom {
		profiles, err = h.Storage.GetBottom(ctx, msg.Chat.ID, opt)
	} else if rating == top {
		profiles, err = h.Storage.GetTop(ctx, msg.Chat.ID, opt)
	}
	if err != nil {
		return fmt.Errorf("failed to get profiles")
	}

	response := message.CreateRating(profiles, opt)
	return c.Send(response)
}
