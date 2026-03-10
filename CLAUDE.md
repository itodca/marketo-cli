# marketo-cli (mrkto)

Lightweight CLI for the Marketo REST API. Built for humans and AI agents.

## Tech Stack
- Python 3.10+, `uv` for deps, `hatchling` for build
- Two runtime deps: `requests`, `python-dotenv`
- `argparse` for CLI (no click/typer)

## Structure
- `src/mrkto/cli.py` — entry point, subcommand routing
- `src/mrkto/client.py` — auth, token cache, HTTP, pagination
- `src/mrkto/config.py` — env var loading, URL derivation
- `src/mrkto/output.py` — JSON/compact formatting
- `src/mrkto/resources/` — one file per API resource group

## Conventions
- JSON output by default
- `--dry-run` default for write operations, `--execute` to apply
- Tests in `tests/`, run with `pytest`
- Commit after each meaningful change
