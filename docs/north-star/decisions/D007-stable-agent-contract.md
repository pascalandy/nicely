# D007 Keep one stable agent contract

Decided 2026-10-05. Machine identifiers stay in English. [Compatibility](../contract/compatibility.md) protects values and their meaning, with versions for protocols and stored records. Exit 78 requires human action. Exit 75 promises that the whole invocation meets [Retry safety](../contract/retry-safety.md).

**Why.** A stable key with a changed unit still breaks consumers. A temporary failure can follow irreversible work. Exit 78 follows `EX_CONFIG` and distinguishes human action from invalid usage.

**Rejected.** Stability of names alone, unversioned breaking changes, and exit 75 after unsafe effects, including a late lock conflict. A missing key needs neither a new exit 3 nor usage exit 2.
