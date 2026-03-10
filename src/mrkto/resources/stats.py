"""API usage and error stats."""


def get_usage(client, weekly=False):
    """Get API usage stats."""
    path = "/v1/stats/usage/last7days.json" if weekly else "/v1/stats/usage.json"
    result = client.get(path)
    return result.get("result", [])


def get_errors(client, weekly=False):
    """Get API error stats."""
    path = "/v1/stats/errors/last7days.json" if weekly else "/v1/stats/errors.json"
    result = client.get(path)
    return result.get("result", [])
