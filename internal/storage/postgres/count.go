package postgres

import (
	"context"
	"time"
)

// Count increment messages counter in storage by given value for user in the chat
func (s *Storage) Count(ctx context.Context, uid, cid int64, dt time.Time, val int) error {
	query := `
insert into word_count (user_id, chat_id, date, val)
values ($1, $2, $3, $4)
on conflict (user_id, chat_id, date)
    do update set val = word_count.val + excluded.val`

	_, err := s.Pool.Exec(
		ctx, query,
		uid, // $1
		cid, // $2
		dt,  // $3
		val, // $4
	)
	return err
}
