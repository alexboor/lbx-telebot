package main

import (
	"log"
	"time"

	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/handler"
	"github.com/alexboor/lbx-telebot/internal/storage"
	"github.com/alexboor/lbx-telebot/internal/storage/postgres"
	tele "gopkg.in/telebot.v3"
)

func main() {
	config := cfg.New()

	pg, err := postgres.New(config.Ctx, config.Dsn)
	if err != nil {
		log.Fatalf("error connection to db: %s\n", err)
	}
	s := storage.NewStorage(pg)

	h, err := handler.New(s, config) // TODO: just pass pg instead of s?
	if err != nil {
		log.Fatalf("error create handler: %s\n", err)
	}

	opts := tele.Settings{
		Token:  config.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(opts)
	if err != nil {
		log.Fatalf("error create bot instance: %s\n", err)
	}

	// Commands handlers
	// Should not handle anything except commands in private messages
	bot.Handle("/help", h.Help)
	bot.Handle("/h", h.Help)
	bot.Handle("/ver", h.Ver)
	bot.Handle("/v", h.Ver)

	bot.Handle("/top", h.GetTop)
	bot.Handle("/bottom", h.GetBottom)
	bot.Handle("/profile", h.GetProfileCount)

	// Handle only messages in allowed groups (msg.Chat.Type = "group" | "supergroup")
	// private messages handles only by command endpoint handler
	bot.Handle(tele.OnText, h.Count)

	log.Println("up and listen")
	bot.Start()
}
