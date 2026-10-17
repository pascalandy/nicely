# D035 Track ownership when synchronizing files

Decided 2026-10-05. Synchronization of persistent destinations tracks the files it manages and detects user changes before replacing or removing them. A sync stages and validates each managed destination, then publishes that destination atomically under its lock. Unrelated destinations are not one transaction. Each future synchronization capability must define its publication boundary and recovery.

**Scope update, 2026-10-17.** [M12](../../milestones/M12-taps.md) starts with immutable installation and refuses existing conflicting destinations. It still publishes a complete installation. [M21](../../milestones/M21-docs-hub.md) reads project docs and builds disposable cache; it creates no persistent mirror to synchronize. Broader synchronization remains in [M99](../../milestones/M99-parking-lot.md). These scope cuts do not relax ownership guarantees when a future command updates persistent user files.

**Why.** A repeated sync must remove stale generated files without deleting unrelated files or edits. Readers need a complete result, and an interrupted update must leave enough evidence to recover. Explicit ownership makes those decisions possible.

**Rejected.** Replacing an entire destination without knowing who owns its files. Preserving every stale output forever. Publishing files one by one while readers can see an incomplete update.
