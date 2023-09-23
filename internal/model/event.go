package model

import (
	"math"
	"time"
)

const (
	EventStatusFinished = "finished"
	EventStatusOpened   = "opened"

	EventCreate = iota
	EventClose
	EventShow
	EventBet
	EventResult
	EventShare
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
	}
)

func NewEvent(cmd int, name string, author int64) Event {
	return Event{
		Cmd:       cmd,
		Name:      name,
		Status:    getStatus(cmd),
		AuthorId:  author,
		WinnerIds: []int64{},
	}
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

func getStatus(cmd int) string {
	var status string
	if cmd == EventCreate {
		status = EventStatusOpened
	} else if cmd == EventClose {
		status = EventStatusFinished
	}
	return status
}
