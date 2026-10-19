"""The JSON Lines messages that the engine exchanges with Go.

Each message is one JSON object on one line, with `protocol_version` and
`type`. Go sends a request, then an ack; the engine answers with a plan, an
intent, and one terminal step, as the protocol spec of transcript defines.
"""

import json
import queue
import threading
from collections.abc import Iterable
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any, TextIO

PROTOCOL_VERSION = 1
STEP = "transcribe"


class ProtocolError(ValueError):
    """A message from Go that breaks the protocol."""


def send(stream: TextIO, kind: str, **fields: Any) -> None:
    """Write one message and flush it, so Go reads it at once."""
    message = {"protocol_version": PROTOCOL_VERSION, "type": kind, **fields}
    stream.write(json.dumps(message) + "\n")
    stream.flush()


def parse(line: str, kind: str) -> dict[str, Any]:
    """The message of `line`, which must be of type `kind`."""
    try:
        message = json.loads(line)
    except json.JSONDecodeError as error:
        raise ProtocolError(f"malformed message: {line[:200]!r}") from error
    if not isinstance(message, dict):
        raise ProtocolError(f"a message is not an object: {line[:200]!r}")
    if message.get("protocol_version") != PROTOCOL_VERSION:
        raise ProtocolError(
            f"unsupported protocol_version: {message.get('protocol_version')!r}"
        )
    if message.get("type") != kind:
        raise ProtocolError(f"expected a {kind}, got {message.get('type')!r}")
    return message


class Inbox:
    """The lines that Go writes on stdin, read by a thread, so a wait for one
    can end after a timeout. A closed stdin ends every later wait at once."""

    def __init__(self, stream: Iterable[str]) -> None:
        self._lines: queue.Queue[str | None] = queue.Queue()
        threading.Thread(target=self._read, args=(stream,), daemon=True).start()
        self._closed = False

    def _read(self, stream: Iterable[str]) -> None:
        for line in stream:
            self._lines.put(line)
        self._lines.put(None)

    def next(self, timeout: float | None) -> str | None:
        """The next line, or None once stdin is closed or `timeout` passes."""
        if self._closed:
            return None
        try:
            line = self._lines.get(timeout=timeout)
        except queue.Empty:
            return None
        self._closed = line is None
        return line


@dataclass(frozen=True)
class Request:
    """The request of Go, checked: one YouTube URL and its folders."""

    url: str
    output: Path
    temp_dir: Path
    browser: str | None
    dry_run: bool
    deadline: datetime

    @classmethod
    def parse(cls, line: str | None) -> "Request":
        if line is None:
            raise ProtocolError("stdin closed before the request")
        message = parse(line, "request")
        urls = message.get("urls")
        if message.get("source") != "youtube" or message.get("zoom") is not None:
            raise ProtocolError("the request names no YouTube source")
        if not isinstance(urls, list) or len(urls) != 1 or not isinstance(urls[0], str):
            raise ProtocolError("the request holds no single URL")
        output, temp_dir = message.get("output"), message.get("temp_dir")
        if not isinstance(output, str) or not isinstance(temp_dir, str):
            raise ProtocolError("the request holds no output or temp_dir")
        if not Path(output).is_absolute() or not Path(temp_dir).is_absolute():
            raise ProtocolError("the request holds a relative folder")
        if message.get("browser") not in {"arc", None}:
            raise ProtocolError(f"unknown browser: {message.get('browser')!r}")
        if not isinstance(message.get("dry_run"), bool):
            raise ProtocolError("the request holds no dry_run boolean")
        try:
            deadline = datetime.fromisoformat(message["deadline"])
        except (KeyError, TypeError, ValueError) as error:
            raise ProtocolError("the request holds no RFC 3339 deadline") from error
        if deadline.tzinfo is None:
            raise ProtocolError("the deadline names no time zone")
        return cls(
            urls[0],
            Path(output),
            Path(temp_dir),
            message["browser"],
            message["dry_run"],
            deadline,
        )


def parse_ack(line: str, item: str) -> Path:
    """The result folder of an ack for `item`, which Go has reserved."""
    message = parse(line, "ack")
    if message.get("item") != item or message.get("step") != STEP:
        raise ProtocolError("the ack names another item or step")
    output_dir = Path(message.get("output_dir") or "")
    if not output_dir.is_absolute() or not output_dir.is_dir():
        raise ProtocolError(f"the ack names no folder: {output_dir}")
    return output_dir
