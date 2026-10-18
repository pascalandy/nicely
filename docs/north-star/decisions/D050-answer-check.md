# D050 Check an extension's answer before the caller sees it

Decided 2026-10-10. Validate captured machine answers before publication. Invalid protocol answers fail with `RUNTIME`, exit 1. Human text passes through with an exit-code check. [Answer check](../contract.md#answer-check) owns result-kind exceptions and [Error codes](../contract.md#error-codes) permits unknown codes.

**Why.** Checking a copy cannot retract an invalid answer. Human forms need the real terminal. Exit codes let older callers handle new error codes.

**Rejected.** Pass-through JSON, a special protocol-failure code with the same recovery as `RUNTIME`, and rejection of unfamiliar core codes.
