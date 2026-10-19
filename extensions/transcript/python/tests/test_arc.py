"""Tests of the Arc cookie adapter that patches yt-dlp's Chrome keyring name."""

from __future__ import annotations

from types import ModuleType

import pytest

import ytdlp_arc


def cookies_module(settings: dict) -> ModuleType:
    """A stand-in for yt-dlp's cookies module whose settings function returns
    a copy of `settings`."""
    module = ModuleType("cookies")
    setattr(module, ytdlp_arc.SETTINGS_FUNCTION, lambda _name: settings.copy())
    return module


def test_arc_adapter_changes_only_the_keyring_name() -> None:
    original_settings = {
        "browser_dir": "/browser",
        "keyring_name": "Chrome",
        "supports_profiles": True,
    }
    module = cookies_module(original_settings)

    ytdlp_arc.install_arc_cookie_adapter(module, runtime_version="2026.07.04")
    adapted = module._get_chromium_based_browser_settings("chrome")

    assert adapted == {**original_settings, "keyring_name": "Arc"}
    assert original_settings["keyring_name"] == "Chrome"


def test_arc_adapter_rejects_unproven_ytdlp_version() -> None:
    module = cookies_module(
        {
            "browser_dir": "/browser",
            "keyring_name": "Chrome",
            "supports_profiles": True,
        }
    )

    with pytest.raises(
        ytdlp_arc.ArcCookieAdapterError,
        match="requires yt-dlp 2026.07.04",
    ):
        ytdlp_arc.install_arc_cookie_adapter(module, runtime_version="2099.01.01")


def test_arc_adapter_fails_loud_when_private_contract_changes() -> None:
    module = cookies_module({"browser_dir": "/browser", "supports_profiles": True})

    with pytest.raises(
        ytdlp_arc.ArcCookieAdapterError,
        match="private cookie settings contract changed",
    ):
        ytdlp_arc.install_arc_cookie_adapter(module, runtime_version="2026.07.04")
