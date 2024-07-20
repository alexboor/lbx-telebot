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

	return "", fmt.Errorf("no response from ChatGPT")
}

// HandleChatGPT receiving onText message and traversing it to chatgpt
func (h Handler) HandleChatGPT(c tele.Context) {
	respond := false

	if c.Message().Entities != nil {
		for _, entity := range c.Message().Entities {
			if entity.Type == tele.EntityMention {
				if strings.Contains(c.Message().Text[entity.Offset:entity.Offset+entity.Length], c.Bot().Me.Username) {
					respond = true
					break
				}
			}
		}
	}

	if c.Message().ReplyTo != nil {
		// Log reply details
		slog.Debug("Reply detected", "reply_to", c.Message().ReplyTo.ID)
		if c.Message().ReplyTo.Sender.ID == c.Bot().Me.ID {
			respond = true
		}
	}

	if respond {
		question := c.Message().Text
		slog.Debug("Detected a message for ChatGPT, querying", "question", question)
		chatGPTResponse, err := QueryChatGPT(h.Config.ChatGPTToken, question)
		if err != nil {
			slog.Error("failed to query ChatGPT", "error", err)
			err := c.Send("Sorry, there was an error processing your request.")
			if err != nil {
				slog.Error("error processing a request: ", err)
			}
		}
		slog.Debug("Received response from ChatGPT", "response", chatGPTResponse)
		err = c.Reply(chatGPTResponse)
		if err != nil {
			slog.Error(fmt.Sprintf("error to reply: %s", err))
		}
	} else {
		slog.Debug("No relevant entity or reply detected")
	}

	return
}
