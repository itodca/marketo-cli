"""Tests for Marketo API client."""

from unittest.mock import patch, MagicMock

from mrkto.config import Config


def make_config():
    return Config(
        munchkin_id="123-ABC-456",
        client_id="test-id",
        client_secret="test-secret",
        rest_url="https://123-ABC-456.mktorest.com/rest",
        identity_url="https://123-ABC-456.mktorest.com/identity",
        profile="default",
    )


def mock_response(data):
    resp = MagicMock()
    resp.json.return_value = data
    resp.raise_for_status = MagicMock()
    return resp


def test_client_authenticates(tmp_path):
    from mrkto.client import MarketoClient

    token_resp = mock_response({"access_token": "tok123", "expires_in": 3600})

    with patch("mrkto.client.requests.get", return_value=token_resp) as mock_get:
        client = MarketoClient(make_config(), token_cache_dir=tmp_path)
        token = client._get_token()
        assert token == "tok123"
        mock_get.assert_called_once()


def test_client_get(tmp_path):
    from mrkto.client import MarketoClient

    token_resp = mock_response({"access_token": "tok123", "expires_in": 3600})
    api_resp = mock_response({"success": True, "result": [{"id": 1}]})

    with patch("mrkto.client.requests.get", side_effect=[token_resp, api_resp]):
        client = MarketoClient(make_config(), token_cache_dir=tmp_path)
        result = client.get("/v1/leads.json", params={"filterType": "email"})
        assert result["success"] is True
        assert result["result"][0]["id"] == 1


def test_client_caches_token_in_memory(tmp_path):
    from mrkto.client import MarketoClient

    token_resp = mock_response({"access_token": "tok123", "expires_in": 3600})
    api_resp = mock_response({"success": True, "result": []})

    with patch(
        "mrkto.client.requests.get", side_effect=[token_resp, api_resp, api_resp]
    ) as mock_get:
        client = MarketoClient(make_config(), token_cache_dir=tmp_path)
        client.get("/v1/leads.json", params={})
        client.get("/v1/leads.json", params={})
        # Token fetched once (1 call), API called twice (2 calls) = 3 total
        assert mock_get.call_count == 3


def test_client_api_error(tmp_path):
    import pytest
    from mrkto.client import MarketoClient, MarketoAPIError

    token_resp = mock_response({"access_token": "tok123", "expires_in": 3600})
    error_resp = mock_response(
        {"success": False, "errors": [{"code": "1001", "message": "Not found"}]}
    )

    with patch("mrkto.client.requests.get", side_effect=[token_resp, error_resp]):
        client = MarketoClient(make_config(), token_cache_dir=tmp_path)
        with pytest.raises(MarketoAPIError, match="1001"):
            client.get("/v1/leads.json", params={})


def test_client_pagination(tmp_path):
    from mrkto.client import MarketoClient

    token_resp = mock_response({"access_token": "tok123", "expires_in": 3600})
    page1 = mock_response(
        {"success": True, "result": [{"id": 1}], "nextPageToken": "abc"}
    )
    page2 = mock_response({"success": True, "result": [{"id": 2}]})

    with patch(
        "mrkto.client.requests.get", side_effect=[token_resp, page1, page2]
    ):
        client = MarketoClient(make_config(), token_cache_dir=tmp_path)
        results = client.get_paginated("/v1/leads.json", params={})
        assert len(results) == 2
        assert results[0]["id"] == 1
        assert results[1]["id"] == 2
