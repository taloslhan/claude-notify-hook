# Claude Code / Codex Telegram Notify Hook

Get Telegram notifications when Claude Code needs your attention or when Claude Code / Codex completes a task.

## Features

- **Claude Needs Input** - Notified when Claude Code is waiting for your response
- **Claude Task Complete** - Notified when a Claude Code task finishes
- **Claude Subagent Complete** - Notified when a Claude Code subagent finishes
- **Codex Task Complete** - Notified when Codex finishes a turn and triggers `notify`
- **Silent Failure** - Never blocks Claude Code, even if notification fails
- **One-Click Install** - Interactive setup with automatic Claude hook registration and Codex notify registration (when available)

## Quick Start

### 1. Create a Telegram Bot

1. Open Telegram and search for **@BotFather**
2. Send `/newbot`
3. Choose a name and username for your bot
4. Copy the **API Token** (looks like `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)
5. **Start a conversation** with your new bot (send `/start`) — this is required for the bot to send you messages

### 2. Get Your Chat ID

**Option A** (Easiest): Send any message to **@userinfobot** on Telegram — it will reply with your Chat ID.

**Option B** (API): After sending a message to your bot, visit:
```
https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates
```
Look for `"chat":{"id": YOUR_CHAT_ID}` in the response.

### 3. Install

```bash
# Build from source
git clone https://github.com/user/claude-notify-hook.git
cd claude-notify-hook
make install

# Run interactive installer
claude-notify-hook install
```

Optional install modes:

```bash
claude-notify-hook install --claude-only  # Only configure Claude Code hooks
claude-notify-hook install --codex-only   # Only configure Codex notify
```

The installer will:
- Ask for your Bot Token and Chat ID
- Offer to reuse saved credentials on reinstall
- Save credentials to `~/.config/claude-notify-hook/.env`
- Register hooks in `~/.claude/settings.json`
- Register `notify` in `~/.codex/config.toml` when Codex is installed
- Send a test notification to verify everything works

## How It Works

```
Claude Code Event / Codex notify → claude-notify-hook notify → Telegram API → Your Phone
```

The installer registers hooks for three Claude Code events:

| Event | When It Fires | What You See |
|-------|---------------|--------------|
| **Notification** | Claude needs your input (permission, question) | "Needs Input" + message |
| **Stop** | A task/session completes | "Task Complete" + summary |
| **SubagentStop** | A subagent finishes | "Subagent Complete" + summary |

For Codex, the installer configures the official `notify` command in `~/.codex/config.toml` so you receive a Telegram message when a Codex turn completes.

> Note: Codex's official `notify` support currently covers turn completion. It does not mirror Claude Code's `Notification` / `SubagentStop` lifecycle events.

### Message Format

```
Claude Code | project-name
Status: Needs Input

Claude is asking for permission to run "npm install"...
```

## Commands

| Command | Description |
|---------|-------------|
| `claude-notify-hook install` | Install Claude Code + Codex integrations |
| `claude-notify-hook install --claude-only` | Only install Claude Code hooks |
| `claude-notify-hook install --codex-only` | Only install Codex notify |
| `claude-notify-hook uninstall` | Remove all hooks and config |
| `claude-notify-hook test` | Send a test notification |
| `claude-notify-hook status` | Check installation status |

## File Structure

```
~/.config/claude-notify-hook/
└── .env                  # Your Telegram credentials (Bot Token + Chat ID)

~/.claude/settings.json   # Hook registrations (managed by install/uninstall)
~/.codex/config.toml      # Codex notify registration (managed when possible)
```

## Troubleshooting

### "Test notification failed: Unauthorized"
Your Bot Token is invalid. Create a new bot with @BotFather and re-run `claude-notify-hook install`.

### "Test notification failed: Bad Request: chat not found"
Your Chat ID is wrong. Get it from @userinfobot and re-run `claude-notify-hook install`.

### "Test notification failed: Forbidden: bot was blocked by the user"
You need to start a conversation with your bot. Open Telegram, find your bot, and send `/start`.

### Notifications not arriving
1. Run `claude-notify-hook status` to check if everything is properly installed
2. Run `claude-notify-hook test` to verify the Telegram connection
3. Check that `~/.claude/settings.json` contains the hook registrations
4. If you use Codex, check that `~/.codex/config.toml` contains the notify configuration

### Reinstalling
Simply run `claude-notify-hook install` again — it will detect the existing installation and offer to reinstall.

If you only want one side, re-run with `--claude-only` or `--codex-only`.

## Uninstall

```bash
claude-notify-hook uninstall
```

This removes hook registrations from `settings.json` and optionally deletes the config directory.

## Build

```bash
make build          # Build for current platform
make install        # Build and copy to ~/go/bin/
make all            # Cross-compile for all platforms
```

## Requirements

- **Go 1.23+** (build only)

## License

MIT
