package handler

import (
	"fmt"
	"github.com/alexboor/lbx-telebot/internal/storage"
	"github.com/jdkato/prose/v2"
	"golang.org/x/exp/maps"
	"log"
	"strings"
	"time"
)

type Handler struct {
	Storage storage.Storage
}

// New create and return the handler instance
func New(s storage.Storage) (*Handler, error) {
	h := &Handler{
		Storage: s,
	}

	return h, nil
}

// Count is counting incoming messages
func (h Handler) Count(uid int64, cid int64, t int64, msg string) error {
	dt := time.Unix(t, 0)

	fmt.Printf("%d %d %s: %s\n", uid, cid, dt, msg)

	doc, err := prose.NewDocument(strings.ToLower(msg))
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]int)
	for _, tok := range doc.Tokens() {
		// filtering words only 2 symbols and longer
		// this is the right place to filtering stopwords
		if len(tok.Text) > 1 {
			m[tok.Text] = 0
		}
	}

	if err := h.Storage.Count(uid, cid, dt, len(maps.Keys(m))); err != nil {
		return nil
	}

	return nil
}

func (h Handler) StoreProfile(id int64, user string, first string, last string) error {
	return nil
}
