"""Raw API escape hatch."""

from __future__ import annotations

from typing import Any


def api_get(client, *, path: str, query: dict[str, Any] | None = None) -> dict:
    return client.get(path, params=query)


def api_post(
    client,
    *,
    path: str,
    query: dict[str, Any] | None = None,
    body: dict[str, Any] | None = None,
) -> dict:
    return client.post(path, params=query, json_body=body)


def api_delete(
    client,
    *,
    path: str,
    query: dict[str, Any] | None = None,
    body: dict[str, Any] | None = None,
) -> dict:
    return client.delete(path, params=query, json_body=body)
