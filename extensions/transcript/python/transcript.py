# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "httpx==0.28.1",
#     "yt-dlp>=2026.7.4",
#     "pycryptodomex==3.24.1",
# ]
# ///
"""Transcribe the audio of one YouTube video with Deepgram.

uv runs this script with the dependencies of its header, which hold every
runtime pin of the engine. It speaks the JSON Lines protocol of transcript on
stdin and stdout, and writes diagnostics only on stderr. Its sibling modules
download the audio with yt-dlp, send it to Deepgram, and save the files.
"""

import logging
import os
import subprocess
import sys
from collections.abc import Iterable, Sequence
from datetime import datetime
from pathlib import Path
from typing import TextIO

from deepgram import DeepgramError, parse_transcript, transcribe_audio
from outputs import Metadata, save_outputs
from protocol import STEP, Inbox, ProtocolError, Request, parse_ack, send
from runtime import RunBudget, retry_request
from youtube import YtDlpError, clean_title, download_audio

__version__ = "1.0.0"

# How long the engine waits for the ack of its intent before it stops
ACK_TIMEOUT = 10.0
# Exit codes for a broken protocol; a reported item outcome exits 0
PROTOCOL_BROKEN = 2
NO_ACK = 3

# The step status that each program code leaves, as the protocol spec lists
STATUS = {
    "download_temporary": "pending",
    "download_failed": "pending",
    "deepgram_unreachable": "failed",
    "deepgram_key_rejected": "failed",
    "deepgram_failed": "failed",
    "deepgram_unknown": "unknown",
    "transcript_unusable": "unknown",
}

log = logging.getLogger("transcript")


def configure_logging() -> None:
    """Send the engine's diagnostics to stderr: debug lines with NCLY_DEBUG,
    progress with NCLY_VERBOSE, and warnings otherwise."""
    handler = logging.StreamHandler(sys.stderr)
    handler.setFormatter(logging.Formatter("transcript: %(message)s"))
    log.addHandler(handler)
    log.propagate = False
    if os.environ.get("NCLY_DEBUG"):
        log.setLevel(logging.DEBUG)
    elif os.environ.get("NCLY_VERBOSE") == "1":
        log.setLevel(logging.INFO)
    else:
        log.setLevel(logging.WARNING)


def ytdlp_version() -> str:
    """The version of the yt-dlp package that the engine runs."""
    from yt_dlp.version import __version__ as version

    return version


def main(stdin: Iterable[str] = sys.stdin, stdout: TextIO = sys.stdout) -> int:
    """Answer one request of Go, then exit: 0 once the plan or the terminal
    step is sent, and nonzero when the protocol breaks."""
    inbox = Inbox(stdin)
    try:
        request = Request.parse(inbox.next(None))
    except ProtocolError as error:
        log.error(f"protocol: {error}")
        return PROTOCOL_BROKEN
    item = request.url
    send(stdout, "plan", items=[{"item": item, "output": str(request.output)}])
    if request.dry_run:
        return 0
    budget = RunBudget.until(request.deadline)

    def step(
        status: str, code: str | None, warnings: Sequence[str] = (), detail: str = ""
    ) -> int:
        send(
            stdout,
            "step",
            item=item,
            step=STEP,
            status=status,
            code=code,
            warnings=list(warnings),
            detail=detail,
        )
        return 0

    try:
        audio = retry_request(
            lambda: download_audio(
                item, request.temp_dir, budget, browser=request.browser
            ),
            deadline=budget.deadline,
        )
    except YtDlpError as error:
        code = "download_temporary" if error.temporary else "download_failed"
        return step("pending", code, error.warnings, str(error))
    except (TimeoutError, subprocess.TimeoutExpired) as error:
        return step("pending", "download_temporary", (), str(error) or "timed out")
    except ValueError as error:
        return step("pending", "download_failed", (), str(error))
    except OSError as error:
        return step("pending", "write_failed", (), str(error))

    tools = {"transcript": __version__, "yt-dlp": ytdlp_version()}
    name = clean_title(audio.title) or audio.video_id
    send(stdout, "intent", item=item, step=STEP, name=name, tools=tools)
    line = inbox.next(ACK_TIMEOUT)
    if line is None:
        log.error("protocol: no ack arrived; stopping before any effect")
        return NO_ACK
    try:
        output_dir = parse_ack(line, item)
    except ProtocolError as error:
        log.error(f"protocol: {error}")
        return PROTOCOL_BROKEN

    meta = Metadata(audio.title, item, audio.method, datetime.now().astimezone())
    outcome = transcribe_into(audio.path, output_dir, budget, meta)
    meta.save(output_dir)
    status, code, detail = outcome
    return step(status, code, audio.warnings, detail)


def transcribe_into(
    audio_path: Path, output_dir: Path, budget: RunBudget, meta: Metadata
) -> tuple[str, str | None, str]:
    """Send the audio to Deepgram once and save the three files: the step's
    status, program code, and diagnostic. From the upload on, a failure may
    have been billed."""
    key = os.environ.get("DEEPGRAM_API_KEY", "")
    try:
        response = transcribe_audio(audio_path, key, budget)
        meta.audio_bytes = audio_path.stat().st_size
        texts = parse_transcript(response)
    except DeepgramError as error:
        meta.transcript_status, meta.error = "failed", str(error)
        return STATUS[error.code], error.code, str(error)
    except ValueError as error:
        meta.transcript_status, meta.error = "failed", str(error)
        return "unknown", "transcript_unusable", str(error)
    try:
        save_outputs(output_dir, *texts)
    except OSError as error:
        meta.transcript_status, meta.error = "failed", str(error)
        return "unknown", "write_failed", str(error)
    meta.transcript_status = "complete"
    return "completed", None, ""


if __name__ == "__main__":
    configure_logging()
    raise SystemExit(main())
