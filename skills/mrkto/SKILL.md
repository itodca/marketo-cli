---
name: mrkto
description: Use when the user needs to interact with Marketo — looking up leads, checking unsubscribe status, campaign membership, list membership, activity history, or any Marketo REST API operation. Triggers on mentions of Marketo, leads, unsubscribe, email marketing, campaign membership, or marketing automation lookups.
---

# mrkto — Marketo REST API CLI

## Quick Reference

Run `mrkto help` for the full command reference with all flags.

## Setup

```bash
mrkto auth setup    # Interactive — prompts for Munchkin ID, Client ID, Client Secret
mrkto auth check    # Verify credentials work
```

## Common Workflows

### Look up a lead by email
```bash
mrkto lead list --email user@example.com
```

### Check unsubscribe / marketing suspended status
```bash
mrkto lead list --email user@example.com --fields id,email,unsubscribed,marketingSuspended,emailInvalid
```

### Look up by SFDC ID
```bash
mrkto lead list --filter sfdcContactId=003xxxxxxxxxxxx
mrkto lead list --filter sfdcLeadId=00Qxxxxxxxxxxxx
```

### Get lead's campaign and list membership
```bash
mrkto lead lists 12345
mrkto lead programs 12345
mrkto lead campaigns 12345
```

### Check activity history
```bash
mrkto activity get 12345 --since 90
mrkto activity get 12345 --types 11,13  # Filter by activity type
mrkto activity types                     # List all type IDs
```

### Watch field changes
```bash
mrkto activity changes --watch email,unsubscribed --since 7
```

### Campaign operations
```bash
mrkto campaign list --name "Welcome"
mrkto campaign get 1234
mrkto campaign trigger 1234 --leads 5001,5002 --execute
```

### Static list operations
```bash
mrkto list list --name "Target"
mrkto list members 1234 --limit 50
mrkto list add 1234 --leads 5001,5002 --execute
mrkto list check 1234 --leads 5001,5002
```

### Company lookup
```bash
mrkto company list --name "ExtraHop"
mrkto company describe
```

### API stats
```bash
mrkto stats usage
mrkto stats errors --weekly
```

## Key Conventions

- **All list commands** support `--limit`, `--fields`, `--json`, `--compact`, `--raw`
- **Write operations** (schedule, trigger, add, remove) require `--execute` — without it they dry-run
- **`--filter key=value`** is the escape hatch for any Marketo filterType not promoted to a named flag
- **Lead default fields:** id, email, firstName, lastName, company, unsubscribed, marketingSuspended, emailInvalid, sfdcLeadId, sfdcContactId, createdAt, updatedAt
- Use `mrkto lead describe` to discover all available lead fields
