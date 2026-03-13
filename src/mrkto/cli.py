"""Typer-based CLI for the Marketo REST API."""

from __future__ import annotations

import json
import sys
from enum import Enum
from importlib import import_module
from pathlib import Path
from typing import Annotated, Any, Callable, TYPE_CHECKING

import typer

from mrkto.output import print_error, print_result

if TYPE_CHECKING:
    from mrkto.client import MarketoClient

app = typer.Typer(
    help="Marketo REST API CLI.",
    no_args_is_help=True,
    add_completion=False,
    pretty_exceptions_enable=False,
)

auth_app = typer.Typer(help="Authentication and profile management.", no_args_is_help=True)
lead_app = typer.Typer(help="Lead lookups and memberships.", no_args_is_help=True)
activity_app = typer.Typer(help="Activity lookups and lead changes.", no_args_is_help=True)
smart_campaign_app = typer.Typer(help="Smart campaign browsing and execution.", no_args_is_help=True)
static_list_app = typer.Typer(help="Static list lookups and membership changes.", no_args_is_help=True)
smart_list_app = typer.Typer(help="Smart list lookups.", no_args_is_help=True)
company_app = typer.Typer(help="Company lookups.", no_args_is_help=True)
program_app = typer.Typer(help="Program lookups.", no_args_is_help=True)
stats_app = typer.Typer(help="Usage and error stats.", no_args_is_help=True)
api_app = typer.Typer(help="Raw API escape hatch.", no_args_is_help=True)
skill_app = typer.Typer(help="Agent skill installation.", no_args_is_help=True)

app.add_typer(auth_app, name="auth")
app.add_typer(lead_app, name="lead")
app.add_typer(activity_app, name="activity")
app.add_typer(smart_campaign_app, name="smart-campaign")
app.add_typer(static_list_app, name="static-list")
app.add_typer(smart_list_app, name="smart-list")
app.add_typer(company_app, name="company")
app.add_typer(program_app, name="program")
app.add_typer(stats_app, name="stats")
app.add_typer(api_app, name="api")
app.add_typer(skill_app, name="skill")


class _LazyModule:
    def __init__(self, module_name: str):
        self._module_name = module_name
        self._module: Any | None = None

    def _load(self) -> Any:
        if self._module is None:
            self._module = import_module(self._module_name)
        return self._module

    def __getattr__(self, name: str) -> Any:
        return getattr(self._load(), name)


client_module = _LazyModule("mrkto.client")
config_module = _LazyModule("mrkto.config")
auth = _LazyModule("mrkto.resources.auth")
lead_resource = _LazyModule("mrkto.resources.lead")
activity = _LazyModule("mrkto.resources.activity")
smart_campaign = _LazyModule("mrkto.resources.smart_campaign")
static_list = _LazyModule("mrkto.resources.static_list")
smart_list = _LazyModule("mrkto.resources.smart_list")
company = _LazyModule("mrkto.resources.company")
program = _LazyModule("mrkto.resources.program")
stats = _LazyModule("mrkto.resources.stats")
api = _LazyModule("mrkto.resources.api")
skill = _LazyModule("mrkto.resources.skill")


class OutputFormat(str, Enum):
    json = "json"
    compact = "compact"
    raw = "raw"


ProfileOption = Annotated[str | None, typer.Option("--profile", help="Named profile to use.")]
FieldsOption = Annotated[str | None, typer.Option("--fields", help="Comma-separated fields to return or display.")]
LimitOption = Annotated[int | None, typer.Option("--limit", min=1, help="Maximum number of records to return.")]
JsonFlag = Annotated[bool, typer.Option("--json", help="Pretty JSON output (default).")]
CompactFlag = Annotated[bool, typer.Option("--compact", help="One JSON object per line.")]
RawFlag = Annotated[bool, typer.Option("--raw", help="Single-line JSON output for the full returned payload.")]
ExecuteFlag = Annotated[bool, typer.Option("--execute", help="Actually perform the write operation.")]


