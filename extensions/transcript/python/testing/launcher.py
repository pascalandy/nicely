"""Run the engine with the test transport of spec/protocol.md.

Only the test binary of ncly runs this file, as uv runs
`--with-requirements <engine>/transcript.py launcher.py <engine>`. yt-dlp
copies the fixture audio, and Deepgram is an httpx.MockTransport that reads
the whole upload, then answers the fixture response. Nothing leaves the
machine.

A scenario steers it through its environment:

- NCLY_TEST_TITLE is the title of the video, "Test video" by default
- NCLY_TEST_DEEPGRAM_STATUS is the HTTP status of Deepgram, 200 by default

It appends one line per download to $WORK/downloads.log and one line per
upload to $WORK/uploads.log, without the key.
"""

from __future__ import annotations

import os
import shutil
import sys
from pathlib import Path
from typing import Any

import httpx

FIXTURES = Path(__file__).resolve().parent.parent / "tests" / "fixtures"


def log(name: str, text: str) -> None:
    with (Path(os.environ["WORK"]) / name).open("a", encoding="utf-8") as file:
        file.write(text + "\n")


def answer(request: httpx.Request) -> httpx.Response:
    body = request.read()
    authorized = request.headers.get("Authorization", "").startswith("Token ")
    log("uploads.log", f"upload {len(body)} bytes, authorized {authorized}")
    status = int(os.environ.get("NCLY_TEST_DEEPGRAM_STATUS", "200"))
    if status != 200:
        return httpx.Response(status, json={"err_msg": "test transport"})
    return httpx.Response(
        200, content=(FIXTURES / "deepgram_response.json").read_bytes()
    )


def main() -> int:
    # The engine's modules sit in the folder that Go extracted, not beside
    # this file.
    sys.path.insert(0, sys.argv[1])
    import transcript
    from youtube import DownloadedAudio

    def download_audio(
        url: str, temp_dir: Path, budget: object, *, browser: str | None
    ) -> DownloadedAudio:
        path = temp_dir / "audio.webm"
        shutil.copyfile(FIXTURES / "audio.webm", path)
        log("downloads.log", url)
        title = os.environ.get("NCLY_TEST_TITLE", "Test video")
        return DownloadedAudio(path, "anonymous", title, "EIEc43CxIvY")

    client = httpx.Client(transport=httpx.MockTransport(answer))

    def post(*args: Any, **kwargs: Any) -> httpx.Response:
        return client.post(*args, **kwargs)

    transcript.download_audio = download_audio
    httpx.post = post
    transcript.configure_logging()
    return transcript.main()


if __name__ == "__main__":
    raise SystemExit(main())
