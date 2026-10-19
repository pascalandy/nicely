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
ARC_BROWSER_PROFILE = Path(
    "~/Library/Application Support/Arc/User Data/Default"
).expanduser()
ARC_ADAPTER_PATH = SCRIPT_DIR / "ytdlp_arc.py"
YTDLP_MODULE_COMMAND = (sys.executable, "-m", "yt_dlp")
YTDLP_ATTEMPT_TIMEOUT = 480

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

YOUTUBE_URL_PATTERNS = (
    r"^https?://(www\.)?youtube\.com/watch\?v=(?P<id>[\w-]+)",
    r"^https?://youtu\.be/(?P<id>[\w-]+)",
    r"^https?://(www\.)?youtube\.com/shorts/(?P<id>[\w-]+)",
)

log = logging.getLogger("transcript")

YtDlpMethod = Literal["anonymous", "arc", "chrome"]


class YtDlpError(RuntimeError):
    """yt-dlp failed with a cleaned diagnostic suitable for CLI output.

    `temporary` is true when the last attempt failed for a network reason.
    """

    def __init__(self, message: str, temporary: bool = False) -> None:
        super().__init__(message)
        self.temporary = temporary


@dataclass(frozen=True)
class YtDlpResult:
    """Captured yt-dlp result and the authentication method that succeeded."""

    process: subprocess.CompletedProcess[str]
    method: YtDlpMethod


@dataclass(frozen=True)
class DownloadedAudio:
    """Downloaded audio, the video it came from, and the YouTube access method."""

    path: Path
    method: YtDlpMethod
    title: str
    video_id: str


def run_ytdlp(args: list[str], url: str, budget: RunBudget) -> YtDlpResult:
    """Run the pinned yt-dlp module with browser authentication first.

    Use Arc when its default profile exists, otherwise use Chrome. Anonymous
    access is a fallback so a valid browser session never pays for a failed
    unauthenticated request first.

    Args:
        args: Extra yt-dlp flags (inserted before the URL).
        url: YouTube URL (appended last).
        budget: Shared workflow deadline used to bound both attempts.

    Raises:
        YtDlpError: If both attempts return a failure status.
        WorkflowTimeoutError: If no workflow time remains before an attempt.
    """
    if ARC_BROWSER_PROFILE.is_dir():
        authenticated_command = [
            sys.executable,
            str(ARC_ADAPTER_PATH),
            "--cookies-from-browser",
            f"chrome:{ARC_BROWSER_PROFILE}",
            *args,
            url,
        ]
        authenticated_operation = "yt-dlp Arc cookie attempt"
        authenticated_label = "Arc"
        authenticated_method: YtDlpMethod = "arc"
    else:
        authenticated_command = [
            *YTDLP_MODULE_COMMAND,
            "--cookies-from-browser",
            "chrome",
            *args,
            url,
        ]
        authenticated_operation = "yt-dlp Chrome cookie attempt"
        authenticated_label = "Chrome"
        authenticated_method = "chrome"
    authenticated = run_child(
        authenticated_command,
        timeout=budget.remaining(authenticated_operation, YTDLP_ATTEMPT_TIMEOUT),
    )
    if authenticated.returncode == 0:
        return YtDlpResult(authenticated, authenticated_method)

    authenticated_diagnostic = _clean_subprocess_diagnostic(
        authenticated.stderr or authenticated.stdout
    )
    log.info(
        f"{authenticated_label} YouTube access failed; retrying anonymously. "
        f"Sign in to YouTube in {authenticated_label} to use your session"
    )
    log.debug(f"{authenticated_label} attempt: {authenticated_diagnostic}")
    anonymous = run_child(
        [*YTDLP_MODULE_COMMAND, *args, url],
        timeout=budget.remaining("yt-dlp anonymous fallback", YTDLP_ATTEMPT_TIMEOUT),
    )
    if anonymous.returncode == 0:
        return YtDlpResult(anonymous, "anonymous")

    anonymous_diagnostic = _clean_subprocess_diagnostic(
        anonymous.stderr or anonymous.stdout
    )
    raise YtDlpError(
        "yt-dlp failed: "
        f"{authenticated_label} attempt: "
        f"{authenticated_diagnostic or 'no diagnostic was returned'}; "
        "anonymous fallback: "
        f"{anonymous_diagnostic or 'no diagnostic was returned'}",
        temporary=bool(YTDLP_TEMPORARY.search(anonymous_diagnostic)),
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


def download_audio(url: str, output_dir: Path, budget: RunBudget) -> DownloadedAudio:
    """Download native YouTube audio, and its video's title and ID, in one
    yt-dlp call."""
    result = run_ytdlp(
        [
            "-f",
            YTDLP_AUDIO_FORMAT,
            "-o",
            str(output_dir / "audio.%(ext)s"),
            "--no-simulate",
            "--print",
            "after_move:%(.{id,title,filepath})j",
        ],
        url,
        budget,
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
        raise FileNotFoundError("Audio download failed")
    return DownloadedAudio(audio_path, result.method, title, video_id)


def youtube_video_id(url: str) -> str | None:
    """The video ID a YouTube URL names, or None for any other URL."""
    for pattern in YOUTUBE_URL_PATTERNS:
        if match := re.match(pattern, url):
            return match["id"]
    return None


def validate_youtube_url(url: str) -> bool:
    """Validate YouTube URL format."""
    return youtube_video_id(url) is not None
