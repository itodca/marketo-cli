"""Skill commands — delegate to the npx skills ecosystem."""

import shutil
import subprocess
import sys


def _skills_command(action: str, *, global_install: bool = False) -> list[str]:
    cmd = ["npx", "skills", action, "itodca/marketo-cli"]
    if global_install:
        cmd.append("--global")
    return cmd


def _run_skills_command(action: str, *, global_install: bool = False) -> None:
    cmd = _skills_command(action, global_install=global_install)

    if not shutil.which("npx"):
        print("npx not found. Install Node.js or run manually:", file=sys.stderr)
        print(f"  {' '.join(cmd)}", file=sys.stderr)
        sys.exit(1)

    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd)
    if result.returncode != 0:
        sys.exit(result.returncode)


def install_skill(*, global_install: bool = False):
    """Install the mrkto agent skill via npx skills."""
    _run_skills_command("add", global_install=global_install)
    return None


def uninstall_skill(*, global_install: bool = False):
    """Uninstall the mrkto agent skill via npx skills."""
    _run_skills_command("remove", global_install=global_install)
    return None
