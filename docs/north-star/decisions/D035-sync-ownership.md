# D035 Track ownership when synchronizing files

Decided 2026-10-05. Docs synchronization tracks its managed files and detects user changes before replacement or removal, while taps use [D061](D061-tap-publications.md)'s whole-folder ownership. A sync stages and validates each managed destination, then publishes that destination atomically under its lock. Unrelated destinations are not one transaction. [M12](../../milestones/M12-taps.md) and [M21](../../milestones/M21-docs-hub.md) define their publication boundaries and recovery.

**Why.** A repeated sync must remove stale generated files without deleting unrelated files or edits. Readers need a complete result, and an interrupted update must leave enough evidence to recover. Explicit ownership makes those decisions possible.

**Rejected.** Replacing an entire destination without knowing who owns its files. Preserving every stale output forever. Publishing files one by one while readers can see an incomplete update.
