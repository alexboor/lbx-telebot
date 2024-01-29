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
//		It could have variants of payload
//		- create with name of event
//		- close with name of event and result
//		- list to get all events with statuses
//		- bet with name of event and value of bet
//		- result to get result of closed event
//		- share to send event to administered group or channel
//	 - my to handle own bets in the given event
func (h Handler) EventCmd(c tele.Context) error {
	msg := c.Message()

	// bot can check private and group messages
	if !h.IsAllowedGroup(msg) && !h.IsAllowedChat(msg) {
		return nil
	}

	newEvent, ok := model.GetNewEvent(msg.Sender.ID, msg.Payload)
	if !ok {
		_, err := c.Bot().Send(c.Sender(), message.GetEventInstruction(), internal.MarkdownOpt)
		return err
	}

	administeredGroup := map[int64]string{}
	// check admin rights
	if newEvent.Cmd == model.EventCreate ||
		newEvent.Cmd == model.EventClose ||
		//newEvent.Cmd == model.EventList  || // I think everyone could use the list commend
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
	case model.EventList:
		return h.eventList(c, newEvent)
	case model.EventResult:
		return h.eventResult(c, newEvent)
	case model.EventBet:
		return h.eventBet(c, newEvent, msg.Sender.ID)
	case model.EventShare:
		return h.eventShare(c, newEvent, administeredGroup)
	case model.EventMy:
		return h.eventMy(c, newEvent)
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

// eventList shows list of events
func (h Handler) eventList(c tele.Context, e model.Event) error {
	ctx := context.Background()

	showAll := len(e.Opts) > 1 && (e.Opts[1] == "-a" || e.Opts[1] == "all")

	events, err := h.Storage.GetAllEvents(ctx, showAll)
	if err != nil {
		resp := message.GetErrorMessage("getting list of events")
		_, err := c.Bot().Send(c.Sender(), resp, internal.MarkdownOpt)
		return err
	}

	resp := message.GetEventList(events, showAll)
	err = c.Send(resp, internal.MarkdownOpt)
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
	err := c.Send(resp, internal.MarkdownOpt)
	return err
}

// eventShare sends message for sharing event to group or channel
func (h Handler) eventShare(c tele.Context, newEvent model.Event, administeredGroup map[int64]string) error {
	resp, keyboard := message.GetEventShareKeyboard(newEvent.Name, administeredGroup)
	_, err := c.Bot().Send(c.Sender(), resp, keyboard)
	return err
}

// eventMy handle /event my command to work with own bets in the given event
// This command can let see the requester own bet for the requested event and also remove it
// The requester can see his bet from any events including finished, but can delete a bet
// from opened event only.
func (h Handler) eventMy(c tele.Context, e model.Event) error {
	ctx := context.Background()
	errMsg := "Incorrect event name _%s_. You can try use `/event list` command to see all ongoing events"

	if len(e.Opts) > 1 {
		name := e.Opts[1]
		event, _ := h.Storage.GetEventByName(ctx, e.Opts[1])
		if event.Name == "" {
			_ = c.Send(fmt.Sprintf(errMsg, name), internal.MarkdownOpt)
			return errors.New("wrong event in /event my command")
		}

		if len(e.Opts) == 2 {
			//case /event my e.Opts[1]
			parts, err := h.Storage.GetEventParticipantByEventName(ctx, name)
			if err != nil {
				return err
			}

			var bet int64
			var isSet = false
			for _, p := range parts {
				if p.UserId == c.Sender().ID {
					bet = p.Bet
					isSet = true
					break
				}
			}

			if isSet == false {
				if err := c.Send("You haven't been placing a bet in the requesting event"); err != nil {
					return err
				}
				return nil
			}

			if err := c.Send(message.GetMyBets(name, bet)); err != nil {
				return err
			}
			return nil
		}

		if len(e.Opts) == 3 && e.Opts[2] == "rm" {
			//case /event my e.Opts[1] rm
			//TODO do not delete bet from finished event
			fmt.Println("remove my bet from event ", e.Opts[1])

		}

	}

	return nil
}
