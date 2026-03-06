package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/claude-notify-hook/internal/config"
	"github.com/user/claude-notify-hook/internal/event"
	"github.com/user/claude-notify-hook/internal/sound"
	"github.com/user/claude-notify-hook/internal/telegram"
)

func NewNotifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notify [EventName]",
		Short: "Process hook event and send Telegram notification",
		Args:  cobra.MaximumNArgs(2),
		// Silent failure: never block Claude/Codex
		Run: func(cmd *cobra.Command, args []string) {
			defer func() { recover() }()
			runNotify(args)
		},
	}
}

func runNotify(args []string) {
	cfg, err := config.Load()
	if err != nil || cfg.BotToken == "" || cfg.ChatID == "" {
		return
	}

	// Determine argHint and possible payload arg
	var argHint string
	var payloadArg string
	if len(args) >= 1 {
		argHint = args[0]
	}
	if len(args) >= 2 {
		payloadArg = args[len(args)-1]
	}

	// Read payload: from last arg (if JSON) or stdin
	var raw []byte
	if payloadArg != "" && strings.HasPrefix(strings.TrimSpace(payloadArg), "{") {
		raw = []byte(payloadArg)
	} else {
		raw = readStdin()
	}

	if len(raw) == 0 {
		logMsg("WARN: payload empty")
		return
	}

	payload := event.ParsePayload(raw)
	info := event.Detect(argHint, payload)
	if info == nil {
		logMsg("WARN: could not determine event type")
		return
	}

	if info.Message == "" {
		return
	}

	// Play sound alert (async, non-blocking)
	if cfg.SoundEnabled {
		sound.Play(cfg.SoundFile)
	}

	client := &telegram.Client{Token: cfg.BotToken, ChatID: cfg.ChatID}
	result, err := client.SendMessage(info.Message)

	httpCode := 0
	if result != nil {
		httpCode = result.StatusCode
	}
	logMsg(fmt.Sprintf("Sent: event=%s project=%s http=%d err=%v",
		info.Event, info.Project, httpCode, err))
}

// readStdin reads from stdin with a timeout, returns nil if stdin is a terminal.
func readStdin() []byte {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil
	}
	// Skip if stdin is a terminal
	if fi.Mode()&os.ModeCharDevice != 0 {
		return nil
	}

	done := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(os.Stdin)
		done <- data
	}()

	select {
	case data := <-done:
		return data
	case <-time.After(5 * time.Second):
		return nil
	}
}

func logMsg(msg string) {
	f, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}
