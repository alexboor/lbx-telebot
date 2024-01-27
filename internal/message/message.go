package message

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/wikimedia"
	"gopkg.in/telebot.v3"

	"github.com/alexboor/lbx-telebot/internal/model"
)

// GetHelp returns message for /h or /help command
func GetHelp() string {
	return `
*Available commands:*

/help or /h
Show this help.

/ver or /v
Show the current version.

/profile \[name] \[period]
Show the stored profile of the requester or another user.
Options:
	name - target chat participant
	period - custom period of statistic (e.g. 7d, 72h), should be > 0

/top \[num] \[period]
Show top users.
Options:
	num - custom number of positions to show, should be > 0
	period - custom period of statistic (e.g. 7d, 72h), should be > 0

/bottom \[num] \[period]
Show reversed rating
Options:
	num - custom number of positions to show, should be > 0
	period - custom period of statistic (e.g. 7d, 72h), should be > 0

/topic text
Set new title in the group

/event
Command for event. Send command without params for detailed instructions.

/today
Returns what happened on this day

I live here: https://github.com/alexboor/lbx-telebot
`
}

// GetEventInstruction returns message with instructions for /event command without payload
func GetEventInstruction() string {
	return `
Available commands:

/event create \[name]
Create new event with \[name] option. It could be sent in group chat or in a direct chat with Valera.
You should have admin rights.
Option is required:
	name - Uniq name for new event. Should be one word with chars and digits only 

/event close \[name] \[result]
Close event with \[name] and \[result] options. It could be sent in group chat or in a direct chat with Valera.
You should have admin rights.
Options are required:
	name - Uniq name for existing event. Should be one word with chars and digits only 
	result - Result of the event. Should be number

/event show
Show all event. It could be sent in group chat or in a direct chat with Valera.
You should have admin rights.

/event result \[name]
Show result for event with given name. It could be sent in group chat or in a direct chat with Valera.
Option is required:
	name - Uniq name for existing event. Should be one word with chars and digits only 

/event bet \[name] \[value]
Make your bet with value. It could be sent in group chat or in a direct chat with Valera.
Options are required:
	name - Uniq name for existing event. Should be one word with chars and digits only 
	value - Your bet for this event. Should be number

/event share \[name]
Share event in administered groups
Option is required:
	name - Uniq name for existing event. Should be one word with chars and digits only`
}

// GetEventShareKeyboard returns message and keyboard for `/event share` cmd with groups
func GetEventShareKeyboard(eventName string, groups map[int64]string) (string, *telebot.ReplyMarkup) {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Choose group to share event %s:\n", eventName))

	buttons := &telebot.ReplyMarkup{ResizeKeyboard: true}
	for chatId, title := range groups {
		callbackData := fmt.Sprintf("%v %v", eventName, strconv.FormatInt(chatId, 10))
		btn := buttons.Data(title, internal.ShareBtn, callbackData)
		buttons.Inline(buttons.Row(btn))
	}

	return msg.String(), buttons
}

// GetEventShare returns message for `share` callback
func GetEventShare(event model.Event) string {
	var msg strings.Builder

	switch event.Status {

	case model.EventStatusOpened:
		msg.WriteString(fmt.Sprintf("Event `%v` has been started!\n", event.Name))
		msg.WriteString("Ladies and gentlemen place your bets!\n")
		msg.WriteString(fmt.Sprintf("To place your bet type \n`/event bet %v value`\n", event.Name))
		msg.WriteString("Where value is your bet. It should be integer.\n")
		msg.WriteString("It could be sent in group chat or in a direct chat with Valera.\n")

	case model.EventStatusFinished:
		msg.WriteString(GetEventResult(event))
	}

	return msg.String()
}

