package cmd

import (
	"path/filepath"
	"testing"
)

func TestChooseManagedBinaryPath_PrefersStableCurrentExecutable(t *testing.T) {
	current := filepath.Join(string(filepath.Separator), "usr", "local", "bin", managedBinaryName)
	installed := filepath.Join(string(filepath.Separator), "opt", "homebrew", "bin", managedBinaryName)

	got := chooseManagedBinaryPath(current, installed)
	if got != current {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, current)
	}
}

func TestChooseManagedBinaryPath_FallsBackToInstalledBinaryForGoRun(t *testing.T) {
	current := filepath.Join(string(filepath.Separator), "Users", "me", "Library", "Caches", "go-build", "ab", "tmp", managedBinaryName)
	installed := filepath.Join(string(filepath.Separator), "Users", "me", "go", "bin", managedBinaryName)

	got := chooseManagedBinaryPath(current, installed)
	if got != installed {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, installed)
	}
}

func TestChooseManagedBinaryPath_FallsBackToCurrentExecutableWhenNoInstalledBinary(t *testing.T) {
	current := filepath.Join(string(filepath.Separator), "Users", "me", "Library", "Caches", "go-build", "ab", "tmp", managedBinaryName)

	got := chooseManagedBinaryPath(current, "")
	if got != current {
		t.Fatalf("chooseManagedBinaryPath() = %q, want %q", got, current)
	}
}

func TestChooseManagedBinaryPath_FallsBackToBinaryName(t *testing.T) {
	got := chooseManagedBinaryPath("", "")
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
