"""Company lookups."""

from __future__ import annotations


def list_companies(
    client,
    *,
    filter_type: str,
    filter_values: str,
    fields: str | None = None,
    limit: int | None = None,
) -> dict:
    params = {
        "filterType": filter_type,
        "filterValues": filter_values,
    }
    if fields:
        params["fields"] = fields
    return client.get_all_pages("/v1/companies.json", params=params, limit=limit)


def describe_company(client) -> dict:
    return client.get("/v1/companies/describe.json")
