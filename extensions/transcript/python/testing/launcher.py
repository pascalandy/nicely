"""Run the engine with the test transport of spec/test-transport.md.

Only the test binary of ncly runs this file, as uv runs
`--with-requirements <engine>/transcript.py launcher.py <engine>`. yt-dlp
copies the fixture audio, and Deepgram is an httpx.MockTransport that reads
the whole upload, then answers the fixture response. Nothing leaves the
machine.

A scenario steers it through its environment, where an empty variable
counts as unset:

- NCLY_TEST_TITLE is the title of the video, "Test video" by default
- NCLY_TEST_DOWNLOAD=temporary or failed makes yt-dlp fail, for a network
  reason or for another one
- NCLY_TEST_COOKIES=skipped downloads anonymously, with the warning of
  cookies that could not be read
- NCLY_TEST_DEEPGRAM_STATUS is the HTTP status of Deepgram, 200 by default
- NCLY_TEST_DEEPGRAM_BODY=garbage makes Deepgram answer 200 with no JSON
- NCLY_TEST_PAUSE=upload makes Deepgram wait before it answers the whole
  upload. It writes the process IDs of Python and of uv, its parent, to
  $WORK/engine.pid and $WORK/uv.pid, then waits until $WORK/release exists,
  for at most 60 seconds

It appends one line per start to $WORK/starts.log, one per download to
$WORK/downloads.log, and one per upload to $WORK/uploads.log, without the
key.
"""

from __future__ import annotations

import os
import shutil
import sys
import time
from pathlib import Path
from typing import Any

import httpx

FIXTURES = Path(__file__).resolve().parent.parent / "tests" / "fixtures"


def log(name: str, text: str) -> None:
    with (Path(os.environ["WORK"]) / name).open("a", encoding="utf-8") as file:
        file.write(text + "\n")


def pause() -> None:
    work = Path(os.environ["WORK"])
    (work / "uv.pid").write_text(f"{os.getppid()}\n")
    (work / "engine.pid").write_text(f"{os.getpid()}\n")
    for _ in range(1200):
        if (work / "release").exists():
            return
        time.sleep(0.05)


def answer(request: httpx.Request) -> httpx.Response:
    body = request.read()
    authorized = request.headers.get("Authorization", "").startswith("Token ")
    log("uploads.log", f"upload {len(body)} bytes, authorized {authorized}")
    if os.environ.get("NCLY_TEST_PAUSE") == "upload":
        pause()
    status = int(os.environ.get("NCLY_TEST_DEEPGRAM_STATUS") or "200")
    if status != 200:
        return httpx.Response(status, json={"err_msg": "test transport"})
    if os.environ.get("NCLY_TEST_DEEPGRAM_BODY") == "garbage":
        return httpx.Response(200, content=b"no transcript")
    return httpx.Response(
        200, content=(FIXTURES / "deepgram_response.json").read_bytes()
    )


def main() -> int:
    # The engine's modules sit in the folder that Go extracted, not beside
    # this file.
    sys.path.insert(0, sys.argv[1])
    log("starts.log", "start")
    import transcript
    from youtube import COOKIES_SKIPPED, DownloadedAudio, YtDlpError

    def download_audio(
        url: str, temp_dir: Path, budget: object, *, browser: str | None
    ) -> DownloadedAudio:
        path = temp_dir / "audio.webm"
        shutil.copyfile(FIXTURES / "audio.webm", path)
        log("downloads.log", url)
        failure = os.environ.get("NCLY_TEST_DOWNLOAD")
        if failure in ("temporary", "failed"):
            raise YtDlpError("test transport", temporary=failure == "temporary")
        warnings = ()
        if os.environ.get("NCLY_TEST_COOKIES") == "skipped":
            warnings = (COOKIES_SKIPPED,)
        title = os.environ.get("NCLY_TEST_TITLE") or "Test video"
        return DownloadedAudio(path, "anonymous", title, "EIEc43CxIvY", warnings)

    client = httpx.Client(transport=httpx.MockTransport(answer))

    def post(*args: Any, **kwargs: Any) -> httpx.Response:
        return client.post(*args, **kwargs)

    transcript.download_audio = download_audio
    httpx.post = post
    transcript.configure_logging()
    return transcript.main()


if __name__ == "__main__":
    raise SystemExit(main())
