package hook

import (
	"os"
	"strings"
)

const (
	managedBegin = "# BEGIN claude-notify-hook managed block"
	managedEnd   = "# END claude-notify-hook managed block"
)

// AddCodexHook adds the notify command to config.toml inside a managed block.
func AddCodexHook(tomlPath, binaryPath string) error {
	lines, err := readLines(tomlPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove existing managed block if present
	lines = removeManagedBlock(lines)

	// Build the managed block
	block := []string{
		managedBegin,
		`notify = ["` + binaryPath + `", "notify", "Codex"]`,
		managedEnd,
	}

	insertAt := len(lines)
	for i, line := range lines {
		if isTableHeader(line) {
			insertAt = i
			break
		}
	}

	prefix := append([]string{}, lines[:insertAt]...)
	suffix := append([]string{}, lines[insertAt:]...)

	if len(prefix) > 0 && strings.TrimSpace(prefix[len(prefix)-1]) != "" {
		prefix = append(prefix, "")
	}
	prefix = append(prefix, block...)
	if len(suffix) > 0 && strings.TrimSpace(suffix[0]) != "" {
		prefix = append(prefix, "")
	}
	lines = append(prefix, suffix...)

	return writeLines(tomlPath, lines)
}

// RemoveCodexHook removes the managed block from config.toml.
func RemoveCodexHook(tomlPath string) error {
	lines, err := readLines(tomlPath)
	if err != nil {
		return nil // nothing to remove
	}
	cleaned := removeManagedBlock(lines)
	return writeLines(tomlPath, cleaned)
}

// HasCodexHook checks if the managed block exists in config.toml.
func HasCodexHook(tomlPath string) bool {
	lines, err := readLines(tomlPath)
	if err != nil {
		return false
	}
	inTopLevel := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isTableHeader(trimmed) {
			inTopLevel = false
			continue
		}
		if trimmed == managedBegin {
			return inTopLevel
		}
	}
	return false
}

func isTableHeader(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}
	return strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
}

func removeManagedBlock(lines []string) []string {
	var result []string
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == managedBegin {
			inBlock = true
			continue
		}
		if trimmed == managedEnd {
			inBlock = false
			continue
		}
		if !inBlock {
			result = append(result, line)
		}
	}
	return result
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	if content == "" {
		return nil, nil
	}
	return strings.Split(content, "\n"), nil
}

func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}
