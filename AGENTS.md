# Agent guide for the Spork CLI

You are likely an AI agent (Claude Code, Cursor, Cline, Windsurf, etc.) helping a user manage uptime monitoring with the `spork` CLI. This file tells you how to drive it well.

## What Spork does

Uptime monitoring. Users add HTTP / DNS / SSL / keyword / TCP / ping monitors against URLs or hosts they care about, attach alert channels (email, Slack, webhooks), and optionally publish a status page.

## Install

```sh
# macOS
brew install sporkops/tap/spork
# Linux / macOS
curl -sSL https://sporkops.com | sh
```

Verify with `spork --version`.

## Authentication

In order of preference for agents:

1. **`SPORK_API_KEY` environment variable** — preferred. Bypasses the OS keyring, works in headless / container / CI contexts. Get a key from <https://sporkops.com/settings/api-keys> or by running `spork apikey create --json` once a human is logged in.
2. `spork login` — opens a browser for OAuth. Only useful when a human is at the keyboard. Do **not** invoke this autonomously.

The API base URL can be overridden with `SPORK_API_URL` (rarely needed).

## Machine-readable output

Pass `--json` (or the equivalent `--output json`) on any command. JSON is the contract agents should rely on; the default table output is human-only and can change without notice.

```sh
spork monitor list --json
spork monitor add https://example.com --json
```

## Errors and exit codes

- Non-zero exit always means failure.
- Today, error messages are human-readable text on stderr. Inspect the exit code, not the stderr string, to branch on outcomes.
- Common conditions: unauthorized (fix `SPORK_API_KEY`), forbidden (key lacks permission), payment required (feature needs a paid plan), not found (resource doesn't exist).

A future release will add structured JSON error envelopes on stderr and a `--agent` global flag that turns on JSON-everywhere mode. Until then, use `--json` per command.

## Command tree

Top-level: `login` · `logout` · `monitor` · `apikey` · `alertchannel` · `incident` · `webhook` · `statuspage` · `members` · `org` · `maintenance`.

Discover flags with `spork <command> --help`. The flag list below is a starting point — confirm with `--help` if a flag is missing.

## Multi-org

Spork is multi-tenant: an account can own or belong to several organizations. Every org-scoped command (`monitor list`, `alert-channel ...`, etc.) needs to know which tenant to target. The CLI resolves it in this order, highest precedence first:

1. `--org ORG_ID` on the command line
2. `SPORK_ORG_ID` env var
3. Active org saved by `spork org use ORG_ID` (in `~/.config/spork/config.json`)
4. SDK auto-resolve — works when an API key is in use (keys are bound to one org); errors with a candidate list for Firebase users in multiple orgs

For agents, the simplest setup is to export `SPORK_ORG_ID` alongside `SPORK_API_KEY` in your environment and let it flow through every command. For interactive users, `spork org use` is the one-shot equivalent.

| Command | Purpose |
|---|---|
| `spork org list` | Show every org you belong to, with the active one marked |
| `spork org current` | Print the active org ID (and where the value came from) |
| `spork org use ORG_ID` | Persist an org as the active default |
| `spork org use --clear` | Drop the saved preference; fall back to env / auto-resolve |

## Common workflows

### Create an HTTP monitor

```sh
spork monitor add https://example.com \
  --name "Marketing site" \
  --interval 60 \
  --json
```

Always pass `--name` so the monitor is identifiable in the dashboard. `--interval` is in seconds.

### Find a monitor by URL (don't guess IDs)

```sh
spork monitor list --json | jq '.[] | select(.target == "https://example.com")'
```

Prefer list-and-filter over assuming an ID. Many `monitor` subcommands accept either an ID or a URL as their argument.

### Pause, resume, delete

```sh
spork monitor pause <id-or-url> --json
spork monitor rm   <id-or-url> --force --json
```

`--force` skips the interactive confirmation. Agents should always pass it after the user has confirmed deletion.

### Create an alert channel and attach it

```sh
spork alertchannel create --json   # discover required flags with --help first
spork monitor update <id> --json   # see --help for the alert-channel attachment flag
```

## Idempotency

The underlying API supports idempotency keys. The CLI does not yet expose a `--idempotency-key` flag for every write; until it does, **list-then-create** to avoid duplicates rather than blindly retrying a failed create.

## What NOT to do

- **Don't run `spork login` without asking** — it opens a browser.
- **Don't parse table output.** Always pass `--json`.
- **Don't loop the CLI for bulk work without throttling** — the API is rate-limited and you will be 429'd.
- **Don't store the API key in committed config files.** Read from env.
- **Don't delete resources without explicit user confirmation** naming the specific monitor / channel.

## When to reach for something else

- **Infrastructure-as-code (committed, PR-reviewed, CI-applied)** → use the [Spork Terraform provider](https://github.com/sporkops/terraform-provider-sporkops).
- **Custom Go service** → use the [spork-go SDK](https://github.com/sporkops/spork-go) directly.
- **One-off interactive change** → this CLI is the right tool.

## Reporting issues

File bugs at <https://github.com/sporkops/cli/issues>. Include the command, the `--json` output, the exit code, and `spork --version`.
