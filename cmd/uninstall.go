package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/claude-notify-hook/internal/config"
	"github.com/user/claude-notify-hook/internal/hook"
	"github.com/user/claude-notify-hook/internal/ui"
)

func NewUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove notification hooks",
		RunE:  runUninstall,
	}
}

func runUninstall(cmd *cobra.Command, args []string) error {
	ui.Header("🗑️  claude-notify-hook 卸载")

	if !ui.Confirm("确定要卸载所有 hook 吗?", false) {
		ui.Info("已取消")
		return nil
	}

	binaryPath := resolveManagedBinaryPath()

	// Remove Claude Code hooks
	err := hook.RemoveClaudeHook(config.SettingsJSON, config.HookEvents, binaryPath)
	if err != nil {
		ui.Warn("Claude Code hook 移除失败: " + err.Error())
	} else {
		ui.Success("Claude Code hook 已移除")
	}

	// Remove Codex hooks
	err = hook.RemoveCodexHook(config.CodexTOML)
	if err != nil {
		ui.Warn("Codex hook 移除失败: " + err.Error())
	} else {
		ui.Success("Codex hook 已移除")
	}

	// Ask about config removal
	if ui.Confirm("是否同时删除配置文件?", false) {
		if err := os.RemoveAll(config.ConfigDir); err != nil {
			ui.Warn("配置目录删除失败: " + err.Error())
		} else {
			ui.Success("配置目录已删除")
		}
	}

	fmt.Println()
	ui.Success("卸载完成")
	return nil
}
