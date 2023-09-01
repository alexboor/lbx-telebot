package postgres

import (
	"context"
	"fmt"
	"github.com/alexboor/lbx-telebot/internal/model"
)

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
func (s *Storage) GetTop(ctx context.Context, chatId int64, opt model.Option) ([]model.Profile, error) {
	query := `
select id,
       first_name,
       last_name,
       user_name,
       (select coalesce(sum(val), 0) as cnt
        from word_count
        where date >= $1
          and user_id = id
          and chat_id = $2) as cnt
from profile
where id in (select user_id from word_count where chat_id = $2)
group by id, first_name, last_name, user_name
order by cnt desc
limit $3`

	rows, err := s.Pool.Query(
		ctx, query,
		opt.Date,  // $1
		chatId,    // $2
		opt.Limit, // $3
	)
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
func (s *Storage) GetBottom(ctx context.Context, chatId int64, opt model.Option) ([]model.Profile, error) {
	query := `
select id,
       first_name,
       last_name,
       user_name,
       (select coalesce(sum(val), 0) as cnt
        from word_count
        where date >= $1
          and user_id = id
          and chat_id = $2) as cnt
from profile
where id in (select user_id from word_count where chat_id = $2)
group by id, first_name, last_name, user_name
order by cnt
limit $3`

	rows, err := s.Pool.Query(
		ctx, query,
		opt.Date,  // $1
		chatId,    // $2
		opt.Limit, // $3
	)
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
func (s *Storage) GetProfileByName(ctx context.Context, chatId int64, opt model.Option) (model.Profile, error) {
	query := `
select id,
       first_name,
       last_name,
       user_name,
       (select coalesce(sum(val), 0) as cnt
        from word_count
        where date >= $1
          and user_id = id
          and chat_id = $2) as cnt
from profile
where user_name = $3
group by id, first_name, last_name, user_name`

	var p model.Profile
	err := s.Pool.QueryRow(ctx, query, opt.Date, chatId, opt.Profile).Scan(&p.Id, &p.FirstName, &p.LastName, &p.UserName, &p.Count)
	return p, err
}

// GetProfileById returns profile by given user id and chat id
func (s *Storage) GetProfileById(ctx context.Context, id, chatId int64, opt model.Option) (model.Profile, error) {
	query := `
select id,
       first_name,
       last_name,
       user_name,
       (select coalesce(sum(val), 0) as cnt
        from word_count
        where date >= $1
          and user_id = id
          and chat_id = $2) as cnt
from profile
where id = $3
group by id, first_name, last_name, user_name`

	var p model.Profile
	err := s.Pool.QueryRow(ctx, query, opt.Date, chatId, id).Scan(&p.Id, &p.FirstName, &p.LastName, &p.UserName, &p.Count)
	return p, err
}

// getLen returns number of profiles in database by given chat id
func (s *Storage) getLen(ctx context.Context, chatId int64) (int, error) {
	query := `
select count(*)
from profile
where id in (select distinct user_id
             from word_count
             where chat_id = $1)`

	var cnt int
	err := s.Pool.QueryRow(ctx, query, chatId).Scan(&cnt)
	return cnt, err
}

// GetProfileIdsByChatId returns all uniq user ids for given chat id
func (s *Storage) GetProfileIdsByChatId(ctx context.Context, chatId int64) ([]int64, error) {
	query := `
select distinct user_id
from word_count
where chat_id = $1`

	rows, err := s.Pool.Query(ctx, query, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("on scan: %v", err)
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}