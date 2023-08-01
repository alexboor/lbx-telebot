package handler

import (
	"fmt"
	"time"
)

type Handler struct {
}

// New create and return the handler instance
func New() (*Handler, error) {
	h := &Handler{}

	return h, nil
}

// Count is counting incoming messages
func (h Handler) Count(uid int64, cid int64, t int64, msg string) error {
	dt := time.Unix(t, 0)

	fmt.Printf("%d %d %s: %s\n", uid, cid, dt, msg)

	return nil
}

func (h Handler) StoreProfile(id int64, user string, first string, last string) error {

	return nil
}
