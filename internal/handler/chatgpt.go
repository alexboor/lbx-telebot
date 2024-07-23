package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	tele "gopkg.in/telebot.v3"
	"log/slog"
	"net/http"
	"strings"
)

type ChatGPTRequest struct {
	Model    string    `json:"model"`
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

// HandleChatGPT receiving onText message and traversing it to chatgpt
func (h Handler) HandleChatGPT(c tele.Context) error {
	shouldProcess := false

	if c.Message().Entities != nil {
		for _, entity := range c.Message().Entities {
			if entity.Type == tele.EntityMention {
				if strings.Contains(c.Message().Text[entity.Offset:entity.Offset+entity.Length], c.Bot().Me.Username) {
					shouldProcess = true
					break
				}
			}
		}
	}

	if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender.ID == c.Bot().Me.ID {
		shouldProcess = true
	}

	if shouldProcess {
		if err := processRequest(c, h.Config.ChatGPTToken); err != nil {
			slog.Error("error process request", "error", err)
		}
	}

	return nil
}

// processRequest processing the request
func processRequest(c tele.Context, token string) error {
	prompt := fmt.Sprintf("You are a chat bot named Valera. You are a computer universe pirate. "+
		"Be funny and sarcastic in your responses. Joke with people who ask for your help. Answer only in Russian language. "+
		"Never say that you are not a human. "+
		"Chat member is asking you: %s", c.Message().Text)

	chatGPTResponse, err := queryChatGPT(token, prompt)
	if err != nil {
		return err
	}

	if len(chatGPTResponse) > 0 {
		err = c.Reply(chatGPTResponse)
		if err != nil {
			return err
		}
	}

	return nil
}

// queryChatGPT do integration to ChatGPT API
func queryChatGPT(token, prompt string) (string, error) {
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
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var chatGPTResp ChatGPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatGPTResp); err != nil {
		return "", err
	}

	if len(chatGPTResp.Choices) > 0 {
		return chatGPTResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from ChatGPT")
}
