package postgres

import (
	"context"
	"time"
)

// Count increment messages counter in storage by given value for user in the chat
func (s *Storage) Count(uid int64, cid int64, dt time.Time, val int) error {
	ctx := context.TODO()
	q := `
		INSERT INTO word_count (user_id, chat_id, date, val)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, chat_id, date) DO UPDATE
        SET val = word_count.val + excluded.val`

	_, err := s.Pool.Exec(ctx, q, uid, cid, dt, val)
	if err != nil {
		return err
	}

	return nil
}
