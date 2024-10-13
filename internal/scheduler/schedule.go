// Package scheduler provides functionality for scheduling tasks and events.
// It uses gocron library to manage cron jobs. See https://github.com/go-co-op/gocron for more details.
package scheduler

import (
	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/storage"
	"github.com/alexboor/lbx-telebot/internal/storage/memory"
	"github.com/go-co-op/gocron/v2"
	"gopkg.in/telebot.v3"
	"time"
)

type Schedule struct {
	Config        *cfg.Cfg
	Storage       storage.Storage
	Memory        *memory.InMemoryStorage
	Bot           *telebot.Bot
	CronScheduler gocron.Scheduler
}

// New creates and returns a new instance of Schedule.
func New(storage storage.Storage, mem *memory.InMemoryStorage, cfg *cfg.Cfg, bot *telebot.Bot) (*Schedule, error) {
	cronScheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	s := &Schedule{
		Storage:       storage,
		Memory:        mem,
		Config:        cfg,
		Bot:           bot,
		CronScheduler: cronScheduler,
	}

	// Add every 5-min job example
	//
	//_, err = cronScheduler.NewJob(gocron.DurationJob(5*time.Second), gocron.NewTask(s.Ping, time.Now()))

	// Example of job creating which uses im-memory storage package to store counter value
	// See examples.go for details
	//
	//_, err = cronScheduler.NewJob(gocron.DurationJob(5*time.Second), gocron.NewTask(s.InMemoryStorageExample))

	_, err = cronScheduler.NewJob(gocron.DurationJob(10*time.Minute), gocron.NewTask(s.MeteoalarmTask))
	s.MeteoalarmTask()

	return s, nil
}

// Start runs all jobs
func (s *Schedule) Start() {
	s.CronScheduler.Start()
}
