package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/user/claude-notify-hook/internal/config"
)

const managedBinaryName = "claude-notify-hook"

func resolveManagedBinaryPath() string {
	managedPath := existingManagedBinaryPath()
	currentExe, _ := os.Executable()
	currentExe = normalizeBinaryPath(currentExe)

	installedPath := ""
	if path, err := exec.LookPath(managedBinaryName); err == nil {
		installedPath = normalizeBinaryPath(path)
	}

	return chooseManagedBinaryPath(currentExe, installedPath, managedPath)
}

func chooseManagedBinaryPath(currentExe, installedPath, managedPath string) string {
	if managedPath != "" {
		return managedPath
	}
	if currentExe != "" && !looksLikeGoToolTemporaryBinary(currentExe) {
		return currentExe
	}
	if installedPath != "" {
		return installedPath
	}
	if currentExe != "" {
		return currentExe
	}
	return managedBinaryName
}

func installManagedBinary() (string, error) {
	currentExe, _ := os.Executable()
	currentExe = normalizeBinaryPath(currentExe)
	return installManagedBinaryFrom(currentExe)
}

func installManagedBinaryFrom(currentExe string) (string, error) {
	target := normalizeBinaryPath(managedBinaryInstallPath())
	if currentExe == "" {
		return target, fmt.Errorf("could not determine current executable path")
	}
	if currentExe == target {
		return target, nil
	}
	if err := copyExecutable(currentExe, target); err != nil {
		return target, err
	}
	return target, nil
}

func managedBinaryInstallPath() string {
	name := managedBinaryName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(config.ConfigDir, "bin", name)
}

func existingManagedBinaryPath() string {
	path := normalizeBinaryPath(managedBinaryInstallPath())
	if fileExists(path) {
		return path
	}
	return ""
}

func normalizeBinaryPath(path string) string {
	if path == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return path
}

func looksLikeGoToolTemporaryBinary(path string) bool {
	if path == "" {
		return false
	}
	clean := filepath.Clean(path)
	sep := string(filepath.Separator)
	return strings.Contains(clean, sep+"go-build"+sep)
}

func copyExecutable(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	mode := info.Mode().Perm()
	if mode == 0 {
		mode = 0755
	}
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Chmod(dst, mode)
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
