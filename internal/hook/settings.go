package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// HookEntry represents a single hook command.
type HookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// HookGroup represents a Claude Code hook entry with an optional matcher.
type HookGroup struct {
	Matcher any         `json:"matcher,omitempty"`
	Hooks   []HookEntry `json:"hooks"`
}

// AddClaudeHook adds the notify hook to settings.json for the given events.
func AddClaudeHook(settingsPath string, events []string, binaryPath string) error {
	data := readOrInit(settingsPath)

	hooks, _ := data["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}

	for _, ev := range events {
		entry := HookEntry{
			Type:    "command",
			Command: binaryPath + " notify " + ev,
		}
		groups := getHookGroups(hooks, ev)
		hooks[ev] = upsertManagedHookGroup(groups, entry, binaryPath)
	}

	data["hooks"] = hooks
	return writeJSON(settingsPath, data)
}

// RemoveClaudeHook removes the notify hook from settings.json.
func RemoveClaudeHook(settingsPath string, events []string, binaryPath string) error {
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil // nothing to remove
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	hooks, _ := data["hooks"].(map[string]any)
	if hooks == nil {
		return nil
	}

	for _, ev := range events {
		groups := getHookGroups(hooks, ev)
		filtered := filterHookGroups(groups, binaryPath)
		if len(filtered) == 0 {
			delete(hooks, ev)
		} else {
			hooks[ev] = filtered
		}
	}

	data["hooks"] = hooks
	return writeJSON(settingsPath, data)
}

// HasClaudeHook checks if the hook is registered for any event.
func HasClaudeHook(settingsPath string, binaryPath string) bool {
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		return false
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return false
	}
	hooks, _ := data["hooks"].(map[string]any)
	if hooks == nil {
		return false
	}
	for _, ev := range []string{"Notification", "Stop", "SubagentStop"} {
		groups := getHookGroups(hooks, ev)
		if containsHookInGroups(groups, binaryPath) {
			return true
		}
	}
	return false
}

func readOrInit(path string) map[string]any {
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return map[string]any{}
	}
	return data
}

func writeJSON(path string, data map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0644)
}

func getHookGroups(hooks map[string]any, event string) []HookGroup {
	raw, ok := hooks[event]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	var entries []HookGroup
	for _, item := range arr {
		b, _ := json.Marshal(item)
		var group HookGroup
		if json.Unmarshal(b, &group) == nil && group.Hooks != nil {
			entries = append(entries, group)
		}
	}
	return entries
}

func containsHookInGroups(groups []HookGroup, binaryPath string) bool {
	for _, group := range groups {
		for _, h := range group.Hooks {
			if h.Type == "command" && containsBinary(h.Command, binaryPath) {
				return true
			}
		}
	}
	return false
}

func filterHookGroups(groups []HookGroup, binaryPath string) []HookGroup {
	var result []HookGroup
	for _, group := range groups {
		var filtered []HookEntry
		for _, h := range group.Hooks {
			if !(h.Type == "command" && containsBinary(h.Command, binaryPath)) {
				filtered = append(filtered, h)
			}
		}
		if len(filtered) > 0 {
			group.Hooks = filtered
			result = append(result, group)
		}
	}
	return result
}

func upsertManagedHookGroup(groups []HookGroup, entry HookEntry, binaryPath string) []HookGroup {
	filtered := filterHookGroups(groups, binaryPath)
	return append(filtered, HookGroup{
		Hooks: []HookEntry{entry},
	})
}

func containsBinary(command, binaryPath string) bool {
	return len(command) >= len(binaryPath) &&
		(command == binaryPath ||
			len(command) > len(binaryPath) && command[:len(binaryPath)+1] == binaryPath+" ") ||
		// Also match by binary name only
		containsBinaryName(command)
}

func containsBinaryName(command string) bool {
	return len(command) >= len("claude-notify-hook") &&
		(command[:len("claude-notify-hook")] == "claude-notify-hook" ||
			indexOf(command, "claude-notify-hook") >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
