package storage

import (
	"context"
	"time"

	"github.com/alexboor/lbx-telebot/internal/model"
)

type Storage interface {
	Count(ctx context.Context, uid, cid int64, dt time.Time, val int) error
	StoreProfile(ctx context.Context, profile model.Profile) error
	GetTop(ctx context.Context, chatId int64, opt model.Option) ([]model.Profile, error)
	GetBottom(ctx context.Context, chatId int64, opt model.Option) ([]model.Profile, error)
	GetProfileByName(ctx context.Context, chatId int64, opt model.Option) (model.Profile, error)
	GetProfileById(ctx context.Context, id, chatId int64, opt model.Option) (model.Profile, error)
	GetProfileIdsByChatId(ctx context.Context, chatId int64) ([]int64, error)
}

func NewStorage(s Storage) Storage {
	return s
}
