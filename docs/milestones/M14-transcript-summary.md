# M14 Transcript summary

Status: planned
Version: v0.14.0

## Demo

A transcript run summarizes through `ncly agent run` with tools off, and `ncly transcript summary run` summarizes a saved run again without a second transcription.

## Scope

Transcript composes with the `agent` extension through the CLI, the same way an agent would. The agent extension owns profiles, adapters, tools-off mode, and the depth limit, and transcript owns the summary step and its record. See [extensions/transcript/spec.md](../../extensions/transcript/spec.md#ncly-transcript-summary-run) and [D037](../north-star/decisions/D037-summary-operation.md).

## Open questions

Settle each one in [extensions/transcript/spec.md](../../extensions/transcript/spec.md), then delete this section.

- Which fields of the agent's answer transcript records beside the child run ID that its intent already holds, such as the harness version as a tool

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M14-T1 | Summary through the agent | agent | — | todo |
| M14-T2 | Summary failures | agent | M14-T1 | todo |
| M14-T3 | Summary run | agent | M14-T1 | todo |
| M14-T4 | Mixed and busy sources | agent | M14-T3 | todo |

### M14-T1 Summary through the agent

- **Read:** [ncly transcript](../../extensions/transcript/spec.md#ncly-transcript), [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume), [ncly agent](../../extensions/agent/spec.md#ncly-agent)
- **Proves:** `extensions/transcript/testdata/script/transcript_summary.txtar`

Preparation calls `ncly agent run --dry-run --json` with the profile before any transcription dispatch, and records the resolved profile, which the summary passes back to `ncly agent run` as flags, with the prompt as `--task-file`, a copy of the transcript as `--input`, the item's `output_dir` as `--cwd`, and a child run ID that the step's intent records first. The summary is its own recorded step, written as `<prompt-name>-<run-id>.md`. Without a profile, the run saves the transcript and adds `TRANSCRIPT_SUMMARY_SKIPPED`. `--no-summary` excludes `--profile` and `--prompt`. With a profile but without the agent extension, the run fails with `PREREQ_MISSING` before transcription.

### M14-T2 Summary failures

- **Read:** [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume)
- **Proves:** `extensions/transcript/testdata/script/transcript_summary_failures.txtar`

Stop after saving the transcript and before starting the summary. An agent run that exits 2 or 78 on the attempt's first call fails the step. Any other end without a saved summary leaves it `unknown`, with no silent retry, and its hint and `next` name `ncly transcript summary run <run-id> --key <item key>`.

### M14-T3 Summary run

- **Read:** [ncly transcript summary run](../../extensions/transcript/spec.md#ncly-transcript-summary-run)
- **Proves:** `extensions/transcript/testdata/script/transcript_summary_run.txtar`

`ncly transcript summary run <run-id>` resolves its profile and prompt, records the verified transcript references in the evidence of each `summarize` step of a new record with `source_run`, without a `transcribe` step, and writes `<prompt-name>-<new-run-id>.md` without transcribing again. `--key` selects items, and an unknown key is `NOT_FOUND` before any effect.

### M14-T4 Mixed and busy sources

- **Read:** [ncly transcript summary run](../../extensions/transcript/spec.md#ncly-transcript-summary-run)
- **Proves:** `extensions/transcript/testdata/script/transcript_summary_mixed.txtar`, `extensions/transcript/testdata/script/transcript_summary_busy.txtar`

Summarize a mixed source batch: rejected items stay failed summaries that nothing ever executes, and eligible ones run. A source with no eligible item creates no record and returns no `run_id`. While the source's executor is active, a new summary exits 75 with `TEMPORARY` and a `run view` hint, before any effect. Once a new summary has recorded its references, it releases the source, and `ncly run view` of the source reports `active` as `false` while the summary still runs.
