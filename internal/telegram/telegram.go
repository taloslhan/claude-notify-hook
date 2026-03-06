package telegram

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiBase = "https://api.telegram.org/bot"

// Client sends messages via Telegram Bot API.
type Client struct {
	Token  string
	ChatID string
}

// SendResult holds the HTTP status from the API call.
type SendResult struct {
	StatusCode int
	Body       string
}

// SendMessage sends an HTML message. Returns the result or error.
func (c *Client) SendMessage(text string) (*SendResult, error) {
	u := apiBase + c.Token + "/sendMessage"

	resp, err := (&http.Client{Timeout: 10 * time.Second}).PostForm(u, url.Values{
		"chat_id":    {c.ChatID},
		"parse_mode": {"HTML"},
		"text":       {text},
	})
	if err != nil {
		return nil, fmt.Errorf("telegram: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return &SendResult{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}, nil
}
