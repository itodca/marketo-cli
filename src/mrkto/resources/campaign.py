"""Campaign list, get, schedule, trigger."""

import sys


def list_campaigns(client, name=None, program_name=None, limit=None):
    """List campaigns, optionally filtered by name or program."""
    params = {}
    if name:
        params["name"] = name
    if program_name:
        params["programName"] = program_name
    results = client.get_paginated("/v1/campaigns.json", params=params)
    if limit:
        results = results[:limit]
    return results


def get_campaign(client, campaign_id):
    """Get a campaign by ID."""
    result = client.get(f"/v1/campaigns/{campaign_id}.json")
    return result.get("result", [])


def schedule_campaign(client, campaign_id, run_at=None, dry_run=True):
    """Schedule a batch campaign."""
    body = {}
    if run_at:
        body["runAt"] = run_at
    if dry_run:
        print(f"[dry-run] Would schedule campaign {campaign_id}", file=sys.stderr)
        print(f"[dry-run] Body: {body}", file=sys.stderr)
        return {"dry_run": True, "campaign_id": campaign_id, "body": body}
    result = client.post(f"/v1/campaigns/{campaign_id}/schedule.json", json_body=body)
    return result.get("result", [])


def trigger_campaign(client, campaign_id, lead_ids, dry_run=True):
    """Trigger a smart campaign for specific leads."""
    body = {"input": {"leads": [{"id": lid} for lid in lead_ids]}}
    if dry_run:
        print(f"[dry-run] Would trigger campaign {campaign_id} for leads {lead_ids}", file=sys.stderr)
        return {"dry_run": True, "campaign_id": campaign_id, "lead_ids": lead_ids}
    result = client.post(f"/v1/campaigns/{campaign_id}/trigger.json", json_body=body)
    return result.get("result", [])
