# mrkto

Public beta Marketo REST API CLI for humans and agents.

The CLI uses explicit resource names, structured JSON output, dry-run defaults for writes, and a raw `api` escape hatch for unsupported endpoints.

## Installation

Primary install path for macOS and Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/itodca/marketo-cli/main/install.sh | bash
```

The installer:

- downloads the matching binary from GitHub Releases
- installs it to `~/.local/bin` by default
- adds that directory to your shell `PATH` if needed

Useful options:

```bash
# Install a specific release tag
curl -fsSL https://raw.githubusercontent.com/itodca/marketo-cli/main/install.sh | bash -s -- --version v0.1.0

# Install somewhere else and leave PATH alone
curl -fsSL https://raw.githubusercontent.com/itodca/marketo-cli/main/install.sh | bash -s -- --install-dir "$HOME/bin" --no-modify-path
```

Other install options:

- download the release artifact directly from GitHub Releases
- install from PyPI with `pipx install marketo-cli`
- install the latest main branch with `pipx install git+https://github.com/itodca/marketo-cli.git`
- install from a checkout with `pip install .`

## Quick Start

```bash
# First-time setup
mrkto setup

# Or configure a profile explicitly
mrkto auth setup --profile production

# Check auth
mrkto auth check

# Look up a lead
mrkto lead list --email user@example.com

# Browse smart campaigns
mrkto smart-campaign list --active --limit 10

# Dry-run a campaign trigger
mrkto smart-campaign trigger 1234 --lead 1001 --lead 1002

# Raw API request
mrkto api get /v1/leads.json --query filterType=email --query filterValues=user@example.com
```

## Configuration

Credentials are stored in `~/.config/mrkto/`.

- Default profile: `~/.config/mrkto/config`
- Named profiles: `~/.config/mrkto/profiles/<name>`

Profile resolution order:

1. `--profile`
2. `MRKTO_PROFILE`
3. `.mrkto-profile` in the current directory tree
4. `default`

Environment variables override file-based config:

| Variable | Required | Description |
| --- | --- | --- |
| `MARKETO_MUNCHKIN_ID` | Yes | Marketo munchkin id |
| `MARKETO_CLIENT_ID` | Yes | LaunchPoint client id |
| `MARKETO_CLIENT_SECRET` | Yes | LaunchPoint client secret |
| `MARKETO_REST_URL` | No | Override REST endpoint |
| `MARKETO_IDENTITY_URL` | No | Override identity endpoint |

## Command Shape

The CLI uses singular resource nouns:

```text
mrkto auth setup
mrkto auth list
mrkto auth check

mrkto lead get
mrkto lead list
mrkto lead describe
mrkto lead static-lists
mrkto lead programs
mrkto lead smart-campaigns

mrkto activity list
mrkto activity types
mrkto activity changes

mrkto smart-campaign list
mrkto smart-campaign get
mrkto smart-campaign schedule
mrkto smart-campaign trigger

mrkto static-list list
mrkto static-list get
mrkto static-list members
mrkto static-list add
mrkto static-list remove
mrkto static-list check

mrkto smart-list list
mrkto smart-list get

mrkto company list
mrkto company describe

mrkto program list
mrkto program get

mrkto stats usage
mrkto stats errors

mrkto api get
mrkto api post
mrkto api delete
```

## Output

The default output is pretty JSON. Every command also supports:

- `--compact` for one JSON object per line
- `--raw` for single-line JSON of the full returned payload
- `--fields` to limit displayed fields on structured results

## Write Safety

Commands that modify data default to dry-run mode and require `--execute` to actually make changes:

- `mrkto smart-campaign schedule`
- `mrkto smart-campaign trigger`
- `mrkto static-list add`
- `mrkto static-list remove`

## Agent Skill

The repo still ships with a skills-based installer for supported coding agents:

```bash
mrkto skill install
mrkto skill install --scope project
```

## Release Automation

- CI runs tests, compile checks, and package builds on pushes and pull requests
- GitHub Releases build macOS and Linux binaries on `v*` tags
- PyPI publishing is prepared through GitHub Actions trusted publishing once the PyPI project is configured

## Source Contracts

The CLI is implemented against Adobe's published Marketo OpenAPI specs:

- [`swagger-mapi.json`](https://raw.githubusercontent.com/AdobeDocs/marketo-apis/main/static/swagger-mapi.json)
- [`swagger-asset.json`](https://raw.githubusercontent.com/AdobeDocs/marketo-apis/main/static/swagger-asset.json)
- [`swagger-identity.json`](https://raw.githubusercontent.com/AdobeDocs/marketo-apis/main/static/swagger-identity.json)
- [`swagger-user.json`](https://raw.githubusercontent.com/AdobeDocs/marketo-apis/main/static/swagger-user.json)

## Binary Releases

Build a release artifact locally with PyInstaller:

```bash
python3 -m pip install '.[build]'
./scripts/build-binary.sh
```

This creates assets under `dist/releases/` using the installer's expected naming scheme:

- `mrkto-darwin-arm64.tar.gz`
- `mrkto-darwin-x64.tar.gz`
- `mrkto-linux-arm64.tar.gz`
- `mrkto-linux-x64.tar.gz`

Each archive is accompanied by a `.sha256` checksum file.

## License

MIT
