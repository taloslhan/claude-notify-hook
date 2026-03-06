package event

import (
	"fmt"
	"strings"
)

// buildMessage constructs the HTML notification message.
func buildMessage(info *Info, payload map[string]interface{}) string {
	var b strings.Builder

	switch info.Event {
	case Notification:
		content := escapeHTML(jsonStr(payload, "message"))
		b.WriteString("<b>🔔 Claude Code 通知</b>\n")
		writeHeader(&b, info)
		b.WriteString("<b>状态：</b>等待输入\n")
		if content != "" {
			b.WriteString("\n")
			b.WriteString(content)
		}

	case Stop:
		summary := escapeHTML(claudeSummary(payload))
		b.WriteString("<b>✅ Claude Code 通知</b>\n")
		writeHeader(&b, info)
		b.WriteString("<b>状态：</b>任务完成")
		if summary != "" {
			b.WriteString("\n\n")
			b.WriteString(summary)
		}

	case SubagentStop:
		summary := escapeHTML(claudeSummary(payload))
		b.WriteString("<b>⚙️ Claude Code 通知</b>\n")
		writeHeader(&b, info)
		b.WriteString("<b>状态：</b>子代理完成")
		if summary != "" {
			b.WriteString("\n\n")
			b.WriteString(summary)
		}

	case AgentTurnComplete:
		summary := escapeHTML(codexSummary(payload))
		b.WriteString("<b>🤖 Codex 通知</b>\n")
		writeHeader(&b, info)
		b.WriteString("<b>状态：</b>任务完成")
		if summary != "" {
			b.WriteString("\n\n")
			b.WriteString(summary)
		}

	default:
		return ""
	}

	return b.String()
}

func writeHeader(b *strings.Builder, info *Info) {
	fmt.Fprintf(b, "<b>主机：</b><code>%s</code>\n", escapeHTML(truncate(info.Hostname, 120)))
	fmt.Fprintf(b, "<b>项目：</b><code>%s</code>\n", escapeHTML(truncate(info.Project, 240)))
	if info.SessionID != "" {
		fmt.Fprintf(b, "<b>会话ID：</b><code>%s</code>\n",
			escapeHTML(truncate(info.SessionID, 120)))
	}
	if info.TurnID != "" {
		fmt.Fprintf(b, "<b>回合ID：</b><code>%s</code>\n",
			escapeHTML(truncate(info.TurnID, 120)))
	}
}

func claudeSummary(payload map[string]interface{}) string {
	return firstOf(payload,
		"last_assistant_message",
		"last-assistant-message",
		"transcript_summary",
	)
}

// codexSummary extracts a summary from Codex payload.
func codexSummary(payload map[string]interface{}) string {
	if s := jsonStr(payload, "last-assistant-message"); s != "" {
		return s
	}
	if s := jsonStr(payload, "message"); s != "" {
		return s
	}
	// Try input-messages[-1].content
	if msgs, ok := payload["input-messages"].([]interface{}); ok && len(msgs) > 0 {
		last := msgs[len(msgs)-1]
		switch v := last.(type) {
		case string:
			return v
		case map[string]interface{}:
			if c, ok := v["content"].(string); ok {
				return c
			}
		}
	}
	return ""
}

// --- Helpers ---

func jsonStr(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func firstOf(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if s := jsonStr(m, k); s != "" {
			return s
		}
	}
	return ""
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func truncate(s string, max int) string {
	if len([]rune(s)) > max {
		return string([]rune(s)[:max]) + "..."
	}
	return s
}
