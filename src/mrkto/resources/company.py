"""Company lookups."""


def get_companies(client, filter_type="company", filter_values="", fields=None):
    """Get companies by filter."""
    params = {
        "filterType": filter_type,
        "filterValues": filter_values,
    }
    if fields:
        params["fields"] = fields
    result = client.get("/v1/companies.json", params=params)
    return result.get("result", [])


def describe_company(client):
    """Return company field schema."""
    result = client.get("/v1/companies/describe.json")
    return result.get("result", [])
