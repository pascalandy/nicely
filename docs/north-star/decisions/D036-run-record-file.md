# D036 Record each run in one JSON file that its executor locks

Decided 2026-10-07, revised 2026-10-16. Store one durable JSON file per run. The executor locks before reading and holds ownership separately from saved status. [Record format](../contract/record-format.md) owns persistence, locking, statuses, and verification. [Run inspection](../core-spec.md#ncly-run) owns bounded listings and evidence.

**Why.** JSON needs no second storage model. Atomic writes protect readers; flushes protect acknowledged intent within platform guarantees. OS locks release on process death. Lock-before-read prevents duplicate paid starts, and undeleted lock files prevent two owners under one name. A brief wait tolerates inspection without waiting for an entire run.

Pascal chose `unknown` over `interrupted` when an effect remains unresolved. The next session first needs to know whether work may have billed; the exit code still records the signal.

A source consumer holds the source lock while it verifies and records what it will use. It releases that lock before its first effect.

**Rejected.** SQLite, reusable process IDs as ownership, heartbeat workers, unreadable records treated as absent, and unbounded lock waits.
