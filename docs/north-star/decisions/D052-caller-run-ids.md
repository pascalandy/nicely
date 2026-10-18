# D052 Let the caller name a run, so that a repeated call continues it

Decided 2026-10-12, revised 2026-10-16. Let callers select `--run-id` before execution. [Run IDs](../contract/run-ids.md) owns replay, input comparison, and refusal of unsafe repeats. Successful answers come from committed records and verified artifacts.

**Scope update, 2026-10-17.** M10 and M13 prove one item and one paid step. [M99](../../milestones/M99-parking-lot.md#resume) defers explicit resume, child-run recovery, and policy-migration fixtures.

A future parent step must persist its child's ID before dispatch. It may repeat that call only while the child's record exists. A deleted record could belong to a child that already billed.

**Why.** A caller-chosen ID survives lost output without a record search. Compare stated inputs, not resolved defaults, so a config edit does not invalidate the same call. Rebuild answers from saved data so presentation can change without execution.

**Rejected.** Child-only parent references, IDs reused with different inputs, resolved-default comparison, saved rendered answers, and IDs treated as permission to repeat unknown effects.
