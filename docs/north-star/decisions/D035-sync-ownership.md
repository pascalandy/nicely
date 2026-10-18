# D035 Track ownership when synchronizing files

Decided 2026-10-05. Persistent synchronization tracks managed files and detects user edits before replacement or removal. Stage and validate each destination, then publish it atomically under its lock. Unrelated destinations are not one transaction. Each synchronization capability must specify its publication boundary and recovery.

**Scope update, 2026-10-17.** [M99](../../milestones/M99-parking-lot/taps.md) defers persistent synchronization. [M21](../../milestones/M21-docs-hub.md) builds disposable cache without a persistent mirror. Future persistent writes still require these guarantees.

**Why.** Ownership permits removal of stale outputs without deleting user work. Atomic publication prevents incomplete results from reaching readers.

**Rejected.** Whole-destination replacement without ownership, indefinite retention of stale outputs, and visible file-by-file publication.
