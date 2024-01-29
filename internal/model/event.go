package model

import (
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	EventStatusFinished = "finished"
	EventStatusOpened   = "opened"

	EventCreate = iota
	EventClose
	EventList
	EventBet
	EventResult
	EventShare
	EventMy
	EventInfo
)

type (
	Event struct {
		Cmd            int
		AuthorId       int64
		AuthorProfile  Profile
		Name           string
		Status         string
		Result         int64
		Bet            int64
		CreatedAt      time.Time
		FinishedAt     time.Time
		WinnerIds      []int64
		WinnerProfiles []Profile
		Opts           []string
	}
)

func GetNewEvent(author int64, payload string) (Event, bool) {
	var result Event

	opts := strings.Split(payload, " ")
	if len(payload) == 0 || len(opts) == 0 {
		return result, false
	}

	switch opts[0] {
	case "create":
		if len(opts) != 2 {
			return result, false
		}
		result = newEvent(EventCreate, opts[1], author, opts)

	case "close":
		if len(opts) != 3 {
			return result, false
		}

		evRes, err := strconv.ParseInt(opts[2], 10, 64)
		if err != nil {
			return result, false
		}
		result = newEvent(EventClose, opts[1], author, opts)
		result.Result = evRes

	case "list":
		if len(opts) > 2 {
			return result, false
		}
		result = newEvent(EventList, "", 0, opts)

	case "result":
		if len(opts) != 2 {
			return result, false
		}
		result = newEvent(EventResult, opts[1], 0, opts)

	case "share":
		if len(opts) != 2 {
			return result, false
		}
		result = newEvent(EventShare, opts[1], 0, opts)

	case "bet":
		if len(opts) != 3 {
			return result, false
		}

		bet, err := strconv.ParseInt(opts[2], 10, 64)
		if err != nil {
			return result, false
		}
		result = newEvent(EventBet, opts[1], author, opts)
		result.Bet = bet

	case "my":
		if len(opts) < 2 || len(opts) > 3 {
			return result, false
		}
		result = newEvent(EventMy, opts[1], author, opts)

	case "info":
		if len(opts) != 2 {
			return result, false
		}
		result = newEvent(EventInfo, opts[1], 0, opts)

	default:
		return result, false
	}

	if (result.Cmd == EventCreate || result.Cmd == EventClose || result.Cmd == EventBet) &&
		(len(result.Name) > 500 || len(result.Name) == 0) {
		return result, false
	}

	return result, true
}

// SetWinners calculates winners for closed event
func (e *Event) SetWinners(participants []Participant) {
	closestResult := math.MaxInt64
	var closestParticipant []Participant
	for _, p := range participants {
		tmpRes := int(math.Abs(float64(e.Result - p.Bet)))
		if tmpRes < closestResult {
			closestResult = tmpRes
			closestParticipant = []Participant{p}
		} else if tmpRes == closestResult {
			closestParticipant = append(closestParticipant, p)
		}
	}

	for _, p := range closestParticipant {
		e.WinnerIds = append(e.WinnerIds, p.UserId)
	}
}

func newEvent(cmd int, name string, author int64, opts []string) Event {
	return Event{
		Cmd:       cmd,
		Name:      name,
		Status:    getStatus(cmd),
		AuthorId:  author,
		WinnerIds: []int64{},
		Opts:      opts,
	}
}

func getStatus(cmd int) string {
	var status string
	if cmd == EventCreate {
		status = EventStatusOpened
	} else if cmd == EventClose {
		status = EventStatusFinished
	}
	return status
}
