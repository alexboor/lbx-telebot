package handler

import (
	"github.com/alexboor/lbx-telebot/internal"
	tele "gopkg.in/telebot.v3"
	"os"
	"regexp"
)

// Ver is handler for command internal.VerCmd
//
//	it returns version to chat
func (h Handler) Ver(c tele.Context) error {
	d, err := os.ReadFile(internal.VersionFile)
	if err != nil {
		return err
	}

	ver := string(d)

	match, err := regexp.MatchString("[0-9]+\\.[0-9]+\\.[0-9]+", ver)
	if err != nil {
		return err
	}

	if match {
		return c.Send(ver)
	}

	return c.Send("unknown version")
}
