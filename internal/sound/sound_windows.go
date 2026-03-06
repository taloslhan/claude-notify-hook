//go:build windows

package sound

import "os/exec"

func buildCmd(file string) *exec.Cmd {
	script := `(New-Object Media.SoundPlayer '` + file + `').PlaySync()`
	return exec.Command("powershell", "-c", script)
}
