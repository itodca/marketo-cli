# marketo-cli (mrkto)

Marketo REST API CLI for humans and AI agents.

## Tech Stack
- Go 1.25 for the CLI implementation
- Cobra for the command tree
- Standard library `net/http` + internal client helpers for HTTP/auth
- GoReleaser for release artifacts
- Python remains in the repo as the reference implementation during migration

## Structure
- `cmd/mrkto/main.go` — binary entrypoint
- `internal/cmd/` — Cobra command tree and CLI wiring
- `internal/client/` — auth, token cache, HTTP helpers, retries, pagination
- `internal/config/` and `internal/profile/` — config file + env var loading, profile support
- `internal/output/` — JSON/compact/raw formatting
- `skills/mrkto/SKILL.md` — agent skill for using the CLI
- `scripts/build-binary.sh` — local native archive build
- `.goreleaser.yaml` — native release packaging
- `.github/workflows/release.yml` — tagged native release publishing

## Command Contract
- Use singular top-level resource nouns: `lead`, `activity`, `program`, `company`
- Keep Marketo-specific overlapping concepts explicit: `smart-campaign`, `static-list`, `smart-list`
- Preserve the raw escape hatch: `mrkto api get|post|delete`
- Default output is pretty JSON; `--compact` and `--raw` are first-class
- Write operations stay dry-run by default and require `--execute`

## Installation
- Primary install path is the binary installer: `curl -fsSL .../install.sh | bash`
- Keep native source install viable via `go build ./cmd/mrkto`
- Binary naming and checksums must stay aligned with `install.sh`

## Testing
- Run Go tests with `go test ./...`
- Keep Python reference tests runnable with `uv run pytest`
- When changing packaging or release flow, also check `bash -n install.sh` and `bash -n scripts/build-binary.sh`

## Maintenance Notes
- Do not reintroduce the old `argparse` command model
- Keep `README.md`, this file, and `skills/mrkto/SKILL.md` in sync when the command tree changes
- Historical design notes under `../docs/marketo-cli/` may describe the old CLI shape; treat the code and README as source of truth
