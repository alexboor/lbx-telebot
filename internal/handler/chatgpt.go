package handler

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
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

func QueryChatGPT(token, prompt string) (string, error) {
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
