# mrkto

Lightweight CLI for the Marketo REST API. Built for humans and AI agents.

## Installation

```bash
pip install git+https://github.com/itodca/marketo-cli.git
```

## Quick Start

```bash
# Set credentials
export MARKETO_MUNCHKIN_ID=123-ABC-456
export MARKETO_CLIENT_ID=your-client-id
export MARKETO_CLIENT_SECRET=your-client-secret

# Look up a lead
mrkto lead user@example.com

# Check API health
mrkto stats usage
```

## Configuration

Set these environment variables (or use a `.env` file):

| Variable | Required | Description |
|----------|----------|-------------|
| `MARKETO_MUNCHKIN_ID` | Yes | Your Marketo munchkin ID (e.g., `123-ABC-456`) |
| `MARKETO_CLIENT_ID` | Yes | API client ID from LaunchPoint |
| `MARKETO_CLIENT_SECRET` | Yes | API client secret |
| `MARKETO_REST_URL` | No | Override REST endpoint |
| `MARKETO_IDENTITY_URL` | No | Override identity endpoint |

## Commands

```
mrkto lead <email>              Look up a lead by email
mrkto lead get --id <id>        Look up by Marketo ID
mrkto lead search --name "Name" Search by name
mrkto lead describe             Show lead field schema
mrkto lead lists <id>           Lists a lead belongs to
mrkto lead programs <id>        Programs a lead is in
mrkto lead campaigns <id>       Smart campaigns for a lead

mrkto activity types            List activity types
mrkto activity get <lead-id>    Activities for a lead

mrkto campaign list             List campaigns
mrkto campaign get <id>         Get campaign by ID

mrkto list list                 List static lists
mrkto list members <id>         Get list members
mrkto list check <id> <lead-id> Check membership

mrkto company get --name "Name" Look up a company

mrkto stats usage               API usage stats
mrkto stats errors              API error stats

mrkto auth check                Verify credentials
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | JSON output (default) |
| `--compact` | One-line-per-record output |
| `--fields f1,f2` | Select specific fields |
| `--raw` | Raw API response |

## License

MIT
