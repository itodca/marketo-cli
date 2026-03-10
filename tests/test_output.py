"""Tests for output formatting."""

import json


def test_json_output(capsys):
    from mrkto.output import print_result

    data = [{"id": 1, "email": "a@b.com"}]
    print_result(data, fmt="json")
    out = capsys.readouterr().out
    assert json.loads(out) == data


def test_compact_output(capsys):
    from mrkto.output import print_result

    data = [{"id": 1, "email": "a@b.com"}, {"id": 2, "email": "c@d.com"}]
    print_result(data, fmt="compact")
    out = capsys.readouterr().out
    lines = out.strip().split("\n")
    assert len(lines) == 2
    assert json.loads(lines[0])["id"] == 1


def test_field_selection(capsys):
    from mrkto.output import print_result

    data = [{"id": 1, "email": "a@b.com", "company": "Acme"}]
    print_result(data, fmt="json", fields=["id", "email"])
    out = json.loads(capsys.readouterr().out)
    assert "company" not in out[0]
    assert out[0]["email"] == "a@b.com"


def test_error_output(capsys):
    from mrkto.output import print_error

    print_error("something went wrong")
    err = capsys.readouterr().err
    assert json.loads(err)["error"] == "something went wrong"
