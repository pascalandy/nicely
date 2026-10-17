# M10 Agent run

Status: active
Version: v0.10.0

## Demo

`ncly agent run "Review this branch" --tools` saves one Codex answer, and `ncly run view` inspects that run without launching it again.

## Scope

An agent task can bill and modify its working directory. Preserve a record before launch, a verified answer after success, and an honest result after interruption. This milestone implements one item, `task`, and one step, `paid`, through the record support in `sdk/`. [The agent spec](../../extensions/agent/spec.md#records-and-replay) fixes its inputs and evidence.

The record format and run-ID rules still apply. Partial multi-step continuation, child-run recovery, effect-policy migration fixtures, and explicit resume stay in [M99](M99-parking-lot.md#resume). Do not implement them to prepare for transcript or another future consumer.

Core keeps `run list` and `run view`: a human can recover a generated ID after losing output. They read the same records and need no index or database. Their text output takes the looks that [ncly run](../north-star/core-spec.md#ncly-run) records from M07; add no new design.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M10-T1 | One-step records | agent | — | done |
| M10-T2 | Codex execution and replay | agent | M10-T1 | todo |
| M10-T3 | Run list | agent | M10-T1 | todo |
| M10-T4 | Run view | agent | M10-T1 | todo |
| M10-T5 | Cancellation and failures | agent | M10-T2, M10-T4 | todo |

### M10-T1 One-step records

- **Read:** [Records and evidence](../north-star/contract.md#records-and-evidence), [Record format](../north-star/contract.md#record-format), [Run IDs](../north-star/contract.md#run-ids), [Records and replay](../../extensions/agent/spec.md#records-and-replay), [Codex invocation](../../extensions/agent/spec.md#codex-invocation)
- **Proves:** `extensions/agent/testdata/script/agent_run_record.txtar`

Add record support in `sdk/`, using the durable writes and locks of `sdk/platform/`. Record one `task` item and one `paid` step, and run one Codex task under that record through `sdk/runner`, with the preparation of M09. A counted stub Codex stands for it in the scenario. Persist intent and complete the durability barrier before the stub starts. A failed barrier starts nothing. Publish verified `answer.txt` and completion before returning success.

Prove the run lock, unsupported or unreadable record refusal, and saved artifact fingerprints. Keep the shared record format; no generic workflow executor or multi-step fixture belongs to this card. Until M10-T2, a call whose run ID a record holds starts nothing.

### M10-T2 Codex execution and replay

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent), [Records and replay](../../extensions/agent/spec.md#records-and-replay), [Run IDs](../north-star/contract.md#run-ids), [Retry safety](../north-star/contract.md#retry-safety)
- **Proves:** `extensions/agent/testdata/script/agent_run.txtar`

Answer a repeated call from the record of M10-T1. The same completed invocation and run ID return the verified saved result after a configuration change, with no second harness call, not even the version probe.

Different stated inputs under that ID fail with `USAGE_INVALID`. Unknown paid work fails with `RESUME_UNSAFE`. Concurrent calls under one ID start at most one task; the contender exits 75 while the run lock is held. Never infer from a missing answer that no effect occurred. Prove that the task receives `NCLY_AGENT_DEPTH` plus one, as [Depth limit](../../extensions/agent/spec.md#depth-limit) defines, and no key.

### M10-T3 Run list

- **Read:** [ncly run](../north-star/core-spec.md#ncly-run)
- **Proves:** `testdata/script/run_list.txtar`

Add `ncly run list` in `internal/cli`. Filter before the limit and report `more`, so lost output remains discoverable without guessing the newest run. Read records directly and preserve the bounded item-key preview.

### M10-T4 Run view

- **Read:** [ncly run](../north-star/core-spec.md#ncly-run), [Record format](../north-star/contract.md#record-format)
- **Proves:** `testdata/script/run_view.txtar`

Add `ncly run view` in `internal/cli`. Return the saved record and lock-based `active`. A failed saved operation still inspects with exit 0. Missing and unsupported records have explicit errors. Teach `run list`, `run view`, and caller-chosen run IDs in the `nicely` skill.

### M10-T5 Cancellation and failures

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs), [Retry safety](../north-star/contract.md#retry-safety)
- **Proves:** `extensions/agent/testdata/script/agent_run_cancel.txtar`, `extensions/agent/testdata/script/agent_run_failures.txtar`, `extensions/agent/testdata/script/agent_run_boundaries.txtar`

Ctrl-C, SIGTERM, and timeout stop the process tree. A task that may have started remains `unknown`; nothing retries it automatically. Signals retain exits 130 and 143. A failed start can be `failed`, and a nonzero exit or missing answer after launch is `unknown`.

Use the test-only boundary hook on this one paid step. Kill at every real boundary of this one step: before intent, after durable intent, after the counted harness effect but before durable completion, and after verified completion. The effect-to-completion window must remain unknown and must not launch a second task. Inspect the record from another process and compare counted calls. Preserve the difference between saved status and lock ownership.

## After this milestone

Pascal authorizes one small Codex task when ready. An independent agent then checks dry run, execution, the saved answer, and replay with no second request on the installed harness. Record the version and evidence before claiming real Codex compatibility. A stub accepting flags is not that proof. No Claude, Pi, Grok, or OpenCode validation belongs to this check.
