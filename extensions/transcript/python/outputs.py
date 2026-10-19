"""Save the three transcript texts into the result folder."""

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
