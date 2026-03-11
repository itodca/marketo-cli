"""Skill install command — copies Claude Code skill to user or project scope."""

import shutil
import sys
from pathlib import Path
from importlib.resources import files


SKILL_SOURCE = files("mrkto.skill").joinpath("SKILL.md")

USER_SKILL_DIR = Path.home() / ".claude" / "skills" / "mrkto"
PROJECT_SKILL_DIR = Path.cwd() / ".claude" / "skills" / "mrkto"


def install_skill(scope="user"):
    """Install the mrkto skill for Claude Code."""
    if scope == "project":
        target_dir = PROJECT_SKILL_DIR
    else:
        target_dir = USER_SKILL_DIR

    target_dir.mkdir(parents=True, exist_ok=True)
    target = target_dir / "SKILL.md"

    source_text = SKILL_SOURCE.read_text()
    target.write_text(source_text)

    print(f"Skill installed to {target}")
    print(f"Scope: {scope}")
    print("Claude Code will now use 'mrkto' for Marketo operations.")
    return {"status": "installed", "path": str(target), "scope": scope}
