package postgres

import (
	"context"
	"fmt"

	"github.com/alexboor/lbx-telebot/internal/model"
)

const limit = 5

// StoreProfile stores user data to the storage
func (s *Storage) StoreProfile(ctx context.Context, profile model.Profile) error {
	query := `
insert into profile (id, first_name, last_name, user_name)
values ($1, $2, $3, $4)
on conflict (id) do update set first_name = excluded.first_name,
                               last_name  = excluded.last_name,
                               user_name  = excluded.user_name`

	_, err := s.Pool.Exec(
		ctx, query,
		profile.Id,        // $1
		profile.FirstName, // $2
		profile.LastName,  // $3
		profile.UserName,  // $4
	)

	return err
}

// GetTop returns top profiles with position by count in the given chat id
func (s *Storage) GetTop(ctx context.Context, chatId int64) ([]model.Profile, error) {
	query := `
select p.id, p.first_name, p.last_name, p.user_name, wc.val
from profile p
         inner join word_count wc on p.id = wc.user_id and wc.chat_id = $1
order by wc.val desc 
limit $2`

	rows, err := s.Pool.Query(ctx, query, chatId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []model.Profile
	for rows.Next() {
		var p model.Profile
		err := rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.UserName, &p.Count)
		if err != nil {
			return nil, fmt.Errorf("on scan: %v", err)
		}

		profiles = append(profiles, p)
	}

	for i := range profiles {
		profiles[i].Position = i + 1
	}

	return profiles, rows.Err()
}

// GetBottom returns bottom profiles with position by count in given chat id
func (s *Storage) GetBottom(ctx context.Context, chatId int64) ([]model.Profile, error) {
	query := `
select p.id, p.first_name, p.last_name, p.user_name, wc.val
from profile p
         inner join word_count wc on p.id = wc.user_id and wc.chat_id = $1
order by wc.val
limit $2`

	rows, err := s.Pool.Query(ctx, query, chatId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []model.Profile
	for rows.Next() {
		var p model.Profile
		err := rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.UserName, &p.Count)
		if err != nil {
			return nil, fmt.Errorf("on scan: %v", err)
		}

		profiles = append(profiles, p)
	}

	cnt, err := s.getLen(ctx, chatId)
	if err != nil {
		return nil, fmt.Errorf("failed to get number of profiles")
	}

	for i := range profiles {
		profiles[i].Position = cnt
		cnt--
	}

	return profiles, rows.Err()
}

// GetProfileByName returns profile by given username and chat id
func (s *Storage) GetProfileByName(ctx context.Context, userName string, chatId int64) (model.Profile, error) {
	query := `
select p.id, p.first_name, p.last_name, p.user_name, wc.val
from profile p
         inner join word_count wc on p.id = wc.user_id and wc.chat_id = $1 and p.user_name = $2`

	var p model.Profile
	err := s.Pool.QueryRow(ctx, query, chatId, userName).Scan(&p.Id, &p.FirstName, &p.LastName, &p.UserName, &p.Count)
	return p, err
}

// GetProfileById returns profile by given user id and chat id
func (s *Storage) GetProfileById(ctx context.Context, id, chatId int64) (model.Profile, error) {
	query := `
select p.id, p.first_name, p.last_name, p.user_name, wc.val
from profile p
         inner join word_count wc on p.id = wc.user_id and wc.chat_id = $1 and p.id = $2`

	var p model.Profile
	err := s.Pool.QueryRow(ctx, query, chatId, id).Scan(&p.Id, &p.FirstName, &p.LastName, &p.UserName, &p.Count)
	return p, err
}

// getLen returns number of profiles in database by given chat id
func (s *Storage) getLen(ctx context.Context, chatId int64) (int, error) {
	query := `
select count(*)
from profile
where id in (select user_id
             from word_count
             where chat_id = $1)`

	var cnt int
	err := s.Pool.QueryRow(ctx, query, chatId).Scan(&cnt)
	return cnt, err
}
