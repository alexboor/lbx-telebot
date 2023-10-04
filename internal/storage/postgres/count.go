package postgres

import (
	"context"
	"github.com/alexboor/lbx-telebot/internal/model"
	"time"
)

// Count increment messages counter in storage by given value for user in the chat
func (s *Storage) Count(ctx context.Context, uid, cid int64, dt time.Time, count model.Count) error {
	query := `
insert into counting (user_id, chat_id, date, word, reply, forward, media, sticker)
values ($1, $2, $3, $4, $5, $6, $7, $8)
on conflict (user_id, chat_id, date)
    do update set word    = counting.word + excluded.word,
                  reply   = counting.reply + excluded.reply,
                  forward = counting.forward + excluded.forward,
                  media   = counting.media + excluded.media,
                  sticker = counting.sticker + excluded.sticker`

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
	)
	return err
}
