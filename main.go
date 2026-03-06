package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/user/claude-notify-hook/cmd"
)

func main() {
	root := &cobra.Command{
		Use:   "claude-notify-hook",
		Short: "Telegram notification hook for Claude Code & Codex",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		cmd.NewNotifyCmd(),
		cmd.NewInstallCmd(),
		cmd.NewUninstallCmd(),
		cmd.NewStatusCmd(),
		cmd.NewTestCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
