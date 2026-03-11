"""Full command reference for agent and human consumption."""

HELP_TEXT = """\
mrkto — Marketo REST API CLI

SETUP:
  mrkto setup                             First-time setup (auth + skill install)
  mrkto setup --profile <name>            Setup a named profile
  mrkto auth setup                        Configure default credentials
  mrkto auth setup --profile <name>       Configure a named profile
  mrkto auth list                         List configured profiles
  mrkto auth check                        Verify credentials work

LEADS:
  mrkto lead get <id>                     Get lead by Marketo ID
  mrkto lead list --email <addr>          List leads by email
  mrkto lead list --id <ids>              List leads by ID(s), comma-separated
  mrkto lead list --filter <key=value>    List leads by any filter field
  mrkto lead describe                     Show lead field schema
  mrkto lead lists <id>                   Static lists a lead belongs to
  mrkto lead programs <id>                Programs a lead is in
  mrkto lead campaigns <id>               Smart campaigns for a lead

ACTIVITIES:
  mrkto activity types                    List all activity types
  mrkto activity get <lead_id>            Get activities for a lead
    --types <ids>                           Filter by activity type IDs
    --since <days>                          Days back (default: 30)
  mrkto activity changes --watch <fields> Get lead field changes
    --since <days>                          Days back (default: 30)

CAMPAIGNS:
  mrkto campaign list                     List campaigns
    --name <name>                           Filter by name
    --program <name>                        Filter by program name
  mrkto campaign get <id>                 Get campaign by ID
  mrkto campaign schedule <id>            Schedule a batch campaign
    --run-at <datetime>                     ISO datetime to run at
    --execute                               Actually schedule (default: dry-run)
  mrkto campaign trigger <id>             Trigger campaign for leads
    --leads <ids>                           Comma-separated lead IDs (required)
    --execute                               Actually trigger (default: dry-run)

STATIC LISTS:
  mrkto list list                         List all static lists
    --name <name>                           Filter by name
  mrkto list get <id>                     Get list by ID
  mrkto list members <id>                 Get list members
  mrkto list add <id> --leads <ids>       Add leads to list
    --execute                               Actually add (default: dry-run)
  mrkto list remove <id> --leads <ids>    Remove leads from list
    --execute                               Actually remove (default: dry-run)
  mrkto list check <id> --leads <ids>     Check list membership

COMPANIES:
  mrkto company list --name <name>        List companies by name
  mrkto company list --filter <key=value> List companies by any filter field
  mrkto company describe                  Company field schema

STATS:
  mrkto stats usage                       API usage stats
  mrkto stats errors                      API error stats
    --weekly                                Last 7 days

PROFILES:
  Profile resolution: --profile flag > MRKTO_PROFILE env var > .mrkto-profile file > default
  Drop a .mrkto-profile file in your project directory to auto-select a profile.

GLOBAL FLAGS (available on all commands):
  --profile <name>        Use a named profile
  --fields <f1,f2,...>    Limit response fields
  --json                  JSON output (default)
  --compact               One-line-per-record
  --raw                   Raw API response
  --limit <n>             Max results (on list commands)

WRITE SAFETY:
  Commands that modify data (schedule, trigger, add, remove) require --execute.
  Without it, they run in dry-run mode and print what would happen.
"""


def print_help():
    """Print full command reference."""
    print(HELP_TEXT, end="")
