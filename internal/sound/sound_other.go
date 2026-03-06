//go:build !darwin && !linux && !windows

package sound

import "os/exec"

func buildCmd(_ string) *exec.Cmd {
	return nil
}
