package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/alexboor/lbx-telebot/internal/message"
	"github.com/alexboor/lbx-telebot/internal/model"
	tele "gopkg.in/telebot.v3"
)

// GetProfileCount is handler for internal.ProfileCmd command
//
//	if payload is set then handler returns profile information about username in payload
//	otherwise returns information about sender
func (h Handler) GetProfileCount(c tele.Context) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) {
		return nil
	}

	var profile model.Profile
	var err error

	ctx := context.Background()
	opt, ok := model.NewProfileOption(msg.Payload)
	if (ok && len(opt.Profile) == 0) || !ok {
		profile, err = h.Storage.GetProfileStatisticById(ctx, msg.Sender.ID, msg.Chat.ID, opt)
	} else if ok && len(opt.Profile) != 0 {
		profile, err = h.Storage.GetProfileStatisticByName(ctx, msg.Chat.ID, opt)
	}
	if err != nil {
		return fmt.Errorf("failed to get user")
	}

	filename, err := message.GenerateProfileRatingImage(profile, opt)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	defer func() { _ = os.Remove(filename) }()
	image := &tele.Photo{File: tele.FromDisk(filename)}
	return c.Send(image)
}
