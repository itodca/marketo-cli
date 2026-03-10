"""Activity types, lead activities, changes."""

from datetime import datetime, timedelta


def get_activity_types(client):
    """List all activity types."""
    result = client.get("/v1/activities/types.json")
    return result.get("result", [])


def get_lead_activities(client, lead_id, activity_type_ids=None, since_days=30):
    """Get activities for a lead."""
    since = (datetime.utcnow() - timedelta(days=since_days)).strftime("%Y-%m-%dT%H:%M:%SZ")
    paging = client.get("/v1/activities/pagingtoken.json", params={"sinceDatetime": since})
    token = paging["nextPageToken"]

    params = {
        "activityTypeIds": activity_type_ids or "",
        "leadIds": str(lead_id),
        "nextPageToken": token,
    }
    return client.get_paginated("/v1/activities.json", params=params)


def get_lead_changes(client, fields, since_days=30):
    """Get lead field changes."""
    since = (datetime.utcnow() - timedelta(days=since_days)).strftime("%Y-%m-%dT%H:%M:%SZ")
    paging = client.get("/v1/activities/pagingtoken.json", params={"sinceDatetime": since})
    token = paging["nextPageToken"]

    params = {
        "fields": fields,
        "nextPageToken": token,
    }
    return client.get_paginated("/v1/activities/leadchanges.json", params=params)
