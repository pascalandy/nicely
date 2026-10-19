"""Tests of the YouTube download: titles, yt-dlp calls, and cookie access."""

from __future__ import annotations

import json
import logging
import sys
from types import SimpleNamespace

import pytest

import runtime
import youtube

FOREVER = runtime.RunBudget(float("inf"))


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


def recorder(monkeypatch, *responses: SimpleNamespace) -> list:
    """Make run_child return `responses` in order, and record each call."""
    calls = []
    queue = list(responses)

    def fake_run(command, **kwargs):
        calls.append((command, kwargs))
        return queue.pop(0)

    monkeypatch.setattr(youtube, "run_child", fake_run)
    return calls


def ok(stdout: str = "Title\nabc\n") -> SimpleNamespace:
    return SimpleNamespace(returncode=0, stdout=stdout, stderr="")


def failed(stderr: str) -> SimpleNamespace:
    return SimpleNamespace(returncode=1, stdout="", stderr=stderr)


@pytest.fixture
def arc_profile(tmp_path, monkeypatch):
    profile = tmp_path / "Arc" / "User Data" / "Default"
    profile.mkdir(parents=True)
    monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", profile)
    return profile


class TestDownloadAudio:
    def test_one_call_downloads_native_audio_and_describes_its_video(
        self, tmp_path, monkeypatch
    ) -> None:
        audio = tmp_path / "audio.webm"
        audio.write_bytes(b"audio")
        info = {"id": "abc", "title": "A video", "filepath": str(audio)}
        calls = recorder(monkeypatch, ok(f"[download] 100%\n{json.dumps(info)}\n"))

        result = youtube.download_audio("url", tmp_path, FOREVER, browser=None)

        ((command, kwargs),) = calls
        assert result == youtube.DownloadedAudio(audio, "anonymous", "A video", "abc")
        assert command[command.index("-f") + 1] == "ba[abr<=64]/ba/b"
        assert command[command.index("-o") + 1] == str(tmp_path / "audio.%(ext)s")
        assert "--no-simulate" in command
        assert "-x" not in command
        assert kwargs["env"] == {"TMPDIR": str(tmp_path)}

    def test_a_video_without_a_title_is_unavailable(
        self, tmp_path, monkeypatch
    ) -> None:
        recorder(monkeypatch, ok(json.dumps({"id": "abc"})))

        with pytest.raises(ValueError, match="private, deleted, or unavailable"):
            youtube.download_audio("url", tmp_path, FOREVER, browser=None)


def test_subprocess_is_not_started_without_remaining_budget(
    tmp_path, monkeypatch
) -> None:
    monkeypatch.setattr(runtime.time, "monotonic", lambda: 10.0)
    monkeypatch.setattr(
        youtube,
        "run_child",
        lambda *_args, **_kwargs: pytest.fail("subprocess started after deadline"),
    )

    with pytest.raises(runtime.WorkflowTimeoutError, match="budget"):
        youtube.run_ytdlp(
            [],
            "https://youtu.be/abc",
            runtime.RunBudget(10),
            browser=None,
            temp_dir=tmp_path,
        )


