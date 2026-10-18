# D031 Stop the whole process tree on cancel and keep finished work

Decided 2026-10-05. Cancellation stops reachable descendants, preserves finished work, and exits 130 or 143 without authorizing retry. [Programs that ncly runs](../contract/programs-that-ncly-runs.md) owns cleanup, pipe limits, and normal exits.

**Why.** Transcript children can start separate sessions, beyond a process-group signal. Linux subreaping keeps orphaned descendants reachable; macOS cannot recover descendants whose parent already exited. The two-second output drain prevents inherited pipes from blocking forever while allowing final output to arrive. Core-call and completion limits use their own pipe rules.

**Rejected.** Direct-child signals alone, reliance on Python cleanup after SIGKILL, and unbounded stream reads. After normal exit, stopping or waiting for every descendant would close or block on a user's detached viewer. Interrupted output may already be billed, and a stopped process does not prove that an external request had no effect.
