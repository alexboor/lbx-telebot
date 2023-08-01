package postgres

import "time"

// Count increment messages counter in storage by given value for user in the chat
func (s Storage) Count(uid int64, cid int64, dt time.Time, val int) error {
	return nil
}
