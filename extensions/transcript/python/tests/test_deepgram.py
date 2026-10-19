"""Tests of the Deepgram request, its parser, and the three saved files."""

from __future__ import annotations

import logging
from pathlib import Path
from types import SimpleNamespace

import httpx
import pytest

import deepgram
import outputs
import runtime
import youtube

TESTS = Path(__file__).parent
FIXTURES = TESTS / "fixtures"
GOLDENS = TESTS / "goldens"


def deepgram_response(transcript_text: str = "Hello world") -> dict:
    """Build the smallest valid Deepgram response used by pipeline tests."""
    return {
        "results": {
            "channels": [
                {
                    "alternatives": [
                        {
                            "transcript": transcript_text,
                            "paragraphs": {
                                "paragraphs": [
                                    {
                                        "sentences": [
                                            {
                                                "start": 1.0,
                                                "end": 2.0,
                                                "text": transcript_text,
                                            }
                                        ]
                                    }
                                ]
                            },
                        }
                    ]
                }
            ]
        }
    }


def test_one_recorded_response_saves_the_golden_files(tmp_path, monkeypatch) -> None:
    audio = FIXTURES / "audio.webm"
    body = (FIXTURES / "deepgram_response.json").read_bytes()
    requests: list[httpx.Request] = []

    def answer(request: httpx.Request) -> httpx.Response:
        requests.append(request)
        return httpx.Response(
            200, content=body, headers={"Content-Type": "application/json"}
        )

    client = httpx.Client(transport=httpx.MockTransport(answer))
    monkeypatch.setattr(httpx, "post", client.post)

    response = deepgram.transcribe_audio(audio, "fake-key", runtime.RunBudget.start())
    outputs.save_outputs(tmp_path, *deepgram.parse_transcript(response))

    (request,) = requests
    assert request.url.path == "/v1/listen"
    assert request.headers["Authorization"] == "Token fake-key"
    assert request.headers["Content-Type"] == "audio/webm"
    assert request.content == audio.read_bytes()
    for name in ("raw_transcript.txt", "raw_sentences.txt", "raw_transcript.json"):
        assert (tmp_path / name).read_bytes() == (GOLDENS / name).read_bytes(), name
    assert sorted(path.name for path in tmp_path.iterdir()) == [
        "raw_sentences.txt",
        "raw_transcript.json",
        "raw_transcript.txt",
    ]


