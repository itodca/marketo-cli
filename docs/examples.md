# Examples

## First-Time Setup

```bash
mrkto setup
mrkto auth check
```

Use a named profile:

```bash
mrkto auth setup --profile sandbox
mrkto auth check --profile sandbox
```

## Lead Lookups

Look up a lead by email:

```bash
mrkto lead list --email user@example.com
```

Look up by Marketo id:

```bash
mrkto lead get 12345
```

Look up by a custom filter:

```bash
mrkto lead list --filter sfdcContactId=003xxxxxxxxxxxx
```

Return a narrower field set:

```bash
mrkto lead list --email user@example.com --fields id,email,unsubscribed,marketingSuspended
```

## Membership Lookups

Static lists for a lead:

```bash
mrkto lead static-lists 12345
```

Programs for a lead:

```bash
mrkto lead programs 12345
```

Smart campaigns for a lead:

```bash
mrkto lead smart-campaigns 12345
```

## Activity Lookups

Recent activities for a lead:

```bash
mrkto activity list 12345 --since 30
```

Filter to activity types:

```bash
mrkto activity list 12345 --type-id 11 --type-id 13
```

Watch lead field changes:

```bash
mrkto activity changes --watch email --watch unsubscribed --since 7
```

## Smart Campaigns

Browse active campaigns:

```bash
mrkto smart-campaign list --active --limit 10
```

Dry-run a trigger:

```bash
mrkto smart-campaign trigger 1234 --lead 1001 --lead 1002
```

Execute it:

```bash
mrkto smart-campaign trigger 1234 --lead 1001 --lead 1002 --execute
```

## Static Lists

List static lists:

```bash
mrkto static-list list --name "Newsletter"
```

Inspect members:

```bash
mrkto static-list members 456 --limit 50
```

Dry-run an add:

```bash
mrkto static-list add 456 --lead 1001 --lead 1002
```

Execute the add:

```bash
mrkto static-list add 456 --lead 1001 --lead 1002 --execute
```

## Raw API Escape Hatch

Get leads directly:

```bash
mrkto api get /v1/leads.json --query filterType=email --query filterValues=user@example.com
```

Post JSON from a file:

```bash
mrkto api post /v1/leads/push.json --input payload.json
```

## Redirecting Output To Files

Because `mrkto` writes results to `stdout`, normal shell redirection works out of the box.

Write pretty JSON to a file:

```bash
mrkto lead list --email user@example.com > lead.json
```

Write full raw payloads:

```bash
mrkto lead list --email user@example.com --raw > lead-raw.json
```

Write line-oriented JSON:

```bash
mrkto activity list 12345 --compact > activities.ndjson
```

Pipe to `jq`:

```bash
mrkto lead list --email user@example.com --raw | jq .
```
