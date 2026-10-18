# D039 Let yt-dlp update without a release

Decided 2026-10-07. Declare a minimum `yt-dlp` version and run with `--upgrade-package yt-dlp`. Keep other dependencies pinned. The Arc adapter checks its target function, falling back to anonymous access with a warning if absent.

**Why.** YouTube fixes must reach users without an `ncly` release. Pascal uses Arc for YouTube.

**Rejected.** Exact pins delay fixes. A `yt-dlp` executable on `PATH` cannot receive the in-process Arc adapter.
