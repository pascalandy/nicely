"""Send one audio file to Deepgram and parse its answer into three texts."""

import json
import logging
import time
from collections.abc import Iterator
from pathlib import Path

import httpx

from runtime import RunBudget

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


class DeepgramResponseError(ValueError):
    """Deepgram returned a successful response with an unusable schema."""


def get_audio_content_type(audio_path: Path) -> str:
    """Return the best Deepgram content type for a local audio file."""
    return AUDIO_CONTENT_TYPES.get(
        audio_path.suffix.lower(), "application/octet-stream"
    )


def transcribe_audio(
    audio_path: Path,
    api_key: str,
    budget: RunBudget,
) -> dict:
    """Transcribe audio using Deepgram API (nova-3 model)."""
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

        started = time.monotonic()
        try:
            response = httpx.post(
                url,
                params=params,
                headers=headers,
                content=chunks(),
                timeout=budget.remaining("Deepgram transcription", DEEPGRAM_TIMEOUT),
            )
        except httpx.WriteTimeout as error:
            raise httpx.WriteTimeout(
                "Deepgram audio upload stalled. "
                f"At least {sent_bytes}/{total_bytes} bytes sent; "
                "the current block may be partially sent."
            ) from error

    log.debug(
        "Deepgram answered %s after %.1fs",
        getattr(response, "status_code", "?"),
        time.monotonic() - started,
    )
    response.raise_for_status()
    if sent_bytes != total_bytes:
        raise ValueError(
            f"Deepgram responded before the complete audio was sent: "
            f"{sent_bytes}/{total_bytes} bytes"
        )
    result = response.json()

    if "error" in result:
        raise ValueError(f"Deepgram error: {result['error']}")

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
