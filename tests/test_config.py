"""Tests for config module."""

import os
from unittest.mock import patch

import pytest


def test_config_from_munchkin_id():
    env = {
        "MARKETO_MUNCHKIN_ID": "123-ABC-456",
        "MARKETO_CLIENT_ID": "test-client-id",
        "MARKETO_CLIENT_SECRET": "test-secret",
    }
    with patch.dict(os.environ, env, clear=True):
        from mrkto.config import load_config

        cfg = load_config()
        assert cfg.munchkin_id == "123-ABC-456"
        assert cfg.client_id == "test-client-id"
        assert cfg.client_secret == "test-secret"
        assert cfg.rest_url == "https://123-ABC-456.mktorest.com/rest"
        assert cfg.identity_url == "https://123-ABC-456.mktorest.com/identity"


def test_config_url_override():
    env = {
        "MARKETO_MUNCHKIN_ID": "123-ABC-456",
        "MARKETO_CLIENT_ID": "id",
        "MARKETO_CLIENT_SECRET": "secret",
        "MARKETO_REST_URL": "https://custom.example.com/rest",
    }
    with patch.dict(os.environ, env, clear=True):
        from mrkto.config import load_config

        cfg = load_config()
        assert cfg.rest_url == "https://custom.example.com/rest"


def test_config_missing_required():
    with patch.dict(os.environ, {}, clear=True):
        from mrkto.config import load_config

        with pytest.raises(SystemExit):
            load_config()
