# Spork CLI

[![Build](https://github.com/sporkops/cli/actions/workflows/test.yml/badge.svg)](https://github.com/sporkops/cli/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/sporkops/cli)](https://github.com/sporkops/cli/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Know when your site goes down before your customers do.**

Spork is uptime monitoring that lives in your terminal. Add a monitor in seconds, get alerted when things break, and check status without leaving your workflow.

## Quickstart

```sh
brew install sporkops/tap/spork && spork login && spork monitor add https://yoursite.com
```

That's it. You're monitoring.

## Install

**macOS (Homebrew):**

```sh
brew install sporkops/tap/spork
```

**Linux / macOS (curl):**

```sh
curl -sSL https://sporkops.com | sh
```

## Commands

| Command | Description |
|---|---|
| `spork login` | Log in via browser |
| `spork monitor add <url>` | Add an uptime monitor |
| `spork monitor list` | List all monitors |
| `spork monitor status` | Show current status |
| `spork monitor history <id\|url>` | Show check history |
| `spork monitor rm <id\|url>` | Remove a monitor |
| `spork monitor pause <id\|url>` | Pause a monitor |
| `spork monitor update <id\|url>` | Update monitor settings |
| `spork apikey` | Manage API keys |
| `spork alertchannel` | Manage alert channels |

`spork ping` remains as a deprecated alias for `spork monitor` and will be removed in a future release.

### Flags

- `--json` — output as JSON (on any command)
- `--name` — set monitor name (on `monitor add`)
- `--method` — HTTP method, default GET (on `monitor add`)
- `--interval` — check interval in seconds, 60 or 30 (on `monitor add`)
- `--limit` — number of history records (on `monitor history`)
- `--force` — skip confirmation (on `monitor rm`)

## Infrastructure as Code

Prefer Terraform? Use the [Spork Terraform Provider](https://registry.terraform.io/providers/sporkops/sporkops/latest) to manage monitors as code.

## Configuration

Credentials are stored in the operating system's secure credential store (macOS Keychain, Linux Secret Service / GNOME Keyring, Windows Credential Manager) after `spork login`.

For headless / CI environments, set the `SPORK_API_KEY` environment variable instead — it is honored by every command and bypasses the keyring entirely.

The API base URL can be overridden with the `SPORK_API_URL` environment variable.

## Documentation

Full docs: [sporkops.com/docs](https://sporkops.com/docs)

---

**Free to start. No credit card required.** [Sign up at sporkops.com →](https://sporkops.com)

## License

MIT
