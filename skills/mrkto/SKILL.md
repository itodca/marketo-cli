---
name: mrkto
description: Use when the user needs to interact with Marketo — looking up leads, checking unsubscribe status, campaign membership, static list membership, smart lists, activity history, or any Marketo REST API operation. Triggers on mentions of Marketo, leads, unsubscribe, email marketing, campaign membership, or marketing automation lookups.
---

# mrkto — Marketo REST API CLI

## Quick Reference

Run `mrkto --help` for the full command reference with all flags.

## Setup

```bash
mrkto setup         # Interactive shortcut
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
mrkto lead static-lists 12345
mrkto lead programs 12345
mrkto lead smart-campaigns 12345
```

### Check activity history
```bash
mrkto activity list 12345 --since 90
mrkto activity list 12345 --type-id 11 --type-id 13
mrkto activity types
```

### Watch field changes
```bash
mrkto activity changes --watch email --watch unsubscribed --since 7
```

### Smart campaign operations
```bash
mrkto smart-campaign list --name "Welcome"
mrkto smart-campaign get 1234
mrkto smart-campaign trigger 1234 --lead 5001 --lead 5002 --execute
```

### Static list operations
```bash
mrkto static-list list --name "Target"
mrkto static-list members 1234 --limit 50
mrkto static-list add 1234 --lead 5001 --lead 5002 --execute
mrkto static-list check 1234 --lead 5001 --lead 5002
```

### Smart list operations
```bash
mrkto smart-list list --name "MQL"
mrkto smart-list get 1234
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

### Raw API escape hatch
```bash
mrkto api get /v1/leads.json --query filterType=email --query filterValues=user@example.com
```

## Key Conventions

- **The CLI uses singular resource nouns** with explicit Marketo concepts like `smart-campaign`, `static-list`, and `smart-list`
- **Structured output flags** are `--json`, `--compact`, and `--raw`
- **Write operations** (schedule, trigger, add, remove) require `--execute` — without it they dry-run
- **`--filter key=value`** is the escape hatch for lead and company filters not promoted to a named flag
- **Lead default fields:** id, email, firstName, lastName, company, unsubscribed, marketingSuspended, emailInvalid, sfdcLeadId, sfdcContactId, createdAt, updatedAt
- Use `mrkto lead describe` to discover all available lead fields
