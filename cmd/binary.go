package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const managedBinaryName = "claude-notify-hook"

func resolveManagedBinaryPath() string {
	currentExe, _ := os.Executable()
	currentExe = normalizeBinaryPath(currentExe)

	if currentExe != "" && !looksLikeGoToolTemporaryBinary(currentExe) {
		return currentExe
	}

	installedPath := ""
	if path, err := exec.LookPath(managedBinaryName); err == nil {
		installedPath = normalizeBinaryPath(path)
	}

	return chooseManagedBinaryPath(currentExe, installedPath)
}

func chooseManagedBinaryPath(currentExe, installedPath string) string {
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
