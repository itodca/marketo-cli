# Changelog

## [0.1.3] — 2026-03-13

Binary performance release.

- Switch GitHub release artifacts from PyInstaller `--onefile` to `--onedir` bundles to avoid multi-second self-extraction on every run
- Update the installer to place the app bundle under `~/.local/share/mrkto` and symlink `mrkto` into `~/.local/bin`
- Add an uninstall path to the installer for removing the app bundle and command symlink
- Defer importing resource modules and the HTTP client until commands actually run, reducing startup work for lightweight commands like `--help`

## [0.1.2] — 2026-03-12

Documentation-focused release.

- Rewrite the README around purpose, profiles, command shape, and stdout-based file output
- Add deeper docs for command model, profiles, and common examples under `docs/`
- Keep the PyPI package page aligned with the improved public documentation

## [0.1.1] — 2026-03-12

Documentation-only follow-up release.

- Refresh the published PyPI project page so the install guidance points at `pipx install marketo-cli`
- Keep the GitHub README and PyPI long description aligned after the initial public release

## [0.1.0] — 2026-03-12

Initial public release.

- Typer-based CLI with singular resource nouns and explicit Marketo concepts like `smart-campaign`, `static-list`, and `smart-list`
- Lead lookups by email, id, or any filter field, plus static list, program, and smart campaign memberships
- Activity history and lead field change tracking with paging token support
- Smart campaign browse, get, schedule, and trigger commands with dry-run by default
- Static list browse, membership, add, remove, and membership check commands
- Company and program lookups, usage and error stats, and a raw `mrkto api` escape hatch
- Profile-scoped config and token caching under `~/.config/mrkto/`, plus `.mrkto-profile` project selection
- macOS/Linux binary installer, PyInstaller release packaging, and source distribution build support
- `mrkto skill install` for the tracked agent skill in this repo
