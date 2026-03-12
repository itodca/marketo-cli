"""Tests for the Marketo API client."""

from __future__ import annotations

import pytest

from mrkto.client import MarketoAPIError, MarketoClient
from mrkto.config import Config


def make_config(profile: str = "default") -> Config:
    return Config(
        munchkin_id="123-ABC-456",
        client_id="test-id",
        client_secret="test-secret",
        rest_url="https://123-ABC-456.mktorest.com/rest",
        identity_url="https://123-ABC-456.mktorest.com/identity",
        profile=profile,
    )


class FakeResponse:
    def __init__(self, *, status_code: int = 200, json_data=None, text: str = ""):
        self.status_code = status_code
        self._json_data = json_data if json_data is not None else {}
        self.text = text
        self.reason = "OK" if status_code < 400 else "Error"

    def json(self):
        return self._json_data

    def raise_for_status(self):
        if self.status_code >= 400:
            import requests

            raise requests.HTTPError(f"{self.status_code} {self.reason}", response=self)


class FakeSession:
    def __init__(self, *, token_responses: list[FakeResponse] | None = None, request_responses: list[FakeResponse] | None = None):
        self.token_responses = token_responses or []
        self.request_responses = request_responses or []
        self.request_calls = []
        self.token_calls = []

    def get(self, url, *, params=None, timeout=None):
        self.token_calls.append({"url": url, "params": params, "timeout": timeout})
        if not self.token_responses:
            raise AssertionError("Unexpected token request")
        return self.token_responses.pop(0)

    def request(self, method, url, *, headers=None, params=None, json=None, timeout=None):
        self.request_calls.append(
            {
                "method": method,
                "url": url,
                "headers": headers,
                "params": params,
                "json": json,
                "timeout": timeout,
            }
        )
        if not self.request_responses:
            raise AssertionError("Unexpected API request")
        return self.request_responses.pop(0)


def token_response(token: str = "tok123", expires_in: int = 3600) -> FakeResponse:
    return FakeResponse(json_data={"access_token": token, "expires_in": expires_in})


def api_response(payload: dict) -> FakeResponse:
    return FakeResponse(json_data=payload)


def test_client_authenticates_and_caches_token_in_memory(tmp_path):
    session = FakeSession(
        token_responses=[token_response()],
        request_responses=[
            api_response({"success": True, "result": [{"id": 1}]}),
            api_response({"success": True, "result": [{"id": 2}]}),
        ],
    )

    client = MarketoClient(make_config(), token_cache_dir=tmp_path, session=session)
    first = client.get("/v1/leads.json")
    second = client.get("/v1/leads.json")

    assert first["result"][0]["id"] == 1
    assert second["result"][0]["id"] == 2
    assert len(session.token_calls) == 1
    assert len(session.request_calls) == 2


def test_client_uses_profile_scoped_token_cache(tmp_path):
    cache_dir = tmp_path / "cache"

    first_session = FakeSession(
        token_responses=[token_response(token="profile-a-token")],
        request_responses=[api_response({"success": True, "result": []})],
    )
    first_client = MarketoClient(make_config(profile="profile-a"), token_cache_dir=cache_dir, session=first_session)
    first_client.get("/v1/leads.json")

    second_session = FakeSession(
        token_responses=[token_response(token="profile-b-token")],
        request_responses=[api_response({"success": True, "result": []})],
    )
    second_client = MarketoClient(make_config(profile="profile-b"), token_cache_dir=cache_dir, session=second_session)
    second_client.get("/v1/leads.json")

    token_files = sorted(path.name for path in cache_dir.iterdir())
    assert token_files == ["token-profile-a.json", "token-profile-b.json"]


def test_client_refreshes_token_after_401(tmp_path):
    session = FakeSession(
        token_responses=[token_response(token="stale-token"), token_response(token="fresh-token")],
        request_responses=[
            FakeResponse(status_code=401, text="Unauthorized"),
            api_response({"success": True, "result": [{"id": 1}]}),
        ],
    )

    client = MarketoClient(make_config(), token_cache_dir=tmp_path, session=session)
    result = client.get("/v1/leads.json")

    assert result["result"][0]["id"] == 1
    assert len(session.token_calls) == 2
    assert len(session.request_calls) == 2


def test_client_retries_rate_limit_errors(tmp_path):
    sleeps = []
    session = FakeSession(
        token_responses=[token_response()],
        request_responses=[
            api_response({"success": False, "errors": [{"code": "606", "message": "Rate limit"}]}),
            api_response({"success": True, "result": [{"id": 1}]}),
        ],
    )

    client = MarketoClient(
        make_config(),
        token_cache_dir=tmp_path,
        session=session,
        sleep_fn=lambda seconds: sleeps.append(seconds),
    )

    result = client.get("/v1/leads.json")

    assert result["result"][0]["id"] == 1
    assert sleeps == [20]


def test_client_get_all_pages_follows_next_page_token(tmp_path):
    session = FakeSession(
        token_responses=[token_response()],
        request_responses=[
            api_response({"success": True, "result": [{"id": 1}], "nextPageToken": "abc"}),
            api_response({"success": True, "result": [{"id": 2}]}),
        ],
    )

    client = MarketoClient(make_config(), token_cache_dir=tmp_path, session=session)
    results = client.get_all_pages("/v1/leads.json")

    assert results["result"] == [{"id": 1}, {"id": 2}]
    assert session.request_calls[1]["params"]["nextPageToken"] == "abc"


def test_client_delete_supports_query_and_json_body(tmp_path):
    session = FakeSession(
        token_responses=[token_response()],
        request_responses=[api_response({"success": True, "result": []})],
    )

    client = MarketoClient(make_config(), token_cache_dir=tmp_path, session=session)
    client.delete("/v1/lists/10/leads.json", params={"id": [1, 2]}, json_body={"input": [{"id": 1}, {"id": 2}]})

    request = session.request_calls[0]
    assert request["method"] == "DELETE"
    assert request["params"] == {"id": [1, 2]}
    assert request["json"] == {"input": [{"id": 1}, {"id": 2}]}


def test_client_raises_marketo_api_error_for_invalid_json_response(tmp_path):
    class InvalidJsonResponse(FakeResponse):
        def json(self):
            raise ValueError("not json")

    session = FakeSession(
        token_responses=[token_response()],
        request_responses=[InvalidJsonResponse()],
    )

    client = MarketoClient(make_config(), token_cache_dir=tmp_path, session=session)

    with pytest.raises(MarketoAPIError, match="Response was not valid JSON"):
        client.get("/v1/leads.json")
