package postgres

import (
	"context"
	"time"
)

// StoreMessageAuthor records the author and date of a message the first time it's seen,
// so a later reaction on it can be credited to the right person on the right day.
// One row per message, never updated once written.
func (s *Storage) StoreMessageAuthor(ctx context.Context, chatId, messageId, userId int64, messageDate time.Time) error {
	query := `
insert into message_author (chat_id, message_id, user_id, message_date)
values ($1, $2, $3, $4)
on conflict (chat_id, message_id) do nothing`

	_, err := s.Pool.Exec(ctx, query, chatId, messageId, userId, messageDate)
	return err
}

// GetMessageAuthor returns the author and original date of a previously stored message.
func (s *Storage) GetMessageAuthor(ctx context.Context, chatId, messageId int64) (int64, time.Time, error) {
	query := `select user_id, message_date from message_author where chat_id = $1 and message_id = $2`

	var userId int64
	var messageDate time.Time
	err := s.Pool.QueryRow(ctx, query, chatId, messageId).Scan(&userId, &messageDate)
	return userId, messageDate, err
}
