package cfg

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Cfg struct {
	Ctx          context.Context
	Token        string
	AllowedChats []int64
	Dsn          string
}

// New creates new app config
func New() *Cfg {
	return &Cfg{
		Ctx:          context.Background(),
		Token:        os.Getenv("TELEGRAM_BOT_TOKEN"),
		AllowedChats: initAllowedChats(),
		Dsn:          initDsn(),
	}
}

// initAllowedChats reads env ALLOW_CHATS, split value by `,` and parses it from string to int
func initAllowedChats() []int64 {
	var result []int64

	allowedChatsEnv := os.Getenv("ALLOW_CHATS")
	for _, chat := range strings.Split(allowedChatsEnv, ",") {
		if len(chat) > 0 {
			id, err := strconv.ParseInt(chat, 10, 64)
			if err != nil {
				log.Printf("warning: parse ALLOW_CHATS env error: %s\n", err)
			}

			result = append(result, id)
		}
	}

	return result
}

// initDsn reads DB_USER, DB_PASSWORD, DB_HOST, DB_PORT and DB_NAME envs and creates DSN for connecting to database
func initDsn() string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	return dsn
}
