"""Tests for config module."""

import os
from unittest.mock import patch

import pytest

from mrkto.config import load_config


@patch("mrkto.config._load_config_file", return_value={})
def test_config_from_env_vars(_mock_file):
    env = {
        "MARKETO_MUNCHKIN_ID": "123-ABC-456",
        "MARKETO_CLIENT_ID": "test-client-id",
        "MARKETO_CLIENT_SECRET": "test-secret",
    }
    with patch.dict(os.environ, env, clear=True):
        cfg = load_config()
        assert cfg.munchkin_id == "123-ABC-456"
        assert cfg.client_id == "test-client-id"
        assert cfg.client_secret == "test-secret"
        assert cfg.rest_url == "https://123-ABC-456.mktorest.com/rest"
        assert cfg.identity_url == "https://123-ABC-456.mktorest.com/identity"
        assert cfg.profile == "default"


@patch("mrkto.config._load_config_file", return_value={})
def test_config_url_override(_mock_file):
    env = {
        "MARKETO_MUNCHKIN_ID": "123-ABC-456",
        "MARKETO_CLIENT_ID": "id",
        "MARKETO_CLIENT_SECRET": "secret",
        "MARKETO_REST_URL": "https://custom.example.com/rest",
    }
    with patch.dict(os.environ, env, clear=True):
        cfg = load_config()
        assert cfg.rest_url == "https://custom.example.com/rest"


@patch("mrkto.config._load_config_file", return_value={})
def test_config_missing_required(_mock_file):
    with patch.dict(os.environ, {}, clear=True):
        with pytest.raises(SystemExit):
            load_config()


@patch("mrkto.config._load_config_file", return_value={
    "MARKETO_MUNCHKIN_ID": "999-XYZ-111",
    "MARKETO_CLIENT_ID": "file-client",
    "MARKETO_CLIENT_SECRET": "file-secret",
})
def test_config_from_file(_mock_file):
    with patch.dict(os.environ, {}, clear=True):
        cfg = load_config()
        assert cfg.munchkin_id == "999-XYZ-111"
        assert cfg.client_id == "file-client"


@patch("mrkto.config._load_config_file", return_value={
    "MARKETO_MUNCHKIN_ID": "from-file",
    "MARKETO_CLIENT_ID": "from-file",
    "MARKETO_CLIENT_SECRET": "from-file",
})
def test_env_vars_override_config_file(_mock_file):
    env = {
        "MARKETO_MUNCHKIN_ID": "from-env",
        "MARKETO_CLIENT_ID": "from-env",
        "MARKETO_CLIENT_SECRET": "from-env",
    }
    with patch.dict(os.environ, env, clear=True):
        cfg = load_config()
        assert cfg.munchkin_id == "from-env"