class TestYouTubeAccess:
    """yt-dlp ignores its config, takes one video, and reads cookies only
    from Arc, when the config asks for it."""

    def test_every_call_ignores_the_config_and_takes_one_video(
        self, tmp_path, monkeypatch
    ) -> None:
        calls = recorder(monkeypatch, ok())

        result = youtube.run_ytdlp(
            ["--get-id"], "url", FOREVER, browser=None, temp_dir=tmp_path
        )

        ((command, _kwargs),) = calls
        assert command[:3] == [sys.executable, "-m", "yt_dlp"]
        assert command[3:7] == [
            "--ignore-config",
            "--no-playlist",
            "--cache-dir",
            str(youtube.CACHE_DIR / "yt-dlp"),
        ]
        assert command[-2:] == ["--get-id", "url"]
        assert "--cookies-from-browser" not in command
        assert (result.method, result.warnings) == ("anonymous", ())

    def test_arc_reads_its_cookies_and_skips_anonymous_when_it_succeeds(
        self, tmp_path, monkeypatch, arc_profile
    ) -> None:
        calls = recorder(monkeypatch, ok())

        result = youtube.run_ytdlp([], "url", FOREVER, browser="arc", temp_dir=tmp_path)

        ((command, _kwargs),) = calls
        assert command[:4] == [
            sys.executable,
            str(youtube.ARC_ADAPTER_PATH),
            "--cookies-from-browser",
            f"chrome:{arc_profile}",
        ]
        assert "--ignore-config" in command and "--no-playlist" in command
        assert (result.method, result.warnings) == ("arc", ())

    def test_a_failed_arc_attempt_falls_back_to_anonymous_with_a_warning(
        self, tmp_path, monkeypatch, caplog, arc_profile
    ) -> None:
        caplog.set_level(logging.INFO, logger="transcript")
        calls = recorder(
            monkeypatch,
            failed("ERROR: Arc cookie adapter failed: settings function missing"),
            ok(),
        )

        result = youtube.run_ytdlp([], "url", FOREVER, browser="arc", temp_dir=tmp_path)

        assert calls[0][0][1] == str(youtube.ARC_ADAPTER_PATH)
        assert calls[1][0][:3] == [sys.executable, "-m", "yt_dlp"]
        assert "--cookies-from-browser" not in calls[1][0]
        assert (result.method, result.warnings) == (
            "anonymous",
            ("browser_cookies_skipped",),
        )
        assert caplog.messages == ["Arc YouTube access failed; retrying anonymously"]

    def test_a_missing_arc_profile_goes_anonymous_with_a_warning(
        self, tmp_path, monkeypatch
    ) -> None:
        monkeypatch.setattr(youtube, "ARC_BROWSER_PROFILE", tmp_path / "missing")
        calls = recorder(monkeypatch, ok())

        result = youtube.run_ytdlp([], "url", FOREVER, browser="arc", temp_dir=tmp_path)

        ((command, _kwargs),) = calls
        assert command[:3] == [sys.executable, "-m", "yt_dlp"]
        assert result.warnings == ("browser_cookies_skipped",)

    def test_failed_attempts_keep_labeled_redacted_diagnostics(
        self, tmp_path, monkeypatch, arc_profile
    ) -> None:
        recorder(
            monkeypatch,
            failed(
                "\x1b[31mArc cookies expired\x1b[0m\n"
                "\x1b[33mAuthorization:\x1b[0m arc-secret\n"
            ),
            failed("anonymous blocked\nCookie: SID=anon-secret\n"),
        )

        with pytest.raises(youtube.YtDlpError) as error:
            youtube.run_ytdlp([], "url", FOREVER, browser="arc", temp_dir=tmp_path)

        diagnostic = str(error.value)
        assert "Arc attempt: Arc cookies expired" in diagnostic
        assert "anonymous fallback: yt-dlp failed: anonymous blocked" in diagnostic
        assert diagnostic.index("Arc attempt") < diagnostic.index("anonymous fallback")
        assert "anon-secret" not in diagnostic
        assert "arc-secret" not in diagnostic
        assert "\x1b" not in diagnostic
        assert error.value.warnings == ("browser_cookies_skipped",)

    @pytest.mark.parametrize(
        ("stderr", "temporary"),
        [
            ("ERROR: HTTP Error 503: Service Unavailable", True),
            ("ERROR: Private video", False),
        ],
    )
    def test_a_network_failure_is_temporary(
        self, tmp_path, monkeypatch, stderr, temporary
    ) -> None:
        recorder(monkeypatch, failed(stderr))

        with pytest.raises(youtube.YtDlpError) as error:
            youtube.run_ytdlp([], "url", FOREVER, browser=None, temp_dir=tmp_path)

        assert error.value.temporary is temporary
