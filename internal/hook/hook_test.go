package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Claude Hook (settings.json) tests ---

func TestAddClaudeHook_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	events := []string{"Notification", "Stop", "SubagentStop"}
	bin := "/usr/local/bin/claude-notify-hook"

	if err := AddClaudeHook(path, events, bin); err != nil {
		t.Fatalf("AddClaudeHook error: %v", err)
	}

	if !HasClaudeHook(path, bin) {
		t.Error("HasClaudeHook should be true after add")
	}

	// Verify hook structure omits matcher for all-event registrations.
	raw, _ := os.ReadFile(path)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	hooks, _ := data["hooks"].(map[string]any)
	if hooks == nil {
		t.Fatal("hooks key missing")
	}
	for _, ev := range events {
		arr, ok := hooks[ev].([]any)
		if !ok || len(arr) == 0 {
			t.Errorf("event %q not found or empty in hooks", ev)
			continue
		}
		first, _ := arr[0].(map[string]any)
		if _, exists := first["matcher"]; exists {
			t.Errorf("event %q: matcher should be omitted for all-event hooks", ev)
		}
		innerHooks, ok := first["hooks"].([]any)
		if !ok || len(innerHooks) == 0 {
			t.Errorf("event %q: missing or empty hooks array", ev)
		}
	}
}

func TestAddClaudeHook_NormalizesLegacyManagedMatcher(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	bin := "/usr/local/bin/claude-notify-hook"

	legacy := map[string]any{
		"hooks": map[string]any{
			"Notification": []any{
				map[string]any{
					"matcher": map[string]any{},
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": bin + " notify Notification",
						},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(legacy)
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatalf("seed settings.json: %v", err)
	}

	if err := AddClaudeHook(path, []string{"Notification"}, bin); err != nil {
		t.Fatalf("AddClaudeHook error: %v", err)
	}

	raw, _ = os.ReadFile(path)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	hooks := data["hooks"].(map[string]any)
	arr := hooks["Notification"].([]any)
	if len(arr) != 1 {
		t.Fatalf("expected 1 normalized hook group, got %d", len(arr))
	}
	first := arr[0].(map[string]any)
	if _, exists := first["matcher"]; exists {
		t.Fatal("normalized hook group should omit matcher")
	}
}

func TestAddClaudeHook_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	events := []string{"Notification"}
	bin := "/usr/local/bin/claude-notify-hook"

	AddClaudeHook(path, events, bin)
	AddClaudeHook(path, events, bin) // second call

	raw, _ := os.ReadFile(path)
	var data map[string]any
	json.Unmarshal(raw, &data)
	hooks := data["hooks"].(map[string]any)
	arr := hooks["Notification"].([]any)
	if len(arr) != 1 {
		t.Errorf("expected 1 matcher entry, got %d (not idempotent)", len(arr))
	}
}

func TestRemoveClaudeHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	events := []string{"Notification", "Stop"}
	bin := "/usr/local/bin/claude-notify-hook"

	AddClaudeHook(path, events, bin)
	if err := RemoveClaudeHook(path, events, bin); err != nil {
		t.Fatalf("RemoveClaudeHook error: %v", err)
	}
	if HasClaudeHook(path, bin) {
		t.Error("HasClaudeHook should be false after remove")
	}
}

