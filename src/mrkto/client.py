"""Marketo REST API client — auth, token cache, pagination, rate limits."""

import json
import time
from pathlib import Path

import requests

from mrkto.config import Config

DEFAULT_TOKEN_DIR = Path.home() / ".config" / "mrkto"


class MarketoClient:
    def __init__(self, config: Config, token_cache_dir: Path | None = None):
        self.config = config
        self._token_cache_dir = token_cache_dir or DEFAULT_TOKEN_DIR
        self._token: str | None = None
        self._token_expiry: float = 0

    def _token_cache_path(self) -> Path:
        return self._token_cache_dir / "token.json"

    def _load_cached_token(self) -> str | None:
        path = self._token_cache_path()
        if not path.exists():
            return None
        try:
            data = json.loads(path.read_text())
            if data.get("expiry", 0) > time.time():
                return data["access_token"]
        except (json.JSONDecodeError, KeyError):
            pass
        return None

    def _save_token(self, token: str, expires_in: int) -> None:
        self._token_cache_dir.mkdir(parents=True, exist_ok=True)
        path = self._token_cache_path()
        data = {
            "access_token": token,
            "expiry": time.time() + expires_in - 60,  # 60s buffer
        }
        path.write_text(json.dumps(data))

    def _get_token(self) -> str:
        if self._token and time.time() < self._token_expiry:
            return self._token

        cached = self._load_cached_token()
        if cached:
            self._token = cached
            return cached

        resp = requests.get(
            f"{self.config.identity_url}/oauth/token",
            params={
                "grant_type": "client_credentials",
                "client_id": self.config.client_id,
                "client_secret": self.config.client_secret,
            },
        )
        resp.raise_for_status()
        data = resp.json()
        token = data["access_token"]
        expires_in = data.get("expires_in", 3600)

        self._token = token
        self._token_expiry = time.time() + expires_in - 60
        self._save_token(token, expires_in)
        return token

    def _headers(self) -> dict:
        return {"Authorization": f"Bearer {self._get_token()}"}

    def get(self, path: str, params: dict | None = None) -> dict:
        url = f"{self.config.rest_url}{path}"
        resp = requests.get(url, headers=self._headers(), params=params or {})
        resp.raise_for_status()
        data = resp.json()
        self._check_errors(data)
        return data

    def post(self, path: str, json_body: dict | None = None) -> dict:
        url = f"{self.config.rest_url}{path}"
        resp = requests.post(url, headers=self._headers(), json=json_body)
        resp.raise_for_status()
        data = resp.json()
        self._check_errors(data)
        return data

    def get_paginated(self, path: str, params: dict | None = None) -> list:
        """Fetch all pages and return combined results."""
        params = dict(params or {})
        all_results = []
        while True:
            data = self.get(path, params)
            results = data.get("result", [])
            all_results.extend(results)
            next_token = data.get("nextPageToken")
            if not next_token:
                break
            params["nextPageToken"] = next_token
        return all_results

    def _check_errors(self, data: dict) -> None:
        if not data.get("success", True):
            errors = data.get("errors", [])
            if errors:
                code = errors[0].get("code", "")
                msg = errors[0].get("message", "Unknown error")
                if code == "606":
                    # Rate limit — wait and caller can retry
                    time.sleep(20)
                raise MarketoAPIError(code, msg)


class MarketoAPIError(Exception):
    def __init__(self, code: str, message: str):
        self.code = code
        self.message = message
        super().__init__(f"[{code}] {message}")
