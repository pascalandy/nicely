# D035 Track ownership when synchronizing files

Decided 2026-10-05. Tap and docs synchronization track the files they manage and detect user changes before replacing or removing them. A sync stages and validates each managed destination, then publishes that destination atomically under its lock. Unrelated destinations are not one transaction. [M12](../../milestones/M12-taps.md) and [M21](../../milestones/M21-docs-hub.md) define their publication boundaries and recovery.

**Why.** A repeated sync must remove stale generated files without deleting unrelated files or edits. Readers need a complete result, and an interrupted update must leave enough evidence to recover. Explicit ownership makes those decisions possible.

**Rejected.** Replacing an entire destination without knowing who owns its files. Preserving every stale output forever. Publishing files one by one while readers can see an incomplete update.