func TestRemoveClaudeHook_PreservesForeignMatchedHooks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	bin := "/usr/local/bin/claude-notify-hook"

	seed := map[string]any{
		"hooks": map[string]any{
			"Notification": []any{
				map[string]any{
					"matcher": "permission_prompt",
					"hooks": []any{
						map[string]any{"type": "command", "command": "echo foreign"},
					},
				},
				map[string]any{
					"matcher": map[string]any{},
					"hooks": []any{
						map[string]any{"type": "command", "command": bin + " notify Notification"},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(seed)
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatalf("seed settings.json: %v", err)
	}

	if err := RemoveClaudeHook(path, []string{"Notification"}, bin); err != nil {
		t.Fatalf("RemoveClaudeHook error: %v", err)
	}

	raw, _ = os.ReadFile(path)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	hooks := data["hooks"].(map[string]any)
	arr := hooks["Notification"].([]any)
	if len(arr) != 1 {
		t.Fatalf("expected foreign hook group to remain, got %d groups", len(arr))
	}
	first := arr[0].(map[string]any)
	if first["matcher"] != "permission_prompt" {
		t.Fatalf("expected foreign matcher to be preserved, got %#v", first["matcher"])
	}
}

func TestRemoveClaudeHook_FileNotExist(t *testing.T) {
	err := RemoveClaudeHook("/nonexistent/path.json", []string{"Stop"}, "bin")
	if err != nil {
		t.Errorf("RemoveClaudeHook nonexistent should return nil, got %v", err)
	}
}

func TestHasClaudeHook_FileNotExist(t *testing.T) {
	if HasClaudeHook("/nonexistent/path.json", "bin") {
		t.Error("HasClaudeHook should be false for nonexistent file")
	}
}

func TestHasClaudeHook_NoHooks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	os.WriteFile(path, []byte(`{"other":"data"}`), 0644)
	if HasClaudeHook(path, "bin") {
		t.Error("HasClaudeHook should be false when no hooks key")
	}
}

// --- Codex Hook (config.toml) tests ---

func TestAddCodexHook_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	bin := "/usr/local/bin/claude-notify-hook"

	if err := AddCodexHook(path, bin); err != nil {
		t.Fatalf("AddCodexHook error: %v", err)
	}
	if !HasCodexHook(path) {
		t.Error("HasCodexHook should be true after add")
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if strings.Contains(content, "[hooks]") {
		t.Error("managed notify should be written at top level, not inside [hooks]")
	}
	if !strings.Contains(content, managedBegin) {
		t.Error("missing managed block begin")
	}
	if !strings.Contains(content, managedEnd) {
		t.Error("missing managed block end")
	}
	if !strings.Contains(content, bin) {
		t.Error("missing binary path in block")
	}
}

func TestAddCodexHook_ExistingHooksSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	existing := "[hooks]\nexisting = \"value\"\n"
	os.WriteFile(path, []byte(existing), 0644)

	bin := "/usr/local/bin/claude-notify-hook"
	if err := AddCodexHook(path, bin); err != nil {
		t.Fatalf("AddCodexHook error: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if !strings.Contains(content, "existing = \"value\"") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(content, managedBegin) {
		t.Error("managed block should be added")
	}
	if strings.Index(content, managedBegin) > strings.Index(content, "[hooks]") {
		t.Error("managed block should be inserted before the first table so notify stays top-level")
	}
}

func TestAddCodexHook_RepairsLegacyNestedManagedBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	legacy := strings.Join([]string{
		"[hooks]",
		managedBegin,
		`notify = ["/old/path", "notify", "Codex"]`,
		managedEnd,
		"existing = \"value\"",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(legacy), 0644); err != nil {
		t.Fatalf("seed config.toml: %v", err)
	}

	bin := "/usr/local/bin/claude-notify-hook"
	if err := AddCodexHook(path, bin); err != nil {
		t.Fatalf("AddCodexHook error: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if strings.Contains(content, "/old/path") {
		t.Error("legacy nested notify should be replaced")
	}
	if strings.Index(content, managedBegin) > strings.Index(content, "[hooks]") {
		t.Error("repaired managed block should move to top level before [hooks]")
	}
	if !HasCodexHook(path) {
		t.Error("HasCodexHook should detect repaired top-level managed block")
	}
	if !strings.Contains(content, `notify = ["`+bin+`", "notify", "Codex"]`) {
		t.Error("new binary path should be present in top-level notify")
	}
}

func TestAddCodexHook_ReplacesExistingBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	bin := "/usr/local/bin/claude-notify-hook"

	AddCodexHook(path, "/old/path")
	AddCodexHook(path, bin) // replace

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if strings.Contains(content, "/old/path") {
		t.Error("old path should be replaced")
	}
	if !strings.Contains(content, bin) {
		t.Error("new path should be present")
	}
	// Only one managed block
	if strings.Count(content, managedBegin) != 1 {
		t.Errorf("expected 1 managed block, got %d", strings.Count(content, managedBegin))
	}
}

func TestRemoveCodexHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	bin := "/usr/local/bin/claude-notify-hook"

	AddCodexHook(path, bin)
	if err := RemoveCodexHook(path); err != nil {
		t.Fatalf("RemoveCodexHook error: %v", err)
	}
	if HasCodexHook(path) {
		t.Error("HasCodexHook should be false after remove")
	}
}

func TestRemoveCodexHook_FileNotExist(t *testing.T) {
	err := RemoveCodexHook("/nonexistent/config.toml")
	if err != nil {
		t.Errorf("RemoveCodexHook nonexistent should return nil, got %v", err)
	}
}

func TestHasCodexHook_FileNotExist(t *testing.T) {
	if HasCodexHook("/nonexistent/config.toml") {
		t.Error("HasCodexHook should be false for nonexistent file")
	}
}

func TestHasCodexHook_LegacyNestedBlockFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := strings.Join([]string{
		"[hooks]",
		managedBegin,
		`notify = ["bin", "notify", "Codex"]`,
		managedEnd,
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("seed config.toml: %v", err)
	}

	if HasCodexHook(path) {
		t.Error("HasCodexHook should reject legacy nested managed blocks")
	}
}

// --- containsBinary tests ---

func TestContainsBinary(t *testing.T) {
	tests := []struct {
		command string
		binary  string
		want    bool
	}{
		{"/usr/bin/claude-notify-hook", "/usr/bin/claude-notify-hook", true},
		{"/usr/bin/claude-notify-hook notify Stop", "/usr/bin/claude-notify-hook", true},
		{"other-tool", "/usr/bin/claude-notify-hook", false},
		{"claude-notify-hook notify", "/usr/bin/claude-notify-hook", true}, // containsBinaryName fallback
		{"", "/usr/bin/claude-notify-hook", false},
	}
	for _, tt := range tests {
		if got := containsBinary(tt.command, tt.binary); got != tt.want {
			t.Errorf("containsBinary(%q, %q) = %v, want %v",
				tt.command, tt.binary, got, tt.want)
		}
	}
}

// --- removeManagedBlock tests ---

func TestRemoveManagedBlock(t *testing.T) {
	lines := []string{
		"[hooks]",
		managedBegin,
		"notify = [\"bin\"]",
		managedEnd,
		"other = true",
	}
	result := removeManagedBlock(lines)
	if len(result) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(result), result)
	}
	if result[0] != "[hooks]" || result[1] != "other = true" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestRemoveManagedBlock_NoBlock(t *testing.T) {
	lines := []string{"[hooks]", "other = true"}
	result := removeManagedBlock(lines)
	if len(result) != 2 {
		t.Errorf("expected 2 lines unchanged, got %d", len(result))
	}
}
