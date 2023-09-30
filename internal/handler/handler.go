package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"strconv"
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
	version = "2.5.0"
	limit   = 5

	markdownOpt = "Markdown" // TODO change to v2
)

type Handler struct {
	Config  *cfg.Cfg
	Storage storage.Storage
}

// New create and return the handler instance
func New(s storage.Storage, cfg *cfg.Cfg) *Handler {
	return &Handler{Storage: s, Config: cfg}
}

// Help handler print help text to private of requester
func (h Handler) Help(c tele.Context) error {
	_, err := c.Bot().Send(c.Sender(), message.GetHelp(), markdownOpt)
	return err
}

// Ver is handler for command internal.VerCmd
//
//	it returns version to chat
func (h Handler) Ver(c tele.Context) error {
	return c.Send(version)
}

// Count counts incoming messages and stores sender to database
func (h Handler) Count(c tele.Context) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) {
		return nil
	}

	// Store user profile data
	ctx := context.Background()
	profile := model.NewProfile(msg.Sender)
	err := h.Storage.StoreProfile(ctx, profile)
	if err != nil {
		return err
	}

	// Count message
	dt := time.Unix(msg.Unixtime, 0)

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

	if err := h.Storage.Count(ctx, msg.Sender.ID, msg.Chat.ID, dt, len(maps.Keys(m))); err != nil {
		return err
	}

	return nil
}

// GetTop is handler for internal.TopCmd command, it returns top profiles by count
func (h Handler) GetTop(c tele.Context) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) {
		return nil
	}

	opt, ok := parseTopAndBottomPayload(msg.Payload)
	if !ok || opt.Limit <= 0 {
		opt.Limit = limit
	}

	profiles, err := h.Storage.GetTop(context.Background(), msg.Chat.ID, opt)
	if err != nil {
		return fmt.Errorf("failed to get profiles")
	}

	resp := message.CreateRating(profiles, opt)
	return c.Send(resp)
}

// GetBottom is handler for internal.BottomCmd command, it returns bottom profiles by count
func (h Handler) GetBottom(c tele.Context) error {
	msg := c.Message()

	if !h.IsAllowedGroup(msg) {
		return nil
	}

	opt, ok := parseTopAndBottomPayload(msg.Payload)
	if !ok || opt.Limit <= 0 {
		opt.Limit = limit
	}

	profiles, err := h.Storage.GetBottom(context.Background(), msg.Chat.ID, opt)
	if err != nil {
		return fmt.Errorf("failed to get profiles")
	}

	response := message.CreateRating(profiles, opt)
	return c.Send(response)
}

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
	opt, ok := parseProfilePayload(msg.Payload)
	if (ok && len(opt.Profile) == 0) || !ok {
		profile, err = h.Storage.GetProfileStatisticById(ctx, msg.Sender.ID, msg.Chat.ID, opt)
	} else if ok && len(opt.Profile) != 0 {
		profile, err = h.Storage.GetProfileStatisticByName(ctx, msg.Chat.ID, opt)
	}
	if err != nil {
		return fmt.Errorf("failed to get user")
	}

	response := message.CreateUserCount(profile, opt)
	return c.Send(response)
}

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

