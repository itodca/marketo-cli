"""Marketo REST API client with profile-scoped auth and pagination helpers."""

from __future__ import annotations

import json
import re
import time
from pathlib import Path
from typing import Any

import requests

from mrkto.config import Config

DEFAULT_TOKEN_DIR = Path.home() / ".config" / "mrkto"
DEFAULT_TIMEOUT_SECONDS = 30
DEFAULT_MAX_RETRIES = 2
DEFAULT_RATE_LIMIT_SLEEP_SECONDS = 20


class MarketoAPIError(Exception):
    """Application-facing error for HTTP and Marketo API failures."""

    def __init__(self, code: str, message: str, http_status: int | None = None):
        self.code = code
        self.message = message
        self.http_status = http_status
        super().__init__(f"[{code}] {message}")


class MarketoClient:
    def __init__(
        self,
        config: Config,
        token_cache_dir: Path | None = None,
        session: requests.Session | None = None,
        timeout: int = DEFAULT_TIMEOUT_SECONDS,
        max_retries: int = DEFAULT_MAX_RETRIES,
        rate_limit_sleep_seconds: int = DEFAULT_RATE_LIMIT_SLEEP_SECONDS,
        sleep_fn=time.sleep,
    ):
        self.config = config
        self._token_cache_dir = token_cache_dir or DEFAULT_TOKEN_DIR
        self._session = session or requests.Session()
        self._timeout = timeout
        self._max_retries = max_retries
        self._rate_limit_sleep_seconds = rate_limit_sleep_seconds
        self._sleep_fn = sleep_fn
        self._token: str | None = None
        self._token_expiry: float = 0

    def _profile_slug(self) -> str:
        return re.sub(r"[^A-Za-z0-9_.-]+", "-", self.config.profile or "default")

    def _token_cache_path(self) -> Path:
        return self._token_cache_dir / f"token-{self._profile_slug()}.json"

    def _load_cached_token(self) -> tuple[str, float] | None:
        path = self._token_cache_path()
        if not path.exists():
            return None

        try:
            data = json.loads(path.read_text())
        except (OSError, json.JSONDecodeError):
            return None

        token = data.get("access_token")
        expiry = data.get("expiry")
        if not token or not isinstance(expiry, (int, float)):
            return None
        if expiry <= time.time():
            return None
        return token, float(expiry)

    def _save_token(self, token: str, expires_in: int) -> None:
        self._token_cache_dir.mkdir(parents=True, exist_ok=True)
        path = self._token_cache_path()
        expiry = time.time() + expires_in - 60
        path.write_text(json.dumps({"access_token": token, "expiry": expiry}))

    def _token_is_valid(self) -> bool:
        return bool(self._token and time.time() < self._token_expiry)

    def _fetch_token(self) -> tuple[str, float]:
        try:
            response = self._session.get(
                f"{self.config.identity_url}/oauth/token",
                params={
                    "grant_type": "client_credentials",
                    "client_id": self.config.client_id,
                    "client_secret": self.config.client_secret,
                },
                timeout=self._timeout,
            )
            response.raise_for_status()
        except requests.RequestException as exc:
            raise MarketoAPIError("auth_request_failed", str(exc)) from exc

        try:
            data = response.json()
        except ValueError as exc:
            raise MarketoAPIError("auth_response_invalid", "Identity response was not valid JSON") from exc

        token = data.get("access_token")
        if not token:
            raise MarketoAPIError("auth_response_invalid", "Identity response did not include an access token")

        expires_in = int(data.get("expires_in", 3600))
        expiry = time.time() + expires_in - 60
        self._save_token(token, expires_in)
        return token, expiry

    def _get_token(self, *, force_refresh: bool = False) -> str:
        if not force_refresh and self._token_is_valid():
            return self._token or ""

        if not force_refresh:
            cached = self._load_cached_token()
            if cached:
                self._token, self._token_expiry = cached
                return self._token

        self._token, self._token_expiry = self._fetch_token()
        return self._token

    def _build_url(self, path: str) -> str:
        normalized = path.strip()
        if not normalized.startswith("/"):
            normalized = f"/{normalized}"
        if normalized.startswith("/rest/"):
            normalized = normalized.removeprefix("/rest")
        return f"{self.config.rest_url.rstrip('/')}{normalized}"

    def _headers(self, *, force_refresh: bool = False) -> dict[str, str]:
        return {"Authorization": f"Bearer {self._get_token(force_refresh=force_refresh)}"}

    def _parse_response_json(self, response: requests.Response) -> dict[str, Any]:
        try:
            return response.json()
        except ValueError as exc:
            raise MarketoAPIError(
                str(response.status_code),
                "Response was not valid JSON",
                http_status=response.status_code,
            ) from exc

    def _request(
        self,
        method: str,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json_body: dict[str, Any] | None = None,
        retries: int | None = None,
    ) -> dict[str, Any]:
        attempts = 0
        max_attempts = 1 + (self._max_retries if retries is None else retries)
        force_refresh = False

        while attempts < max_attempts:
            attempts += 1
            try:
                response = self._session.request(
                    method.upper(),
                    self._build_url(path),
                    headers=self._headers(force_refresh=force_refresh),
                    params=params,
                    json=json_body,
                    timeout=self._timeout,
                )
            except requests.RequestException as exc:
                raise MarketoAPIError("request_failed", str(exc)) from exc

            if response.status_code == 401 and attempts < max_attempts:
                force_refresh = True
                continue

            try:
                response.raise_for_status()
            except requests.HTTPError as exc:
                message = response.text.strip() or response.reason or "HTTP request failed"
                raise MarketoAPIError(
                    str(response.status_code),
                    message,
                    http_status=response.status_code,
                ) from exc

            data = self._parse_response_json(response)
            if data.get("success", True):
                return data

            errors = data.get("errors", [])
            code = errors[0].get("code", "unknown_error") if errors else "unknown_error"
            message = errors[0].get("message", "Unknown Marketo error") if errors else "Unknown Marketo error"

            if code == "606" and attempts < max_attempts:
                self._sleep_fn(self._rate_limit_sleep_seconds)
                continue

            raise MarketoAPIError(code, message)

        raise MarketoAPIError("request_failed", "Request exhausted retries")

    def get(self, path: str, params: dict[str, Any] | None = None) -> dict[str, Any]:
        return self._request("GET", path, params=params)

    def post(
        self,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json_body: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        return self._request("POST", path, params=params, json_body=json_body)

    def delete(
        self,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json_body: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        return self._request("DELETE", path, params=params, json_body=json_body)

    def get_all_pages(
        self,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        limit: int | None = None,
        batch_size: int | None = None,
    ) -> dict[str, Any]:
        """Follow Marketo nextPageToken pagination and aggregate results."""

        page_params = dict(params or {})
        if batch_size is None and limit:
            batch_size = min(limit, 300)
        if batch_size is not None:
            page_params["batchSize"] = min(batch_size, 300)

        results: list[Any] = []
        warnings: list[Any] = []
        request_id: str | None = None

        while True:
            data = self.get(path, params=page_params)
            request_id = data.get("requestId", request_id)
            warnings.extend(data.get("warnings", []))
            results.extend(data.get("result", []))

            if limit and len(results) >= limit:
                results = results[:limit]
                break

            next_page_token = data.get("nextPageToken")
            if not next_page_token:
                break

            page_params["nextPageToken"] = next_page_token

        return {
            "success": True,
            "requestId": request_id,
            "warnings": warnings,
            "result": results,
        }

    def get_all_offset_pages(
        self,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        limit: int | None = None,
        page_size: int = 200,
    ) -> dict[str, Any]:
        """Follow Marketo offset pagination and aggregate results."""

        page_size = min(page_size, 200)
        page_params = dict(params or {})
        offset = int(page_params.get("offset", 0))

        results: list[Any] = []
        warnings: list[Any] = []
        request_id: str | None = None

        while True:
            remaining = None if limit is None else max(limit - len(results), 0)
            if remaining == 0:
                break

            current_page_size = page_size if remaining is None else min(page_size, remaining)
            page_params["offset"] = offset
            page_params["maxReturn"] = current_page_size

            data = self.get(path, params=page_params)
            request_id = data.get("requestId", request_id)
            warnings.extend(data.get("warnings", []))
            page_results = data.get("result", [])
            results.extend(page_results)

            if len(page_results) < current_page_size:
                break

            offset += len(page_results)

        return {
            "success": True,
            "requestId": request_id,
            "warnings": warnings,
            "result": results,
        }
