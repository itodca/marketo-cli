"""Smart list lookups."""

from __future__ import annotations

import json


def _folder_value(folder_id: int | None, folder_type: str | None) -> str | None:
    if folder_id is None and folder_type is None:
        return None
    if folder_id is None or folder_type is None:
        raise ValueError("Both folder_id and folder_type are required together")
    return json.dumps({"id": folder_id, "type": folder_type})


def list_smart_lists(
    client,
    *,
    name: str | None = None,
    folder_id: int | None = None,
    folder_type: str | None = None,
    limit: int | None = None,
) -> dict:
    if name:
        return client.get("/asset/v1/smartList/byName.json", params={"name": name})

    params = {}
    folder = _folder_value(folder_id, folder_type)
    if folder:
        params["folder"] = folder
    return client.get_all_offset_pages("/asset/v1/smartLists.json", params=params or None, limit=limit)


def get_smart_list(client, *, list_id: int, include_rules: bool = False) -> dict:
    params = {"includeRules": include_rules} if include_rules else None
    return client.get(f"/asset/v1/smartList/{list_id}.json", params=params)
