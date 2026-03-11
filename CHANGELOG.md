# Changelog

## [1.0.0] — 2026-03-11

Initial release.

- Lead lookups by email, ID, or any filter field (`--filter key=value` escape hatch)
- Activity history and field change tracking (`--watch`, `--since`)
- Campaign list, get, schedule, trigger (dry-run by default, `--execute` to apply)
- Static list management — members, add, remove, membership check
- Company lookups by name or custom filter, schema discovery
- API usage and error stats
- Profile support — multiple Marketo instances, `.mrkto-profile` project files (like `.tool-versions`)
- Config at `~/.config/mrkto/` — no `.env` dependency
- `mrkto setup` — first-run experience (auth + Claude Code skill install)
- `mrkto help` — full command reference for agents
- `mrkto skill install` — Claude Code skill from package
- gh-style flags: `--fields`, `--limit`, `--json`, `--compact`, `--raw`
- One runtime dependency (`requests`)
