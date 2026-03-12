"""Program lookups."""

from __future__ import annotations


def list_programs(client, *, name: str | None = None, limit: int | None = None) -> dict:
    if name:
        return client.get("/asset/v1/program/byName.json", params={"name": name})
    return client.get_all_offset_pages("/asset/v1/programs.json", limit=limit)


def get_program(client, *, program_id: int) -> dict:
    return client.get(f"/asset/v1/program/{program_id}.json")
