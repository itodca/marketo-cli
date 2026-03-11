"""Configuration loading with profile support.

Config lives at ~/.config/mrkto/:
  config                  — default profile
  profiles/<name>         — named profiles

Priority: env vars > profile config file.
Select profile: --profile flag or MRKTO_PROFILE env var.
"""

import os
import sys
from dataclasses import dataclass
from pathlib import Path

CONFIG_DIR = Path.home() / ".config" / "mrkto"
PROFILES_DIR = CONFIG_DIR / "profiles"


@dataclass
class Config:
    munchkin_id: str
    client_id: str
    client_secret: str
    rest_url: str
    identity_url: str
    profile: str


def config_file_for(profile: str | None = None) -> Path:
    """Return the config file path for a profile."""
    if profile:
        return PROFILES_DIR / profile
    return CONFIG_DIR / "config"


def list_profiles() -> list[str]:
    """List all available profiles."""
    profiles = []
    if (CONFIG_DIR / "config").exists():
        profiles.append("default")
    if PROFILES_DIR.exists():
        profiles.extend(sorted(p.name for p in PROFILES_DIR.iterdir() if p.is_file()))
    return profiles


def _load_config_file(profile: str | None = None) -> dict:
    """Read key=value pairs from a config file."""
    path = config_file_for(profile)
    if not path.exists():
        return {}
    result = {}
    for line in path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        key, _, value = line.partition("=")
        if value:
            result[key.strip()] = value.strip()
    return result


def _find_profile_file() -> str | None:
    """Walk up from cwd looking for .mrkto-profile, like .tool-versions."""
    cwd = Path.cwd()
    for directory in [cwd, *cwd.parents]:
        candidate = directory / ".mrkto-profile"
        if candidate.exists():
            name = candidate.read_text().strip()
            if name:
                return name
        if directory == directory.parent:
            break
    return None


def load_config(profile: str | None = None) -> Config:
    """Load config from profile file, with env var overrides."""
    # Resolve profile: flag > env var > .mrkto-profile file > default
    profile = profile or os.environ.get("MRKTO_PROFILE") or _find_profile_file() or None
    file_cfg = _load_config_file(profile)

    munchkin_id = os.environ.get("MARKETO_MUNCHKIN_ID") or file_cfg.get("MARKETO_MUNCHKIN_ID", "")
    client_id = os.environ.get("MARKETO_CLIENT_ID") or file_cfg.get("MARKETO_CLIENT_ID", "")
    client_secret = os.environ.get("MARKETO_CLIENT_SECRET") or file_cfg.get("MARKETO_CLIENT_SECRET", "")

    if not all([munchkin_id, client_id, client_secret]):
        hint = f" (profile: {profile})" if profile else ""
        print(
            f"Error: Marketo credentials not found{hint}.",
            file=sys.stderr,
        )
        print("Run 'mrkto setup' or set MARKETO_MUNCHKIN_ID, MARKETO_CLIENT_ID, MARKETO_CLIENT_SECRET as env vars.", file=sys.stderr)
        sys.exit(1)

    rest_url = (
        os.environ.get("MARKETO_REST_URL")
        or file_cfg.get("MARKETO_REST_URL")
        or f"https://{munchkin_id}.mktorest.com/rest"
    )
    identity_url = (
        os.environ.get("MARKETO_IDENTITY_URL")
        or file_cfg.get("MARKETO_IDENTITY_URL")
        or f"https://{munchkin_id}.mktorest.com/identity"
    )

    return Config(
        munchkin_id=munchkin_id,
        client_id=client_id,
        client_secret=client_secret,
        rest_url=rest_url,
        identity_url=identity_url,
        profile=profile or "default",
    )
