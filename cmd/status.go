package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/user/claude-notify-hook/internal/config"
	"github.com/user/claude-notify-hook/internal/hook"
	"github.com/user/claude-notify-hook/internal/ui"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check installation health",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Header("📊 claude-notify-hook 状态检查")

			score := 0
			total := 0

			// 1. Binary in PATH
			total++
			binaryPath := resolveManagedBinaryPath()
			if binaryPath != "" {
				if _, err := os.Stat(binaryPath); err == nil {
					ui.StatusLine("二进制文件:", "已就绪", ui.Green)
					ui.StatusLine("  路径:", binaryPath, ui.Dim)
					score++
				} else if path, err := exec.LookPath(managedBinaryName); err == nil {
					ui.StatusLine("二进制文件:", "已在 PATH 中", ui.Green)
					ui.StatusLine("  路径:", path, ui.Dim)
					score++
				} else {
					ui.StatusLine("二进制文件:", "未找到可执行文件", ui.Red)
				}
			} else if path, err := exec.LookPath(managedBinaryName); err == nil {
				ui.StatusLine("二进制文件:", "已在 PATH 中", ui.Green)
				ui.StatusLine("  路径:", path, ui.Dim)
				score++
			} else {
				ui.StatusLine("二进制文件:", "未找到可执行文件", ui.Red)
			}

			// 2. Config file
			total++
			cfg, err := config.Load()
			if err == nil && cfg.BotToken != "" && cfg.ChatID != "" {
				ui.StatusLine("配置文件:", "已配置", ui.Green)
				ui.StatusLine("  Bot Token:", ui.MaskToken(cfg.BotToken), ui.Dim)
				ui.StatusLine("  Chat ID:", cfg.ChatID, ui.Dim)
				score++
			} else {
				ui.StatusLine("配置文件:", "未配置或不完整", ui.Red)
			}

			// 3. Claude Code hook
			total++
			if hook.HasClaudeHook(config.SettingsJSON, binaryPath) {
				ui.StatusLine("Claude Code Hook:", "已注册", ui.Green)
				score++
			} else {
				ui.StatusLine("Claude Code Hook:", "未注册", ui.Yellow)
			}

			// 4. Codex hook
			total++
			if hook.HasCodexHook(config.CodexTOML) {
				ui.StatusLine("Codex Hook:", "已注册", ui.Green)
				score++
			} else {
				ui.StatusLine("Codex Hook:", "未注册", ui.Yellow)
			}

			// Summary
			fmt.Println()
			pct := score * 100 / total
			color := ui.Green
			if pct < 75 {
				color = ui.Yellow
			}
			if pct < 50 {
				color = ui.Red
			}
			ui.StatusLine("健康评分:", fmt.Sprintf("%d/%d (%d%%)", score, total, pct), color)

			return nil
		},
	}
}
