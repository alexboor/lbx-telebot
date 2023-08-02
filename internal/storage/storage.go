package storage

import (
	"time"
)

type Storage interface {
	Count(uid int64, cid int64, dt time.Time, val int) error
	StoreProfile(uid int64, user string, first string, last string) error
}

func NewStorage(s Storage) Storage {
	return s
}
