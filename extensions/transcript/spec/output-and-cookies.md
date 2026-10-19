# Output and cookies

The result folder holds `raw_transcript.txt`, `raw_sentences.txt`, and `raw_transcript.json`.
These are the required transcript artifacts.
The program also writes `meta.txt` beside them, an advisory file that the record never lists and replay never checks.
Temporary audio and child scratch files stay in Go's `temp_dir`, so no other file reaches the result folder.

`--output` selects the output parent.
Without it, `transcript.youtube.output` selects the parent, or the current folder when no value is configured.
The parent must be an existing folder: `--output` refuses with `USAGE_INVALID` otherwise, and the config with `CONFIG_INVALID`.
In JSON, `output` names the parent and `output_dir` names the result folder.

```toml
[transcript]
browser = "arc"

[transcript.youtube]
output = "~/Documents/transcripts"
```

Machine-specific paths belong in local config.
The retained cookie integration is the source program's Arc path on macOS.
`browser = "arc"` selects it: yt-dlp then runs through the source's `ytdlp_arc.py`, with Arc's cookies.
Without a browser setting, use anonymous access.
Any other `browser` value is `CONFIG_INVALID`.
Scenarios stub cookie reading, and Pascal's paid check after M16 proves the real Arc path.
Additional browser integrations are not required for this release.

Unreadable configured cookies fall back to anonymous access with `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` in the answer's top-level `warnings`.
Anonymous access can still fail to download the video; the warning never means transcription succeeded.
The Arc adapter checks the `yt-dlp` function it adapts, not the `yt-dlp` version, since the header accepts newer releases.
A missing function produces the same fallback warning.
[ncly transcript status](../spec.md#ncly-transcript-status) reports that problem as a `warn` of `ncly doctor transcript`.

Go selects a result-folder name from the video title, adding `-2`, `-3`, and so on when taken.
The intent proposes the name, from the source's title cleanup, which turns spaces into underscores.
A name that holds a slash, is empty, or is `.` or `..` is a protocol failure.
Go tries at most `name-1000`.
Go records each candidate path with the intent, then creates the folder with an exclusive `mkdir`.
When that `mkdir` finds the name taken, Go records the next candidate before it tries again.
Two new runs never share a result folder or write into an existing one.
A run that recorded a candidate and died before its `mkdir` never writes there, so a later run may create that folder.
A folder that cannot be created ends the step `failed`, since no `ack` left, and no audio either.
After an interrupted reservation, another invocation does not write into that folder.
Verified-result replay only reads its existing artifacts.
A new run selects a new folder and may bill Deepgram again.

Extract the embedded Python program through a temporary folder and atomic rename into `~/.cache/nicely/python/<hash>/`, following XDG overrides.
`<hash>` is the SHA-256 of the embedded files, so two builds that embed different files never share a folder.
Run it with `uv run --script --upgrade-package yt-dlp --no-python-downloads --cache-dir <folder>`.
Here, `<folder>` is `uv` inside Nicely's cache, such as `~/.cache/nicely/uv`.
The explicit flag overrides an inherited `UV_CACHE_DIR`.
Python 3.12 or later must already be installed; `uv` never downloads it.
The engine runs with `PYTHONDONTWRITEBYTECODE=1`, so the folder of a hash holds only the embedded files.
Only the test binary adds `--offline`, as [Test transport](test-transport.md) says.
yt-dlp runs with `--ignore-config`, `--no-playlist`, and a `--cache-dir` named `yt-dlp` inside Nicely's cache.
The engine finds that cache two folders above its own, `python/<hash>/`.

The program uses the newest `yt-dlp` release at or above its declared minimum and pins its other Python dependencies.
The manifest declares the programs that a cold run needs: `uv`, and `python3` with `version = "3.12"`.
[Cold download](#cold-download) shows that the YouTube path needs neither `ffmpeg` nor a JavaScript runtime.
A dry run may contact the package index to prepare Nicely's cache.
Dependency preparation failure returns `RUNTIME` with a setup hint.
Doctor probes installed tools without upgrading them.
The source CLI's `--output-dir` becomes `--output`; debug output uses `NCLY_DEBUG`.

## Cold download

Spike SP1 ran on 2026-10-18 on Omarchy, with uv 0.13.1 and Python 3.14.8.
`yt-dlp>=2026.7.4` resolved to yt-dlp 2026.08.19.
Each run started from an empty uv cache and an empty yt-dlp cache, with `--ignore-config` and no cookies.
It downloaded the audio of the test video `EIEc43CxIvY` through the production `uv` line, with the source's format `ba[abr<=64]/ba/b`.
`ffmpeg` was not on `PATH` in any run.

| Run | Result | Format | Time |
|---|---|---|---|
| Without Deno, twice | Success, with one yt-dlp warning: YouTube extraction without a JavaScript runtime is deprecated, and some formats may be missing | `250`, Opus in WebM, 61 kbit/s, 2,264,567 bytes | 1.6 to 2.6 s of download, 2.6 s in total |
| With Deno 2.9.7 | Success, with no warning | The same | 1.8 s of download, 3.0 s in total |

The same audio then went once to Deepgram through the source's `transcribe_audio`, with a fake key.
Deepgram read all 2,264,567 bytes, then answered 401 with `INVALID_AUTH`.
httpx 0.28.1 raised it as `HTTPStatusError`, not as a write error.
So a fake key gives `deepgram_key_rejected`, and the Verify of M13-T3 expects the item code `AUTH_REJECTED`.
Add Deno to the manifest when yt-dlp stops downloading YouTube audio without a JavaScript runtime.
