package main

import (
	"context"
	"log/slog"

	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/handler"
	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/alexboor/lbx-telebot/internal/storage/postgres"
	tele "gopkg.in/telebot.v3"
	"log"
)

func main() {
	cfg.InitLogger()

	slog.Info("starting...")
	defer slog.Info("finished")

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
		Poller: &tele.LongPoller{Timeout: internal.Timeout},
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
			slog.Error("failed to get profile ids for chat",
				slog.Any("chat", chatId), slog.Any("error", err))
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
				slog.Error("failed to get profile info for id",
					slog.Any("id", id), slog.Any("error", err))
				continue
			}

			p := model.NewProfile(profile.User)
			if err := pg.StoreProfile(ctx, p); err != nil {
				slog.Error("failed to store profile with id",
					slog.Any("id", profile.User.ID), slog.Any("error", err))
			}
		}
	}
	uniqUserIds = nil

	// Commands handlers
	// Should not handle anything except commands in private messages
	bot.Handle(internal.HelpCmd, h.Help)
	bot.Handle(internal.HCmd, h.Help)
	bot.Handle(internal.StartCmd, h.Help)
	bot.Handle(internal.VerCmd, h.Ver)
	bot.Handle(internal.VCmd, h.Ver)

	bot.Handle(internal.TopCmd, h.GetTop)
	bot.Handle(internal.BottomCmd, h.GetBottom)
	bot.Handle(internal.ProfileCmd, h.GetProfileCount)
	bot.Handle(internal.TopicCmd, h.SetTopic)
	bot.Handle(internal.EventCmd, h.EventCmd)
	bot.Handle(internal.TodayCmd, h.TodayCmd)

	// Button handlers
	bot.Handle("\f"+internal.ShareBtn, h.EventCallback)

	// Handle only messages in allowed groups (msg.Chat.Type = "group" | "supergroup")
	// private messages handles only by command endpoint handler
	bot.Handle(tele.OnText, h.Count)
	bot.Handle(tele.OnText, h.HandleChatGPT)
	bot.Handle(tele.OnAudio, h.Count)
	bot.Handle(tele.OnVideo, h.Count)
	bot.Handle(tele.OnAnimation, h.Count)
	bot.Handle(tele.OnDocument, h.Count)
	bot.Handle(tele.OnPhoto, h.Count)
	bot.Handle(tele.OnVoice, h.Count)
	bot.Handle(tele.OnSticker, h.Count)

	slog.Info("up and listen")
	bot.Start()
}
