"""Configuration loading from environment variables."""

import os
import sys
from dataclasses import dataclass

from dotenv import load_dotenv


@dataclass
class Config:
    munchkin_id: str
    client_id: str
    client_secret: str
    rest_url: str
    identity_url: str


def load_config() -> Config:
    load_dotenv()

    munchkin_id = os.environ.get("MARKETO_MUNCHKIN_ID", "")
    client_id = os.environ.get("MARKETO_CLIENT_ID", "")
    client_secret = os.environ.get("MARKETO_CLIENT_SECRET", "")

    if not all([munchkin_id, client_id, client_secret]):
        print(
            "Error: MARKETO_MUNCHKIN_ID, MARKETO_CLIENT_ID, and MARKETO_CLIENT_SECRET are required.",
            file=sys.stderr,
        )
        print("Set them in .env or as environment variables.", file=sys.stderr)
        sys.exit(1)

    rest_url = os.environ.get(
        "MARKETO_REST_URL", f"https://{munchkin_id}.mktorest.com/rest"
    )
    identity_url = os.environ.get(
        "MARKETO_IDENTITY_URL", f"https://{munchkin_id}.mktorest.com/identity"
    )

    return Config(
        munchkin_id=munchkin_id,
        client_id=client_id,
        client_secret=client_secret,
        rest_url=rest_url,
        identity_url=identity_url,
    )
