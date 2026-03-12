# marketo-cli (mrkto)

Marketo REST API CLI for humans and AI agents.

## Tech Stack
- Python 3.10+, `uv` for local workflows, `hatchling` for packaging
- Typer for the CLI
- `requests` for HTTP
- PyInstaller for binary release artifacts

## Structure
- `src/mrkto/cli.py` — Typer entrypoint and command tree
- `src/mrkto/client.py` — auth, token cache, HTTP helpers, retries, pagination
- `src/mrkto/config.py` — config file + env var loading, profile support
- `src/mrkto/output.py` — JSON/compact/raw formatting
- `src/mrkto/resources/` — one file per resource group
- `skills/mrkto/SKILL.md` — agent skill for using the CLI
- `scripts/build-binary.sh` — PyInstaller release packaging
- `.github/workflows/release.yml` — macOS/Linux release publishing

## Command Contract
- Use singular top-level resource nouns: `lead`, `activity`, `program`, `company`
- Keep Marketo-specific overlapping concepts explicit: `smart-campaign`, `static-list`, `smart-list`
- Preserve the raw escape hatch: `mrkto api get|post|delete`
- Default output is pretty JSON; `--compact` and `--raw` are first-class
- Write operations stay dry-run by default and require `--execute`

## Installation
- Primary install path is the binary installer: `curl -fsSL .../install.sh | bash`
- Keep source install working via `pip install .` and the `mrkto` console script
- Binary naming and checksums must stay aligned with `install.sh`

## Testing
- Run tests with `uv run pytest`
- Validate syntax with `python3 -m compileall src tests`
- When changing packaging or release flow, also check `bash -n install.sh` and `bash -n scripts/build-binary.sh`

## Maintenance Notes
- Do not reintroduce the old `argparse` command model
- Keep `README.md`, this file, and `skills/mrkto/SKILL.md` in sync when the command tree changes
- Historical design notes under `../docs/marketo-cli/` may describe the old CLI shape; treat the code and README as source of truth
