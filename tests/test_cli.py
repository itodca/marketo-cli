"""Tests for the Typer CLI contract."""

from __future__ import annotations

import json

from typer.testing import CliRunner

from mrkto.cli import app

runner = CliRunner()


def test_lead_list_uses_email_filter(monkeypatch):
    calls = {}

    monkeypatch.setattr("mrkto.cli.get_client", lambda profile: object())

    def fake_list_leads(client, *, filter_type, filter_values, fields=None, limit=None):
        calls.update(
            {
                "filter_type": filter_type,
                "filter_values": filter_values,
                "fields": fields,
                "limit": limit,
            }
        )
        return {"success": True, "result": [{"id": 1, "email": "user@example.com"}]}

    monkeypatch.setattr("mrkto.cli.lead_resource.list_leads", fake_list_leads)

    result = runner.invoke(app, ["lead", "list", "--email", "user@example.com", "--limit", "5"])

    assert result.exit_code == 0
    assert calls == {
        "filter_type": "email",
        "filter_values": "user@example.com",
        "fields": None,
        "limit": 5,
    }
    assert json.loads(result.stdout) == [{"id": 1, "email": "user@example.com"}]


def test_smart_campaign_trigger_defaults_to_dry_run(monkeypatch):
    calls = {}

    monkeypatch.setattr("mrkto.cli.get_client", lambda profile: object())

    def fake_trigger(client, *, campaign_id, lead_ids, dry_run):
        calls.update({"campaign_id": campaign_id, "lead_ids": lead_ids, "dry_run": dry_run})
        return {"dry_run": dry_run, "campaign_id": campaign_id, "lead_ids": lead_ids}

    monkeypatch.setattr("mrkto.cli.smart_campaign.trigger_smart_campaign", fake_trigger)

    result = runner.invoke(app, ["smart-campaign", "trigger", "42", "--lead", "1", "--lead", "2"])

    assert result.exit_code == 0
    assert calls == {"campaign_id": 42, "lead_ids": [1, 2], "dry_run": True}
    assert json.loads(result.stdout) == {"dry_run": True, "campaign_id": 42, "lead_ids": [1, 2]}


def test_api_post_rejects_body_and_input_together(monkeypatch, tmp_path):
    payload = tmp_path / "payload.json"
    payload.write_text('{"name":"from-file"}')

    monkeypatch.setattr("mrkto.cli.get_client", lambda profile: object())

    result = runner.invoke(
        app,
        ["api", "post", "/v1/leads/push.json", "--body", "name=from-flag", "--input", str(payload)],
    )

    assert result.exit_code == 1
    assert json.loads(result.stderr)["error"] == "Use either --body or --input, not both"


def test_skill_install_defaults_to_project_scope(monkeypatch):
    calls = {}

    def fake_install_skill(*, global_install=False):
        calls["global_install"] = global_install

    monkeypatch.setattr("mrkto.cli.skill.install_skill", fake_install_skill)

    result = runner.invoke(app, ["skill", "install"])

    assert result.exit_code == 0
    assert calls == {"global_install": False}


def test_skill_uninstall_can_target_global_scope(monkeypatch):
    calls = {}

    def fake_uninstall_skill(*, global_install=False):
        calls["global_install"] = global_install

    monkeypatch.setattr("mrkto.cli.skill.uninstall_skill", fake_uninstall_skill)

    result = runner.invoke(app, ["skill", "uninstall", "--global"])

    assert result.exit_code == 0
    assert calls == {"global_install": True}
