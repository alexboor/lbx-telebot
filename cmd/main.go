package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/alexboor/lbx-telebot/internal"
    "github.com/alexboor/lbx-telebot/internal/cfg"
    "github.com/alexboor/lbx-telebot/internal/handler"
    "github.com/alexboor/lbx-telebot/internal/model"
    "github.com/alexboor/lbx-telebot/internal/storage/postgres"
    tele "gopkg.in/telebot.v3"
    "github.com/joho/godotenv"
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
    log.Printf("Sending request to ChatGPT with prompt: %s", prompt)
    url := "https://api.openai.com/v1/chat/completions"
    reqBody := ChatGPTRequest{
        Model: "gpt-4o",
        Messages: []Message{
            {
                Role:    "user",
                Content: prompt,
            },
        },
    }
    jsonReq, err := json.Marshal(reqBody)
    if err != nil {
        log.Printf("Error marshalling request body: %s", err)
        return "", err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
    if err != nil {
        log.Printf("Error creating new HTTP request: %s", err)
        return "", err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending request to ChatGPT: %s", err)
        return "", err
    }
    defer resp.Body.Close()

    var chatGPTResp ChatGPTResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatGPTResp); err != nil {
        log.Printf("Error decoding ChatGPT response: %s", err)
        return "", err
    }

    if len(chatGPTResp.Choices) > 0 {
        log.Printf("Received response from ChatGPT: %s", chatGPTResp.Choices[0].Message.Content)
        return chatGPTResp.Choices[0].Message.Content, nil
    }

    log.Printf("No response from ChatGPT")
    return "", fmt.Errorf("no response from ChatGPT")
}

func main() {
    // Загружаем переменные окружения из файла .env
    if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file")
    }

    cfg.InitLogger()

    log.Println("starting...")
    defer log.Println("finished")

    config := cfg.New()
    ctx := context.Background()

    log.Println("Initializing PostgreSQL connection")
    pg, err := postgres.New(ctx, config.Dsn)
    if err != nil {
        log.Fatalf("error connection to db: %s\n", err)
    }
    log.Println("Connected to PostgreSQL")

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
    log.Println("Bot instance created")

    // getting information about profiles
    uniqUserIds := map[int64]struct{}{}
    for _, chatId := range config.AllowedChats {
        // Convert old group ID to supergroup ID if needed
        if chatId > 0 {
            chatId = -100 + chatId
        }

        log.Printf("Fetching profile IDs for chat: %d", chatId)
        profileIds, err := pg.GetProfileIdsByChatId(ctx, chatId)
        if err != nil {
            log.Printf("failed to get profile ids for chat %d: %s", chatId, err)
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
                log.Printf("failed to get profile info for id %d: %s", id, err)
                continue
            }

            p := model.NewProfile(profile.User)
            if err := pg.StoreProfile(ctx, p); err != nil {
                log.Printf("failed to store profile with id %d: %s", profile.User.ID, err)
            } else {
                log.Printf("Stored profile with id %d", profile.User.ID)
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
        if c.Message().Entities != nil {
            for _, entity := range c.Message().Entities {
                if entity.Type == tele.EntityMention && entity.User != nil && entity.User.ID == bot.Me.ID {
                    question := c.Message().Text
                    log.Printf("Mention detected, querying ChatGPT with question: %s", question)
                    chatGPTResponse, err := queryChatGPT(config.ChatGPTToken, question)
                    if err != nil {
                        log.Printf("failed to query ChatGPT: %s", err)
                        return err
                    }
                    log.Printf("Sending response to chat: %s", chatGPTResponse)
                    return c.Send(chatGPTResponse)
                }
            }
        }
        return nil
    })
    bot.Handle(tele.OnAudio, h.Count)
    bot.Handle(tele.OnVideo, h.Count)
    bot.Handle(tele.OnAnimation, h.Count)
    bot.Handle(tele.OnDocument, h.Count)
    bot.Handle(tele.OnPhoto, h.Count)
    bot.Handle(tele.OnVoice, h.Count)
    bot.Handle(tele.OnSticker, h.Count)

    log.Println("up and listen")
    bot.Start()
}
