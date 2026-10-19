"""Tests of the YouTube download: titles, URLs, yt-dlp calls, and access order."""

from __future__ import annotations

import json
import logging
import sys
from pathlib import Path
from types import SimpleNamespace

import pytest

import runtime
import youtube

SCRIPT_PATH = Path(__file__).parents[1] / "transcript.py"


class TestCleanTitle:
    def test_removes_special_characters(self) -> None:
        assert youtube.clean_title("Hello, World!") == "Hello_World"

    def test_truncates_to_50_chars(self) -> None:
        result = youtube.clean_title("A" * 100)
        assert len(result) == 50

    def test_replaces_spaces_with_underscores(self) -> None:
        assert youtube.clean_title("foo bar baz") == "foo_bar_baz"

    def test_empty_string(self) -> None:
        assert youtube.clean_title("") == ""


class TestValidateYoutubeUrl:
    def test_standard_url(self) -> None:
        assert youtube.validate_youtube_url(
            "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
        )

    def test_short_url(self) -> None:
        assert youtube.validate_youtube_url("https://youtu.be/dQw4w9WgXcQ")

    def test_shorts_url(self) -> None:
        assert youtube.validate_youtube_url(
            "https://www.youtube.com/shorts/dQw4w9WgXcQ"
        )

    def test_invalid_url(self) -> None:
        assert not youtube.validate_youtube_url("https://example.com/video")

    def test_empty_string(self) -> None:
        assert not youtube.validate_youtube_url("")


class TestDownloadAudio:
    def _ytdlp_prints(self, tmp_path, monkeypatch, info: dict) -> list:
        """Make the Arc attempt succeed and print `info` as yt-dlp's last line."""
        arc_profile = tmp_path / "Arc" / "User Data" / "Default"
        arc_profile.mkdir(parents=True)
        commands = []

        def fake_run(command, **_kwargs):
            commands.append(command)
            stdout = f"[download] 100%\n{json.dumps(info)}\n"
            return SimpleNamespace(returncode=0, stdout=stdout, stderr="")

        monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", arc_profile)
        monkeypatch.setattr(youtube, "run_child", fake_run)
        return commands

    def test_one_call_downloads_native_audio_and_describes_its_video(
        self, tmp_path, monkeypatch
    ) -> None:
        audio = tmp_path / "audio.webm"
        audio.write_bytes(b"audio")
        info = {"id": "abc", "title": "A video", "filepath": str(audio)}
        commands = self._ytdlp_prints(tmp_path, monkeypatch, info)

        result = youtube.download_audio(
            "url", tmp_path, runtime.RunBudget(float("inf"))
        )

        (command,) = commands
        assert result == youtube.DownloadedAudio(audio, "arc", "A video", "abc")
        assert command[command.index("-f") + 1] == "ba[abr<=64]/ba/b"
        assert "--no-simulate" in command
        assert "-x" not in command

    def test_a_video_without_a_title_is_unavailable(
        self, tmp_path, monkeypatch
    ) -> None:
        self._ytdlp_prints(tmp_path, monkeypatch, {"id": "abc"})

        with pytest.raises(ValueError, match="private, deleted, or unavailable"):
            youtube.download_audio("url", tmp_path, runtime.RunBudget(float("inf")))


def test_subprocess_is_not_started_without_remaining_budget(monkeypatch) -> None:
    monkeypatch.setattr(runtime.time, "monotonic", lambda: 10.0)
    monkeypatch.setattr(
        youtube,
        "run_child",
        lambda *_args, **_kwargs: pytest.fail("subprocess started after deadline"),
    )

    with pytest.raises(runtime.WorkflowTimeoutError, match="budget"):
        youtube.run_ytdlp([], "https://youtu.be/abc", runtime.RunBudget(10))