// EventCmd is handler for internal.EventCmd command
//
//	It could have variants of payload
//	- create with name of event
//	- close with name of event and result
//	- show to get all events with statuses
//	- bet with name of event and value of bet
//	- result to get result of closed event
//	- share to send event to administered group or channel
func (h Handler) EventCmd(c tele.Context) error {
	msg := c.Message()

	// bot can check private and group messages
	if !h.IsAllowedGroup(msg) && !h.IsAllowedChat(msg) {
		return nil
	}

	newEvent, ok := parseEventPayload(msg.Sender.ID, msg.Payload)
	if !ok {
		_, err := c.Bot().Send(c.Sender(), message.GetEventInstruction(), markdownOpt)
		return err
	}

	administeredGroup := map[int64]string{}
	// check admin rights
	if newEvent.Cmd == model.EventCreate || newEvent.Cmd == model.EventClose || newEvent.Cmd == model.EventShow ||
		newEvent.Cmd == model.EventShare {
		var isAdmin bool

		for _, chatId := range h.Config.AllowedChats {
			member, err := c.Bot().ChatMemberOf(tele.ChatID(chatId), &tele.User{ID: msg.Sender.ID})
			if err != nil {
				continue
			}
			if member.Role == tele.Creator || member.Role == tele.Administrator {
				chat, err := c.Bot().ChatByID(chatId)
				if err != nil {
					continue
				}
				administeredGroup[chatId] = chat.Title
				isAdmin = true
			}
		}

		if !isAdmin {
			_, err := c.Bot().Send(c.Sender(), "You do not have permissions to use `/event` command", markdownOpt)
			return err
		}
	}

	ctx := context.Background()

	// creating event
	if newEvent.Cmd == model.EventCreate {
		event, err := h.Storage.GetEventByName(ctx, newEvent.Name)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) { // it is okay that there are no rows with given name
			_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("checking event"), markdownOpt)
			return err
		}

		// check existing event, it shouldn't be event with given name
		if event.Name == newEvent.Name {
			resp := fmt.Sprintf("Event with name `%v` already exists", newEvent.Name)
			return c.Send(resp, markdownOpt)
		}

		// create new event in db
		if err := h.Storage.CreateNewEvent(ctx, newEvent); err != nil {
			_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("creating event"), markdownOpt)
			return err
		}

		resp := message.GetEventCreate(newEvent)
		return c.Send(resp, markdownOpt)
	}

	// closing event
	if newEvent.Cmd == model.EventClose {
		event, err := h.Storage.GetEventByName(ctx, newEvent.Name)
		if errors.Is(err, pgx.ErrNoRows) { // it is not ok that there is no event with given name in db
			resp := fmt.Sprintf("There is no event with name %v", newEvent.Name)
			return c.Send(resp, markdownOpt)
		}

		// event should not be closed
		if event.Status == model.EventStatusFinished {
			resp := fmt.Sprintf("Event %v is already closed!", newEvent.Name)
			return c.Send(resp, markdownOpt)
		}

		// getting participant for event
		participants, err := h.Storage.GetEventParticipantByEventName(ctx, newEvent.Name)
		if err != nil {
			_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("getting participants"), markdownOpt)
			return err
		}
		newEvent.SetWinners(participants)

		// update event in db
		if err := h.Storage.CloseEvent(ctx, newEvent); err != nil {
			_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("closing event"), markdownOpt)
			return err
		}

		// get profiles for winners by ids
		if len(newEvent.WinnerIds) != 0 {
			newEvent.WinnerProfiles, err = h.Storage.GetProfilesById(ctx, newEvent.WinnerIds)
			if err != nil {
				_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("getting winners"), markdownOpt)
				return err
			}
		}

		resp := message.GetEventResult(newEvent)
		return c.Send(resp, markdownOpt)
	}

	// showing list of events
	if newEvent.Cmd == model.EventShow {
		events, err := h.Storage.GetAllEvents(ctx)
		if err != nil {
			_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("getting list of events"), markdownOpt)
			return err
		}

		resp := message.GetEventShow(events)
		_, err = c.Bot().Send(c.Sender(), resp, markdownOpt)
		return err
	}

	if newEvent.Cmd == model.EventResult {
		event, err := h.Storage.GetEventWithWinnersByName(ctx, newEvent.Name)
		if err != nil {
			_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("with getting event"), markdownOpt)
			return err
		}

		// event should be closed
		if event.Status != model.EventStatusFinished {
			resp := fmt.Sprintf("Event %v is still opened", newEvent.Name)
			return c.Send(resp, markdownOpt)
		}

		resp := message.GetEventResult(event)
		return c.Send(resp, markdownOpt)
	}

	// betting value for event
	if newEvent.Cmd == model.EventBet {
		if err := h.Storage.StoreBet(ctx, newEvent, msg.Sender.ID); err != nil {
			_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("saving bet"), markdownOpt)
			return err
		}

		resp := fmt.Sprintf("Your bet `%v` for event `%v` is accepted!", newEvent.Bet, newEvent.Name)
		_, err := c.Bot().Send(c.Sender(), resp, markdownOpt)
		return err
	}

	// send message for sharing event to group or channel
	if newEvent.Cmd == model.EventShare {
		resp, keyboard := message.GetEventShareKeyboard(newEvent.Name, administeredGroup)
		_, err := c.Bot().Send(c.Sender(), resp, keyboard)
		return err
	}

	return nil
}

// EventCallback is handler for internal.ShareBtn callback. It sends event to chosen group
func (h Handler) EventCallback(c tele.Context) error {
	data := strings.Split(c.Data(), " ")
	if len(data) != 2 {
		return fmt.Errorf("failed to get event callback data, callback=%v", data)
	}
	eventName := data[0]
	chatId, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil || len(data[1]) < 2 {
		return fmt.Errorf("failed to get chat_id in event callback, chat_id=%v", data[1])
	}

	event, err := h.Storage.GetEventWithWinnersByName(context.Background(), eventName)
	if err != nil {
		_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("getting event"), markdownOpt)
		return err
	}

	resp := message.GetEventShare(event)
	_, err = c.Bot().Send(&tele.Chat{ID: chatId}, resp, markdownOpt)
	if err != nil {
		_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("sharing event"), markdownOpt)
		return err
	}

	err = c.Delete()
	if err != nil {
		return err
	}

	_, err = c.Bot().Send(c.Sender(), "event shared", markdownOpt)
	return err
}
