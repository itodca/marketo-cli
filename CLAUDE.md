# marketo-cli (mrkto)

Lightweight CLI for the Marketo REST API. Built for humans and AI agents.

## Tech Stack
- Python 3.10+, `uv` for deps, `hatchling` for build
- One runtime dep: `requests`
- `argparse` for CLI (no click/typer)

## Structure
- `src/mrkto/cli.py` — entry point, subcommand routing
- `src/mrkto/client.py` — auth, token cache, HTTP, pagination
- `src/mrkto/config.py` — config file + env var loading, profile support
- `src/mrkto/help.py` — full command reference
- `src/mrkto/output.py` — JSON/compact/raw formatting
- `src/mrkto/resources/` — one file per API resource group
- `src/mrkto/skill/SKILL.md` — Claude Code skill (ships with package)

## Conventions
- JSON output by default
- Write operations require `--execute` (dry-run by default)
- Config at `~/.config/mrkto/` with profile support
- Tests in `tests/`, run with `uv run pytest`
- Commit after each meaningful change
