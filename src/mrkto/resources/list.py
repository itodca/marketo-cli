"""Static list management."""

import sys


def list_lists(client, name=None, limit=None):
    """List all static lists."""
    params = {}
    if name:
        params["name"] = name
    results = client.get_paginated("/v1/lists.json", params=params)
    if limit:
        results = results[:limit]
    return results


def get_list(client, list_id):
    """Get a list by ID."""
    result = client.get(f"/v1/lists/{list_id}.json")
    return result.get("result", [])


def get_list_members(client, list_id, fields=None, limit=None):
    """Get all members of a list (paginated)."""
    params = {}
    if fields:
        params["fields"] = fields
    results = client.get_paginated(f"/v1/lists/{list_id}/leads.json", params=params)
    if limit:
        results = results[:limit]
    return results


def add_to_list(client, list_id, lead_ids, dry_run=True):
    """Add leads to a static list."""
    if dry_run:
        print(f"[dry-run] Would add leads {lead_ids} to list {list_id}", file=sys.stderr)
        return {"dry_run": True, "list_id": list_id, "lead_ids": lead_ids}
    body = [{"id": lid} for lid in lead_ids]
    result = client.post(f"/v1/lists/{list_id}/leads.json", json_body=body)
    return result.get("result", [])


def remove_from_list(client, list_id, lead_ids, dry_run=True):
    """Remove leads from a static list."""
    if dry_run:
        print(f"[dry-run] Would remove leads {lead_ids} from list {list_id}", file=sys.stderr)
        return {"dry_run": True, "list_id": list_id, "lead_ids": lead_ids}
    # Marketo uses DELETE for this, but we use the JSON body approach
    params = {"id": ",".join(str(lid) for lid in lead_ids)}
    result = client.get(f"/v1/lists/{list_id}/leads.json", params={"_method": "DELETE", **params})
    return result.get("result", [])


def is_member(client, list_id, lead_ids):
    """Check if leads are members of a list."""
    params = {"id": ",".join(str(lid) for lid in lead_ids)}
    result = client.get(f"/v1/lists/{list_id}/leads/ismember.json", params=params)
    return result.get("result", [])
