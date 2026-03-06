package telegram

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"
)

const apiBase = "https://api.telegram.org/bot"

const maxTelegramTextRunes = 4000

const continuationHeader = "<b>↪️ 续页</b>\n"

// Client sends messages via Telegram Bot API.
type Client struct {
	Token      string
	ChatID     string
	BaseURL    string
	HTTPClient *http.Client
}

// SendResult holds the HTTP status from the API call.
type SendResult struct {
	StatusCode int
	Body       string
}

// SendMessage sends an HTML message. Returns the result or error.
func (c *Client) SendMessage(text string) (*SendResult, error) {
	chunks := splitHTMLMessage(text, maxTelegramTextRunes)
	var result *SendResult
	for _, chunk := range chunks {
		res, err := c.sendMessageChunk(chunk)
		if err != nil {
			return result, err
		}
		result = res
	}
	return result, nil
}

func (c *Client) sendMessageChunk(text string) (*SendResult, error) {
	u := c.baseURL() + "/sendMessage"

	resp, err := c.httpClient().PostForm(u, url.Values{
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

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return apiBase + c.Token
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func splitHTMLMessage(text string, maxRunes int) []string {
	if text == "" {
		return nil
	}
	if runeLen(text) <= maxRunes {
		return []string{text}
	}

	prefix, body, ok := splitPrefixAndBody(text)
	if !ok {
		return []string{text}
	}

	firstLimit := maxRunes - runeLen(prefix)
	continuedLimit := maxRunes - runeLen(continuationHeader)
	if firstLimit <= 0 || continuedLimit <= 0 {
		return []string{text}
	}

	bodyChunks := splitPlainText(htmlUnescape(body), firstLimit, continuedLimit)
	if len(bodyChunks) <= 1 {
		return []string{text}
	}

	chunks := make([]string, 0, len(bodyChunks))
	chunks = append(chunks, prefix+escapeHTML(bodyChunks[0]))
	for _, chunk := range bodyChunks[1:] {
		chunks = append(chunks, continuationHeader+escapeHTML(chunk))
	}
	return chunks
}

func splitPrefixAndBody(text string) (prefix string, body string, ok bool) {
	idx := strings.Index(text, "\n\n")
	if idx < 0 {
		return "", "", false
	}
	return text[:idx+2], text[idx+2:], true
}

func splitPlainText(text string, firstLimit, continuedLimit int) []string {
	if text == "" {
		return []string{""}
	}
	runes := []rune(text)
	parts := make([]string, 0, 2)
	start := 0
	limit := firstLimit
	for start < len(runes) {
		remaining := len(runes) - start
		if remaining <= limit {
			parts = append(parts, string(runes[start:]))
			break
		}

		cut := bestSplitIndex(runes[start:], limit)
		if cut <= 0 {
			cut = limit
		}
		parts = append(parts, string(runes[start:start+cut]))
		start += cut
		limit = continuedLimit
	}
	return parts
}

func bestSplitIndex(runes []rune, limit int) int {
	if len(runes) <= limit {
		return len(runes)
	}
	min := limit / 2
	if min < 1 {
		min = 1
	}
	for i := limit; i >= min; i-- {
		if runes[i-1] == '\n' {
			return i
		}
	}
	for i := limit; i >= min; i-- {
		if unicode.IsSpace(runes[i-1]) {
			return i
		}
	}
	return limit
}

func runeLen(s string) int {
	return len([]rune(s))
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func htmlUnescape(s string) string {
	replacer := strings.NewReplacer(
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&",
	)
	return replacer.Replace(s)
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
