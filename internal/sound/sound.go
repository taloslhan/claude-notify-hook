// Package sound provides cross-platform audio notification playback.
// It uses native OS commands (afplay/paplay/powershell) with zero external dependencies.
package sound

import "runtime"

// goos is the OS identifier used by DefaultSoundFile. Tests may override it.
var goos = runtime.GOOS

// buildCmdFn wraps buildCmd for testability.
var buildCmdFn = buildCmd

// DefaultSoundFile returns the platform default system sound path.
func DefaultSoundFile() string {
	switch goos {
	case "darwin":
		return "/System/Library/Sounds/Glass.aiff"
	case "linux":
		return "/usr/share/sounds/freedesktop/stereo/complete.oga"
	default:
		return ""
	}
}

// Play plays the given sound file asynchronously (fire-and-forget).
// If file is empty, it falls back to the platform default.
// Errors are silently ignored to never block the caller.
func Play(file string) {
	if file == "" {
		file = DefaultSoundFile()
	}
	if file == "" {
		return
	}

	cmd := buildCmdFn(file)
	if cmd == nil {
		return
	}
	// Fire and forget — don't block notification delivery
	_ = cmd.Start()
}
