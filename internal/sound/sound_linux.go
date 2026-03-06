//go:build linux

package sound

import "os/exec"

func buildCmd(file string) *exec.Cmd {
	// Try paplay (PulseAudio) first, fall back to aplay (ALSA)
	if path, err := exec.LookPath("paplay"); err == nil {
		return exec.Command(path, file)
	}
	if path, err := exec.LookPath("aplay"); err == nil {
		return exec.Command(path, file)
	}
	return nil
}
