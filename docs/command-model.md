# Command Model

`marketo-cli` uses a noun/action CLI shape through the `mrkto` command:

```text
mrkto <resource> <action> [flags]
```

Examples:

```bash
mrkto lead get 12345
mrkto lead list --email user@example.com
mrkto static-list add 456 --lead 1001 --lead 1002 --execute
mrkto smart-campaign trigger 789 --lead 1001 --execute
mrkto api get /v1/leads.json --query filterType=email --query filterValues=user@example.com
```

## Why This Shape

The goal is to keep the CLI:

- easy to scan for humans
- predictable for agents
- stable even when Marketo's API grouping is inconsistent

The command names follow Marketo concepts, but not Marketo's API casing. That is why the CLI uses `smart-campaign` instead of `smartCampaign`.

## Resource Names

General nouns:

- `lead`
- `activity`
- `company`
- `program`
- `auth`
- `stats`
- `api`

Explicit Marketo concepts:

- `smart-campaign`
- `static-list`
- `smart-list`

These explicit names avoid ambiguity. A generic `list` command would be confusing because Marketo has both static lists and smart lists.

## Output Contract

The CLI is JSON-first.

- default output is pretty JSON
- `--compact` emits one JSON object per line
- `--raw` emits the full response payload
- `--fields` narrows displayed fields on structured results

Result data goes to `stdout`. Errors go to `stderr`. That keeps the CLI composable with tools like `jq` and shell redirection.

Examples:

```bash
mrkto lead list --email user@example.com > lead.json
mrkto lead list --email user@example.com --raw | jq .
mrkto activity list 12345 --compact > activities.ndjson
```

## Write Safety

Mutating commands are dry-run by default. Add `--execute` to perform the change.

Examples:

```bash
mrkto static-list add 456 --lead 1001 --lead 1002
mrkto static-list add 456 --lead 1001 --lead 1002 --execute
mrkto smart-campaign trigger 789 --lead 1001
mrkto smart-campaign trigger 789 --lead 1001 --execute
```

This keeps the default behavior safe for both humans and agents.

## Raw API Escape Hatch

When a higher-level command does not exist yet, use the raw API namespace:

```bash
mrkto api get /v1/leads.json --query filterType=email --query filterValues=user@example.com
mrkto api post /v1/leads/push.json --input payload.json
mrkto api delete /v1/lists/456/leads.json --query id=1001
```

That keeps the CLI useful even when the resource surface is still growing.
