package main

import (
	"fmt"
	"github.com/alexboor/lbx-telebot/internal/handler"
	"github.com/alexboor/lbx-telebot/internal/storage"
	"github.com/alexboor/lbx-telebot/internal/storage/postgres"
	"golang.org/x/exp/slices" // remember to update after v21.0
	tele "gopkg.in/telebot.v3"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "2.0.0"

var (
	token      string
	allowChats []int64
	dsn        string
)

func init() {
	token = os.Getenv("TELEGRAM_BOT_TOKEN")

	for _, i := range strings.Split(os.Getenv("ALLOW_CHATS"), ",") {
		if len(i) > 0 {
			n, err := strconv.ParseInt(i, 10, 64)
			if err != nil {
				log.Printf("warning: parse ALLOW_CHATS env error: %s\n", err)
			}

			allowChats = append(allowChats, n)
		}
	}

	dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

}

func main() {
	pg, err := postgres.New(dsn)
	if err != nil {
		log.Fatalf("error connection to db: %s\n", err)
	}
	s := storage.NewStorage(pg)

	h, err := handler.New(&s)
	if err != nil {
		log.Fatalf("error create handler: %s\n", err)
	}

	opts := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(opts)
	if err != nil {
		log.Fatalf("error create bot instance: %s\n", err)
	}

	// Commands handlers
	// Should not handle anything except commands in private messages
	bot.Handle("/v", func(c tele.Context) error {
		return c.Send(version)
	})

	// Handle only messages in allowed groups (msg.Chat.Type = "group" | "supergroup")
	// private messages handles only by command endpoint handler
	bot.Handle(tele.OnText, func(c tele.Context) error {
		msg := c.Message()

		if (msg.Chat.Type == "group" || msg.Chat.Type == "supergroup") && slices.Contains(allowChats, msg.Chat.ID) {

			// Store user profile data
			if err := h.StoreProfile(msg.Sender.ID, msg.Sender.Username, msg.Sender.FirstName, msg.Sender.LastName); err != nil {
				return err
			}

			// Count messages
			if err := h.Count(msg.Sender.ID, msg.Chat.ID, msg.Unixtime, msg.Text); err != nil {
				return err
			}
		}

		return nil
	})

	log.Println("up and listen")
	bot.Start()
}
