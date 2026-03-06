package sound

import (
	"os/exec"
	"runtime"
	"slices"
	"testing"
)

// withGOOS temporarily overrides goos for testing, restores after.
func withGOOS(t *testing.T, os string) {
	t.Helper()
	orig := goos
	goos = os
	t.Cleanup(func() { goos = orig })
}

// withNilBuildCmd overrides buildCmdFn to return nil.
func withNilBuildCmd(t *testing.T) {
	t.Helper()
	orig := buildCmdFn
	buildCmdFn = func(string) *exec.Cmd { return nil }
	t.Cleanup(func() { buildCmdFn = orig })
}

func TestDefaultSoundFile_Darwin(t *testing.T) {
	withGOOS(t, "darwin")
	got := DefaultSoundFile()
	want := "/System/Library/Sounds/Glass.aiff"
	if got != want {
		t.Errorf("DefaultSoundFile() = %q, want %q", got, want)
	}
}

func TestDefaultSoundFile_Linux(t *testing.T) {
	withGOOS(t, "linux")
	got := DefaultSoundFile()
	want := "/usr/share/sounds/freedesktop/stereo/complete.oga"
	if got != want {
		t.Errorf("DefaultSoundFile() = %q, want %q", got, want)
	}
}

func TestDefaultSoundFile_Unsupported(t *testing.T) {
	withGOOS(t, "freebsd")
	got := DefaultSoundFile()
	if got != "" {
		t.Errorf("DefaultSoundFile() = %q, want empty", got)
	}
}

func TestBuildCmd_Darwin_UsesAfplay(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only test")
	}
	cmd := buildCmd("/test/file.aiff")
	if cmd == nil {
		t.Fatal("buildCmd() returned nil")
	}
	if !slices.Contains(cmd.Args, "/test/file.aiff") {
		t.Errorf("cmd.Args = %v, want to contain /test/file.aiff", cmd.Args)
	}
}

func TestPlay_EmptyFile_SupportedPlatform(t *testing.T) {
	withGOOS(t, "darwin")
	// On darwin, empty file → DefaultSoundFile → non-empty → buildCmd → Start
	Play("")
}

func TestPlay_EmptyFile_UnsupportedPlatform(t *testing.T) {
	withGOOS(t, "freebsd")
	// On unsupported, empty file → DefaultSoundFile="" → second check → return
	Play("")
}

func TestPlay_CustomFile(t *testing.T) {
	// Play with a non-existent file should not panic
	Play("/nonexistent/sound.wav")
}

func TestPlay_BuildCmdReturnsNil(t *testing.T) {
	withGOOS(t, "darwin")
	withNilBuildCmd(t)
	// buildCmdFn returns nil → cmd==nil branch → return
	Play("/some/file.aiff")
}

func TestPlay_ExplicitFile_FullPath(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only: needs afplay")
	}
	// Covers the full happy path: non-empty file → buildCmd → cmd.Start()
	Play("/System/Library/Sounds/Glass.aiff")
}

func TestBuildCmd_ReturnsNonNil_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only test")
	}
	cmd := buildCmd("/test/sound.aiff")
	if cmd == nil {
		t.Fatal("buildCmd() returned nil on darwin")
	}
}