class TestYouTubeAuthenticationOrder:
    """YouTube prefers browser authentication before anonymous access."""

    def test_runtime_pins_the_proven_ytdlp_version(self) -> None:
        script_text = SCRIPT_PATH.read_text(encoding="utf-8")

        assert '"yt-dlp==2026.7.4"' in script_text

    def test_ytdlp_uses_arc_first_and_skips_anonymous_when_it_succeeds(
        self, tmp_path, monkeypatch
    ) -> None:
        arc_profile = tmp_path / "Arc" / "User Data" / "Default"
        arc_profile.mkdir(parents=True)
        calls = []

        def fake_run(command, **kwargs):
            calls.append((command, kwargs))
            return SimpleNamespace(returncode=0, stdout="Title\nabc\n", stderr="")

        monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", arc_profile)
        monkeypatch.setattr(youtube, "run_child", fake_run)

        result = youtube.run_ytdlp(
            ["--get-title", "--get-id"],
            "url",
            runtime.RunBudget(float("inf")),
        )

        assert result.method == "arc"
        assert result.process.stdout == "Title\nabc\n"
        assert len(calls) == 1
        assert calls[0][0][:2] == [sys.executable, str(youtube.ARC_ADAPTER_PATH)]

    def test_ytdlp_falls_back_to_anonymous_after_arc_failure(
        self, tmp_path, monkeypatch, caplog
    ) -> None:
        caplog.set_level(logging.INFO, logger="transcript")
        arc_profile = tmp_path / "Arc" / "User Data" / "Default"
        arc_profile.mkdir(parents=True)
        responses = [
            SimpleNamespace(returncode=1, stdout="", stderr="Arc failed"),
            SimpleNamespace(returncode=0, stdout="Title\nabc\n", stderr=""),
        ]
        calls = []

        def fake_run(command, **kwargs):
            calls.append((command, kwargs))
            return responses.pop(0)

        monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", arc_profile)
        monkeypatch.setattr(youtube, "run_child", fake_run)

        result = youtube.run_ytdlp(
            ["--get-title", "--get-id"],
            "url",
            runtime.RunBudget(float("inf")),
        )

        assert result.method == "anonymous"
        assert calls[0][0][:2] == [sys.executable, str(youtube.ARC_ADAPTER_PATH)]
        assert calls[0][0][2:4] == [
            "--cookies-from-browser",
            f"chrome:{arc_profile}",
        ]
        assert calls[1][0][:3] == [sys.executable, "-m", "yt_dlp"]
        assert caplog.messages == [
            (
                "Arc YouTube access failed; retrying anonymously. "
                "Sign in to YouTube in Arc to use your session"
            )
        ]

    def test_ytdlp_uses_chrome_first_when_arc_is_absent(
        self, tmp_path, monkeypatch
    ) -> None:
        commands = []

        monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", tmp_path / "missing")
        monkeypatch.setattr(
            youtube,
            "run_child",
            lambda command, **_kwargs: (
                commands.append(command)
                or SimpleNamespace(returncode=0, stdout="ok", stderr="")
            ),
        )

        result = youtube.run_ytdlp([], "url", runtime.RunBudget(float("inf")))

        assert result.method == "chrome"
        assert len(commands) == 1
        assert commands[0][:5] == [
            sys.executable,
            "-m",
            "yt_dlp",
            "--cookies-from-browser",
            "chrome",
        ]

    def test_ytdlp_falls_back_to_anonymous_after_chrome_failure(
        self, tmp_path, monkeypatch, caplog
    ) -> None:
        caplog.set_level(logging.INFO, logger="transcript")
        responses = [
            SimpleNamespace(returncode=1, stdout="", stderr="Chrome failed"),
            SimpleNamespace(returncode=0, stdout="Title\nabc\n", stderr=""),
        ]
        commands = []

        def fake_run(command, **_kwargs):
            commands.append(command)
            return responses.pop(0)

        monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", tmp_path / "missing")
        monkeypatch.setattr(youtube, "run_child", fake_run)

        result = youtube.run_ytdlp([], "url", runtime.RunBudget(float("inf")))

        assert result.method == "anonymous"
        assert caplog.messages == [
            (
                "Chrome YouTube access failed; retrying anonymously. "
                "Sign in to YouTube in Chrome to use your session"
            )
        ]
        assert commands[0][3:5] == ["--cookies-from-browser", "chrome"]
        assert commands[1][:3] == [sys.executable, "-m", "yt_dlp"]

    def test_failed_attempts_keep_labeled_redacted_diagnostics(
        self, tmp_path, monkeypatch
    ) -> None:
        arc_profile = tmp_path / "Arc" / "User Data" / "Default"
        arc_profile.mkdir(parents=True)
        failures = [
            SimpleNamespace(
                returncode=1,
                stdout="",
                stderr=(
                    "\x1b[31mArc cookies expired\x1b[0m\n"
                    "\x1b[33mAuthorization:\x1b[0m arc-secret\n"
                ),
            ),
            SimpleNamespace(
                returncode=1,
                stdout="",
                stderr="anonymous blocked\nCookie: SID=anon-secret\n",
            ),
        ]

        monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", arc_profile)
        monkeypatch.setattr(
            youtube, "run_child", lambda *_args, **_kwargs: failures.pop(0)
        )

        with pytest.raises(youtube.YtDlpError) as error:
            youtube.run_ytdlp([], "url", runtime.RunBudget(float("inf")))

        diagnostic = str(error.value)
        assert "Arc attempt: Arc cookies expired" in diagnostic
        assert "anonymous fallback: anonymous blocked" in diagnostic
        assert diagnostic.index("Arc attempt") < diagnostic.index("anonymous fallback")
        assert "anon-secret" not in diagnostic
        assert "arc-secret" not in diagnostic
        assert "\x1b" not in diagnostic
