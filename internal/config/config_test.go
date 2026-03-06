package config

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempEnv sets up a temp dir as ConfigDir/EnvFile, restores originals after test.
func withTempEnv(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	origDir := ConfigDir
	origEnv := EnvFile
	dir = t.TempDir()
	ConfigDir = dir
	EnvFile = filepath.Join(dir, ".env")
	return dir, func() {
		ConfigDir = origDir
		EnvFile = origEnv
	}
}

func TestParseEnvLine(t *testing.T) {
	tests := []struct {
		line    string
		wantK   string
		wantV   string
		wantOK  bool
	}{
		{"KEY=VALUE", "KEY", "VALUE", true},
		{`KEY="quoted value"`, "KEY", "quoted value", true},
		{"KEY='single'", "KEY", "single", true},
		{"KEY=", "KEY", "", true},
		{"noequals", "", "", false},
		{"=value", "", "", false},
		{"  SPACED = hello  ", "SPACED", "hello", true},
	}
	for _, tt := range tests {
		k, v, ok := parseEnvLine(tt.line)
		if ok != tt.wantOK || k != tt.wantK || v != tt.wantV {
			t.Errorf("parseEnvLine(%q) = (%q,%q,%v), want (%q,%q,%v)",
				tt.line, k, v, ok, tt.wantK, tt.wantV, tt.wantOK)
		}
	}
}

func TestLoadAndSave(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	orig := &Config{
		BotToken:       "123:ABC",
		ChatID:         "-100999",
		InstallTargets: "claude,codex",
	}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if loaded.BotToken != orig.BotToken {
		t.Errorf("BotToken = %q, want %q", loaded.BotToken, orig.BotToken)
	}
	if loaded.ChatID != orig.ChatID {
		t.Errorf("ChatID = %q, want %q", loaded.ChatID, orig.ChatID)
	}
	if loaded.InstallTargets != orig.InstallTargets {
		t.Errorf("InstallTargets = %q, want %q", loaded.InstallTargets, orig.InstallTargets)
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	_, err := Load()
	if err == nil {
		t.Error("Load should fail when .env does not exist")
	}
}

func TestLoad_CommentsAndBlankLines(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	content := "# comment\n\nTELEGRAM_BOT_TOKEN=mytoken\n  \n# another\nTELEGRAM_CHAT_ID=123\n"
	if err := os.WriteFile(EnvFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	c, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if c.BotToken != "mytoken" {
		t.Errorf("BotToken = %q, want mytoken", c.BotToken)
	}
	if c.ChatID != "123" {
		t.Errorf("ChatID = %q, want 123", c.ChatID)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	origDir := ConfigDir
	origEnv := EnvFile
	defer func() { ConfigDir = origDir; EnvFile = origEnv }()

	tmp := t.TempDir()
	ConfigDir = filepath.Join(tmp, "nested", "dir")
	EnvFile = filepath.Join(ConfigDir, ".env")

	c := &Config{BotToken: "tok", ChatID: "id", InstallTargets: "claude"}
	if err := c.Save(); err != nil {
		t.Fatalf("Save with nested dir: %v", err)
	}
	if _, err := os.Stat(EnvFile); err != nil {
		t.Errorf("EnvFile not created: %v", err)
	}
}

func TestWantsClaude(t *testing.T) {
	c := &Config{InstallTargets: "claude,codex"}
	if !c.WantsClaude() {
		t.Error("WantsClaude should be true")
	}
	c.InstallTargets = "codex"
	if c.WantsClaude() {
		t.Error("WantsClaude should be false")
	}
}

func TestWantsCodex(t *testing.T) {
	c := &Config{InstallTargets: "claude,codex"}
	if !c.WantsCodex() {
		t.Error("WantsCodex should be true")
	}
	c.InstallTargets = "claude"
	if c.WantsCodex() {
		t.Error("WantsCodex should be false")
	}
}

func TestLoad_SoundEnabled(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	tests := []struct {
		name    string
		envVal  string
		want    bool
	}{
		{"true string", "true", true},
		{"1 string", "1", true},
		{"false string", "false", false},
		{"0 string", "0", false},
		{"empty string", "", false},
		{"random string", "yes", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := "TELEGRAM_BOT_TOKEN=tok\nTELEGRAM_CHAT_ID=id\n"
			if tt.envVal != "" {
				content += "NOTIFY_SOUND_ENABLED=" + tt.envVal + "\n"
			}
			if err := os.WriteFile(EnvFile, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			c, err := Load()
			if err != nil {
				t.Fatalf("Load error: %v", err)
			}
			if c.SoundEnabled != tt.want {
				t.Errorf("SoundEnabled = %v, want %v", c.SoundEnabled, tt.want)
			}
		})
	}
}

func TestLoad_SoundFile(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	content := "TELEGRAM_BOT_TOKEN=tok\nTELEGRAM_CHAT_ID=id\nNOTIFY_SOUND_FILE=/custom/sound.wav\n"
	if err := os.WriteFile(EnvFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if c.SoundFile != "/custom/sound.wav" {
		t.Errorf("SoundFile = %q, want /custom/sound.wav", c.SoundFile)
	}
}

func TestSave_SoundFields(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	orig := &Config{
		BotToken:     "tok",
		ChatID:       "id",
		SoundEnabled: true,
		SoundFile:    "/my/sound.aiff",
	}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if !loaded.SoundEnabled {
		t.Error("SoundEnabled should be true after roundtrip")
	}
	if loaded.SoundFile != "/my/sound.aiff" {
		t.Errorf("SoundFile = %q, want /my/sound.aiff", loaded.SoundFile)
	}
}

func TestSave_SoundDisabled(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	orig := &Config{
		BotToken:     "tok",
		ChatID:       "id",
		SoundEnabled: false,
		SoundFile:    "",
	}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if loaded.SoundEnabled {
		t.Error("SoundEnabled should be false after roundtrip")
	}
	if loaded.SoundFile != "" {
		t.Errorf("SoundFile = %q, want empty", loaded.SoundFile)
	}
}

func TestLoadAndSave_AllFieldsRoundtrip(t *testing.T) {
	_, cleanup := withTempEnv(t)
	defer cleanup()

	orig := &Config{
		BotToken:       "123:ABC",
		ChatID:         "-100999",
		InstallTargets: "claude,codex",
		SoundEnabled:   true,
		SoundFile:      "/path/to/sound.wav",
	}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if loaded.BotToken != orig.BotToken {
		t.Errorf("BotToken = %q, want %q", loaded.BotToken, orig.BotToken)
	}
	if loaded.ChatID != orig.ChatID {
		t.Errorf("ChatID = %q, want %q", loaded.ChatID, orig.ChatID)
	}
	if loaded.InstallTargets != orig.InstallTargets {
		t.Errorf("InstallTargets = %q, want %q", loaded.InstallTargets, orig.InstallTargets)
	}
	if loaded.SoundEnabled != orig.SoundEnabled {
		t.Errorf("SoundEnabled = %v, want %v", loaded.SoundEnabled, orig.SoundEnabled)
	}
	if loaded.SoundFile != orig.SoundFile {
		t.Errorf("SoundFile = %q, want %q", loaded.SoundFile, orig.SoundFile)
	}
}
