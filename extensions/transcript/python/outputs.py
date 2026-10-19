"""Save the three transcript texts and the advisory meta.txt into the result
folder."""

from contextlib import suppress
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

SAVE_OUTPUT_NAMES = {
    "transcript": "raw_transcript.txt",
    "sentences": "raw_sentences.txt",
    "json": "raw_transcript.json",
}


def write_text_atomic(path: Path, content: str) -> None:
    """Replace a text artifact only after its complete content is on disk."""
    temporary_path = path.with_name(f".{path.name}.tmp")
    try:
        temporary_path.write_text(content, encoding="utf-8")
        temporary_path.replace(path)
    finally:
        temporary_path.unlink(missing_ok=True)


def save_outputs(
    output_dir: Path,
    transcript: str,
    sentences: str,
    json_data: str,
) -> dict[str, Path]:
    """Save all output files."""
    files = {}

    transcript_path = output_dir / SAVE_OUTPUT_NAMES["transcript"]
    write_text_atomic(transcript_path, transcript)
    files["transcript"] = transcript_path

    sentences_path = output_dir / SAVE_OUTPUT_NAMES["sentences"]
    write_text_atomic(sentences_path, sentences)
    files["sentences"] = sentences_path

    json_path = output_dir / SAVE_OUTPUT_NAMES["json"]
    write_text_atomic(json_path, json_data)
    files["json"] = json_path

    return files


@dataclass
class Metadata:
    """The run state that meta.txt shows to a human. No record lists the file,
    and replay never reads it."""

    title: str
    source: str
    method: str
    started: datetime
    transcript_status: str = "in progress"
    audio_bytes: int | None = None
    error: str | None = None

    def save(self, output_dir: Path) -> None:
        """Write meta.txt. A write failure is ignored, since the file is
        advisory and the reported outcome must stay the original one."""
        lines = [
            f"Title: {self.title}",
            f"Date: {self.started.strftime('%Y_%m_%d %Hh%M')}",
            f"Source: {self.source}",
            f"YouTube audio method: {self.method}",
            f"Transcript status: {self.transcript_status}",
        ]
        if self.audio_bytes is not None:
            lines.append(f"Audio upload: complete ({self.audio_bytes} bytes)")
        if self.error:
            lines.append(f"Transcript error: {self.error}")
        with suppress(OSError):
            write_text_atomic(output_dir / "meta.txt", "\n".join(lines) + "\n")
