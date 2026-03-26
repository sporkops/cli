# spork-cli

CLI for [Ping by Spork](https://sporkops.com) — uptime monitoring from your terminal.

## Install

**macOS (Homebrew):**

```sh
brew install sporkops/tap/spork
```

**Linux / macOS (curl):**

```sh
curl -sSL https://sporkops.com | sh
```

## Quick Start

```sh
# Log in (opens browser)
spork login

# Add a monitor
spork ping add https://example.com

# List monitors
spork ping list

# Check status
spork ping status
```

## Commands

| Command | Description |
|---|---|
| `spork login` | Log in via browser |
| `spork ping add <url>` | Add an uptime monitor |
| `spork ping list` | List all monitors |
| `spork ping status` | Show current status |
| `spork ping rm <id\|url>` | Remove a monitor |
| `spork ping history <id\|url>` | Show check history |

### Flags

- `--json` — output as JSON (on any command)
- `--name` — set monitor name (on `ping add`)
- `--method` — HTTP method, default GET (on `ping add`)
- `--interval` — check interval in seconds, 60 or 30 (on `ping add`)
- `--limit` — number of history records (on `ping history`)
- `--force` — skip confirmation (on `ping rm`)

## Configuration

Credentials are stored in `~/.config/spork/credentials.json` after login.

The API base URL can be overridden with the `SPORK_API_URL` environment variable.

## Docs

Full documentation: [https://sporkops.com/docs](https://sporkops.com/docs)

## License

MIT
