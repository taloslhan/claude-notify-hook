package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/claude-notify-hook/internal/config"
	"github.com/user/claude-notify-hook/internal/hook"
	"github.com/user/claude-notify-hook/internal/ui"
)

func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install notification hooks for Claude Code and/or Codex",
		RunE:  runInstall,
	}
	cmd.Flags().Bool("claude-only", false, "Only install Claude Code hook")
	cmd.Flags().Bool("codex-only", false, "Only install Codex hook")
	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	ui.Header("🔧 claude-notify-hook 安装向导")

	claudeOnly, _ := cmd.Flags().GetBool("claude-only")
	codexOnly, _ := cmd.Flags().GetBool("codex-only")

	// Load existing config or create new
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Collect credentials if missing
	if cfg.BotToken == "" {
		cfg.BotToken = ui.Prompt("Telegram Bot Token", "")
		if cfg.BotToken == "" {
			ui.Error("Bot Token 不能为空")
			return fmt.Errorf("empty bot token")
		}
	} else {
		ui.Info("已有 Bot Token: " + ui.MaskToken(cfg.BotToken))
		if ui.Confirm("是否更新 Bot Token?", false) {
			cfg.BotToken = ui.Prompt("新的 Bot Token", "")
		}
	}

	if cfg.ChatID == "" {
		cfg.ChatID = ui.Prompt("Telegram Chat ID", "")
		if cfg.ChatID == "" {
			ui.Error("Chat ID 不能为空")
			return fmt.Errorf("empty chat id")
		}
	} else {
		ui.Info("已有 Chat ID: " + cfg.ChatID)
		if ui.Confirm("是否更新 Chat ID?", false) {
			cfg.ChatID = ui.Prompt("新的 Chat ID", "")
		}
	}

	// Determine install targets
	installClaude := !codexOnly
	installCodex := !claudeOnly

	if !claudeOnly && !codexOnly {
		installClaude = ui.Confirm("安装 Claude Code hook?", true)
		installCodex = ui.Confirm("安装 Codex hook?", true)
	}

	targets := ""
	if installClaude && installCodex {
		targets = "claude,codex"
	} else if installClaude {
		targets = "claude"
	} else if installCodex {
		targets = "codex"
	}
	cfg.InstallTargets = targets

	// Save config
	if err := cfg.Save(); err != nil {
		ui.Error("保存配置失败: " + err.Error())
		return err
	}
	ui.Success("配置已保存到 " + config.EnvFile)

	binaryPath, err := installManagedBinary()
	if err != nil {
		ui.Warn("托管二进制安装失败，将继续使用当前可执行文件: " + err.Error())
		binaryPath = resolveManagedBinaryPath()
	} else {
		ui.Success("二进制已安装到 " + binaryPath)
	}

	// Install Claude Code hook
	if installClaude {
		err := hook.AddClaudeHook(config.SettingsJSON, config.HookEvents, binaryPath)
		if err != nil {
			ui.Error("Claude Code hook 注册失败: " + err.Error())
		} else {
			ui.Success("Claude Code hook 已注册")
		}
	}

	// Install Codex hook
	if installCodex {
		if err := os.MkdirAll(config.CodexDir, 0755); err == nil {
			err := hook.AddCodexHook(config.CodexTOML, binaryPath)
			if err != nil {
				ui.Error("Codex hook 注册失败: " + err.Error())
			} else {
				ui.Success("Codex hook 已注册")
			}
		}
	}

	fmt.Println()
	ui.Success("安装完成！运行 'claude-notify-hook test' 验证配置。")
	return nil
}
