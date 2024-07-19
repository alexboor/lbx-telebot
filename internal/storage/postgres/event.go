package postgres

import (
	"context"
	"fmt"
	"github.com/alexboor/lbx-telebot/internal/model"
)

// CreateNewEvent inserts new event to db without updating
func (s *Storage) CreateNewEvent(ctx context.Context, event model.Event) error {
	query := `
insert into event (name, finished_at, author, result, status, winners)
values ($1, $2, $3, $4, $5, $6)
on conflict (name) do nothing`

	_, err := s.Pool.Exec(
		ctx, query,
		event.Name,       // $1
		event.FinishedAt, // $2
		event.AuthorId,   // $3
		event.Result,     // $4
		event.Status,     // $5
		event.WinnerIds,  // $6
	)

	return err
}

// CloseEvent updates result, winners and status of existing event
func (s *Storage) CloseEvent(ctx context.Context, event model.Event) error {
	query := `
update event
set finished_at = now(),
    result      = $2,
    winners     = $3,
    status      = $4
where name = $1`

	_, err := s.Pool.Exec(
		ctx, query,
		event.Name,      // $1
		event.Result,    // $2
		event.WinnerIds, // $3
		event.Status,    // $4
	)

	return err
}

// StoreBet inserts participant bet for opened event
func (s *Storage) StoreBet(ctx context.Context, event model.Event, userId int64) error {
	query := `
insert into event_participant(event_name, user_id, bet)
values ($1, $2, $3)
on conflict (event_name,user_id) do update set bet = excluded.bet
where (select status from event where name = excluded.event_name) != 'finished';`

	_, err := s.Pool.Exec(
		ctx, query,
		event.Name, // $1
		userId,     // $2
		event.Bet,  // $3
	)

	return err
}

// GetEventParticipantByEventName returns participants of event by given name of event
func (s *Storage) GetEventParticipantByEventName(ctx context.Context, name string) ([]model.Participant, error) {
	query := `
select event_name, user_id, bet
from event_participant
where event_name = $1`

	rows, err := s.Pool.Query(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Participant
	for rows.Next() {
		var p model.Participant
		err := rows.Scan(
			&p.EventName,
			&p.UserId,
			&p.Bet,
		)
		if err != nil {
			return nil, fmt.Errorf("on scan: %v", err)
		}

		result = append(result, p)
	}

	return result, rows.Err()
}

// GetAllEvents returns all events
func (s *Storage) GetAllEvents(ctx context.Context, all bool) ([]model.Event, error) {
	query := `select name, created_at, finished_at, author, result, status, winners
					from event where status = 'opened'`
	if all {
		query = `select name, created_at, finished_at, author, result, status, winners
					from event`
	}

	rows, err := s.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Event
	for rows.Next() {
		var e model.Event
		err := rows.Scan(
			&e.Name,
			&e.CreatedAt,
			&e.FinishedAt,
			&e.AuthorId,
			&e.Result,
			&e.Status,
			&e.WinnerIds,
		)
		if err != nil {
			return nil, fmt.Errorf("on scan: %v", err)
		}

		// it is unnecessary at the moment, but could be used later
		//e.AuthorProfile, err = s.GetProfileById(ctx, e.AuthorId)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get author by id: %v", err)
		//}
		//e.WinnerProfiles, err = s.GetProfilesById(ctx, e.WinnerIds)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get winners by id: %v", err)
		//}

		result = append(result, e)
	}

	return result, rows.Err()
}

// GetEventByName returns event by given name of event
func (s *Storage) GetEventByName(ctx context.Context, name string) (model.Event, error) {
	query := `
select name, created_at, finished_at, author, result, status, winners
from event
where name = $1`

	var event model.Event
	row := s.Pool.QueryRow(ctx, query, name)
	err := row.Scan(
		&event.Name,
		&event.CreatedAt,
		&event.FinishedAt,
		&event.AuthorId,
		&event.Result,
		&event.Status,
		&event.WinnerIds,
	)
	if err != nil {
		return model.Event{}, err
	}

	return event, nil
}

func (s *Storage) GetEventWithWinnersByName(ctx context.Context, name string) (model.Event, error) {
	event, err := s.GetEventByName(ctx, name)
	if err != nil {
		return model.Event{}, fmt.Errorf("failed to get event by name=%v: %v", name, err)
	}

	if len(event.WinnerIds) != 0 {
		event.WinnerProfiles, err = s.GetProfilesById(ctx, event.WinnerIds)
		if err != nil {
			return model.Event{}, fmt.Errorf("failed to get winners for event with name=%v: %v", event.Name, err)
		}
	}

	return event, nil
}

// RemoveBet for the particular user from the given event
func (s *Storage) RemoveBet(ctx context.Context, event model.Event, userId int64) error {
	query := `DELETE FROM event_participant WHERE event_name = $1 AND user_id = $2`
	_, err := s.Pool.Exec(ctx, query, event.Name, userId)
	return err
}
