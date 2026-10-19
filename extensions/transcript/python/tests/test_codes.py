"""Each row of the program-code table of the protocol spec: the code, the
step status, and the effects that the engine reaches before it."""

from __future__ import annotations

import httpx
import pytest
from harness import ack, fake_deepgram, fake_ytdlp, request, run

import youtube


def step_of(messages: list[dict]) -> dict:
    assert messages[-1]["type"] == "step"
    return messages[-1]


@pytest.mark.parametrize(
    ("stderr", "code"),
    [
        ("ERROR: HTTP Error 503: Service Unavailable", "download_temporary"),
        (
            "ERROR: [youtube] Video unavailable. This video is private",
            "download_failed",
        ),
    ],
)
def test_a_failed_download_stays_pending_and_never_intends(
    tmp_path, monkeypatch, stderr, code
) -> None:
    fake_ytdlp(monkeypatch, stderr=stderr)
    uploads = fake_deepgram(monkeypatch)

    exit_code, messages = run(request(tmp_path), ack(tmp_path))

    assert exit_code == 0
    assert [m["type"] for m in messages] == ["plan", "step"]
    assert (step_of(messages)["status"], step_of(messages)["code"]) == ("pending", code)
    assert uploads == []


def test_a_download_that_cannot_write_is_write_failed_and_pending(
    tmp_path, monkeypatch
) -> None:
    def refuse(_command, **_kwargs):
        raise PermissionError("cannot write the scratch folder")

    monkeypatch.setattr(youtube, "run_child", refuse)

    exit_code, messages = run(request(tmp_path), ack(tmp_path))

    assert exit_code == 0
    step = step_of(messages)
    assert (step["status"], step["code"]) == ("pending", "write_failed")


@pytest.mark.parametrize(
    ("status", "code", "step_status"),
    [
        (429, "deepgram_unreachable", "failed"),
        (408, "deepgram_unreachable", "failed"),
        (401, "deepgram_key_rejected", "failed"),
        (400, "deepgram_failed", "failed"),
        (500, "deepgram_unknown", "unknown"),
        (504, "deepgram_unknown", "unknown"),
    ],
)
def test_a_deepgram_refusal_counts_one_upload_and_saves_no_transcript(
    tmp_path, monkeypatch, status, code, step_status
) -> None:
    fake_ytdlp(monkeypatch)
    uploads = fake_deepgram(monkeypatch, httpx.Response(status, json={"x": 1}))

    exit_code, messages = run(request(tmp_path), ack(tmp_path))

    assert exit_code == 0
    step = step_of(messages)
    assert (step["status"], step["code"]) == (step_status, code)
    assert len(uploads) == 1
    folder = tmp_path / "out" / "A_video"
    assert sorted(path.name for path in folder.iterdir()) == ["meta.txt"]
    assert "Transcript status: failed" in (folder / "meta.txt").read_text()


def test_an_unusable_transcript_is_unknown(tmp_path, monkeypatch) -> None:
    fake_ytdlp(monkeypatch)
    uploads = fake_deepgram(monkeypatch, httpx.Response(200, json={"results": {}}))

    exit_code, messages = run(request(tmp_path), ack(tmp_path))

    assert exit_code == 0
    step = step_of(messages)
    assert (step["status"], step["code"]) == ("unknown", "transcript_unusable")
    assert len(uploads) == 1


def test_a_result_that_cannot_be_saved_is_write_failed_and_unknown(
    tmp_path, monkeypatch
) -> None:
    fake_ytdlp(monkeypatch)
    uploads = fake_deepgram(monkeypatch)
    acknowledged = ack(tmp_path)
    folder = tmp_path / "out" / "A_video"
    folder.chmod(0o500)
    try:
        exit_code, messages = run(request(tmp_path), acknowledged)
    finally:
        folder.chmod(0o700)

    assert exit_code == 0
    step = step_of(messages)
    assert (step["status"], step["code"]) == ("unknown", "write_failed")
    assert len(uploads) == 1


def test_skipped_cookies_warn_in_the_step(tmp_path, monkeypatch) -> None:
    monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", tmp_path / "missing")
    fake_ytdlp(monkeypatch)
    fake_deepgram(monkeypatch)

    exit_code, messages = run(request(tmp_path, browser="arc"), ack(tmp_path))

    assert exit_code == 0
    step = step_of(messages)
    assert (step["status"], step["warnings"]) == (
        "completed",
        ["browser_cookies_skipped"],
    )
