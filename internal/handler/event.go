package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/message"
	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/jackc/pgx/v4"
	tele "gopkg.in/telebot.v3"
)

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

	newEvent, ok := model.GetNewEvent(msg.Sender.ID, msg.Payload)
	if !ok {
		resp := message.GetEventInstruction()
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
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
			resp := "You do not have permissions to use `/event` command"
			_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
			return err
		}
	}

	switch newEvent.Cmd {
	case model.EventCreate:
		return h.eventCreate(c, newEvent)
	case model.EventClose:
		return h.eventClose(c, newEvent)
	case model.EventShow:
		return h.eventShow(c)
	case model.EventResult:
		return h.eventResult(c, newEvent)
	case model.EventBet:
		return h.eventBet(c, newEvent, msg.Sender.ID)
	case model.EventShare:
		return h.eventShare(c, newEvent, administeredGroup)
	}

	resp := message.GetEventInstruction()
	_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
	return err
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
		_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("getting event"), internal.MarkdownOpt)
		return err
	}

	resp := message.GetEventShare(event)
	_, err = c.Bot().Send(&tele.Chat{ID: chatId}, resp, internal.MarkdownOpt)
	if err != nil {
		_, err := c.Bot().Send(c.Sender(), message.GetErrorMessage("sharing event"), internal.MarkdownOpt)
		return err
	}

	err = c.Delete()
	if err != nil {
		return err
	}

	_, err = c.Bot().Send(c.Sender(), "Event is shared", internal.MarkdownOpt)
	return err
}

// eventCreate creates given new event
func (h Handler) eventCreate(c tele.Context, newEvent model.Event) error {
	ctx := context.Background()

	event, err := h.Storage.GetEventByName(ctx, newEvent.Name)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) { // it is okay that there are no rows with given name
		resp := message.GetErrorMessage("checking event")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	// check existing event, it shouldn't be event with given name
	if event.Name == newEvent.Name {
		resp := fmt.Sprintf("Event with name `%v` already exists", newEvent.Name)
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	// create new event in db
	if err := h.Storage.CreateNewEvent(ctx, newEvent); err != nil {
		resp := message.GetErrorMessage("creating event")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	resp := message.GetEventCreate(newEvent)
	return c.Send(resp, internal.MarkdownOpt)
}

// eventClose closes given event
func (h Handler) eventClose(c tele.Context, newEvent model.Event) error {
	ctx := context.Background()

	event, err := h.Storage.GetEventByName(ctx, newEvent.Name)
	if errors.Is(err, pgx.ErrNoRows) { // it is not ok that there is no event with given name in db
		resp := fmt.Sprintf("There is no event with name %v", newEvent.Name)
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	// event should not be closed
	if event.Status == model.EventStatusFinished {
		resp := fmt.Sprintf("Event %v is already closed!", newEvent.Name)
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	// getting participant for event
	participants, err := h.Storage.GetEventParticipantByEventName(ctx, newEvent.Name)
	if err != nil {
		resp := message.GetErrorMessage("getting participants")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}
	newEvent.SetWinners(participants)

	// update event in db
	if err := h.Storage.CloseEvent(ctx, newEvent); err != nil {
		resp := message.GetErrorMessage("closing event")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	// get profiles for winners by ids
	if len(newEvent.WinnerIds) != 0 {
		newEvent.WinnerProfiles, err = h.Storage.GetProfilesById(ctx, newEvent.WinnerIds)
		if err != nil {
			resp := message.GetErrorMessage("getting winners")
			_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
			return err
		}
	}

	resp := message.GetEventResult(newEvent)
	return c.Send(resp, internal.MarkdownOpt)
}

// eventShow shows list of events
func (h Handler) eventShow(c tele.Context) error {
	ctx := context.Background()

	events, err := h.Storage.GetAllEvents(ctx)
	if err != nil {
		resp := message.GetErrorMessage("getting list of events")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	resp := message.GetEventShow(events)
	_, err = c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
	return err
}

func (h Handler) eventResult(c tele.Context, newEvent model.Event) error {
	ctx := context.Background()

	event, err := h.Storage.GetEventWithWinnersByName(ctx, newEvent.Name)
	if err != nil {
		resp := message.GetErrorMessage("with getting event")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	// event should be closed
	if event.Status != model.EventStatusFinished {
		resp := fmt.Sprintf("Event %v is still opened", newEvent.Name)
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	resp := message.GetEventResult(event)
	return c.Send(resp, internal.MarkdownOpt)
}

// eventBet stores betting value for event to db
func (h Handler) eventBet(c tele.Context, newEvent model.Event, userId int64) error {
	ctx := context.Background()

	if err := h.Storage.StoreBet(ctx, newEvent, userId); err != nil {
		resp := message.GetErrorMessage("saving bet")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	resp := fmt.Sprintf("Your bet `%v` for event `%v` is accepted!", newEvent.Bet, newEvent.Name)
	_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
	return err
}

// eventShare sends message for sharing event to group or channel
func (h Handler) eventShare(c tele.Context, newEvent model.Event, administeredGroup map[int64]string) error {
	resp, keyboard := message.GetEventShareKeyboard(newEvent.Name, administeredGroup)
	_, err := c.Bot().Send(c.Sender(), resp, keyboard)
	return err
}
