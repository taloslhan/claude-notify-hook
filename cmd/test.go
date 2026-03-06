package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/claude-notify-hook/internal/config"
	"github.com/user/claude-notify-hook/internal/telegram"
	"github.com/user/claude-notify-hook/internal/ui"
)

func NewTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Send a test notification to verify Telegram setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Header("🧪 发送测试通知")

			cfg, err := config.Load()
			if err != nil {
				ui.Error("无法加载配置: " + err.Error())
				return err
			}
			if cfg.BotToken == "" || cfg.ChatID == "" {
				ui.Error("配置不完整，请先运行 install")
				return fmt.Errorf("missing config")
			}

			ui.Info("Token: " + ui.MaskToken(cfg.BotToken))
			ui.Info("Chat ID: " + cfg.ChatID)

			msg := "<b>🧪 测试通知</b>\n这是一条来自 claude-notify-hook 的测试消息。\n如果您看到此消息，说明配置正确！"

			client := &telegram.Client{Token: cfg.BotToken, ChatID: cfg.ChatID}
			result, err := client.SendMessage(msg)
			if err != nil {
				ui.Error("发送失败: " + err.Error())
				return err
			}

			if result.StatusCode == 200 {
				ui.Success("测试通知发送成功！请检查 Telegram。")
			} else {
				ui.Error(fmt.Sprintf("Telegram API 返回 HTTP %d", result.StatusCode))
				ui.Error("响应: " + result.Body)
				return fmt.Errorf("HTTP %d", result.StatusCode)
			}
			return nil
		},
	}
}
