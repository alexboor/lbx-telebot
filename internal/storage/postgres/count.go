package postgres

import (
	"context"
	"time"

	"github.com/alexboor/lbx-telebot/internal/model"
)

// Count increment messages counter in storage by given value for user in the chat
func (s *Storage) Count(ctx context.Context, uid, cid int64, dt time.Time, count model.Count) error {
	query := `
insert into counting (user_id, chat_id, date, word, reply, forward, media, sticker, message)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
on conflict (user_id, chat_id, date)
    do update set word    = counting.word + excluded.word,
                  reply   = counting.reply + excluded.reply,
                  forward = counting.forward + excluded.forward,
                  media   = counting.media + excluded.media,
                  sticker = counting.sticker + excluded.sticker,
                  message = counting.message + excluded.message`

	_, err := s.Pool.Exec(
		ctx, query,
		uid,           // $1
		cid,           // $2
		dt,            // $3
		count.Word,    // $4
		count.Reply,   // $5
		count.Forward, // $6
		count.Media,   // $7
		count.Sticker, // $8
		count.Message, // $9
	)
	return err
}

// GetAllIds returns all user ids in the chat
func (s *Storage) GetAllIds(ctx context.Context, chatId int64) ([]int64, error) {
	query := `select distinct user_id from counting where chat_id = $1`

	rows, err := s.Pool.Query(ctx, query, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIds []int64
	for rows.Next() {
		var userId int64
		if err := rows.Scan(&userId); err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}

	return userIds, nil
}

// GetAllCountsByUser returns all counts for a user in a chat
func (s *Storage) GetAllCountsByUser(ctx context.Context, chatId int64, userId int64) ([]model.DateCount, error) {
	query := `select date, word, reply, forward, media, sticker, message 
	from counting where chat_id = $1 and user_id = $2 order by date asc`

	rows, err := s.Pool.Query(ctx, query, chatId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []model.DateCount
	for rows.Next() {
		var date time.Time
		var count model.Count
		if err := rows.Scan(&date, &count.Word, &count.Reply, &count.Forward, &count.Media, &count.Sticker, &count.Message); err != nil {
			return nil, err
		}
		counts = append(counts, model.DateCount{Date: date, Count: count})
	}

	return counts, nil
}
