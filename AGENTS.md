# marketo-cli Agent Guide

This repo ships `mrkto`, a Go-based Marketo CLI.

## Source Of Truth
- Command behavior: `internal/cmd/`
- HTTP/client behavior: `internal/client/`
- Config and profile behavior: `internal/config/`, `internal/profile/`
- User-facing contract: `README.md`
- Agent usage examples: `skills/mrkto/SKILL.md`

## Project Rules
- Keep the CLI noun/action shape stable
- Use singular top-level nouns and explicit Marketo concepts:
  - `lead`, `activity`, `company`, `program`
  - `smart-campaign`, `static-list`, `smart-list`
- Preserve `mrkto api ...` as the fallback for unsupported endpoints
- Default output is JSON; do not replace structured output with prose
- Mutating commands must remain dry-run-first and require `--execute`

## Development
- Run Go tests: `go test ./...`
- Run Python reference tests: `uv run pytest`
- Validate installer scripts: `bash -n install.sh && bash -n scripts/build-binary.sh`
- Build native artifact: `./scripts/build-binary.sh`

## Sync Points
- If you change commands, update:
  - `README.md`
  - `CLAUDE.md`
  - `skills/mrkto/SKILL.md`
- If you change release artifact names or install locations, update:
  - `install.sh`
  - `scripts/build-binary.sh`
  - `.github/workflows/release.yml`

## Avoid
- Reintroducing the old `argparse`-era command names like `campaign` or `list`
- Adding hidden prompts to mutating commands
- Reintroducing PyInstaller-era app bundle assumptions into the installer or release flow
