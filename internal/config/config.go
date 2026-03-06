package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Paths
var (
	ConfigDir  = filepath.Join(homeDir(), ".config", "claude-notify-hook")
	EnvFile    = filepath.Join(ConfigDir, ".env")
	LogFile    = filepath.Join(ConfigDir, "hook.log")
	ClaudeDir  = filepath.Join(homeDir(), ".claude")
	SettingsJSON = filepath.Join(ClaudeDir, "settings.json")
	CodexDir   = filepath.Join(homeDir(), ".codex")
	CodexTOML  = filepath.Join(CodexDir, "config.toml")
)

// Hook event types registered in settings.json
var HookEvents = []string{"Notification", "Stop", "SubagentStop"}

// Config holds the .env configuration.
type Config struct {
	BotToken       string
	ChatID         string
	InstallTargets string // "claude", "codex", "claude,codex"
	SoundEnabled   bool   // play audio alert on notification
	SoundFile      string // custom sound file path (empty = platform default)
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// Load reads the .env file and returns a Config.
func Load() (*Config, error) {
	f, err := os.Open(EnvFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := &Config{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := parseEnvLine(line)
		if !ok {
			continue
		}
		switch k {
		case "TELEGRAM_BOT_TOKEN":
			c.BotToken = v
		case "TELEGRAM_CHAT_ID":
			c.ChatID = v
		case "NOTIFY_INSTALL_TARGETS":
			c.InstallTargets = v
		case "NOTIFY_SOUND_ENABLED":
			c.SoundEnabled = v == "true" || v == "1"
		case "NOTIFY_SOUND_FILE":
			c.SoundFile = v
		}
	}
	return c, scanner.Err()
}

// Save writes the Config to the .env file with 0600 permissions.
func (c *Config) Save() error {
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return err
	}
	soundEnabled := "false"
	if c.SoundEnabled {
		soundEnabled = "true"
	}
	content := fmt.Sprintf(
		"TELEGRAM_BOT_TOKEN=%q\nTELEGRAM_CHAT_ID=%q\nNOTIFY_INSTALL_TARGETS=%q\nNOTIFY_SOUND_ENABLED=%s\nNOTIFY_SOUND_FILE=%q\n",
		c.BotToken, c.ChatID, c.InstallTargets, soundEnabled, c.SoundFile,
	)
	return os.WriteFile(EnvFile, []byte(content), 0600)
}

// WantsClaude returns true if install targets include claude.
func (c *Config) WantsClaude() bool {
	return strings.Contains(c.InstallTargets, "claude")
}

// WantsCodex returns true if install targets include codex.
func (c *Config) WantsCodex() bool {
	return strings.Contains(c.InstallTargets, "codex")
}

// parseEnvLine parses KEY="VALUE" or KEY=VALUE.
func parseEnvLine(line string) (key, value string, ok bool) {
	idx := strings.IndexByte(line, '=')
	if idx < 1 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	// Strip surrounding quotes
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return key, value, true
}
