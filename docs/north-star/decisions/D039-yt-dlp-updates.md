# D039 Let yt-dlp update without a release

Decided 2026-10-07. The Python program names a minimum `yt-dlp` version, and `ncly` runs it with `--upgrade-package yt-dlp`, so each run uses the newest release. The other dependencies stay pinned. The Arc cookie adapter checks the `yt-dlp` function that it changes instead of an exact version, and a missing function falls back to anonymous access with a warning.

**Why.** YouTube breaks `yt-dlp` every few weeks, and a fix must reach users without an `ncly` release. Pascal keeps Arc for YouTube, so the adapter stays.

**Rejected.** An exact pin, which needs an `ncly` release for every YouTube change. The `yt-dlp` on `PATH`, because the Arc adapter changes `yt-dlp` inside the same process.
