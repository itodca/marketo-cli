# Changelog

## [0.1.0] — 2026-03-12

Public beta release.

- Typer-based CLI with singular resource nouns and explicit Marketo concepts like `smart-campaign`, `static-list`, and `smart-list`
- Lead lookups by email, id, or any filter field, plus static list, program, and smart campaign memberships
- Activity history and lead field change tracking with paging token support
- Smart campaign browse, get, schedule, and trigger commands with dry-run by default
- Static list browse, membership, add, remove, and membership check commands
- Company and program lookups, usage and error stats, and a raw `mrkto api` escape hatch
- Profile-scoped config and token caching under `~/.config/mrkto/`, plus `.mrkto-profile` project selection
- macOS/Linux binary installer, PyInstaller release packaging, and source distribution build support
- `mrkto skill install` for the tracked agent skill in this repo
