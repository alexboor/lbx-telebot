package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    // "os"

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

    log.Printf("Request sent to ChatGPT, awaiting response...")

    var chatGPTResp ChatGPTResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatGPTResp); err !=
