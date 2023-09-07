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
	GetProfileStatisticByName(ctx context.Context, chatId int64, opt model.Option) (model.Profile, error)
	GetProfileStatisticById(ctx context.Context, id, chatId int64, opt model.Option) (model.Profile, error)
	GetProfileIdsByChatId(ctx context.Context, chatId int64) ([]int64, error)
	GetEventByName(ctx context.Context, name string) (model.Event, error)
	CreateNewEvent(ctx context.Context, event model.Event) error
	GetEventParticipantByEventName(ctx context.Context, name string) ([]model.Participant, error)
	CloseEvent(ctx context.Context, event model.Event) error
	GetProfilesById(ctx context.Context, ids []int64) ([]model.Profile, error)
	GetProfileById(ctx context.Context, id int64) (model.Profile, error)
	GetAllEvents(ctx context.Context) ([]model.Event, error)
	StoreBet(ctx context.Context, event model.Event, userId int64) error
}
