package event

import (
	"os"
	"strings"
	"testing"
)

// --- Helper tests ---

func TestJsonStr(t *testing.T) {
	m := map[string]interface{}{"key": "value", "num": 42}

	if got := jsonStr(m, "key"); got != "value" {
		t.Errorf("jsonStr(m, 'key') = %q, want 'value'", got)
	}
	if got := jsonStr(m, "num"); got != "" {
		t.Errorf("jsonStr(m, 'num') = %q, want empty (non-string)", got)
	}
	if got := jsonStr(m, "missing"); got != "" {
		t.Errorf("jsonStr(m, 'missing') = %q, want empty", got)
	}
	if got := jsonStr(nil, "key"); got != "" {
		t.Errorf("jsonStr(nil, 'key') = %q, want empty", got)
	}
}

func TestFirstOf(t *testing.T) {
	m := map[string]interface{}{"b": "found"}

	if got := firstOf(m, "a", "b", "c"); got != "found" {
		t.Errorf("firstOf = %q, want 'found'", got)
	}
	if got := firstOf(m, "x", "y"); got != "" {
		t.Errorf("firstOf no match = %q, want empty", got)
	}
	if got := firstOf(nil, "a"); got != "" {
		t.Errorf("firstOf nil map = %q, want empty", got)
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<script>", "&lt;script&gt;"},
		{"a & <b>", "a &amp; &lt;b&gt;"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := escapeHTML(tt.in); got != tt.want {
			t.Errorf("escapeHTML(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate short = %q, want 'short'", got)
	}
	if got := truncate("abcdef", 3); got != "abc..." {
		t.Errorf("truncate long = %q, want 'abc...'", got)
	}
	if got := truncate("", 5); got != "" {
		t.Errorf("truncate empty = %q, want empty", got)
	}
	// Unicode: 3 runes "你好世" truncated to 2
	if got := truncate("你好世", 2); got != "你好..." {
		t.Errorf("truncate unicode = %q, want '你好...'", got)
	}
}

func TestCodexSummary(t *testing.T) {
	// Priority 1: last-assistant-message
	p1 := map[string]interface{}{
		"last-assistant-message": "from assistant",
		"message":               "from message",
	}
	if got := codexSummary(p1); got != "from assistant" {
		t.Errorf("codexSummary priority1 = %q", got)
	}

	// Priority 2: message
	p2 := map[string]interface{}{"message": "from message"}
	if got := codexSummary(p2); got != "from message" {
		t.Errorf("codexSummary priority2 = %q", got)
	}

	// Priority 3: input-messages[-1] as string
	p3 := map[string]interface{}{
		"input-messages": []interface{}{"first", "last msg"},
	}
	if got := codexSummary(p3); got != "last msg" {
		t.Errorf("codexSummary priority3 string = %q", got)
	}

	// Priority 3: input-messages[-1] as map with content
	p4 := map[string]interface{}{
		"input-messages": []interface{}{
			map[string]interface{}{"content": "map content"},
		},
	}
	if got := codexSummary(p4); got != "map content" {
		t.Errorf("codexSummary priority3 map = %q", got)
	}

	// Empty payload
	if got := codexSummary(map[string]interface{}{}); got != "" {
		t.Errorf("codexSummary empty = %q", got)
	}
	if got := codexSummary(nil); got != "" {
		t.Errorf("codexSummary nil = %q", got)
	}
}

// --- ParsePayload tests ---

func TestParsePayload(t *testing.T) {
	// Valid JSON
	m := ParsePayload([]byte(`{"type":"Stop","message":"hello"}`))
	if m == nil {
		t.Fatal("ParsePayload valid JSON returned nil")
	}
	if m["type"] != "Stop" {
		t.Errorf("ParsePayload type = %v, want Stop", m["type"])
	}

	// Invalid JSON
	if got := ParsePayload([]byte("not json")); got != nil {
		t.Errorf("ParsePayload invalid = %v, want nil", got)
	}

	// Empty input
	if got := ParsePayload([]byte("")); got != nil {
		t.Errorf("ParsePayload empty = %v, want nil", got)
	}

	// Nil input
	if got := ParsePayload(nil); got != nil {
		t.Errorf("ParsePayload nil = %v, want nil", got)
	}
}

// --- Detect tests ---

func TestDetect_Tier1_ArgHint(t *testing.T) {
	payload := map[string]interface{}{"message": "test"}

	tests := []struct {
		argHint    string
		wantEvent  string
		wantSource Source
	}{
		{Notification, Notification, SourceClaude},
		{Stop, Stop, SourceClaude},
		{SubagentStop, SubagentStop, SourceClaude},
		{"Codex", AgentTurnComplete, SourceCodex},
		{AgentTurnComplete, AgentTurnComplete, SourceCodex},
	}
	for _, tt := range tests {
		info := Detect(tt.argHint, payload)
		if info == nil {
			t.Fatalf("Detect(%q) returned nil", tt.argHint)
		}
		if info.Event != tt.wantEvent {
			t.Errorf("Detect(%q).Event = %q, want %q", tt.argHint, info.Event, tt.wantEvent)
		}
		if info.Source != tt.wantSource {
			t.Errorf("Detect(%q).Source = %q, want %q", tt.argHint, info.Source, tt.wantSource)
		}
	}
}

func TestDetect_Tier2_EnvVar(t *testing.T) {
	os.Setenv("CLAUDE_HOOK_EVENT", "Notification")
	defer os.Unsetenv("CLAUDE_HOOK_EVENT")

	payload := map[string]interface{}{"message": "test"}
	info := Detect("", payload)
	if info == nil {
		t.Fatal("Detect with env var returned nil")
	}
	if info.Event != "Notification" {
		t.Errorf("Detect env event = %q, want Notification", info.Event)
	}
}

func TestDetect_Tier3_InferFromJSON(t *testing.T) {
	// type field
	info := Detect("", map[string]interface{}{"type": "Stop"})
	if info == nil || info.Event != "Stop" {
		t.Errorf("Detect infer type field failed")
	}

	// type = agent-turn-complete → Codex
	info = Detect("", map[string]interface{}{"type": AgentTurnComplete})
	if info == nil || info.Event != AgentTurnComplete || info.Source != SourceCodex {
		t.Errorf("Detect infer codex type failed")
	}

	// message field → Notification
	info = Detect("", map[string]interface{}{"message": "hello"})
	if info == nil || info.Event != Notification {
		t.Errorf("Detect infer message failed")
	}

	// transcript_summary → Stop
	info = Detect("", map[string]interface{}{"transcript_summary": "done"})
	if info == nil || info.Event != Stop {
		t.Errorf("Detect infer transcript_summary failed")
	}
}

func TestDetect_NoEvent_ReturnsNil(t *testing.T) {
	if info := Detect("", map[string]interface{}{}); info != nil {
		t.Errorf("Detect empty payload = %+v, want nil", info)
	}
	if info := Detect("", nil); info != nil {
		t.Errorf("Detect nil payload = %+v, want nil", info)
	}
}

func TestDetect_SessionAndTurnID(t *testing.T) {
	payload := map[string]interface{}{
		"message":    "test",
		"session_id": "sess-123",
		"turn_id":    "turn-456",
	}
	info := Detect(Notification, payload)
	if info.SessionID != "sess-123" {
		t.Errorf("SessionID = %q, want sess-123", info.SessionID)
	}
	if info.TurnID != "turn-456" {
		t.Errorf("TurnID = %q, want turn-456", info.TurnID)
	}
}

func TestDetect_AlternateIDFields(t *testing.T) {
	payload := map[string]interface{}{
		"message":        "test",
		"conversationId": "conv-789",
		"generationId":   "gen-012",
	}
	info := Detect(Notification, payload)
	if info.SessionID != "conv-789" {
		t.Errorf("SessionID alt = %q, want conv-789", info.SessionID)
	}
	if info.TurnID != "gen-012" {
		t.Errorf("TurnID alt = %q, want gen-012", info.TurnID)
	}
}

func TestDetect_CwdFromPayload(t *testing.T) {
	payload := map[string]interface{}{
		"message": "test",
		"cwd":     "/my/project",
	}
	info := Detect(Notification, payload)
	if info.Project != "/my/project" {
		t.Errorf("Project = %q, want /my/project", info.Project)
	}
}

func TestDetect_MessageNotEmpty(t *testing.T) {
	payload := map[string]interface{}{"message": "hello world"}
	info := Detect(Notification, payload)
	if info.Message == "" {
		t.Error("Detect Notification should produce non-empty Message")
	}
}

// --- buildMessage tests ---

func TestBuildMessage_Notification(t *testing.T) {
	info := &Info{Event: Notification, Source: SourceClaude, Hostname: "myhost", Project: "/proj"}
	payload := map[string]interface{}{"message": "waiting for input"}
	msg := buildMessage(info, payload)
	if !strings.Contains(msg, "🔔") {
		t.Error("Notification message missing bell emoji")
	}
	if !strings.Contains(msg, "waiting for input") {
		t.Error("Notification message missing content")
	}
}

func TestBuildMessage_Stop(t *testing.T) {
	info := &Info{Event: Stop, Source: SourceClaude, Hostname: "h", Project: "/p"}
	payload := map[string]interface{}{"transcript_summary": "done stuff"}
	msg := buildMessage(info, payload)
	if !strings.Contains(msg, "✅") {
		t.Error("Stop message missing checkmark")
	}
	if !strings.Contains(msg, "任务完成") {
		t.Error("Stop message missing status")
	}
}

func TestBuildMessage_SubagentStop(t *testing.T) {
	info := &Info{Event: SubagentStop, Source: SourceClaude, Hostname: "h", Project: "/p"}
	msg := buildMessage(info, map[string]interface{}{})
	if !strings.Contains(msg, "⚙️") {
		t.Error("SubagentStop message missing gear emoji")
	}
	if !strings.Contains(msg, "子代理完成") {
		t.Error("SubagentStop message missing status")
	}
}

func TestBuildMessage_AgentTurnComplete(t *testing.T) {
	info := &Info{Event: AgentTurnComplete, Source: SourceCodex, Hostname: "h", Project: "/p"}
	payload := map[string]interface{}{"last-assistant-message": "codex done"}
	msg := buildMessage(info, payload)
	if !strings.Contains(msg, "🤖") {
		t.Error("AgentTurnComplete message missing robot emoji")
	}
	if !strings.Contains(msg, "codex done") {
		t.Error("AgentTurnComplete message missing summary")
	}
}

func TestBuildMessage_UnknownEvent(t *testing.T) {
	info := &Info{Event: "Unknown", Hostname: "h", Project: "/p"}
	if msg := buildMessage(info, nil); msg != "" {
		t.Errorf("Unknown event message = %q, want empty", msg)
	}
}
