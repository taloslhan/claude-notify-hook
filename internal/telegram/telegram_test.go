package telegram

import (
	"testing"
)

func TestSendMessage_InvalidURL(t *testing.T) {
	// Use a URL that will definitely fail to connect
	// by creating a client with a token that makes the URL invalid
	c := &Client{Token: "fake:token", ChatID: "123"}
	result, err := c.SendMessage("test message")
	if err != nil {
		// Network error path - error is wrapped with "telegram:"
		if got := err.Error(); len(got) == 0 {
			t.Error("error message should not be empty")
		}
		return
	}
	// If the request succeeded (e.g. reached Telegram API), check result
	if result == nil {
		t.Fatal("both err and result are nil")
	}
	// Telegram returns 401 for invalid tokens
	if result.StatusCode == 0 {
		t.Error("StatusCode should not be zero")
	}
}

func TestClientFields(t *testing.T) {
	c := &Client{Token: "tok123", ChatID: "chat456"}
	if c.Token != "tok123" {
		t.Errorf("Token = %q, want tok123", c.Token)
	}
	if c.ChatID != "chat456" {
		t.Errorf("ChatID = %q, want chat456", c.ChatID)
	}
}

func TestSendResult_Fields(t *testing.T) {
	r := &SendResult{StatusCode: 200, Body: `{"ok":true}`}
	if r.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", r.StatusCode)
	}
	if r.Body != `{"ok":true}` {
		t.Errorf("Body = %q", r.Body)
	}
}
