# D031 Stop the whole process tree on cancel and keep finished work

Decided 2026-10-05. On Ctrl-C or SIGTERM, `ncly` stops every program it started and every descendant that it can still reach, orphans included on Linux, keeps every finished output, and exits 130 or 143 with the data that still applies, as [contract.md](../contract.md#programs-that-ncly-runs) defines. Run inspection distinguishes verified finished work from uncertain effects. Neither interruption code authorizes an automatic retry.

**Why.** The transcript CLI starts its children in their own process groups, so a signal to its parent alone leaves `yt-dlp`, `ffmpeg`, or a harness running. A harness that hits its time limit sends SIGTERM. A finished transcript may already be billed, so cleanup must never delete it.

`ncly` finds descendants through their parent process, because a signal to a process group misses the children that the transcript CLI starts in their own sessions. On Linux, `ncly` registers as a child subreaper, so an orphaned descendant stays reachable. macOS has no equivalent, so a descendant whose parent exited before the signal is the one exception there.

**Rejected.** Signaling only the direct child. Relying on the Python program alone to stop its children, because the SIGKILL after 15 seconds leaves them running. Deleting partial results on interrupt. Assuming a stopped process proves that its last external request had no effect.
