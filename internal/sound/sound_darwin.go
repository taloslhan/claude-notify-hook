//go:build darwin

package sound

import "os/exec"

func buildCmd(file string) *exec.Cmd {
	return exec.Command("afplay", file)
}
