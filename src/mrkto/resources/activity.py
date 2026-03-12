"""Activity lookups and lead change tracking."""

from __future__ import annotations

from datetime import datetime, timedelta, timezone


def _since_datetime(days: int) -> str:
    since = datetime.now(timezone.utc) - timedelta(days=days)
    return since.strftime("%Y-%m-%dT%H:%M:%SZ")


def _paging_token(client, days: int) -> str:
    paging = client.get("/v1/activities/pagingtoken.json", params={"sinceDatetime": _since_datetime(days)})
    token = paging.get("nextPageToken")
    if not token:
        raise KeyError("Marketo paging token response did not include nextPageToken")
    return token


def get_activity_types(client) -> dict:
    return client.get("/v1/activities/types.json")


def list_activities(
    client,
    *,
    lead_id: int,
    activity_type_ids: list[int] | None = None,
    since_days: int = 30,
    limit: int | None = None,
) -> dict:
    params = {
        "leadIds": str(lead_id),
        "nextPageToken": _paging_token(client, since_days),
    }
    if activity_type_ids:
        params["activityTypeIds"] = ",".join(str(activity_type_id) for activity_type_id in activity_type_ids)
    return client.get_all_pages("/v1/activities.json", params=params, limit=limit)


def get_lead_changes(
    client,
    *,
    fields: list[str],
    since_days: int = 30,
    lead_ids: list[int] | None = None,
    list_id: int | None = None,
    limit: int | None = None,
) -> dict:
    params = {
        "fields": ",".join(fields),
        "nextPageToken": _paging_token(client, since_days),
    }
    if lead_ids:
        params["leadIds"] = ",".join(str(lead_id) for lead_id in lead_ids)
    if list_id is not None:
        params["listId"] = list_id
    return client.get_all_pages("/v1/activities/leadchanges.json", params=params, limit=limit)
