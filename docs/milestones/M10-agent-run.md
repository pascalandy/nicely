# M10 Agent run

Status: planned
Version: v0.10.0

## Demo

`ncly agent run` answers through claude or pi with tools off, and `ncly run view` shows the record of that run.

## Scope

`ncly agent run` can bill, so it records its execution before the harness starts. It is the first and simplest consumer of run records: one item and one step. Record support starts in `sdk/`, as [Records and evidence](../north-star/contract.md#records-and-evidence) defines, and core inspects them with [ncly run](../north-star/core-spec.md#ncly-run) `list` and `view`. Resume waits for [M15](M15-resume.md). `run list` and `run view` use the looks that Pascal picked in M07.

## Open questions

Settle each one in [extensions/agent/spec.md](../../extensions/agent/spec.md) and [Record format](../north-star/contract.md#record-format), then delete this section.

- The item key of an agent run, which `ncly run list --key` filters on, and the `inputs` that it records, such as a fingerprint of the material
- How a caller links a child run to its own step before the child can bill, so the link survives a lost answer: a `--run-id` that the caller chooses and records in its intent, which also lets an agent find a run after lost output and refuses a second run with the same ID, or a parent reference that the child records from its environment, which needs a scan of the records to find a child

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M10-T1 | Records and agent run | agent | — | todo |
| M10-T2 | Run list | agent | M10-T1 | todo |
| M10-T3 | Run view | agent | M10-T1 | todo |
| M10-T4 | Cancellation and failures | agent | M10-T1 | todo |
| M10-T5 | Real harness check | agent | M10-T4 | todo |

### M10-T1 Records and agent run

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent), [Operations](../north-star/contract.md#operations), [Records and evidence](../north-star/contract.md#records-and-evidence), [Record format](../north-star/contract.md#record-format), [Retry safety](../north-star/contract.md#retry-safety)
- **Proves:** `testdata/script/agent_run.txtar`, `testdata/script/agent_run_record.txtar`, `testdata/script/record_retry_safety.txtar`

`ncly agent run` sends the task to a stub harness and answers with `run_id` and the answer. Before the harness starts, the SDK writes the record atomically under its lock and completes the persistence barrier, and the step turns `unknown`. A failed barrier prevents the start.

The step declares the `paid` effect. The SDK writes every step transition and derives retry safety from the declared effects, so `ncly-agent` never reports an effect level of its own. A fixture extension with a `repeatable` step, then a `paid` one, proves the derivation: a temporary failure exits 75 while the paid step is `pending`, and 1 once that step has left `pending`, even as `failed`.

### M10-T2 Run list

- **Read:** [ncly run](../north-star/core-spec.md#ncly-run)
- **Proves:** `testdata/script/run_list.txtar`

`ncly run list` sorts, filters before the limit, and reports `more` and bounded item-key previews. An invalid time is `USAGE_INVALID`.

### M10-T3 Run view

- **Read:** [ncly run](../north-star/core-spec.md#ncly-run), [Inspection and explicit resume](../north-star/contract.md#inspection-and-explicit-resume)
- **Proves:** `testdata/script/run_view.txtar`

`ncly run view` returns the record plus `active`, which comes from the lock and never from the saved status. A failed saved run still exits 0. An unknown run is `NOT_FOUND`, and an unsupported record version is reported, never read as a known format.

### M10-T4 Cancellation and failures

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs), [Retry safety](../north-star/contract.md#retry-safety), [Records and evidence](../north-star/contract.md#records-and-evidence)
- **Proves:** `testdata/script/agent_run_cancel.txtar`, `testdata/script/agent_run_failures.txtar`, `testdata/script/agent_run_boundaries.txtar`

Ctrl-C, SIGTERM, and the timeout stop the harness's process tree and leave the step `unknown`, with exit 130, 143, or 1. The record then reads `interrupted` after a signal and `unknown` after the timeout, which proves that cleanup wrote it after cancellation, because the step was already `unknown` before the harness started. A harness that never started fails the step. A nonzero exit or a missing answer after the start leaves it `unknown` with exit 1, and nothing launches the task again.

Build the boundary hook, which lives only in the test binary and kills the process at a chosen step boundary, as SIGKILL would. The boundary scenario kills `ncly-agent` at each of its step's boundaries. After each kill, core answers exit 1 with `RUNTIME`, and `ncly run view` reports `active` as `false`, with the step `pending` before its intent persists, `unknown` from then until its final status persists, and `completed` after. The counted stub harness ran zero times at every boundary before the effect and once at every boundary after it.

### M10-T5 Real harness check

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** the observations in the pull request, and the check date in `extensions/agent/spec.md`

Recheck the adapters' flags against the installed `claude` and `pi`, then run each one once with material that tries to use tools, hooks, MCP servers, and loaded extensions, and observe that it cannot. A stub accepting flags or a harness returning text is not that proof. Each run may bill a little, so ask Pascal before starting.