class TestDeepgramContract:
    def test_rejects_success_response_without_complete_upload(
        self, tmp_path, monkeypatch
    ):
        audio = tmp_path / "audio.mp3"
        audio.write_bytes(b"audio")
        monkeypatch.setattr(
            httpx,
            "post",
            lambda url, **kwargs: httpx.Response(
                200, json=deepgram_response(), request=httpx.Request("POST", url)
            ),
        )
        with pytest.raises(
            deepgram.DeepgramError, match="complete audio.*0/5 bytes"
        ) as error:
            deepgram.transcribe_audio(audio, "secret", runtime.RunBudget.start())
        assert error.value.code == "deepgram_unknown"

    def test_upload_timeout_reports_completed_bytes(self, tmp_path, monkeypatch):
        audio = tmp_path / "audio.mp3"
        audio.write_bytes(b"a" * 200_000)

        def stalled_post(_url, **kwargs):
            chunks = iter(kwargs["content"])
            next(chunks)
            next(chunks)
            raise httpx.WriteTimeout("The write operation timed out")

        monkeypatch.setattr(httpx, "post", stalled_post)
        with pytest.raises(deepgram.DeepgramError, match="65536/200000 bytes") as error:
            deepgram.transcribe_audio(audio, "secret", runtime.RunBudget.start())
        assert error.value.code == "deepgram_unknown"

    def test_upload_stops_at_workflow_deadline(self, tmp_path, monkeypatch):
        audio = tmp_path / "audio.mp3"
        audio.write_bytes(b"a" * 200_000)
        now = [0.0]
        monkeypatch.setattr(runtime.time, "monotonic", lambda: now[0])

        def slow_post(_url, **kwargs):
            chunks = iter(kwargs["content"])
            next(chunks)
            now[0] = 11.0
            next(chunks)

        monkeypatch.setattr(httpx, "post", slow_post)
        with pytest.raises(deepgram.DeepgramError, match="upload") as error:
            deepgram.transcribe_audio(audio, "secret", runtime.RunBudget(10))
        assert error.value.code == "deepgram_unknown"

    def test_upload_streams_bounded_chunks_without_changing_audio(
        self, tmp_path, monkeypatch, caplog
    ) -> None:
        audio = tmp_path / "audio.mp3"
        payload = b"audio-data" * 100_000
        audio.write_bytes(payload)
        now = [0.0]
        monkeypatch.setattr(runtime.time, "monotonic", lambda: now[0])

        def inspect_post(url, *, params, headers, content, timeout):
            request = httpx.Request(
                "POST", url, params=params, headers=headers, content=content
            )
            assert isinstance(request.stream, httpx.SyncByteStream)
            chunks = []
            for chunk in request.stream:
                transfer_seconds = len(chunk) / 2500
                assert transfer_seconds < timeout
                now[0] += transfer_seconds
                chunks.append(chunk)
            assert now[0] > timeout
            assert max(map(len, chunks)) <= 64 * 1024
            assert b"".join(chunks) == payload
            assert request.headers["Content-Length"] == str(len(payload))
            assert "Transfer-Encoding" not in request.headers
            return httpx.Response(200, json=deepgram_response(), request=request)

        monkeypatch.setattr(httpx, "post", inspect_post)
        caplog.set_level(logging.INFO, logger="transcript")
        deepgram.transcribe_audio(audio, "secret", runtime.RunBudget.start())
        assert caplog.messages == [
            (
                "Audio upload complete: 1000000/1000000 bytes sent. "
                "Waiting for Deepgram transcription."
            )
        ]

    @pytest.mark.parametrize(
        ("name", "content_type"),
        [("audio.webm", "audio/webm"), ("audio.m4a", "audio/mp4")],
    )
    def test_native_audio_keeps_its_content_type(self, name, content_type) -> None:
        assert deepgram.get_audio_content_type(Path(name)) == content_type

    def test_parses_valid_response(self) -> None:
        plain, timestamped, _json_data = deepgram.parse_transcript(deepgram_response())

        assert plain == "Hello world"
        assert timestamped == "[1s - 2s] Hello world"

    @pytest.mark.parametrize(
        "response",
        [
            {},
            deepgram_response(""),
            {
                "results": {
                    "channels": [
                        {"alternatives": [{"transcript": "text", "paragraphs": None}]}
                    ]
                }
            },
        ],
    )
    def test_rejects_incomplete_or_empty_response(self, response) -> None:
        with pytest.raises(ValueError, match="Deepgram response"):
            deepgram.parse_transcript(response)


def test_every_external_call_receives_a_bounded_timeout(tmp_path, monkeypatch) -> None:
    now = [0.0]
    subprocess_timeouts = []
    http_timeouts = []
    audio = tmp_path / "audio.mp3"
    audio.write_bytes(b"audio")
    responses = [
        SimpleNamespace(stdout="", stderr="public failure", returncode=1),
        SimpleNamespace(stdout="Title\nabc\n", stderr="", returncode=0),
    ]

    def fake_run(_command, **kwargs):
        subprocess_timeouts.append(kwargs["timeout"])
        now[0] += 1
        return responses.pop(0)

    def fake_post(*_args, **kwargs):
        http_timeouts.append(kwargs["timeout"])
        assert b"".join(kwargs["content"]) == b"audio"
        return SimpleNamespace(
            is_success=True,
            status_code=200,
            json=lambda: deepgram_response(),
        )

    monkeypatch.setattr(runtime.time, "monotonic", lambda: now[0])
    monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", tmp_path)
    monkeypatch.setattr(youtube, "run_child", fake_run)
    monkeypatch.setattr(httpx, "post", fake_post)
    budget = runtime.RunBudget(20)

    youtube.run_ytdlp(
        ["--get-title", "--get-id"], "url", budget, browser="arc", temp_dir=tmp_path
    )
    deepgram.transcribe_audio(audio, "secret", budget)

    assert all(0 < timeout <= 20 for timeout in subprocess_timeouts)
    assert subprocess_timeouts[1] < subprocess_timeouts[0]
    assert 0 < http_timeouts[0] <= 18
