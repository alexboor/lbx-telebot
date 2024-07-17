package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/alexboor/lbx-telebot/internal"
    "github.com/alexboor/lbx-telebot/internal/cfg"
    "github.com/alexboor/lbx-telebot/internal/handler"
    "github.com/alexboor/lbx-telebot/internal/model"
    "github.com/alexboor/lbx-telebot/internal/storage/postgres"
    tele "gopkg.in/telebot.v3"
    "github.com/joho/godotenv"
    "log/slog"
)

type ChatGPTRequest struct {
    Model    string   `json:"model"`
    Messages []Message `json:"messages"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatGPTResponse struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

func queryChatGPT(token, prompt string) (string, error) {
    slog.Debug("Sending request to ChatGPT", "prompt", prompt)
    url := "https://api.openai.com/v1/chat/completions"
    reqBody := ChatGPTRequest{
        Model: "gpt-3.5-turbo",
        Messages: []Message{
            {
                Role:    "user",
                Content: prompt,
            },
        },
    }
    jsonReq, err := json.Marshal(reqBody)
    if err != nil {
        slog.Error("Error marshalling request body", "error", err)
        return "", err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
    if err != nil {
        slog.Error("Error creating new HTTP request", "error", err)
        return "", err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        slog.Error("Error sending request to ChatGPT", "error", err)
        return "", err
    }
    defer resp.Body.Close()

    slog.Debug("Request sent to ChatGPT, awaiting response...")

    var chatGPTResp ChatGPTResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatGPTResp); err != nil {
        slog.Error("Error decoding ChatGPT response", "error", err)
        return "", err
    }

    slog.Debug("Received response from ChatGPT", "response", chatGPTResp)

    if len(chatGPTResp.Choices) > 0 {
        slog.Debug("ChatGPT response content", "content", chatGPTResp.Choices[0].Message.Content)
        return chatGPTResp.Choices[0].Message.Content, nil
    }

    slog.Error("No response from ChatGPT")
    return "", fmt.Errorf("no response from ChatGPT")
}

func main() {
    // Загружаем переменные окружения из файла .env
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
    bot.Handle(tele.OnText, func(c tele.Context) error {
        h.Count(c)
        slog.Debug("Received text message", "message", c.Message().Text)
        if c.Message().Entities != nil {
            for _, entity := range c.Message().Entities {
                slog.Debug("Entity detected", "entity", entity)
                // Log the entity details
                slog.Debug("Entity details", "type", entity.Type, "user", entity.User)
                if entity.Type == tele.EntityMention && entity.User != nil && entity.User.ID == bot.Me.ID {
                    question := c.Message().Text
                    slog.Debug("Mention detected, querying ChatGPT", "question", question)
                    slog.Debug("Using ChatGPT token", "token", config.ChatGPTToken)

                    // New log message to confirm function entry
                    slog.Debug("Calling queryChatGPT")

                    chatGPTResponse, err := queryChatGPT(config.ChatGPTToken, question)
                    if err != nil {
                        slog.Error("failed to query ChatGPT", "error", err)
                        return err
                    }

                    // New log message to confirm response received
                    slog.Debug("Received response from ChatGPT", "response", chatGPTResponse)

                    slog.Debug("Sending response to chat", "response", chatGPTResponse)
                    err = c.Send(chatGPTResponse)
                    if err != nil {
                        slog.Error("failed to send response", "error", err)
                        return err
                    }
                    return nil
                }
            }
        }
        slog.Debug("No relevant entity detected")
        return nil
    })
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
