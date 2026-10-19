"""Download the audio of one YouTube video with yt-dlp."""

import json
import logging
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Literal

from runtime import RunBudget, run_child

SCRIPT_DIR = Path(__file__).parent.resolve()
# Go extracts the engine into python/<hash>/ inside Nicely's cache
CACHE_DIR = SCRIPT_DIR.parent.parent
ARC_BROWSER_PROFILE = Path(
    "~/Library/Application Support/Arc/User Data/Default"
).expanduser()
ARC_ADAPTER_PATH = SCRIPT_DIR / "ytdlp_arc.py"
YTDLP_MODULE_COMMAND = (sys.executable, "-m", "yt_dlp")
YTDLP_ATTEMPT_TIMEOUT = 480
# Every call ignores the user's yt-dlp config, takes one video only, and keeps
# its cache inside Nicely's cache
YTDLP_FLAGS = (
    "--ignore-config",
    "--no-playlist",
    "--cache-dir",
    str(CACHE_DIR / "yt-dlp"),
)
COOKIES_SKIPPED = "browser_cookies_skipped"

# The best audio-only stream up to 64 kbit/s, usually Opus. Speech needs no more,
# and Deepgram takes it as is, so no transcode slows the run
YTDLP_AUDIO_FORMAT = "ba[abr<=64]/ba/b"

# yt-dlp diagnostics that name a network failure a later retry may fix
YTDLP_TEMPORARY = re.compile(
    r"HTTP Error (?:429|5\d\d)|timed out|Temporary failure in name resolution|"
    r"Name or service not known|Connection (?:reset|refused|aborted)|"
    r"Network is unreachable|Remote end closed connection",
    re.IGNORECASE,
)

log = logging.getLogger("transcript")

YtDlpMethod = Literal["anonymous", "arc"]


class YtDlpError(RuntimeError):
    """yt-dlp failed with a cleaned diagnostic.

    `temporary` is true when the last attempt failed for a network reason, and
    `warnings` holds the program warnings of the attempts.
    """

    def __init__(
        self, message: str, temporary: bool = False, warnings: tuple[str, ...] = ()
    ) -> None:
        super().__init__(message)
        self.temporary = temporary
        self.warnings = warnings


@dataclass(frozen=True)
class YtDlpResult:
    """Captured yt-dlp result, the access method that succeeded, and the
    program warnings of the attempts."""

    process: subprocess.CompletedProcess[str]
    method: YtDlpMethod
    warnings: tuple[str, ...] = ()


@dataclass(frozen=True)
class DownloadedAudio:
    """Downloaded audio, the video it came from, and the YouTube access method."""

    path: Path
    method: YtDlpMethod
    title: str
    video_id: str
    warnings: tuple[str, ...] = ()


def run_ytdlp(
    args: list[str],
    url: str,
    budget: RunBudget,
    *,
    browser: str | None,
    temp_dir: Path,
) -> YtDlpResult:
    """Run the yt-dlp module, with Arc's cookies first when `browser` is arc.

    The Arc attempt needs Arc's default profile. When that attempt cannot run
    or fails, the call tries anonymous access and warns that it skipped the
    cookies. A child keeps its scratch files in `temp_dir`.

    Raises:
        YtDlpError: If the last attempt returns a failure status.
        WorkflowTimeoutError: If no workflow time remains before an attempt.
    """
    tail = [*YTDLP_FLAGS, *args, url]
    env = {"TMPDIR": str(temp_dir)}
    warnings: tuple[str, ...] = ()
    arc_diagnostic = ""
    if browser == "arc":
        warnings = (COOKIES_SKIPPED,)
        if ARC_BROWSER_PROFILE.is_dir():
            arc = run_child(
                [
                    sys.executable,
                    str(ARC_ADAPTER_PATH),
                    "--cookies-from-browser",
                    f"chrome:{ARC_BROWSER_PROFILE}",
                    *tail,
                ],
                timeout=budget.remaining(
                    "yt-dlp Arc cookie attempt", YTDLP_ATTEMPT_TIMEOUT
                ),
                env=env,
            )
            if arc.returncode == 0:
                return YtDlpResult(arc, "arc")
            arc_diagnostic = _clean_subprocess_diagnostic(arc.stderr or arc.stdout)
        else:
            arc_diagnostic = f"no Arc profile at {ARC_BROWSER_PROFILE}"
        log.info("Arc YouTube access failed; retrying anonymously")
        log.debug(f"Arc attempt: {arc_diagnostic}")
    anonymous = run_child(
        [*YTDLP_MODULE_COMMAND, *tail],
        timeout=budget.remaining("yt-dlp anonymous attempt", YTDLP_ATTEMPT_TIMEOUT),
        env=env,
    )
    if anonymous.returncode == 0:
        return YtDlpResult(anonymous, "anonymous", warnings)

    anonymous_diagnostic = _clean_subprocess_diagnostic(
        anonymous.stderr or anonymous.stdout
    )
    message = f"yt-dlp failed: {anonymous_diagnostic or 'no diagnostic was returned'}"
    if arc_diagnostic:
        message = f"Arc attempt: {arc_diagnostic}; anonymous fallback: {message}"
    raise YtDlpError(
        message,
        temporary=bool(YTDLP_TEMPORARY.search(anonymous_diagnostic)),
        warnings=warnings,
    )


def _clean_subprocess_diagnostic(text: str) -> str:
    """Remove terminal controls and collapse a subprocess diagnostic."""
    without_ansi = re.sub(r"\x1b\[[0-?]*[ -/]*[@-~]", "", text)
    without_secrets = re.sub(
        r"(?im)^(\s*(?:authorization|cookie|set-cookie)\s*[:=])[^\r\n]*",
        r"\1 [redacted]",
        without_ansi,
    )
    without_controls = re.sub(r"[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]", " ", without_secrets)
    return " ".join(without_controls.split())[:2000]


def clean_title(title: str) -> str:
    """Clean video title for filesystem use."""
    cleaned = re.sub(r"[^a-zA-Z0-9 ]", "", title)
    cleaned = cleaned.replace(" ", "_")
    return cleaned[:50]


def download_audio(
    url: str, temp_dir: Path, budget: RunBudget, *, browser: str | None
) -> DownloadedAudio:
    """Download native YouTube audio into `temp_dir`, and its video's title
    and ID, in one yt-dlp call."""
    result = run_ytdlp(
        [
            "-f",
            YTDLP_AUDIO_FORMAT,
            "-o",
            str(temp_dir / "audio.%(ext)s"),
            "--no-simulate",
            "--print",
            "after_move:%(.{id,title,filepath})j",
        ],
        url,
        budget,
        browser=browser,
        temp_dir=temp_dir,
    )
    try:
        info = json.loads(result.process.stdout.strip().splitlines()[-1])
    except (IndexError, json.JSONDecodeError) as error:
        raise ValueError("yt-dlp did not describe the downloaded audio") from error
    title, video_id = info.get("title"), info.get("id")
    if not title or not video_id:
        raise ValueError("Video may be private, deleted, or unavailable")
    audio_path = Path(info.get("filepath") or "")
    if not audio_path.is_file():
        raise ValueError("Audio download failed")
    return DownloadedAudio(audio_path, result.method, title, video_id, result.warnings)
