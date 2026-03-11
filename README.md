# mrkto

Lightweight CLI for the Marketo REST API. Built for humans and AI agents.

## Installation

```bash
pip install git+https://github.com/itodca/marketo-cli.git
```

Or with the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/itodca/marketo-cli/main/install.sh | bash
```

## Quick Start

```bash
# First-time setup (auth + Claude Code skill)
mrkto setup

# Or configure credentials manually
mrkto auth setup

# Look up a lead
mrkto lead list --email user@example.com

# Check API health
mrkto stats usage
```

## Configuration

Credentials are stored in `~/.config/mrkto/config`. Run `mrkto setup` to configure interactively.

Env vars override the config file (useful for CI/scripting):

| Variable | Required | Description |
|----------|----------|-------------|
| `MARKETO_MUNCHKIN_ID` | Yes | Your Marketo munchkin ID (e.g., `123-ABC-456`) |
| `MARKETO_CLIENT_ID` | Yes | API client ID from LaunchPoint |
| `MARKETO_CLIENT_SECRET` | Yes | API client secret |
| `MARKETO_REST_URL` | No | Override REST endpoint |
| `MARKETO_IDENTITY_URL` | No | Override identity endpoint |

Docs: https://experienceleague.adobe.com/en/docs/marketo-developer/marketo/rest/rest-api

## Commands

Run `mrkto help` for the full reference. Summary:

```
mrkto setup                               First-time setup (auth + skill)
mrkto auth setup                          Configure credentials
mrkto auth check                          Verify credentials

mrkto lead list --email <addr>            List leads by email
mrkto lead list --id <ids>                List leads by Marketo ID(s)
mrkto lead list --filter <key=value>      List leads by any filter field
mrkto lead get <id>                       Get lead by Marketo ID
mrkto lead describe                       Show lead field schema
mrkto lead lists <id>                     Static lists a lead belongs to
mrkto lead programs <id>                  Programs a lead is in
mrkto lead campaigns <id>                 Smart campaigns for a lead

mrkto activity types                      List activity types
mrkto activity get <lead_id>              Activities for a lead
mrkto activity changes --watch <fields>   Lead field changes

mrkto campaign list                       List campaigns
mrkto campaign get <id>                   Get campaign by ID
mrkto campaign schedule <id> --execute    Schedule a batch campaign
mrkto campaign trigger <id> --execute     Trigger campaign for leads

mrkto list list                           List static lists
mrkto list get <id>                       Get list by ID
mrkto list members <id>                   Get list members
mrkto list add <id> --leads <ids> --execute    Add leads to list
mrkto list remove <id> --leads <ids> --execute Remove leads from list
mrkto list check <id> --leads <ids>       Check membership

mrkto company list --name <name>          List companies by name
mrkto company list --filter <key=value>   List companies by any filter
mrkto company describe                    Company field schema

mrkto stats usage                         API usage stats
mrkto stats errors                        API error stats

mrkto skill install [--scope user|project] Install Claude Code skill
mrkto help                                Full command reference
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--fields f1,f2` | Limit response fields |
| `--limit N` | Max results (on list commands) |
| `--json` | JSON output (default) |
| `--compact` | One-line-per-record output |
| `--raw` | Raw API response |

## Write Safety

Commands that modify data (`schedule`, `trigger`, `add`, `remove`) require `--execute`. Without it, they run in dry-run mode.

## Claude Code Integration

```bash
mrkto skill install              # Install for all projects
mrkto skill install --scope project  # Install for current project only
```

## License

MIT
