"""Static list lookups and membership changes."""

from __future__ import annotations

from mrkto.client import MarketoAPIError


def _lead_inputs(lead_ids: list[int]) -> list[dict[str, int]]:
    return [{"id": lead_id} for lead_id in lead_ids]


def _validate_lead_ids(lead_ids: list[int]) -> None:
    if not lead_ids:
        raise ValueError("At least one lead id is required")
    if len(lead_ids) > 300:
        raise MarketoAPIError("invalid_input", "A maximum of 300 leads is allowed per static list request")


def list_static_lists(
    client,
    *,
    name: str | None = None,
    program_name: str | None = None,
    workspace_name: str | None = None,
    limit: int | None = None,
) -> dict:
    params = {}
    if name:
        params["name"] = name
    if program_name:
        params["programName"] = program_name
    if workspace_name:
        params["workspaceName"] = workspace_name
    return client.get_all_pages("/v1/lists.json", params=params or None, limit=limit)


def get_static_list(client, *, list_id: int) -> dict:
    return client.get(f"/v1/lists/{list_id}.json")


def get_static_list_members(
    client,
    *,
    list_id: int,
    fields: str | None = None,
    limit: int | None = None,
) -> dict:
    params = {}
    if fields:
        params["fields"] = fields
    return client.get_all_pages(f"/v1/lists/{list_id}/leads.json", params=params or None, limit=limit)


def add_to_static_list(
    client,
    *,
    list_id: int,
    lead_ids: list[int],
    dry_run: bool = True,
) -> dict:
    _validate_lead_ids(lead_ids)
    body = {"input": _lead_inputs(lead_ids)}
    if dry_run:
        return {
            "dry_run": True,
            "resource": "static-list",
            "action": "add",
            "list_id": list_id,
            "request": body,
        }
    return client.post(f"/v1/lists/{list_id}/leads.json", json_body=body)


def remove_from_static_list(
    client,
    *,
    list_id: int,
    lead_ids: list[int],
    dry_run: bool = True,
) -> dict:
    _validate_lead_ids(lead_ids)
    body = {"input": _lead_inputs(lead_ids)}
    params = {"id": lead_ids}
    if dry_run:
        return {
            "dry_run": True,
            "resource": "static-list",
            "action": "remove",
            "list_id": list_id,
            "request": body,
            "params": params,
        }
    return client.delete(f"/v1/lists/{list_id}/leads.json", params=params, json_body=body)


def check_static_list_membership(client, *, list_id: int, lead_ids: list[int]) -> dict:
    _validate_lead_ids(lead_ids)
    return client.get(f"/v1/lists/{list_id}/leads/ismember.json", params={"id": lead_ids})
