"""The program code of each failed Deepgram request: one upload at most,
and no retry."""

from __future__ import annotations

import httpx
import pytest

import deepgram
import runtime


def answering(monkeypatch, response: httpx.Response | Exception) -> list:
    """Make Deepgram read the whole upload, then answer `response` or raise
    it, and record each request."""
    requests = []

    def answer(request: httpx.Request) -> httpx.Response:
        requests.append(request)
        if isinstance(response, Exception):
            raise response
        return response

    client = httpx.Client(transport=httpx.MockTransport(answer))
    monkeypatch.setattr(httpx, "post", client.post)
    return requests


@pytest.mark.parametrize(
    ("status", "code"),
    [
        (408, "deepgram_unreachable"),
        (429, "deepgram_unreachable"),
        (401, "deepgram_key_rejected"),
        (403, "deepgram_key_rejected"),
        (400, "deepgram_failed"),
        (404, "deepgram_failed"),
        (500, "deepgram_unknown"),
        (503, "deepgram_unknown"),
        (504, "deepgram_unknown"),
    ],
)
def test_each_refusal_gets_its_program_code_after_one_upload(
    tmp_path, monkeypatch, status, code
) -> None:
    audio = tmp_path / "audio.webm"
    audio.write_bytes(b"audio")
    requests = answering(monkeypatch, httpx.Response(status, json={"err_code": "X"}))

    with pytest.raises(deepgram.DeepgramError) as error:
        deepgram.transcribe_audio(audio, "secret", runtime.RunBudget.start())

    assert error.value.code == code
    assert len(requests) == 1


@pytest.mark.parametrize(
    ("raised", "code"),
    [
        (httpx.ConnectError("refused"), "deepgram_unreachable"),
        (httpx.ConnectTimeout("no route"), "deepgram_unreachable"),
        (httpx.ReadTimeout("no answer"), "deepgram_unknown"),
        (httpx.RemoteProtocolError("lost"), "deepgram_unknown"),
    ],
)
def test_each_transport_failure_gets_its_program_code(
    tmp_path, monkeypatch, raised, code
) -> None:
    audio = tmp_path / "audio.webm"
    audio.write_bytes(b"audio")
    requests = answering(monkeypatch, raised)

    with pytest.raises(deepgram.DeepgramError) as error:
        deepgram.transcribe_audio(audio, "secret", runtime.RunBudget.start())

    assert error.value.code == code
    assert len(requests) == 1


def test_a_deadline_before_the_upload_sends_nothing(tmp_path, monkeypatch) -> None:
    audio = tmp_path / "audio.webm"
    audio.write_bytes(b"audio")
    requests = answering(monkeypatch, httpx.Response(200, json={}))
    monkeypatch.setattr(runtime.time, "monotonic", lambda: 10.0)

    with pytest.raises(deepgram.DeepgramError) as error:
        deepgram.transcribe_audio(audio, "secret", runtime.RunBudget(10))

    assert error.value.code == "deepgram_unreachable"
    assert requests == []


@pytest.mark.parametrize(
    "response",
    [
        httpx.Response(200, content=b"not json"),
        httpx.Response(200, json={"error": "bad audio"}),
        httpx.Response(200, json=["a list"]),
    ],
)
def test_an_unusable_success_answer_is_transcript_unusable(
    tmp_path, monkeypatch, response
) -> None:
    audio = tmp_path / "audio.webm"
    audio.write_bytes(b"audio")
    answering(monkeypatch, response)

    with pytest.raises(deepgram.DeepgramError) as error:
        deepgram.transcribe_audio(audio, "secret", runtime.RunBudget.start())

    assert error.value.code == "transcript_unusable"
