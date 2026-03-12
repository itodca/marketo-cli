"""Usage and error statistics."""

from __future__ import annotations


def get_usage(client, *, weekly: bool = False) -> dict:
    path = "/v1/stats/usage/last7days.json" if weekly else "/v1/stats/usage.json"
    return client.get(path)


def get_errors(client, *, weekly: bool = False) -> dict:
    path = "/v1/stats/errors/last7days.json" if weekly else "/v1/stats/errors.json"
    return client.get(path)
