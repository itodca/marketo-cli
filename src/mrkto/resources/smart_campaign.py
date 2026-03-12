"""Smart campaign browsing and execution."""

from __future__ import annotations

import json

from mrkto.client import MarketoAPIError


def _folder_value(folder_id: int | None, folder_type: str | None) -> str | None:
    if folder_id is None and folder_type is None:
        return None
    if folder_id is None or folder_type is None:
        raise ValueError("Both folder_id and folder_type are required together")
    return json.dumps({"id": folder_id, "type": folder_type})


def list_smart_campaigns(
    client,
    *,
    name: str | None = None,
    folder_id: int | None = None,
    folder_type: str | None = None,
    is_active: bool | None = None,
    limit: int | None = None,
) -> dict:
    if name:
        return client.get("/asset/v1/smartCampaign/byName.json", params={"name": name})

    params = {}
    folder = _folder_value(folder_id, folder_type)
    if folder:
        params["folder"] = folder
    if is_active is not None:
        params["isActive"] = is_active

    return client.get_all_offset_pages("/asset/v1/smartCampaigns.json", params=params or None, limit=limit)


def get_smart_campaign(client, *, campaign_id: int) -> dict:
    return client.get(f"/asset/v1/smartCampaign/{campaign_id}.json")


def schedule_smart_campaign(
    client,
    *,
    campaign_id: int,
    run_at: str | None = None,
    dry_run: bool = True,
) -> dict:
    body = {"input": {}}
    if run_at:
        body["input"]["runAt"] = run_at

    if dry_run:
        return {
            "dry_run": True,
            "resource": "smart-campaign",
            "action": "schedule",
            "campaign_id": campaign_id,
            "request": body,
        }

    return client.post(f"/v1/campaigns/{campaign_id}/schedule.json", json_body=body)


def trigger_smart_campaign(
    client,
    *,
    campaign_id: int,
    lead_ids: list[int],
    dry_run: bool = True,
) -> dict:
    if not lead_ids:
        raise ValueError("At least one lead id is required")
    if len(lead_ids) > 100:
        raise MarketoAPIError("invalid_input", "A maximum of 100 leads is allowed per trigger request")

    body = {
        "input": {
            "leads": [{"id": lead_id} for lead_id in lead_ids],
        }
    }

    if dry_run:
        return {
            "dry_run": True,
            "resource": "smart-campaign",
            "action": "trigger",
            "campaign_id": campaign_id,
            "request": body,
        }

    return client.post(f"/v1/campaigns/{campaign_id}/trigger.json", json_body=body)