// GetEventResult returns message for `/event result` and `/event close` commands
func GetEventResult(event model.Event) string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Event `%v` is closed with result `%v`!\n", event.Name, event.Result))

	switch len(event.WinnerProfiles) {
	case 0:
		msg.WriteString("There are no winners!")
	case 1:
		msg.WriteString(fmt.Sprintf("Winner is %v!", getName(event.WinnerProfiles[0])))
	default:
		msg.WriteString("Winners are ")
		for i, profile := range event.WinnerProfiles {
			if i != 0 {
				msg.WriteString(", ")
			}
			msg.WriteString(getName(profile))
		}
		msg.WriteString("!")
	}

	return msg.String()
}

// GetEventShow returns message for `/event show` command
func GetEventShow(events []model.Event) string {
	var msg strings.Builder

	if len(events) == 0 {
		msg.WriteString("Event list is empty")
	} else {
		msg.WriteString("List of events:\n")
		for _, e := range events {
			msg.WriteString(fmt.Sprintf("`%v` %v\n", e.Name, e.Status))
		}
	}

	return msg.String()
}

// GetErrorMessage return message about internal error with given msg
func GetErrorMessage(msg string) string {
	return fmt.Sprintf("Something went wrong with %v, try again later", msg)
}

// GetEventCreate returns success message for `/event create` message
func GetEventCreate(event model.Event) string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Event `%v` has been created!", event.Name))
	return msg.String()
}

// CreateRating returns message with information about given profiles
func CreateRating(profiles []model.Profile, opt model.Option) string {
	if len(profiles) == 0 {
		return "nothing"
	}

	var result strings.Builder
	if !opt.Date.IsZero() {
		result.WriteString(fmt.Sprintf("statistic since %v", opt.Date.Format("2006-01-02")))
	}
	for _, profile := range profiles {
		if result.Len() != 0 {
			result.WriteString("\n")
		}
		result.WriteString(fmt.Sprintf("%v. %v: %v",
			profile.Position, getName(profile), profile.Count.Total))
	}

	return result.String()
}

func GetTodayMessage(otd wikimedia.OnThisDay) string {
	var result strings.Builder

	switch otd.Type {
	case wikimedia.Birthday:
		result.WriteString("On this day")
		if otd.Year != 0 {
			result.WriteString(fmt.Sprintf(" in %d", otd.Year))
		}
		result.WriteString(fmt.Sprintf(" was born %s", otd.Text))

	case wikimedia.Holiday:
		result.WriteString(fmt.Sprintf("Today is %s", otd.Text))

	case wikimedia.Event:
		result.WriteString("On this day")
		if otd.Year != 0 {
			result.WriteString(fmt.Sprintf(" in %d", otd.Year))
		}
		result.WriteString(fmt.Sprintf(": %s", otd.Text))
	}

	return result.String()
}

// getProfileTitle returns information about given profile
func getProfileTitle(profile model.Profile, opt model.Option) string {
	if profile.Id == 0 {
		return "unknown"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%v", getName(profile)))
	if !opt.Date.IsZero() {
		result.WriteString(fmt.Sprintf(" since %v", opt.Date.Format("2006-01-02")))
	}

	return result.String()
}

// getName checks given profile and returns name
//
//	if there is no first name or last name returns only username
//	if there is no username, but there is first name or last name returns first name + last name
//	otherwise returns 'unknown' + id
func getName(profile model.Profile) string {
	var name string

	switch {

	case len(profile.UserName) != 0 && (len(profile.FirstName) == 0 || len(profile.LastName) == 0):
		name = fmt.Sprintf("@%v", profile.UserName)

	case len(profile.UserName) != 0 && len(profile.FirstName) != 0 && len(profile.LastName) != 0:
		name = strings.TrimSpace(fmt.Sprintf("%v @%v %v", profile.FirstName, profile.UserName, profile.LastName))

	case len(profile.UserName) == 0 && (len(profile.FirstName) != 0 || len(profile.LastName) != 0):
		name = strings.TrimSpace(fmt.Sprintf("%v %v", profile.FirstName, profile.LastName))

	case len(profile.UserName) == 0 && len(profile.FirstName) == 0 && len(profile.LastName) == 0:
		name = fmt.Sprintf("unknown %v", profile.Id)
	}

	return name
}
