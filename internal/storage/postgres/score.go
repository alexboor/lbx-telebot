package postgres

import (
	"context"
	"time"

	"github.com/alexboor/lbx-telebot/internal/model"
)

// StoreScore stores score for a user
func (s *Storage) StoreScore(ctx context.Context, user int64, score int) error {
	query := `insert into score (user_id, score, last_update) values ($1, $2, $3)
		on conflict (user_id) do update set score = excluded.score, last_update = excluded.last_update`
	_, err := s.Pool.Exec(ctx, query, user, score, time.Now())
	return err
}

// GetAllScores returns all scores for all users ordered by score descending
func (s *Storage) GetAllScores(ctx context.Context) ([]model.ProfileWithScore, error) {
	query := `select p.id, p.first_name, p.last_name, p.user_name, s.score, s.last_update from profile p
	join score s on p.id = s.user_id
	order by s.score desc`

	var profiles []model.ProfileWithScore
	rows, err := s.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pws model.ProfileWithScore
		if err := rows.Scan(&pws.Profile.Id, &pws.Profile.FirstName, &pws.Profile.LastName, &pws.Profile.UserName, &pws.Score, &pws.LastUpdate); err != nil {
			return nil, err
		}
		profiles = append(profiles, pws)
	}

	return profiles, nil
}

// GetScore returns score for a user
func (s *Storage) GetScore(ctx context.Context, user int64) (int, error) {
	query := `select p.id, p.first_name, p.last_name, p.user_name, s.score, s.last_update from profile p
	join score s on p.id = s.user_id
	where p.id = $1	`

	var profile model.ProfileWithScore
	if err := s.Pool.QueryRow(ctx, query, user).Scan(&profile.Profile.Id, &profile.Profile.FirstName, &profile.Profile.LastName, &profile.Profile.UserName, &profile.Score, &profile.LastUpdate); err != nil {
		return 0, err
	}

	return profile.Score, nil
}
