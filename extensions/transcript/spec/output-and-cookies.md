# Output and cookies

The result folder holds `raw_transcript.txt`, `raw_sentences.txt`, and `raw_transcript.json`.
These are the required transcript artifacts.
M13 must establish any other output files of the copied program before implementation is ready.

`--output` selects the output parent.
Without it, `transcript.youtube.output` selects the parent, or the current folder when no value is configured.
In JSON, `output` names the parent and `output_dir` names the result folder.

```toml
[transcript]
browser = "arc"

[transcript.youtube]
output = "~/Documents/transcripts"
```

Machine-specific paths belong in local config.
The retained cookie integration is the source program's Arc path on macOS.
`browser = "arc"` selects it.
Without a browser setting, use anonymous access.
The supported path and its dependencies need the evidence listed in M13.
Additional browser integrations are not required for this release.

Unreadable configured cookies fall back to anonymous access with `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` in the answer's top-level `warnings`.
Anonymous access can still fail to download the video; the warning never means transcription succeeded.
The Arc adapter checks the `yt-dlp` function it adapts.
A missing function produces the same fallback warning.
M16 makes `ncly doctor transcript` report the cookie problem as `warn`.

Go selects a result-folder name from the video title, adding `-2`, `-3`, and so on when taken.
Under the output resource lock, it records the candidate path and intent, then creates the folder while rechecking collisions.
Two new runs never share a result folder or write into an existing one.
After an interrupted reservation, another invocation does not write into that folder.
Verified-result replay only reads its existing artifacts.
A new run selects a new folder and may bill Deepgram again.

Extract the embedded Python program through a temporary folder and atomic rename into `~/.cache/nicely/python/<version>/`, following XDG overrides.
Run it with `uv run --script --upgrade-package yt-dlp --no-python-downloads --cache-dir <folder>`.
Here, `<folder>` is `uv` inside Nicely's cache, such as `~/.cache/nicely/uv`.
The explicit flag overrides an inherited `UV_CACHE_DIR`.
Python 3.12 or later must already be installed; `uv` never downloads it.

The program uses the newest `yt-dlp` release at or above its declared minimum and pins its other Python dependencies.
Declare every actual download dependency in the manifest; M13 must prove that set from an empty cache.
Do not assume that `uv` and Python alone suffice.
A dry run may contact the package index to prepare Nicely's cache.
Dependency preparation failure returns `RUNTIME` with a setup hint.
Doctor probes installed tools without upgrading them.
The source CLI's `--output-dir` becomes `--output`; debug output uses `NCLY_DEBUG`.
