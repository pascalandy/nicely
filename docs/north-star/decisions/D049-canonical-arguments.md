# D049 Pass an extension its validated call in one canonical form

Decided 2026-10-10, revised 2026-10-12. Core validates and normalizes extension arguments under [Dispatch](../contract/dispatch.md). Resolved global flags travel through environment variables; `--dry-run` and `--run-id` travel as arguments.

**Why.** Extensions need not reimplement mixed shorthands, global precedence, or quoting. Unlike Git, Nicely already parses the call, so a shell extension can consume the simpler form.

**Rejected.** Unchanged caller arguments duplicate parsing. JSON on stdin would occupy the extension's input.
