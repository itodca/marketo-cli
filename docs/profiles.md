# Profiles

`marketo-cli` stores profiles and token cache data for the `mrkto` command.

A profile is a saved Marketo connection.

Each profile stores the credentials and endpoints for one Marketo connection. In most teams, that usually means one Marketo instance or environment.

Most users only need the `default` profile. Create more profiles when you need to switch between multiple saved connections.

## When To Use Profiles

Profiles are useful when you work with:

- multiple Marketo instances inside the same company
- production and sandbox instances
- multiple clients
- multiple business units
- separate project directories that should automatically target different Marketo accounts

Example profile names:

- `default`
- `sandbox`
- `marketing-prod`
- `marketing-sandbox`
- `na`
- `emea`
- `acme`
- `globex`

The profile name is just a label you choose for the connection.

## Creating Profiles

Create the default profile:

```bash
mrkto auth setup
```

Create a named profile:

```bash
mrkto auth setup --profile sandbox
mrkto auth setup --profile acme
```

Check a profile:

```bash
mrkto auth check --profile sandbox
```

List configured profiles:

```bash
mrkto auth list
```

## Where Profiles Live

`mrkto` stores config under `~/.config/mrkto/`.

- default profile: `~/.config/mrkto/config`
- named profiles: `~/.config/mrkto/profiles/<name>`

Profile-specific token cache files live in the same config area so switching profiles does not reuse the wrong access token.

## How Profile Selection Works

Profile resolution order:

1. `--profile`
2. `MRKTO_PROFILE`
3. `.mrkto-profile` in the current directory tree
4. `default`

That means an explicit command-line profile always wins.

Examples:

```bash
mrkto lead list --email user@example.com --profile sandbox
MRKTO_PROFILE=acme mrkto stats usage
```

## Project-Level Profiles

If you work inside a project directory that should always use a specific profile, create a `.mrkto-profile` file:

```bash
echo "sandbox" > .mrkto-profile
```

Then commands inside that directory tree will automatically use `sandbox` unless you override it with `--profile`.

This is useful when:

- one repo should always target one Marketo instance inside a company
- one repo targets a sandbox instance
- another repo targets a production instance
- you want agents running in that directory to inherit the right Marketo context automatically

## Environment Variable Overrides

Environment variables can override file-based config:

- `MARKETO_MUNCHKIN_ID`
- `MARKETO_CLIENT_ID`
- `MARKETO_CLIENT_SECRET`
- `MARKETO_REST_URL`
- `MARKETO_IDENTITY_URL`

That is useful for CI or ephemeral automation, but for normal day-to-day use profiles are usually the cleaner interface.
