package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/user/claude-notify-hook/internal/config"
)

func TestChooseManagedBinaryPath_PrefersStableCurrentExecutable(t *testing.T) {
	current := filepath.Join(string(filepath.Separator), "usr", "local", "bin", managedBinaryName)
	installed := filepath.Join(string(filepath.Separator), "opt", "homebrew", "bin", managedBinaryName)

	got := chooseManagedBinaryPath(current, installed, "")
	if got != current {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, current)
	}
}

func TestChooseManagedBinaryPath_PrefersManagedInstall(t *testing.T) {
	current := filepath.Join(string(filepath.Separator), "usr", "local", "bin", managedBinaryName)
	installed := filepath.Join(string(filepath.Separator), "opt", "homebrew", "bin", managedBinaryName)
	managed := filepath.Join(string(filepath.Separator), "Users", "me", ".config", "claude-notify-hook", "bin", managedBinaryName)

	got := chooseManagedBinaryPath(current, installed, managed)
	if got != managed {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, managed)
	}
}

func TestChooseManagedBinaryPath_FallsBackToInstalledBinaryForGoRun(t *testing.T) {
	current := filepath.Join(string(filepath.Separator), "Users", "me", "Library", "Caches", "go-build", "ab", "tmp", managedBinaryName)
	installed := filepath.Join(string(filepath.Separator), "Users", "me", "go", "bin", managedBinaryName)

	got := chooseManagedBinaryPath(current, installed, "")
	if got != installed {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, installed)
	}
}

func TestChooseManagedBinaryPath_FallsBackToCurrentExecutableWhenNoInstalledBinary(t *testing.T) {
	current := filepath.Join(string(filepath.Separator), "Users", "me", "Library", "Caches", "go-build", "ab", "tmp", managedBinaryName)

	got := chooseManagedBinaryPath(current, "", "")
	if got != current {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, current)
	}
}

func TestChooseManagedBinaryPath_FallsBackToBinaryName(t *testing.T) {
	got := chooseManagedBinaryPath("", "", "")
	if got != managedBinaryName {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, managedBinaryName)
	}
}

func TestLooksLikeGoToolTemporaryBinary(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{filepath.Join(string(filepath.Separator), "Users", "me", "Library", "Caches", "go-build", "ab", "tmp", managedBinaryName), true},
		{filepath.Join(string(filepath.Separator), "var", "folders", "tmp", "go-build", "ab", "tmp", managedBinaryName), true},
		{filepath.Join(string(filepath.Separator), "Users", "me", "go", "bin", managedBinaryName), false},
		{"", false},
	}

	for _, tt := range tests {
		if got := looksLikeGoToolTemporaryBinary(tt.path); got != tt.want {
			t.Fatalf("looksLikeGoToolTemporaryBinary(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestInstallManagedBinaryFrom_CopiesExecutableToManagedPath(t *testing.T) {
	origConfigDir := config.ConfigDir
	defer func() { config.ConfigDir = origConfigDir }()
	config.ConfigDir = filepath.Join(t.TempDir(), "config")

	srcDir := t.TempDir()
	src := filepath.Join(srcDir, managedBinaryName)
	if runtime.GOOS == "windows" {
		src += ".exe"
	}
	content := []byte("#!/bin/sh\necho ok\n")
	if err := os.WriteFile(src, content, 0755); err != nil {
		t.Fatalf("WriteFile src: %v", err)
	}

	got, err := installManagedBinaryFrom(src)
	if err != nil {
		t.Fatalf("installManagedBinaryFrom error: %v", err)
	}
	want := managedBinaryInstallPath()
	if got != want {
		t.Fatalf("installManagedBinaryFrom() = %q, want %q", got, want)
	}
	raw, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("ReadFile managed binary: %v", err)
	}
	if string(raw) != string(content) {
		t.Fatalf("managed binary content mismatch")
	}
}
