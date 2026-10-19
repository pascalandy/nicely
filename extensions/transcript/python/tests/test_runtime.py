"""Tests of the run deadline, the retry policy, and the child processes."""

from __future__ import annotations

import os
import signal
import subprocess
import time
from collections.abc import Callable
from pathlib import Path

import httpx
import pytest

import runtime

SCRIPT_PATH = Path(__file__).parents[1] / "transcript.py"


class TestRetryPolicy:
    def test_retries_known_transient_error(self, monkeypatch) -> None:
        attempts = 0

        def flaky_call():
            nonlocal attempts
            attempts += 1
            if attempts == 1:
                raise httpx.ConnectError("temporary network failure")
            return "ok"

        monkeypatch.setattr(runtime.time, "sleep", lambda _delay: None)

        assert runtime.retry_request(flaky_call) == "ok"
        assert attempts == 2

    def test_does_not_retry_permanent_error(self, monkeypatch) -> None:
        attempts = 0

        def invalid_call():
            nonlocal attempts
            attempts += 1
            raise ValueError("invalid request")

        monkeypatch.setattr(runtime.time, "sleep", lambda _delay: None)

        with pytest.raises(ValueError, match="invalid request"):
            runtime.retry_request(invalid_call)
        assert attempts == 1


def test_pep723_declares_required_python() -> None:
    source = SCRIPT_PATH.read_text(encoding="utf-8")

    assert '# requires-python = ">=3.12"' in source.split("# ///", 2)[1]


def alive(pid: int) -> bool:
    """Whether `pid` runs; a zombie waiting for init counts as gone."""
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return False
    try:
        state = Path(f"/proc/{pid}/stat").read_text().split(") ")[-1]
    except ProcessLookupError:
        return False
    except FileNotFoundError:
        return not Path("/proc/self/stat").exists()
    return not state.startswith("Z")


def wait_for(condition: Callable[[], object], what: str, seconds: float = 30) -> None:
    deadline = time.monotonic() + seconds
    while not condition():
        assert time.monotonic() < deadline, f"timed out waiting for {what}"
        time.sleep(0.05)


# A child that starts its own child and has no TERM trap, as yt-dlp starts ffmpeg
UNCOOPERATIVE = """\
sleep 60 &
echo "$!" > "{ready}.tmp" && mv "{ready}.tmp" "{ready}"
wait
"""


def test_a_timeout_also_stops_what_the_child_started(tmp_path: Path) -> None:
    ready = tmp_path / "ready"
    grandchild = 0
    try:
        with pytest.raises(subprocess.TimeoutExpired):
            runtime.run_child(
                ["sh", "-c", UNCOOPERATIVE.format(ready=ready)], timeout=3
            )
        grandchild = int(ready.read_text())
        wait_for(lambda: not alive(grandchild), "the grandchild to end", 5)
    finally:
        if grandchild and alive(grandchild):
            os.kill(grandchild, signal.SIGKILL)


# A child that cleans up on SIGTERM, which SIGKILL would prevent
GENTLE = """\
trap ': > "{cleanup}"; exit 0' TERM
sleep 60 &
wait
"""


def test_a_timeout_lets_the_child_clean_up_after_sigterm(tmp_path: Path) -> None:
    cleanup = tmp_path / "cleanup"

    with pytest.raises(subprocess.TimeoutExpired):
        runtime.run_child(["sh", "-c", GENTLE.format(cleanup=cleanup)], timeout=1)

    assert cleanup.exists()
