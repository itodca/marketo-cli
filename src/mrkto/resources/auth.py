"""Auth setup, check, and profile management."""

import sys

from mrkto.config import config_file_for, list_profiles, PROFILES_DIR


def setup_auth(profile=None):
    """Prompt for credentials and write to config file."""
    config_path = config_file_for(profile)
    label = f"profile '{profile}'" if profile else "default profile"

    if config_path.exists():
        print(f"Config for {label} already exists at {config_path}", file=sys.stderr)
        confirm = input("Overwrite? [y/N] ").strip().lower()
        if confirm != "y":
            print("Aborted.", file=sys.stderr)
            return None

    print(f"Marketo API credentials setup ({label}).")
    print("Get these from Admin > Integration > LaunchPoint in your Marketo instance.")
    print("Docs: https://experienceleague.adobe.com/en/docs/marketo-developer/marketo/rest/rest-api\n")
    munchkin_id = input("Munchkin ID (e.g. 123-ABC-456): ").strip()
    client_id = input("Client ID: ").strip()
    client_secret = input("Client Secret: ").strip()

    if not all([munchkin_id, client_id, client_secret]):
        print("All three values are required.", file=sys.stderr)
        sys.exit(1)

    config_path.parent.mkdir(parents=True, exist_ok=True)
    config_path.write_text(
        f"MARKETO_MUNCHKIN_ID={munchkin_id}\n"
        f"MARKETO_CLIENT_ID={client_id}\n"
        f"MARKETO_CLIENT_SECRET={client_secret}\n"
    )
    print(f"\nCredentials saved to {config_path}")
    if profile:
        print(f"Use with: mrkto --profile {profile} <command>")
        print(f"Or set:   export MRKTO_PROFILE={profile}")
    print("Run 'mrkto auth check' to verify.")
    return {"status": "saved", "path": str(config_path), "profile": profile or "default"}


def list_auth():
    """List all configured profiles."""
    profiles = list_profiles()
    if not profiles:
        print("No profiles configured. Run 'mrkto setup' to create one.")
        return None
    return [{"profile": p} for p in profiles]


def check_auth(client):
    """Verify credentials work by fetching activity types (lightweight call)."""
    result = client.get("/v1/activities/types.json")
    count = len(result.get("result", []))
    return {
        "status": "ok",
        "profile": client.config.profile,
        "munchkin_id": client.config.munchkin_id,
        "rest_url": client.config.rest_url,
        "activity_types_available": count,
    }
