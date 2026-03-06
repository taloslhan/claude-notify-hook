package event

import (
	"encoding/json"
	"os"
	"strings"
)

// Event types
const (
	Notification      = "Notification"
	Stop              = "Stop"
	SubagentStop      = "SubagentStop"
	AgentTurnComplete = "agent-turn-complete"
)

// Source identifies the caller.
type Source string

const (
	SourceClaude Source = "Claude Code"
	SourceCodex  Source = "Codex"
)

// Info holds the parsed event context.
type Info struct {
	Event     string
	Source    Source
	Hostname  string
	Project   string
	SessionID string
	TurnID    string
	Message   string // final HTML message to send
}

// Detect determines the event type using a 3-tier fallback:
//  1. argHint (CLI argument)
//  2. CLAUDE_HOOK_EVENT env var
//  3. Infer from JSON fields
func Detect(argHint string, payload map[string]interface{}) *Info {
	info := &Info{Source: SourceClaude}

	// Tier 1: CLI argument
	switch argHint {
	case Notification, Stop, SubagentStop:
		info.Event = argHint
	case "Codex", AgentTurnComplete:
		info.Event = AgentTurnComplete
		info.Source = SourceCodex
	}

	// Tier 2: env var
	if info.Event == "" {
		if ev := os.Getenv("CLAUDE_HOOK_EVENT"); ev != "" {
			info.Event = ev
		}
	}

	// Tier 3: infer from JSON
	if info.Event == "" {
		if t := jsonStr(payload, "type"); t != "" {
			if t == AgentTurnComplete {
				info.Event = t
				info.Source = SourceCodex
			} else {
				info.Event = t
			}
		} else if jsonStr(payload, "message") != "" {
			info.Event = Notification
		} else if jsonStr(payload, "transcript_summary") != "" {
			info.Event = Stop
		}
	}

	if info.Event == "" {
		return nil
	}

	// Extract metadata
	hostname, _ := os.Hostname()
	if idx := strings.IndexByte(hostname, '.'); idx > 0 {
		hostname = hostname[:idx]
	}
	info.Hostname = hostname

	if cwd := jsonStr(payload, "cwd"); cwd != "" {
		info.Project = cwd
	} else if wd, err := os.Getwd(); err == nil {
		info.Project = wd
	} else {
		info.Project = "unknown"
	}

	info.SessionID = firstOf(payload, "session_id", "sessionId",
		"conversation_id", "conversationId", "thread-id", "thread_id", "threadId")
	info.TurnID = firstOf(payload, "generation_id", "generationId",
		"turn-id", "turn_id", "turnId")

	info.Message = buildMessage(info, payload)
	return info
}

// ParsePayload parses JSON from raw bytes.
func ParsePayload(data []byte) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}
