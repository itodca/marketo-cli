"""Authentication setup and profile inspection helpers."""

from __future__ import annotations

import os

from mrkto.config import config_file_for, list_profiles


def write_auth_config(
    *,
    munchkin_id: str,
    client_id: str,
    client_secret: str,
    profile: str | None = None,
    overwrite: bool = False,
) -> dict:
    """Persist Marketo credentials for the target profile."""

    config_path = config_file_for(profile)
    if config_path.exists() and not overwrite:
        raise FileExistsError(str(config_path))

    config_path.parent.mkdir(parents=True, exist_ok=True)
    config_path.write_text(
        f"MARKETO_MUNCHKIN_ID={munchkin_id}\n"
        f"MARKETO_CLIENT_ID={client_id}\n"
        f"MARKETO_CLIENT_SECRET={client_secret}\n"
    )

    try:
        os.chmod(config_path, 0o600)
    except OSError:
        pass

    return {
        "status": "saved",
        "path": str(config_path),
        "profile": profile or "default",
    }


def list_auth() -> dict:
    profiles = [{"profile": profile} for profile in list_profiles()]
    return {
        "success": True,
        "result": profiles,
    }


def check_auth(client) -> dict:
    result = client.get("/v1/activities/types.json")
    return {
        "success": True,
        "result": [
            {
                "status": "ok",
                "profile": client.config.profile,
                "munchkin_id": client.config.munchkin_id,
                "rest_url": client.config.rest_url,
                "activity_types_available": len(result.get("result", [])),
            }
        ],
    }
