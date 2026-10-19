"""Run pinned yt-dlp with Arc's macOS keyring name patched in memory."""

from __future__ import annotations

import sys
from collections.abc import Callable
from types import ModuleType

import yt_dlp
from yt_dlp import cookies, version

SUPPORTED_YTDLP_VERSION = "2026.07.04"
# A private yt-dlp function, so the type checker cannot see it on the module
SETTINGS_FUNCTION = "_get_chromium_based_browser_settings"


class ArcCookieAdapterError(RuntimeError):
    """The proven yt-dlp private cookie contract is unavailable."""


def install_arc_cookie_adapter(
    cookies_module: ModuleType,
    *,
    runtime_version: str,
) -> None:
    """Patch only Chrome's keyring name so yt-dlp can decrypt Arc cookies."""
    if runtime_version != SUPPORTED_YTDLP_VERSION:
        raise ArcCookieAdapterError(
            f"Arc cookie adapter requires yt-dlp {SUPPORTED_YTDLP_VERSION}; "
            f"found {runtime_version}"
        )

    original = getattr(cookies_module, SETTINGS_FUNCTION, None)
    if not callable(original):
        raise ArcCookieAdapterError(
            "yt-dlp private cookie settings contract changed: settings function missing"
        )
    original_settings: Callable[[str], object] = original

    def read_settings(browser_name: str) -> dict[str, object]:
        try:
            settings = original_settings(browser_name)
        except Exception as error:
            raise ArcCookieAdapterError(
                "yt-dlp private cookie settings contract changed: "
                "Chrome settings call failed"
            ) from error
        required_keys = {"browser_dir", "keyring_name", "supports_profiles"}
        if (
            not isinstance(settings, dict)
            or not required_keys.issubset(settings)
            or settings.get("keyring_name") != "Chrome"
        ):
            raise ArcCookieAdapterError(
                "yt-dlp private cookie settings contract changed: "
                "expected Chrome keyring settings"
            )
        return settings

    read_settings("chrome")

    def arc_settings(browser_name: str) -> dict[str, object]:
        settings = read_settings(browser_name)
        if browser_name != "chrome":
            return settings
        adapted = settings.copy()
        adapted["keyring_name"] = "Arc"
        return adapted

    setattr(cookies_module, SETTINGS_FUNCTION, arc_settings)


def main(argv: list[str] | None = None) -> int:
    """Install the checked adapter, then delegate arguments to yt-dlp."""
    try:
        install_arc_cookie_adapter(cookies, runtime_version=version.__version__)
    except ArcCookieAdapterError as error:
        print(f"ERROR: Arc cookie adapter failed: {error}", file=sys.stderr)
        return 1

    try:
        yt_dlp.main(argv)
    except SystemExit as error:
        return error.code if isinstance(error.code, int) else 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
