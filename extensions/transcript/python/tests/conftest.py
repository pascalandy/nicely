"""Test configuration: no test reaches yt-dlp or Deepgram by accident."""

from __future__ import annotations

from typing import Any

import httpx
import pytest

import youtube


@pytest.fixture(autouse=True)
def no_real_processes(monkeypatch: pytest.MonkeyPatch) -> None:
    """Fail an in-process test that would start yt-dlp or call Deepgram; a test
    that needs one replaces it first."""

    def refuse(command: Any, *_args: Any, **_kwargs: Any) -> Any:
        pytest.fail(f"test reached a real external boundary: {command!r}")

    monkeypatch.setattr(youtube, "run_child", refuse)
    monkeypatch.setattr(httpx, "post", refuse)
