"""Lead lookups and membership helpers."""

from __future__ import annotations

DEFAULT_FIELDS = ",".join(
    [
        "id",
        "email",
        "firstName",
        "lastName",
        "company",
        "unsubscribed",
        "marketingSuspended",
        "emailInvalid",
        "sfdcLeadId",
        "sfdcContactId",
        "createdAt",
        "updatedAt",
    ]
)


def _fields_param(fields: str | None) -> str:
    return fields or DEFAULT_FIELDS


def list_leads(client, *, filter_type: str, filter_values: str, fields: str | None = None, limit: int | None = None) -> dict:
    params = {
        "filterType": filter_type,
        "filterValues": filter_values,
        "fields": _fields_param(fields),
    }
    return client.get_all_pages("/v1/leads.json", params=params, limit=limit)


def get_lead(client, *, lead_id: int, fields: str | None = None) -> dict:
    params = {"fields": _fields_param(fields)}
    return client.get(f"/v1/lead/{lead_id}.json", params=params)


def describe_lead(client, *, detailed: bool = True) -> dict:
    path = "/v1/leads/describe2.json" if detailed else "/v1/leads/describe.json"
    return client.get(path)


def get_lead_static_lists(client, *, lead_id: int, limit: int | None = None) -> dict:
    return client.get_all_pages(f"/v1/leads/{lead_id}/listMembership.json", limit=limit)


def get_lead_programs(
    client,
    *,
    lead_id: int,
    limit: int | None = None,
    program_ids: list[int] | None = None,
) -> dict:
    params = {}
    if program_ids:
        params["filterType"] = "programId"
        params["filterValues"] = ",".join(str(program_id) for program_id in program_ids)
    return client.get_all_pages(
        f"/v1/leads/{lead_id}/programMembership.json",
        params=params or None,
        limit=limit,
    )


def get_lead_smart_campaigns(client, *, lead_id: int, limit: int | None = None) -> dict:
    return client.get_all_pages(f"/v1/leads/{lead_id}/smartCampaignMembership.json", limit=limit)
