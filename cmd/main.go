package main

import (
	"github.com/alexboor/lbx-telebot/internal/handler"
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

}

func main() {
	h, err := handler.New()
	if err != nil {
		log.Fatal(err)
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle("/v", func(c tele.Context) error {
		return c.Send(version)
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		msg := c.Message()

		// Handle only messages in allowed groups (msg.Chat.Type = "group" | "supergroup")
		// private messages handles only by command endpoint handler
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
