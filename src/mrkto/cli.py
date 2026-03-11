"""mrkto CLI — entry point and subcommand routing."""

import argparse
import sys

from mrkto.config import load_config
from mrkto.client import MarketoClient, MarketoAPIError
from mrkto.output import print_result, print_error


def add_global_flags(parser):
    parser.add_argument("--json", dest="fmt", action="store_const", const="json", help="JSON output (default)")
    parser.add_argument("--compact", dest="fmt", action="store_const", const="compact", help="One-line-per-record")
    parser.add_argument("--raw", dest="fmt", action="store_const", const="raw", help="Raw API response")
    parser.add_argument("--fields", help="Comma-separated field list")


def main():
    parser = argparse.ArgumentParser(prog="mrkto", description="Marketo REST API CLI")
    sub = parser.add_subparsers(dest="resource")

    # --- auth ---
    auth_p = sub.add_parser("auth", help="Authentication")
    auth_sub = auth_p.add_subparsers(dest="action")
    auth_check = auth_sub.add_parser("check", help="Verify credentials")
    add_global_flags(auth_check)

    # --- lead ---
    lead_p = sub.add_parser("lead", help="Lead operations")
    lead_sub = lead_p.add_subparsers(dest="action")

    lead_get = lead_sub.add_parser("get", help="Get lead by Marketo ID")
    lead_get.add_argument("lead_id", type=int, help="Marketo lead ID")
    add_global_flags(lead_get)

    lead_search = lead_sub.add_parser("search", help="Search leads by email")
    lead_search.add_argument("query", help="Email address")
    add_global_flags(lead_search)

    lead_describe = lead_sub.add_parser("describe", help="Show lead field schema")
    add_global_flags(lead_describe)

    lead_lists = lead_sub.add_parser("lists", help="Lists a lead belongs to")
    lead_lists.add_argument("lead_id", type=int, help="Marketo lead ID")
    add_global_flags(lead_lists)

    lead_programs = lead_sub.add_parser("programs", help="Programs a lead is in")
    lead_programs.add_argument("lead_id", type=int, help="Marketo lead ID")
    add_global_flags(lead_programs)

    lead_campaigns = lead_sub.add_parser("campaigns", help="Smart campaigns for a lead")
    lead_campaigns.add_argument("lead_id", type=int, help="Marketo lead ID")
    add_global_flags(lead_campaigns)

    # --- activity ---
    activity_p = sub.add_parser("activity", help="Activity operations")
    activity_sub = activity_p.add_subparsers(dest="action")

    activity_types = activity_sub.add_parser("types", help="List activity types")
    add_global_flags(activity_types)

    activity_get = activity_sub.add_parser("get", help="Get activities for a lead")
    activity_get.add_argument("lead_id", type=int, help="Marketo lead ID")
    activity_get.add_argument("--types", help="Comma-separated activity type IDs")
    activity_get.add_argument("--since", type=int, default=30, help="Days back (default: 30)")
    add_global_flags(activity_get)

    activity_changes = activity_sub.add_parser("changes", help="Get lead field changes")
    activity_changes.add_argument("--watch", dest="change_fields", required=True, help="Comma-separated field names to watch")
    activity_changes.add_argument("--since", type=int, default=30, help="Days back (default: 30)")
    add_global_flags(activity_changes)

    # --- campaign ---
    campaign_p = sub.add_parser("campaign", help="Campaign operations")
    campaign_sub = campaign_p.add_subparsers(dest="action")

    campaign_list = campaign_sub.add_parser("list", help="List campaigns")
    campaign_list.add_argument("--name", help="Filter by name")
    campaign_list.add_argument("--program", help="Filter by program name")
    add_global_flags(campaign_list)

    campaign_get = campaign_sub.add_parser("get", help="Get campaign by ID")
    campaign_get.add_argument("campaign_id", type=int, help="Campaign ID")
    add_global_flags(campaign_get)

    campaign_schedule = campaign_sub.add_parser("schedule", help="Schedule a batch campaign")
    campaign_schedule.add_argument("campaign_id", type=int, help="Campaign ID")
    campaign_schedule.add_argument("--run-at", help="ISO datetime to run at")
    campaign_schedule.add_argument("--execute", action="store_true", help="Actually schedule (default: dry-run)")
    add_global_flags(campaign_schedule)

    campaign_trigger = campaign_sub.add_parser("trigger", help="Trigger campaign for leads")
    campaign_trigger.add_argument("campaign_id", type=int, help="Campaign ID")
    campaign_trigger.add_argument("--leads", required=True, help="Comma-separated lead IDs")
    campaign_trigger.add_argument("--execute", action="store_true", help="Actually trigger (default: dry-run)")
    add_global_flags(campaign_trigger)

    # --- list ---
    list_p = sub.add_parser("list", help="Static list operations")
    list_sub = list_p.add_subparsers(dest="action")

    list_list = list_sub.add_parser("list", help="List all static lists")
    list_list.add_argument("--name", help="Filter by name")
    add_global_flags(list_list)

    list_get = list_sub.add_parser("get", help="Get list by ID")
    list_get.add_argument("list_id", type=int, help="List ID")
    add_global_flags(list_get)

    list_members = list_sub.add_parser("members", help="Get list members")
    list_members.add_argument("list_id", type=int, help="List ID")
    add_global_flags(list_members)

    list_add = list_sub.add_parser("add", help="Add leads to list")
    list_add.add_argument("list_id", type=int, help="List ID")
    list_add.add_argument("--leads", required=True, help="Comma-separated lead IDs")
    list_add.add_argument("--execute", action="store_true", help="Actually add (default: dry-run)")
    add_global_flags(list_add)

    list_remove = list_sub.add_parser("remove", help="Remove leads from list")
    list_remove.add_argument("list_id", type=int, help="List ID")
    list_remove.add_argument("--leads", required=True, help="Comma-separated lead IDs")
    list_remove.add_argument("--execute", action="store_true", help="Actually remove (default: dry-run)")
    add_global_flags(list_remove)

    list_check = list_sub.add_parser("check", help="Check list membership")
    list_check.add_argument("list_id", type=int, help="List ID")
    list_check.add_argument("--leads", required=True, help="Comma-separated lead IDs")
    add_global_flags(list_check)

    # --- company ---
    company_p = sub.add_parser("company", help="Company operations")
    company_sub = company_p.add_subparsers(dest="action")

    company_get = company_sub.add_parser("get", help="Get company by name")
    company_get.add_argument("name", help="Company name")
    add_global_flags(company_get)

    company_describe = company_sub.add_parser("describe", help="Company field schema")
    add_global_flags(company_describe)

    # --- stats ---
    stats_p = sub.add_parser("stats", help="API usage and errors")
    stats_sub = stats_p.add_subparsers(dest="action")

    stats_usage = stats_sub.add_parser("usage", help="API usage stats")
    stats_usage.add_argument("--weekly", action="store_true", help="Last 7 days")
    add_global_flags(stats_usage)

    stats_errors = stats_sub.add_parser("errors", help="API error stats")
    stats_errors.add_argument("--weekly", action="store_true", help="Last 7 days")
    add_global_flags(stats_errors)

    # --- parse ---
    args = parser.parse_args()

    if args.resource is None:
        parser.print_help()
        sys.exit(1)

    if getattr(args, "action", None) is None:
        # Print help for the resource if no action given
        sub.choices[args.resource].print_help()
        sys.exit(1)

    fmt = getattr(args, "fmt", None) or "json"
    fields = getattr(args, "fields", None)
    field_list = fields.split(",") if fields else None

    try:
        result = dispatch(args)
        if result is not None:
            print_result(result, fmt=fmt, fields=field_list)
    except MarketoAPIError as e:
        print_error(f"[{e.code}] {e.message}")
        sys.exit(1)
    except KeyboardInterrupt:
        sys.exit(130)


