"""Send one audio file to Deepgram and parse its answer into three texts."""

import json
import logging
import time
from collections.abc import Iterator
from pathlib import Path

import httpx

from runtime import RunBudget, WorkflowTimeoutError

DEEPGRAM_TIMEOUT = 300

AUDIO_CONTENT_TYPES = {
    ".m4a": "audio/mp4",
    ".mp4": "audio/mp4",
    ".mp3": "audio/mp3",
    ".webm": "audio/webm",
    ".opus": "audio/ogg",
    ".ogg": "audio/ogg",
}

log = logging.getLogger("transcript")


class DeepgramError(RuntimeError):
    """A failed Deepgram request, with its program code.

    `deepgram_unreachable`, `deepgram_key_rejected`, and `deepgram_failed`
    mean that Deepgram took no audio or refused it. `deepgram_unknown` means
    that Deepgram may have received the audio, and `transcript_unusable` that
    its success answer holds no usable transcript.
    """

    def __init__(self, code: str, message: str) -> None:
        super().__init__(message)
        self.code = code


class DeepgramResponseError(ValueError):
    """Deepgram returned a successful response with an unusable schema."""


def get_audio_content_type(audio_path: Path) -> str:
    """Return the best Deepgram content type for a local audio file."""
    return AUDIO_CONTENT_TYPES.get(
        audio_path.suffix.lower(), "application/octet-stream"
    )


def status_code(status: int) -> str:
    """The program code of a Deepgram answer that is not a success."""
    if status in {408, 429}:
        return "deepgram_unreachable"
    if status in {401, 403}:
        return "deepgram_key_rejected"
    if 400 <= status < 500:
        return "deepgram_failed"
    return "deepgram_unknown"


def transcribe_audio(
    audio_path: Path,
    api_key: str,
    budget: RunBudget,
) -> dict:
    """Transcribe audio using Deepgram API (nova-3 model), in one request.

    No retry ever repeats the upload. Every failure raises a DeepgramError.
    """
    url = "https://api.deepgram.com/v1/listen"
    params = {
        "model": "nova-3",
        "detect_language": "true",
        "punctuate": "true",
        "paragraphs": "true",
    }
    total_bytes = audio_path.stat().st_size
    sent_bytes = 0
    headers = {
        "Authorization": f"Token {api_key}",
        "Content-Type": get_audio_content_type(audio_path),
        "Content-Length": str(total_bytes),
    }

    with open(audio_path, "rb") as audio_file:

        def chunks() -> Iterator[bytes]:
            nonlocal sent_bytes
            while chunk := audio_file.read(64 * 1024):
                budget.remaining("Deepgram audio upload")
                yield chunk
                sent_bytes += len(chunk)
            if sent_bytes != total_bytes:
                raise ValueError("Audio file size changed during Deepgram upload")
            log.info(
                f"Audio upload complete: {sent_bytes}/{total_bytes} bytes sent. "
                "Waiting for Deepgram transcription."
            )

        try:
            timeout = budget.remaining("Deepgram transcription", DEEPGRAM_TIMEOUT)
        except WorkflowTimeoutError as error:
            raise DeepgramError("deepgram_unreachable", str(error)) from error
        started = time.monotonic()
        try:
            response = httpx.post(
                url,
                params=params,
                headers=headers,
                content=chunks(),
                timeout=timeout,
            )
        except (httpx.ConnectError, httpx.ConnectTimeout) as error:
            raise DeepgramError(
                "deepgram_unreachable", f"Deepgram is unreachable: {error}"
            ) from error
        except httpx.WriteTimeout as error:
            raise DeepgramError(
                "deepgram_unknown",
                "Deepgram audio upload stalled. "
                f"At least {sent_bytes}/{total_bytes} bytes sent; "
                "the current block may be partially sent.",
            ) from error
        except (httpx.HTTPError, WorkflowTimeoutError, ValueError) as error:
            raise DeepgramError(
                "deepgram_unknown",
                f"{error or type(error).__name__}; "
                f"{sent_bytes}/{total_bytes} bytes sent",
            ) from error

    log.debug(
        "Deepgram answered %s after %.1fs",
        response.status_code,
        time.monotonic() - started,
    )
    if not response.is_success:
        raise DeepgramError(
            status_code(response.status_code),
            f"Deepgram answered HTTP {response.status_code}",
        )
    if sent_bytes != total_bytes:
        raise DeepgramError(
            "deepgram_unknown",
            f"Deepgram responded before the complete audio was sent: "
            f"{sent_bytes}/{total_bytes} bytes",
        )
    try:
        result = response.json()
    except ValueError as error:
        raise DeepgramError(
            "transcript_unusable", "Deepgram answered with invalid JSON"
        ) from error
    if not isinstance(result, dict) or "error" in result:
        raise DeepgramError(
            "transcript_unusable", f"Deepgram error: {str(result)[:200]}"
        )
    return result


def parse_transcript(response: dict) -> tuple[str, str, str]:
    """Parse Deepgram response into different output formats."""
    try:
        channel = response["results"]["channels"][0]["alternatives"][0]
        transcript_text = channel["transcript"]
    except (KeyError, IndexError, TypeError) as error:
        raise ValueError("Deepgram response is missing a transcript") from error

    if not isinstance(transcript_text, str) or not transcript_text.strip():
        raise ValueError("Deepgram response contains an empty transcript")

    paragraphs_container = channel.get("paragraphs")
    if not isinstance(paragraphs_container, dict):
        raise DeepgramResponseError("Deepgram response is missing paragraph data")
    paragraphs = paragraphs_container.get("paragraphs")
    if not isinstance(paragraphs, list):
        raise DeepgramResponseError("Deepgram response contains invalid paragraph data")

    sentences_lines = []
    for paragraph in paragraphs:
        if not isinstance(paragraph, dict):
            raise DeepgramResponseError(
                "Deepgram response contains an invalid paragraph"
            )
        sentences = paragraph.get("sentences")
        if not isinstance(sentences, list):
            raise DeepgramResponseError(
                "Deepgram response contains invalid sentence data"
            )
        for sentence in sentences:
            try:
                start = int(sentence["start"])
                end = int(sentence["end"])
                text = sentence["text"]
            except (KeyError, TypeError, ValueError) as error:
                raise ValueError(
                    "Deepgram response contains an invalid sentence"
                ) from error
            sentences_lines.append(f"[{start}s - {end}s] {text}")
    sentences_text = "\n".join(sentences_lines)

    # JSON data (paragraphs structure)
    json_data = json.dumps(paragraphs, indent=2)

    return transcript_text, sentences_text, json_data
