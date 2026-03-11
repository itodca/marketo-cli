"""Skill install command — delegates to npx skills ecosystem."""

import shutil
import subprocess
import sys


def install_skill(scope="user"):
    """Install the mrkto agent skill via npx skills."""
    if not shutil.which("npx"):
        print("npx not found. Install Node.js or run manually:", file=sys.stderr)
        print("  npx skills add itodca/marketo-cli", file=sys.stderr)
        sys.exit(1)

    cmd = ["npx", "skills", "add", "itodca/marketo-cli"]
    if scope == "user":
        cmd.append("--global")

    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd)
    if result.returncode != 0:
        sys.exit(result.returncode)

    return None