def dispatch(args):
    """Route to the appropriate resource function."""
    config = load_config()
    client = MarketoClient(config)

    if args.resource == "auth":
        from mrkto.resources.auth import check_auth
        return check_auth(client)

    elif args.resource == "lead":
        from mrkto.resources.lead import (
            get_lead, describe_lead,
            get_lead_lists, get_lead_programs, get_lead_campaigns,
        )
        if args.action == "get":
            return get_lead(client, lead_id=args.lead_id)
        elif args.action == "search":
            return get_lead(client, email=args.query)
        elif args.action == "describe":
            return describe_lead(client)
        elif args.action == "lists":
            return get_lead_lists(client, args.lead_id)
        elif args.action == "programs":
            return get_lead_programs(client, args.lead_id)
        elif args.action == "campaigns":
            return get_lead_campaigns(client, args.lead_id)

    elif args.resource == "activity":
        from mrkto.resources.activity import (
            get_activity_types, get_lead_activities, get_lead_changes,
        )
        if args.action == "types":
            return get_activity_types(client)
        elif args.action == "get":
            type_ids = args.types if args.types else None
            return get_lead_activities(client, args.lead_id, type_ids, args.since)
        elif args.action == "changes":
            return get_lead_changes(client, args.change_fields, args.since)

    elif args.resource == "campaign":
        from mrkto.resources.campaign import (
            list_campaigns, get_campaign, schedule_campaign, trigger_campaign,
        )
        if args.action == "list":
            return list_campaigns(client, name=args.name, program_name=args.program)
        elif args.action == "get":
            return get_campaign(client, args.campaign_id)
        elif args.action == "schedule":
            return schedule_campaign(client, args.campaign_id, args.run_at, dry_run=not args.execute)
        elif args.action == "trigger":
            lead_ids = [int(x) for x in args.leads.split(",")]
            return trigger_campaign(client, args.campaign_id, lead_ids, dry_run=not args.execute)

    elif args.resource == "list":
        from mrkto.resources.list import (
            list_lists, get_list, get_list_members,
            add_to_list, remove_from_list, is_member,
        )
        if args.action == "list":
            return list_lists(client, name=args.name)
        elif args.action == "get":
            return get_list(client, args.list_id)
        elif args.action == "members":
            return get_list_members(client, args.list_id)
        elif args.action == "add":
            lead_ids = [int(x) for x in args.leads.split(",")]
            return add_to_list(client, args.list_id, lead_ids, dry_run=not args.execute)
        elif args.action == "remove":
            lead_ids = [int(x) for x in args.leads.split(",")]
            return remove_from_list(client, args.list_id, lead_ids, dry_run=not args.execute)
        elif args.action == "check":
            lead_ids = [int(x) for x in args.leads.split(",")]
            return is_member(client, args.list_id, lead_ids)

    elif args.resource == "company":
        from mrkto.resources.company import get_companies, describe_company
        if args.action == "get":
            return get_companies(client, filter_values=args.name)
        elif args.action == "describe":
            return describe_company(client)

    elif args.resource == "stats":
        from mrkto.resources.stats import get_usage, get_errors
        if args.action == "usage":
            return get_usage(client, weekly=args.weekly)
        elif args.action == "errors":
            return get_errors(client, weekly=args.weekly)

    print_error("Unknown command. Run 'mrkto --help'.")
    sys.exit(1)
