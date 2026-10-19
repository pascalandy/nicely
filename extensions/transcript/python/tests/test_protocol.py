"""The sequence rules of the protocol, driven through the engine's main()."""

from __future__ import annotations

import io
import json
from datetime import UTC, datetime

import httpx
import pytest
from harness import (
    GOLDENS,
    URL,
    Held,
    ack,
    fake_deepgram,
    fake_ytdlp,
    line,
    request,
    run,
)

import transcript


def test_a_dry_run_answers_its_plan_and_starts_nothing(tmp_path) -> None:
    code, messages = run(request(tmp_path, dry_run=True))

    assert code == 0
    assert messages == [
        {
            "protocol_version": 1,
            "type": "plan",
            "items": [{"item": URL, "output": str(tmp_path / "out")}],
        }
    ]
    assert list((tmp_path / "out").iterdir()) == []


def test_a_run_plans_then_asks_before_it_uploads_and_saves(
    tmp_path, monkeypatch
) -> None:
    downloads = fake_ytdlp(monkeypatch)
    uploads = fake_deepgram(monkeypatch)

    code, messages = run(request(tmp_path), ack(tmp_path))

    assert code == 0
    assert [m["type"] for m in messages] == ["plan", "intent", "step"]
    intent, step = messages[1], messages[2]
    assert (intent["item"], intent["step"], intent["name"]) == (
        URL,
        "transcribe",
        "A_video",
    )
    assert intent["tools"]["transcript"] == transcript.__version__
    assert intent["tools"]["yt-dlp"]
    assert step == {
        "protocol_version": 1,
        "type": "step",
        "item": URL,
        "step": "transcribe",
        "status": "completed",
        "code": None,
        "warnings": [],
        "detail": "",
    }
    assert len(downloads) == 1
    (upload,) = uploads
    assert upload.headers["Authorization"] == "Token fake-key"
    folder = tmp_path / "out" / "A_video"
    for name in ("raw_transcript.txt", "raw_sentences.txt", "raw_transcript.json"):
        assert (folder / name).read_bytes() == (GOLDENS / name).read_bytes(), name
    assert "Transcript status: complete" in (folder / "meta.txt").read_text()
    assert list((tmp_path / "tmp").iterdir()) == [tmp_path / "tmp" / "audio.webm"]


def test_a_key_quoted_by_an_http_error_stays_out_of_the_detail_and_meta(
    tmp_path, monkeypatch
) -> None:
    fake_ytdlp(monkeypatch)
    key = "fake-key\nsecret-part"
    monkeypatch.setenv("DEEPGRAM_API_KEY", key)

    def post(*_args, **_kwargs):
        value = b"Token " + key.encode()
        raise httpx.LocalProtocolError(f"Illegal header value {value!r}")

    monkeypatch.setattr(httpx, "post", post)

    code, messages = run(request(tmp_path), ack(tmp_path))

    assert code == 0
    step = messages[-1]
    meta = (tmp_path / "out" / "A_video" / "meta.txt").read_text()
    assert step["code"] == "deepgram_unknown"
    for text in (step["detail"], meta):
        assert "<redacted>" in text
        assert "fake-key" not in text and "secret-part" not in text


def test_a_title_without_ascii_letters_proposes_the_video_id(
    tmp_path, monkeypatch
) -> None:
    fake_ytdlp(monkeypatch, title="日本語")

    code, messages = run(request(tmp_path))

    assert messages[1]["name"] == "EIEc43CxIvY"
    assert code == transcript.NO_ACK


def test_a_closed_stdin_after_the_intent_stops_before_any_upload(
    tmp_path, monkeypatch
) -> None:
    fake_ytdlp(monkeypatch)
    uploads = fake_deepgram(monkeypatch)

    code, messages = run(request(tmp_path))

    assert code == transcript.NO_ACK
    assert [m["type"] for m in messages] == ["plan", "intent"]
    assert uploads == []
    assert list((tmp_path / "out").iterdir()) == []


def test_no_ack_within_the_timeout_stops_before_any_upload(
    tmp_path, monkeypatch
) -> None:
    fake_ytdlp(monkeypatch)
    uploads = fake_deepgram(monkeypatch)
    monkeypatch.setattr(transcript, "ACK_TIMEOUT", 0.2)
    stdin = Held(request(tmp_path))
    try:
        code, messages = run(stdin=stdin)
    finally:
        stdin.release.set()

    assert code == transcript.NO_ACK
    assert [m["type"] for m in messages] == ["plan", "intent"]
    assert uploads == []


@pytest.mark.parametrize(
    "bad_ack",
    [
        line("ack", item="https://youtu.be/other", step="transcribe", output_dir="/"),
        line("ack", item=URL, step="other", output_dir="/"),
        line("ack", item=URL, step="transcribe", output_dir="relative"),
        line("step", item=URL, step="transcribe"),
        "not json\n",
    ],
)
def test_an_ack_that_breaks_the_protocol_stops_before_any_upload(
    tmp_path, monkeypatch, bad_ack
) -> None:
    fake_ytdlp(monkeypatch)
    uploads = fake_deepgram(monkeypatch)

    code, messages = run(request(tmp_path), bad_ack)

    assert code == transcript.PROTOCOL_BROKEN
    assert [m["type"] for m in messages] == ["plan", "intent"]
    assert uploads == []


@pytest.mark.parametrize(
    "fields",
    [
        {"protocol_version": 2},
        {"type": "ack"},
        {"source": "zoom"},
        {"urls": [URL, URL]},
        {"output": "relative"},
        {"browser": "chrome"},
        {"dry_run": "yes"},
        {"deadline": "tomorrow"},
        {"deadline": datetime.now(UTC).replace(tzinfo=None).isoformat()},
    ],
)
def test_a_request_that_breaks_the_protocol_gets_no_plan(tmp_path, fields) -> None:
    message = json.loads(request(tmp_path))
    code, messages = run(json.dumps({**message, **fields}) + "\n")

    assert code == transcript.PROTOCOL_BROKEN
    assert messages == []


def test_a_closed_stdin_before_the_request_gets_no_plan() -> None:
    code, messages = run(stdin=io.StringIO(""))

    assert (code, messages) == (transcript.PROTOCOL_BROKEN, [])