def get_client(profile: str | None) -> MarketoClient:
    return client_module.MarketoClient(config_module.load_config(profile=profile))


def parse_fields(fields: str | None) -> list[str] | None:
    if not fields:
        return None
    result = [field.strip() for field in fields.split(",") if field.strip()]
    return result or None


def resolve_output_format(json_output: bool, compact_output: bool, raw_output: bool) -> OutputFormat:
    selected = [name for name, enabled in [("json", json_output), ("compact", compact_output), ("raw", raw_output)] if enabled]
    if len(selected) > 1:
        raise ValueError("Choose only one of --json, --compact, or --raw")
    if compact_output:
        return OutputFormat.compact
    if raw_output:
        return OutputFormat.raw
    return OutputFormat.json


def output_payload(data: Any, fmt: OutputFormat) -> Any:
    if fmt == OutputFormat.raw:
        return data
    if isinstance(data, dict) and "success" in data and "result" in data:
        return data["result"]
    return data


def parse_kv_pairs(values: list[str]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for value in values:
        key, separator, raw = value.partition("=")
        if not separator:
            raise ValueError(f"Invalid key=value pair: {value}")
        key = key.strip()
        raw = raw.strip()
        if key in result:
            existing = result[key]
            if isinstance(existing, list):
                existing.append(raw)
            else:
                result[key] = [existing, raw]
        else:
            result[key] = raw
    return result


def load_json_input(input_path: str | None) -> dict[str, Any] | None:
    if not input_path:
        return None
    raw = sys.stdin.read() if input_path == "-" else Path(input_path).read_text()
    data = json.loads(raw)
    if not isinstance(data, dict):
        raise ValueError("JSON input must be an object")
    return data


def prompt_if_missing(value: str | None, prompt: str, *, hide_input: bool = False) -> str:
    if value:
        return value
    return typer.prompt(prompt, hide_input=hide_input).strip()


def run_action(
    action: Callable[[], Any],
    *,
    fields: str | None,
    json_output: bool,
    compact_output: bool,
    raw_output: bool,
) -> None:
    marketo_api_error = client_module.MarketoAPIError
    try:
        fmt = resolve_output_format(json_output, compact_output, raw_output)
        result = action()
    except FileExistsError as exc:
        print_error(f"Config already exists at {exc.args[0]}. Re-run with --overwrite to replace it.")
        raise typer.Exit(1) from exc
    except (ValueError, KeyError, json.JSONDecodeError) as exc:
        print_error(str(exc))
        raise typer.Exit(1) from exc
    except marketo_api_error as exc:
        print_error(f"[{exc.code}] {exc.message}")
        raise typer.Exit(1) from exc

    if result is not None:
        print_result(output_payload(result, fmt), fmt=fmt.value, fields=parse_fields(fields))


@app.command("setup")
def setup_command(
    profile: ProfileOption = None,
    munchkin_id: Annotated[str | None, typer.Option("--munchkin-id", help="Marketo munchkin id.")] = None,
    client_id: Annotated[str | None, typer.Option("--client-id", help="LaunchPoint client id.")] = None,
    client_secret: Annotated[str | None, typer.Option("--client-secret", help="LaunchPoint client secret.")] = None,
    overwrite: Annotated[bool, typer.Option("--overwrite", help="Overwrite the target profile if it already exists.")] = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    """Interactive alias for auth setup."""

    def action() -> dict:
        return auth.write_auth_config(
            profile=profile,
            munchkin_id=prompt_if_missing(munchkin_id, "Munchkin ID"),
            client_id=prompt_if_missing(client_id, "Client ID"),
            client_secret=prompt_if_missing(client_secret, "Client Secret", hide_input=True),
            overwrite=overwrite,
        )

    run_action(action, fields=None, json_output=json_output, compact_output=compact_output, raw_output=raw_output)


@auth_app.command("setup")
def auth_setup(
    profile: ProfileOption = None,
    munchkin_id: Annotated[str | None, typer.Option("--munchkin-id", help="Marketo munchkin id.")] = None,
    client_id: Annotated[str | None, typer.Option("--client-id", help="LaunchPoint client id.")] = None,
    client_secret: Annotated[str | None, typer.Option("--client-secret", help="LaunchPoint client secret.")] = None,
    overwrite: Annotated[bool, typer.Option("--overwrite", help="Overwrite the target profile if it already exists.")] = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    """Write credentials for a profile."""

    def action() -> dict:
        return auth.write_auth_config(
            profile=profile,
            munchkin_id=prompt_if_missing(munchkin_id, "Munchkin ID"),
            client_id=prompt_if_missing(client_id, "Client ID"),
            client_secret=prompt_if_missing(client_secret, "Client Secret", hide_input=True),
            overwrite=overwrite,
        )

    run_action(action, fields=None, json_output=json_output, compact_output=compact_output, raw_output=raw_output)


@auth_app.command("list")
def auth_list(
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(auth.list_auth, fields=None, json_output=json_output, compact_output=compact_output, raw_output=raw_output)


@auth_app.command("check")
def auth_check(
    profile: ProfileOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: auth.check_auth(get_client(profile)),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@lead_app.command("get")
def lead_get(
    lead_id: Annotated[int, typer.Argument(help="Marketo lead id.")],
    profile: ProfileOption = None,
    fields: FieldsOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: lead_resource.get_lead(get_client(profile), lead_id=lead_id, fields=fields),
        fields=fields,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@lead_app.command("list")
def lead_list(
    email: Annotated[str | None, typer.Option("--email", help="Filter by email address.")] = None,
    lead_ids: Annotated[str | None, typer.Option("--id", help="Comma-separated Marketo lead ids.")] = None,
    filter_value: Annotated[str | None, typer.Option("--filter", help="Custom filter as key=value.")] = None,
    profile: ProfileOption = None,
    fields: FieldsOption = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    def action() -> dict:
        if email:
            return lead_resource.list_leads(get_client(profile), filter_type="email", filter_values=email, fields=fields, limit=limit)
        if lead_ids:
            return lead_resource.list_leads(get_client(profile), filter_type="id", filter_values=lead_ids, fields=fields, limit=limit)
        if filter_value:
            filter_type, separator, raw = filter_value.partition("=")
            if not separator:
                raise ValueError("--filter must use key=value form")
            return lead_resource.list_leads(get_client(profile), filter_type=filter_type, filter_values=raw, fields=fields, limit=limit)
        raise ValueError("Provide one of --email, --id, or --filter")

    run_action(action, fields=fields, json_output=json_output, compact_output=compact_output, raw_output=raw_output)


@lead_app.command("describe")
def lead_describe(
    profile: ProfileOption = None,
    legacy: Annotated[bool, typer.Option("--legacy", help="Use the older describe endpoint instead of describe2.")] = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: lead_resource.describe_lead(get_client(profile), detailed=not legacy),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@lead_app.command("static-lists")
def lead_static_lists(
    lead_id: Annotated[int, typer.Argument(help="Marketo lead id.")],
    profile: ProfileOption = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: lead_resource.get_lead_static_lists(get_client(profile), lead_id=lead_id, limit=limit),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@lead_app.command("programs")
def lead_programs(
    lead_id: Annotated[int, typer.Argument(help="Marketo lead id.")],
    profile: ProfileOption = None,
    program_id: Annotated[list[int] | None, typer.Option("--program-id", help="Filter to specific program ids.")] = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: lead_resource.get_lead_programs(
            get_client(profile),
            lead_id=lead_id,
            limit=limit,
            program_ids=program_id or None,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@lead_app.command("smart-campaigns")
def lead_smart_campaigns(
    lead_id: Annotated[int, typer.Argument(help="Marketo lead id.")],
    profile: ProfileOption = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: lead_resource.get_lead_smart_campaigns(get_client(profile), lead_id=lead_id, limit=limit),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@activity_app.command("types")
def activity_types(
    profile: ProfileOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: activity.get_activity_types(get_client(profile)),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@activity_app.command("list")
def activity_list(
    lead_id: Annotated[int, typer.Argument(help="Marketo lead id.")],
    profile: ProfileOption = None,
    activity_type_id: Annotated[list[int] | None, typer.Option("--type-id", help="Filter to activity type ids.")] = None,
    since: Annotated[int, typer.Option("--since", min=0, help="Days back from now.")] = 30,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: activity.list_activities(
            get_client(profile),
            lead_id=lead_id,
            activity_type_ids=activity_type_id or None,
            since_days=since,
            limit=limit,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@activity_app.command("changes")
def activity_changes(
    watch: Annotated[list[str], typer.Option("--watch", help="Lead field name to watch. Repeat the option for multiple fields.")],
    profile: ProfileOption = None,
    lead_id: Annotated[list[int] | None, typer.Option("--lead-id", help="Filter to specific lead ids.")] = None,
    list_id: Annotated[int | None, typer.Option("--list-id", help="Filter to a static list id.")] = None,
    since: Annotated[int, typer.Option("--since", min=0, help="Days back from now.")] = 30,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: activity.get_lead_changes(
            get_client(profile),
            fields=watch,
            since_days=since,
            lead_ids=lead_id or None,
            list_id=list_id,
            limit=limit,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@smart_campaign_app.command("list")
def smart_campaign_list(
    name: Annotated[str | None, typer.Option("--name", help="Lookup by exact smart campaign name.")] = None,
    profile: ProfileOption = None,
    folder_id: Annotated[int | None, typer.Option("--folder-id", help="Parent folder or program id.")] = None,
    folder_type: Annotated[str | None, typer.Option("--folder-type", help="Folder type: Folder or Program.")] = None,
    active: Annotated[bool | None, typer.Option("--active/--all", help="Only active smart campaigns, or all if omitted.")] = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: smart_campaign.list_smart_campaigns(
            get_client(profile),
            name=name,
            folder_id=folder_id,
            folder_type=folder_type,
            is_active=active,
            limit=limit,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@smart_campaign_app.command("get")
def smart_campaign_get(
    campaign_id: Annotated[int, typer.Argument(help="Smart campaign id.")],
    profile: ProfileOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: smart_campaign.get_smart_campaign(get_client(profile), campaign_id=campaign_id),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@smart_campaign_app.command("schedule")
def smart_campaign_schedule(
    campaign_id: Annotated[int, typer.Argument(help="Smart campaign id.")],
    profile: ProfileOption = None,
    run_at: Annotated[str | None, typer.Option("--run-at", help="ISO-8601 datetime to schedule.")] = None,
    execute: ExecuteFlag = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: smart_campaign.schedule_smart_campaign(
            get_client(profile),
            campaign_id=campaign_id,
            run_at=run_at,
            dry_run=not execute,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@smart_campaign_app.command("trigger")
def smart_campaign_trigger(
    campaign_id: Annotated[int, typer.Argument(help="Smart campaign id.")],
    lead: Annotated[list[int], typer.Option("--lead", help="Lead id to pass to the campaign. Repeat the option for multiple leads.")],
    profile: ProfileOption = None,
    execute: ExecuteFlag = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: smart_campaign.trigger_smart_campaign(
            get_client(profile),
            campaign_id=campaign_id,
            lead_ids=lead,
            dry_run=not execute,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@static_list_app.command("list")
def static_list_list(
    name: Annotated[str | None, typer.Option("--name", help="Filter by exact static list name.")] = None,
    program_name: Annotated[str | None, typer.Option("--program", help="Filter by program name.")] = None,
    workspace_name: Annotated[str | None, typer.Option("--workspace", help="Filter by workspace name.")] = None,
    profile: ProfileOption = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: static_list.list_static_lists(
            get_client(profile),
            name=name,
            program_name=program_name,
            workspace_name=workspace_name,
            limit=limit,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@static_list_app.command("get")
def static_list_get(
    list_id: Annotated[int, typer.Argument(help="Static list id.")],
    profile: ProfileOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: static_list.get_static_list(get_client(profile), list_id=list_id),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@static_list_app.command("members")
def static_list_members(
    list_id: Annotated[int, typer.Argument(help="Static list id.")],
    profile: ProfileOption = None,
    fields: FieldsOption = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: static_list.get_static_list_members(get_client(profile), list_id=list_id, fields=fields, limit=limit),
        fields=fields,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@static_list_app.command("add")
def static_list_add(
    list_id: Annotated[int, typer.Argument(help="Static list id.")],
    lead: Annotated[list[int], typer.Option("--lead", help="Lead id to add. Repeat the option for multiple leads.")],
    profile: ProfileOption = None,
    execute: ExecuteFlag = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: static_list.add_to_static_list(
            get_client(profile),
            list_id=list_id,
            lead_ids=lead,
            dry_run=not execute,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@static_list_app.command("remove")
def static_list_remove(
    list_id: Annotated[int, typer.Argument(help="Static list id.")],
    lead: Annotated[list[int], typer.Option("--lead", help="Lead id to remove. Repeat the option for multiple leads.")],
    profile: ProfileOption = None,
    execute: ExecuteFlag = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: static_list.remove_from_static_list(
            get_client(profile),
            list_id=list_id,
            lead_ids=lead,
            dry_run=not execute,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@static_list_app.command("check")
def static_list_check(
    list_id: Annotated[int, typer.Argument(help="Static list id.")],
    lead: Annotated[list[int], typer.Option("--lead", help="Lead id to check. Repeat the option for multiple leads.")],
    profile: ProfileOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: static_list.check_static_list_membership(get_client(profile), list_id=list_id, lead_ids=lead),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@smart_list_app.command("list")
def smart_list_list(
    name: Annotated[str | None, typer.Option("--name", help="Lookup by exact smart list name.")] = None,
    profile: ProfileOption = None,
    folder_id: Annotated[int | None, typer.Option("--folder-id", help="Parent folder or program id.")] = None,
    folder_type: Annotated[str | None, typer.Option("--folder-type", help="Folder type: Folder or Program.")] = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: smart_list.list_smart_lists(
            get_client(profile),
            name=name,
            folder_id=folder_id,
            folder_type=folder_type,
            limit=limit,
        ),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@smart_list_app.command("get")
def smart_list_get(
    list_id: Annotated[int, typer.Argument(help="Smart list id.")],
    profile: ProfileOption = None,
    include_rules: Annotated[bool, typer.Option("--include-rules", help="Include smart list rules in the response.")] = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: smart_list.get_smart_list(get_client(profile), list_id=list_id, include_rules=include_rules),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@company_app.command("list")
def company_list(
    name: Annotated[str | None, typer.Option("--name", help="Filter by company name.")] = None,
    filter_value: Annotated[str | None, typer.Option("--filter", help="Custom filter as key=value.")] = None,
    profile: ProfileOption = None,
    fields: FieldsOption = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    def action() -> dict:
        if name:
            return company.list_companies(get_client(profile), filter_type="company", filter_values=name, fields=fields, limit=limit)
        if filter_value:
            filter_type, separator, raw = filter_value.partition("=")
            if not separator:
                raise ValueError("--filter must use key=value form")
            return company.list_companies(get_client(profile), filter_type=filter_type, filter_values=raw, fields=fields, limit=limit)
        raise ValueError("Provide one of --name or --filter")

    run_action(action, fields=fields, json_output=json_output, compact_output=compact_output, raw_output=raw_output)


@company_app.command("describe")
def company_describe(
    profile: ProfileOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: company.describe_company(get_client(profile)),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@program_app.command("list")
def program_list(
    name: Annotated[str | None, typer.Option("--name", help="Lookup by exact program name.")] = None,
    profile: ProfileOption = None,
    limit: LimitOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: program.list_programs(get_client(profile), name=name, limit=limit),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@program_app.command("get")
def program_get(
    program_id: Annotated[int, typer.Argument(help="Program id.")],
    profile: ProfileOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: program.get_program(get_client(profile), program_id=program_id),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@stats_app.command("usage")
def stats_usage(
    profile: ProfileOption = None,
    weekly: Annotated[bool, typer.Option("--weekly", help="Return the last 7 days instead of the current period.")] = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: stats.get_usage(get_client(profile), weekly=weekly),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@stats_app.command("errors")
def stats_errors(
    profile: ProfileOption = None,
    weekly: Annotated[bool, typer.Option("--weekly", help="Return the last 7 days instead of the current period.")] = False,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: stats.get_errors(get_client(profile), weekly=weekly),
        fields=None,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@api_app.command("get")
def api_get(
    path: Annotated[str, typer.Argument(help="Path under /rest, for example /v1/leads.json.")],
    profile: ProfileOption = None,
    query: Annotated[list[str], typer.Option("--query", help="Query parameter in key=value form.")] = [],
    fields: FieldsOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    run_action(
        lambda: api.api_get(get_client(profile), path=path, query=parse_kv_pairs(query) or None),
        fields=fields,
        json_output=json_output,
        compact_output=compact_output,
        raw_output=raw_output,
    )


@api_app.command("post")
def api_post(
    path: Annotated[str, typer.Argument(help="Path under /rest, for example /v1/leads/push.json.")],
    profile: ProfileOption = None,
    query: Annotated[list[str], typer.Option("--query", help="Query parameter in key=value form.")] = [],
    body: Annotated[list[str], typer.Option("--body", help="JSON body field in key=value form.")] = [],
    input_path: Annotated[str | None, typer.Option("--input", help="Path to a JSON object to send as the request body, or - for stdin.")] = None,
    fields: FieldsOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    def action() -> dict:
        parsed_body = parse_kv_pairs(body)
        file_body = load_json_input(input_path)
        if parsed_body and file_body:
            raise ValueError("Use either --body or --input, not both")
        return api.api_post(
            get_client(profile),
            path=path,
            query=parse_kv_pairs(query) or None,
            body=file_body or parsed_body or None,
        )

    run_action(action, fields=fields, json_output=json_output, compact_output=compact_output, raw_output=raw_output)


@api_app.command("delete")
def api_delete(
    path: Annotated[str, typer.Argument(help="Path under /rest, for example /v1/lists/123/leads.json.")],
    profile: ProfileOption = None,
    query: Annotated[list[str], typer.Option("--query", help="Query parameter in key=value form.")] = [],
    body: Annotated[list[str], typer.Option("--body", help="JSON body field in key=value form.")] = [],
    input_path: Annotated[str | None, typer.Option("--input", help="Path to a JSON object to send as the request body, or - for stdin.")] = None,
    fields: FieldsOption = None,
    json_output: JsonFlag = False,
    compact_output: CompactFlag = False,
    raw_output: RawFlag = False,
) -> None:
    def action() -> dict:
        parsed_body = parse_kv_pairs(body)
        file_body = load_json_input(input_path)
        if parsed_body and file_body:
            raise ValueError("Use either --body or --input, not both")
        return api.api_delete(
            get_client(profile),
            path=path,
            query=parse_kv_pairs(query) or None,
            body=file_body or parsed_body or None,
        )

    run_action(action, fields=fields, json_output=json_output, compact_output=compact_output, raw_output=raw_output)


@skill_app.command("install")
def skill_install(
    scope: Annotated[str, typer.Option("--scope", help="Install scope.", case_sensitive=False)] = "user",
) -> None:
    if scope not in {"user", "project"}:
        print_error("--scope must be one of: user, project")
        raise typer.Exit(1)
    try:
        skill.install_skill(scope=scope)
    except SystemExit as exc:
        raise typer.Exit(exc.code) from exc


def main() -> None:
    try:
        app()
    except KeyboardInterrupt as exc:
        raise typer.Exit(130) from exc


if __name__ == "__main__":
    main()
