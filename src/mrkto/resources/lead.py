"""Lead lookups, describe, memberships."""

DEFAULT_FIELDS = "id,email,firstName,lastName,company,unsubscribed,marketingSuspended,emailInvalid,sfdcLeadId,sfdcContactId,createdAt,updatedAt"


def lookup_lead(client, query, fields=None):
    """Auto-detect query type and look up lead."""
    query = query.strip()
    if query.isdigit():
        return get_lead(client, lead_id=int(query), fields=fields)
    elif "@" in query:
        return get_lead(client, email=query, fields=fields)
    else:
        return search_leads(client, name=query, fields=fields)


def get_lead(client, email=None, lead_id=None, fields=None):
    """Get lead by email or Marketo ID."""
    if email:
        params = {
            "filterType": "email",
            "filterValues": email,
            "fields": fields or DEFAULT_FIELDS,
        }
        result = client.get("/v1/leads.json", params=params)
    elif lead_id:
        result = client.get(
            f"/v1/lead/{lead_id}.json",
            params={"fields": fields or DEFAULT_FIELDS},
        )
    else:
        return []
    return result.get("result", [])


def search_leads(client, name, fields=None):
    """Search leads by name (firstName + lastName)."""
    parts = name.strip().split(None, 1)
    if len(parts) == 2:
        # Search by lastName, then filter by firstName client-side
        params = {
            "filterType": "lastName",
            "filterValues": parts[1],
            "fields": fields or DEFAULT_FIELDS,
        }
        result = client.get("/v1/leads.json", params=params)
        leads = result.get("result", [])
        first = parts[0].lower()
        return [l for l in leads if (l.get("firstName") or "").lower().startswith(first)]
    else:
        # Single name — search as lastName
        params = {
            "filterType": "lastName",
            "filterValues": parts[0],
            "fields": fields or DEFAULT_FIELDS,
        }
        result = client.get("/v1/leads.json", params=params)
        return result.get("result", [])


def describe_lead(client):
    """Return lead field schema."""
    result = client.get("/v1/leads/describe2.json")
    return result.get("result", [])


def get_lead_lists(client, lead_id):
    """Get static lists a lead belongs to."""
    result = client.get(f"/v1/leads/{lead_id}/listMembership.json")
    return result.get("result", [])


def get_lead_programs(client, lead_id):
    """Get programs a lead is in."""
    result = client.get(f"/v1/leads/{lead_id}/programMembership.json")
    return result.get("result", [])


def get_lead_campaigns(client, lead_id):
    """Get smart campaigns for a lead."""
    result = client.get(f"/v1/leads/{lead_id}/smartCampaignMembership.json")
    return result.get("result", [])
