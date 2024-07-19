package cfg

import (
    "fmt"
    "log/slog"
    "os"
    "strconv"
    "strings"
)

type Cfg struct {
    Token        string
    ChatGPTToken string
    AllowedChats []int64
    Dsn          string
    GifUrls      []string
}

// New creates new app config
func New() *Cfg {
    return &Cfg{
        Token:        os.Getenv("TELEGRAM_BOT_TOKEN"),
        ChatGPTToken: os.Getenv("CHATGPT_TOKEN"),
        AllowedChats: initAllowedChats(),
        Dsn:          initDsn(),
        GifUrls:      initGifUrls(),
    }
}

func InitLogger() {
    opts := &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo}
    logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
    slog.SetDefault(logger)
}

// initAllowedChats reads env ALLOW_CHATS, split value by `,` and parses it from string to int
func initAllowedChats() []int64 {
    var result []int64

    allowedChatsEnv := os.Getenv("ALLOW_CHATS")
    for _, chat := range strings.Split(allowedChatsEnv, ",") {
        if len(chat) > 0 {
            id, err := strconv.ParseInt(chat, 10, 64)
            if err != nil {
                slog.Warn("failed to parse ALLOW_CHATS env", slog.Any("error", err))
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

// initGifUrls reads GIF_URLS env and splits it into a slice
func initGifUrls() []string {
    return strings.Split(os.Getenv("GIF_URLS"), ",")
}
