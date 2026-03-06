package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendMessage_SendsSingleHTMLMessage(t *testing.T) {
	var gotText string
	var gotParseMode string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		gotText = r.Form.Get("text")
		gotParseMode = r.Form.Get("parse_mode")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := &Client{Token: "tok", ChatID: "123", BaseURL: server.URL + "/bottok"}
	result, err := c.SendMessage("<b>test</b>")
	if err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}
	if result == nil || result.StatusCode != http.StatusOK {
		t.Fatalf("unexpected result: %+v", result)
	}
	if gotText != "<b>test</b>" {
		t.Fatalf("text = %q, want %q", gotText, "<b>test</b>")
	}
	if gotParseMode != "HTML" {
		t.Fatalf("parse_mode = %q, want HTML", gotParseMode)
	}
}

func TestSendMessage_SplitsLongHTMLMessage(t *testing.T) {
	type request struct {
		Text      string
		ParseMode string
	}
	var requests []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		requests = append(requests, request{
			Text:      r.Form.Get("text"),
			ParseMode: r.Form.Get("parse_mode"),
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	body := strings.Repeat("alpha beta gamma delta ", 320)
	message := "<b>🤖 Codex 通知</b>\n<b>状态：</b>任务完成\n\n" + escapeHTML(body)

	c := &Client{Token: "tok", ChatID: "123", BaseURL: server.URL + "/bottok"}
	if _, err := c.SendMessage(message); err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}
	if len(requests) < 2 {
		t.Fatalf("expected split requests, got %d", len(requests))
	}
	if !strings.Contains(requests[0].Text, "<b>🤖 Codex 通知</b>") {
		t.Fatal("first chunk should contain original header")
	}
	if !strings.HasPrefix(requests[1].Text, continuationHeader) {
		t.Fatal("continuation chunk should start with continuation header")
	}
	for i, req := range requests {
		if req.ParseMode != "HTML" {
			t.Fatalf("request %d parse_mode = %q, want HTML", i, req.ParseMode)
		}
		if runeLen(req.Text) > maxTelegramTextRunes {
			t.Fatalf("request %d too large: %d runes", i, runeLen(req.Text))
		}
	}
	joined := requests[0].Text
	for _, req := range requests[1:] {
		joined += strings.TrimPrefix(req.Text, continuationHeader)
	}
	if !strings.Contains(joined, escapeHTML("alpha beta gamma delta")) {
		t.Fatal("joined chunks should retain escaped body content")
	}
}

func TestSendMessage_RequestError(t *testing.T) {
	c := &Client{
		Token:      "tok",
		ChatID:     "123",
		BaseURL:    "http://127.0.0.1:1/bottok",
		HTTPClient: &http.Client{Timeout: 50 * 1000000},
	}
	if _, err := c.SendMessage("<b>test</b>"); err == nil {
		t.Fatal("expected network error")
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

func TestSplitHTMLMessage_PreservesBodyAcrossChunks(t *testing.T) {
	body := strings.Repeat("第一段 第二段 第三段 ", 80)
	message := "<b>🔔 Claude Code 通知</b>\n<b>状态：</b>等待输入\n\n" + escapeHTML(body)

	chunks := splitHTMLMessage(message, 120)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	for i, chunk := range chunks {
		if runeLen(chunk) > 120 {
			t.Fatalf("chunk %d too large: %d runes", i, runeLen(chunk))
		}
	}
	joined := chunks[0]
	for _, chunk := range chunks[1:] {
		joined += strings.TrimPrefix(chunk, continuationHeader)
	}
	if !strings.Contains(joined, escapeHTML("第一段 第二段 第三段")) {
		t.Fatal("joined chunks should contain full escaped body")
	}
}
