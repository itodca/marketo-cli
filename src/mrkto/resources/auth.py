"""Auth check command."""


def check_auth(client):
    """Verify credentials work by fetching activity types (lightweight call)."""
    result = client.get("/v1/activities/types.json")
    count = len(result.get("result", []))
    return {
        "status": "ok",
        "munchkin_id": client.config.munchkin_id,
        "rest_url": client.config.rest_url,
        "activity_types_available": count,
    }
