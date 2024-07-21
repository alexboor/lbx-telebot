package main

import (
	"context"
	"log"
	"os"

	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/handler"
	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/alexboor/lbx-telebot/internal/storage/postgres"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
	"log/slog"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	cfg.InitLogger()

	opts := &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Info("starting...")
	defer slog.Info("finished")

	config := cfg.New()
	ctx := context.Background()

	slog.Info("Initializing PostgreSQL connection")
	pg, err := postgres.New(ctx, config.Dsn)
	if err != nil {
		log.Fatalf("error connection to db: %s", err)
	}
	slog.Info("Connected to PostgreSQL")

	h := handler.New(pg, config)
	if err != nil {
		log.Fatalf("error create handler: %s", err)
	}

	optsTele := tele.Settings{
		Token:  config.Token,
		Poller: &tele.LongPoller{Timeout: internal.Timeout},
	}

	bot, err := tele.NewBot(optsTele)
	if err != nil {
		log.Fatalf("error create bot instance: %s", err)
	}
	slog.Info("Bot instance created")

	// getting information about profiles
	uniqUserIds := map[int64]struct{}{}
	for _, chatId := range config.AllowedChats {
		// Convert old group ID to supergroup ID if needed
		if chatId > 0 {
			chatId = -1000000000000 + chatId
		}

		slog.Debug("Fetching profile IDs for chat", "chatId", chatId)
		profileIds, err := pg.GetProfileIdsByChatId(ctx, chatId)
		if err != nil {
			slog.Warn("failed to get profile ids for chat", "chatId", chatId, "error", err)
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
				slog.Warn("failed to get profile info for id", "id", id, "error", err)
				continue
			}

			p := model.NewProfile(profile.User)
			if err := pg.StoreProfile(ctx, p); err != nil {
				slog.Warn("failed to store profile with id", "id", profile.User.ID, "error", err)
			} else {
				slog.Debug("Stored profile with id", "id", profile.User.ID)
			}
		}
	}
	uniqUserIds = nil

	// Commands handlers
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
