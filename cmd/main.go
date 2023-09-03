package main

import (
	"context"
	"log"
	"time"

	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/handler"
	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/alexboor/lbx-telebot/internal/storage/postgres"
	tele "gopkg.in/telebot.v3"
)

func main() {
	config := cfg.New()
	ctx := context.Background()

	pg, err := postgres.New(ctx, config.Dsn)
	if err != nil {
		log.Fatalf("error connection to db: %s\n", err)
	}

	h := handler.New(pg, config)
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

	// getting information about profiles
	uniqUserIds := map[int64]struct{}{}
	for _, chatId := range config.AllowedChats {
		profileIds, err := pg.GetProfileIdsByChatId(ctx, chatId)
		if err != nil {
			log.Printf("failed to get profile ids for chat=%v: %v", chatId, err)
			continue
		}

		for _, id := range profileIds {
			if _, ok := uniqUserIds[id]; !ok {
				uniqUserIds[id] = struct{}{}
			} else {
				continue
			}

			profile, err := bot.ChatMemberOf(tele.ChatID(chatId), &tele.User{ID: id})
			if err != nil {
				log.Printf("failed to get profile info for id=%v: %v", id, err)
				continue
			}

			p := model.NewProfile(profile.User)
			if err := pg.StoreProfile(ctx, p); err != nil {
				log.Printf("failed to store profile with id=%v: %v", profile.User.ID, err)
			}
		}
	}
	uniqUserIds = nil

	// Commands handlers
	// Should not handle anything except commands in private messages
	bot.Handle("/help", h.Help)
	bot.Handle("/h", h.Help)
	bot.Handle("/start", h.Help)
	bot.Handle("/ver", h.Ver)
	bot.Handle("/v", h.Ver)

	bot.Handle("/top", h.GetTop)
	bot.Handle("/bottom", h.GetBottom)
	bot.Handle("/profile", h.GetProfileCount)
	bot.Handle("/topic", h.SetTopic)
	bot.Handle("/event", h.Event)

	// Handle only messages in allowed groups (msg.Chat.Type = "group" | "supergroup")
	// private messages handles only by command endpoint handler
	bot.Handle(tele.OnText, h.Count)

	log.Println("up and listen")
	bot.Start()
}
