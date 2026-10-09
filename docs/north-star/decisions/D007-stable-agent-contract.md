# D007 Keep one stable agent contract

Decided 2026-10-05. Command names, flags, JSON keys, error codes, and exit codes stay in English. [Compatibility](../contract.md#compatibility) covers types, units, formats, meaning, required fields, nullability, enum values, and defined array order. Protocols and stored run records have versions from their first use. Each core error code maps to one exit code. Exit `78` means that a human must act. Exit `75` follows the [retry safety rule](../contract.md#retry-safety) for the whole invocation, including finished items in a batch.

**Why.** Agents branch on codes and values, so keeping a key while changing its unit can break a consumer. A temporary failure before a paid request can still follow a file write. Exit `75` therefore requires that all effects so far are safe to repeat, through absence of incompatible effects or proven idempotence. `78` is `EX_CONFIG` in `sysexits.h`, and it lets an agent tell "a human must act" from "the call was wrong".

**Rejected.** Stability of key names alone. Treating every new enum value as compatible without an explicit rule for unknown values. Recording a breaking change as a decision without a version and migration policy. Returning `75` for a lock conflict after effects that make the invocation unsafe to repeat. A dedicated exit code 3 for a missing key, or exit 2 for a step that only a human can do.
