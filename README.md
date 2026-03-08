# Telbot

Go-based tool for managing Telkomsel accounts via **Telegram Bot**, **Terminal CLI**, or **MCP Server** (for AI agents).

## Features

- 🔑 SMS OTP login with session caching
- 📊 Profile, balance, quota checking
- 📦 Browse recommended packages
- 🛍️ Purchase packages (Pulsa / QRIS)
- ⏰ Auto-buy monitor (auto-purchase when quota depleted)
- 🤖 MCP server for AI agent integration

## Quick Start

```bash
git clone https://github.com/0xtbug/telbot
cd telbot
cp .env.example .env   # fill in your tokens
go mod tidy
```

## Modes

| Flag | Description |
|------|-------------|
| `--bot` | Telegram bot (requires `.env` config) |
| `--cli` | Interactive terminal UI (Bubbletea) |
| `--mcp` | MCP server for AI agents (stdio) |

```bash
go run . --bot       # Telegram bot
go run . --cli       # Terminal UI
go run . --mcp       # MCP server
```

## Install Globally

Install as `telbot` command available from anywhere:

```bash
go build -o "$GOPATH/bin/telbot" -ldflags "-s -w" .
```

> **Windows (PowerShell):**
>
> To install the executable in your system-wide Go bin directory, run PowerShell **as Administrator** and execute:
> ```powershell
> go build -o "C:\Program Files\Go\bin\telbot.exe" -ldflags "-s -w" .
> ```
> *(Note: You must have Administrator privileges to write to `C:\Program Files`)*

Then use from any directory:

```bash
telbot --bot
telbot --cli
telbot --mcp
```

### Pre-built Binaries (GitHub Releases)

If you don't want to install Go or build it yourself, you can download the pre-compiled executables for Windows, Linux, and macOS directly from the **[Releases](https://github.com/0xtbug/telbot/releases)** page of this repository.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TELKOMSEL_BOT_TOKEN` | Bot mode | Telegram bot token from BotFather |
| `TELEGRAM_ADMIN_ID` | Bot mode | Your Telegram user ID |

You can either export these directly in your terminal, or place them in a `.env` file located in your platform's standard configuration directory:

- **Windows:** `%APPDATA%\telbot\.env` (e.g., `C:\Users\<User>\AppData\Roaming\telbot\.env`)
- **Linux/macOS:** `~/.config/telbot/.env`

## Documentation

See the [docs/](docs/) folder for detailed guides:

- [Telegram Bot](docs/telegram-bot.md)
- [CLI Mode](docs/cli.md)
- [MCP Server](docs/mcp-server.md)
- [Running as Systemd Service (Linux)](docs/systemd.md)

## Disclaimer ⚠️

This application uses unofficial, reverse-engineered MyTelkomsel API endpoints. Subject to change without warning. Use at your own risk.
