"""Drive the engine's main() in the test process, as Go drives the script:
JSON lines in, JSON lines out, with yt-dlp and Deepgram replaced."""

from __future__ import annotations

import io
import json
import threading
from collections.abc import Iterable, Iterator
from datetime import UTC, datetime, timedelta
from pathlib import Path
from types import SimpleNamespace

import httpx

import transcript
import youtube

URL = "https://www.youtube.com/watch?v=EIEc43CxIvY"
FIXTURES = Path(__file__).parent / "fixtures"
GOLDENS = Path(__file__).parent / "goldens"


def line(kind: str, **fields) -> str:
    return json.dumps({"protocol_version": 1, "type": kind, **fields}) + "\n"


def request(tmp_path: Path, **fields) -> str:
    """A request for URL, with the folders of the test."""
    (tmp_path / "out").mkdir(exist_ok=True)
    (tmp_path / "tmp").mkdir(exist_ok=True)
    deadline = datetime.now(UTC) + timedelta(seconds=600)
    values = {
        "source": "youtube",
        "urls": [URL],
        "zoom": None,
        "output": str(tmp_path / "out"),
        "temp_dir": str(tmp_path / "tmp"),
        "browser": None,
        "dry_run": False,
        "deadline": deadline.isoformat().replace("+00:00", "Z"),
    }
    return line("request", **{**values, **fields})


def ack(tmp_path: Path, item: str = URL) -> str:
    """The ack of Go, which reserved out/A_video."""
    folder = tmp_path / "out" / "A_video"
    folder.mkdir(parents=True, exist_ok=True)
    return line("ack", item=item, step="transcribe", output_dir=str(folder))


class Held:
    """A stdin that gives its lines, then stays open until `release` is set,
    as Go's stdin does while Go prepares the ack."""

    def __init__(self, *lines: str) -> None:
        self.lines = lines
        self.release = threading.Event()

    def __iter__(self) -> Iterator[str]:
        yield from self.lines
        self.release.wait()


def run(*lines: str, stdin: Iterable[str] | None = None) -> tuple[int, list[dict]]:
    """Run main() on `lines`, and return its exit code and its messages."""
    stdout = io.StringIO()
    code = transcript.main(stdin or io.StringIO("".join(lines)), stdout)
    return code, [json.loads(text) for text in stdout.getvalue().splitlines()]


def fake_ytdlp(monkeypatch, stderr: str = "", title: str = "A video") -> list:
    """Make yt-dlp copy the fixture audio where -o asks, or fail with
    `stderr`, and record each command."""
    commands = []

    def fake_run(command, **_kwargs):
        commands.append(command)
        if stderr:
            return SimpleNamespace(returncode=1, stdout="", stderr=stderr)
        target = Path(command[command.index("-o") + 1].replace("%(ext)s", "webm"))
        target.write_bytes((FIXTURES / "audio.webm").read_bytes())
        info = {"id": "EIEc43CxIvY", "title": title, "filepath": str(target)}
        return SimpleNamespace(returncode=0, stdout=json.dumps(info) + "\n", stderr="")

    monkeypatch.setattr(youtube, "run_child", fake_run)
    return commands


def fake_deepgram(monkeypatch, response: httpx.Response | None = None) -> list:
    """Make Deepgram read the whole upload and answer `response`, or the
    fixture response, and record each upload."""
    body = (FIXTURES / "deepgram_response.json").read_bytes()
    uploads = []

    def answer(request: httpx.Request) -> httpx.Response:
        uploads.append(request)
        return response or httpx.Response(200, content=body)

    client = httpx.Client(transport=httpx.MockTransport(answer))
    monkeypatch.setattr(httpx, "post", client.post)
    monkeypatch.setenv("DEEPGRAM_API_KEY", "fake-key")
    return uploads
