package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/message"
	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/alexboor/lbx-telebot/internal/storage"
	"github.com/jdkato/prose/v2"
	"golang.org/x/exp/maps"
	tele "gopkg.in/telebot.v3"
)

const (
	version = "2.2.1"
	limit   = 5
)

type Handler struct {
	Config  *cfg.Cfg
	Storage storage.Storage
}

// New create and return the handler instance
func New(s storage.Storage, cfg *cfg.Cfg) (*Handler, error) {
	h := &Handler{Storage: s, Config: cfg}
	return h, nil // you cannot get error here TODO: change signature of function
}

// Help handler print help text to private of requester
func (h Handler) Help(c tele.Context) error {
	help := `
*Available commands:*

/help or /h
Show this help.

/ver or /v
Show the current version.

/profile [name]
Show the stored profile of the requester or another user.
Options:
      name - target chat participant.

/top
Show top users.

/bottom
Show reversed rating.
`
	_, err := c.Bot().Send(c.Sender(), help, tele.ParseMode("Markdown"))
	if err != nil {
		return err
	}

	return nil
}

// Ver is handler for command `/ver`
//
//	it's returns version to chat
func (h Handler) Ver(c tele.Context) error {
	return c.Send(version)
}

// Count counts incoming messages and stores sender to database
func (h Handler) Count(c tele.Context) error {
	msg := c.Message()

	if !h.IsChatAllowed(msg) {
		return nil
	}

	// Store user profile data
	profile := model.NewProfile(msg.Sender.ID, msg.Sender.Username, msg.Sender.FirstName, msg.Sender.LastName)
	err := h.Storage.StoreProfile(h.Config.Ctx, profile)
	if err != nil {
		return err
	}

	// Count message
	dt := time.Unix(msg.Unixtime, 0)

	fmt.Printf("%d %d %s: %s\n", msg.Sender.ID, msg.Chat.ID, dt, msg.Text) // todo: mb change to log.Printf?

	doc, err := prose.NewDocument(strings.ToLower(msg.Text))
	if err != nil {
		return err
	}

	m := make(map[string]int)
	for _, tok := range doc.Tokens() {
		// filtering words only 2 symbols and longer
		// this is the right place to filtering stopwords
		if len(tok.Text) > 1 {
			m[tok.Text] = 0
		}
	}

	if err := h.Storage.Count(h.Config.Ctx, msg.Sender.ID, msg.Chat.ID, dt, len(maps.Keys(m))); err != nil {
		return nil // if err != nil return nil? TODO change to return err
	}

	return nil
}

// GetTop is handler for `/top` command, it returns top profiles by count
func (h Handler) GetTop(c tele.Context) error {
	msg := c.Message()

	if !h.IsChatAllowed(msg) {
		return nil
	}

	opt, ok := parseTopAndBottomPayload(msg.Payload)
	if !ok || opt.Limit <= 0 {
		opt.Limit = limit
	}

	profiles, err := h.Storage.GetTop(h.Config.Ctx, msg.Chat.ID, opt)
	if err != nil {
		return fmt.Errorf("failed to get profiles")
	}

	response := message.CreateRating(profiles, opt)
	return c.Send(response)
}

// GetBottom is handler for `/bottom` command, it returns bottom profiles by count
func (h Handler) GetBottom(c tele.Context) error {
	msg := c.Message()

	if !h.IsChatAllowed(msg) {
		return nil
	}

	opt, ok := parseTopAndBottomPayload(msg.Payload)
	if !ok || opt.Limit <= 0 {
		opt.Limit = limit
	}

	profiles, err := h.Storage.GetBottom(h.Config.Ctx, msg.Chat.ID, opt)
	if err != nil {
		return fmt.Errorf("failed to get profiles")
	}

	response := message.CreateRating(profiles, opt)
	return c.Send(response)
}

// GetProfileCount is handler for `/profile` command
//
//	if payload is set then handler returns profile information about username in payload
//	otherwise returns information about sender
func (h Handler) GetProfileCount(c tele.Context) error {
	msg := c.Message()

	if !h.IsChatAllowed(msg) {
		return nil
	}

	var profile model.Profile
	var err error

	opt, ok := parseProfilePayload(msg.Payload)
	if (ok && len(opt.Profile) == 0) || !ok {
		profile, err = h.Storage.GetProfileById(h.Config.Ctx, msg.Sender.ID, msg.Chat.ID, opt)
	} else if ok && len(opt.Profile) != 0 {
		profile, err = h.Storage.GetProfileByName(h.Config.Ctx, msg.Chat.ID, opt)
	}
	if err != nil {
		return fmt.Errorf("failed to get user")
	}

	response := message.CreateUserCount(profile, opt)
	return c.Send(response)
}
