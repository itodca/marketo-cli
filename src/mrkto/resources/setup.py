"""First-run setup — auth + optional skill install."""

from mrkto.resources.auth import setup_auth
from mrkto.resources.skill import install_skill


def run_setup(profile=None):
    """Run full first-time setup."""
    print("=== mrkto setup ===\n")

    # Step 1: Auth
    result = setup_auth(profile=profile)
    if result is None:
        return None

    # Step 2: Skill
    print()
    answer = input("Install Claude Code skill? [Y/n] ").strip().lower()
    if answer in ("", "y", "yes"):
        scope = input("Scope — user (all projects) or project (this dir)? [user/project] ").strip().lower()
        if scope not in ("user", "project"):
            scope = "user"
        install_skill(scope=scope)

    print("\nSetup complete. Run 'mrkto auth check' to verify.")
    return {"status": "complete"}
