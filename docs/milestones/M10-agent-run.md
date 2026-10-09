# M10 Agent run

Status: planned
Version: v0.10.0

## Demo

`ncly agent run` answers through claude or pi with tools off, and `ncly run view` shows the record of that run.

## Scope

`ncly agent run` can bill, so it records its execution before the harness starts. It is the first and simplest consumer of run records: one item and one step. Record support starts in `sdk/`, as [Records and evidence](../north-star/cli-spec.md#records-and-evidence) defines, and core inspects them with [ncly run](../north-star/cli-spec.md#ncly-run) `list` and `view`. Resume waits for [M13](M13-resume.md). `run list` and `run view` use the looks that Pascal picked in M07.

## Open questions

Settle each one in [extensions/agent/spec.md](../../extensions/agent/spec.md) and [Record format](../north-star/cli-spec.md#record-format), then delete this section.

- The item key of an agent run, which `ncly run list --key` filters on, and the `inputs` that it records, such as a fingerprint of the material
- Where the SDK's record helpers stop and the extension's own evidence starts

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M10-T1 | Records and agent run | agent | — | todo |
| M10-T2 | Run list | agent | M10-T1 | todo |
| M10-T3 | Run view | agent | M10-T1 | todo |
| M10-T4 | Cancellation and failures | agent | M10-T1 | todo |
| M10-T5 | Real harness check | agent | M10-T4 | todo |

### M10-T1 Records and agent run

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent), [Records and evidence](../north-star/cli-spec.md#records-and-evidence), [Record format](../north-star/cli-spec.md#record-format)
- **Proves:** `testdata/script/agent_run.txtar`, `testdata/script/agent_run_record.txtar`

`ncly agent run` sends the task to a stub harness and answers with `run_id` and the answer. Before the harness starts, the SDK writes the record atomically under its lock and completes the persistence barrier, and the step turns `unknown`. A failed barrier prevents the start.

### M10-T2 Run list

- **Read:** [ncly run](../north-star/cli-spec.md#ncly-run)
- **Proves:** `testdata/script/run_list.txtar`

`ncly run list` sorts, filters before the limit, and reports `more` and bounded item-key previews. An invalid time is `USAGE_INVALID`.

### M10-T3 Run view

- **Read:** [ncly run](../north-star/cli-spec.md#ncly-run), [Inspection and explicit resume](../north-star/cli-spec.md#inspection-and-explicit-resume)
- **Proves:** `testdata/script/run_view.txtar`

`ncly run view` returns the record plus `active`, which comes from the lock and never from the saved status. A failed saved run still exits 0. An unknown run is `NOT_FOUND`, and an unsupported record version is reported, never read as a known format.

### M10-T4 Cancellation and failures

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent), [Programs that ncly runs](../north-star/cli-spec.md#programs-that-ncly-runs), [Retry safety](../north-star/cli-spec.md#retry-safety)
- **Proves:** `testdata/script/agent_run_cancel.txtar`, `testdata/script/agent_run_failures.txtar`

Ctrl-C, SIGTERM, and the timeout stop the harness's process tree and leave the step `unknown`, with exit 130, 143, or 1. A harness that never started fails the step. A nonzero exit or a missing answer after the start leaves it `unknown` with exit 1, and nothing launches the task again.

### M10-T5 Real harness check

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** the observations in the pull request, and the check date in `extensions/agent/spec.md`

Recheck the adapters' flags against the installed `claude` and `pi`, then run each one once with material that tries to use tools, hooks, MCP servers, and loaded extensions, and observe that it cannot. A stub accepting flags or a harness returning text is not that proof. Each run may bill a little, so ask Pascal before starting.
