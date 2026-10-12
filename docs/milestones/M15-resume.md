# M15 Resume

Status: planned
Version: v0.15.0

## Demo

`ncly run resume` continues a saved transcript run without transcribing again, and refuses to repeat uncertain paid work.

## Scope

Core finds the extension that owns a record from its command path and hands it the resume. The extension decides what to reuse and what to execute, as [Inspection and explicit resume](../north-star/contract.md#inspection-and-explicit-resume) and transcript's [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume) define.

## Open questions

Settle each one in [ncly run](../north-star/core-spec.md#ncly-run) and [Extensions](../north-star/contract.md#extensions), then delete this section.

- How core hands a resume to the extension that owns a record: the manifest's `resume` mode, and how the extension receives the run, such as its command path with `--run-id`, since a call with an existing run ID already continues its run
- How core answers `RESUME_UNSAFE` for a command whose manifest declares no resume, before starting the extension
- Whether a record keeps the source of the extension that wrote it, as [M08](M08-grants.md) defines a source, so that resume refuses a run whose domain now comes from another source

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M15-T1 | Resume dispatch | agent | — | todo |
| M15-T2 | Reuse saved work | agent | M15-T1 | todo |
| M15-T3 | Refuse unsafe work | agent | M15-T2 | todo |
| M15-T4 | Resume a summary run | agent | M15-T2 | todo |
| M15-T5 | Recover lost output | agent | M15-T2 | todo |

### M15-T1 Resume dispatch

- **Read:** [ncly run](../north-star/core-spec.md#ncly-run), [Extensions](../north-star/contract.md#extensions)
- **Proves:** `testdata/script/run_resume.txtar`

`ncly run resume <run-id>` hands the run to the extension that owns it, and the `nicely` skill gains its line. An unknown run is `NOT_FOUND`, a run held by another process exits 75 with `TEMPORARY` and a `run view` hint before any effect, and an unsupported command or record version is `RESUME_UNSAFE`. Dry run describes the remaining steps only.

### M15-T2 Reuse saved work

- **Read:** [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume), [Run IDs](../north-star/contract.md#run-ids)
- **Proves:** `extensions/transcript/testdata/script/transcript_resume.txtar`

Resume a saved transcript whose summary never started: counted calls prove that transcription is not repeated, and a summary-only resume needs no `uv` and no Deepgram key. A resume that reuses the completed transcription and fails for a temporary cause before the summary's intent exits 75, because it started no step. The same call with the run's `--run-id` resumes it as `ncly run resume` does. The same ID with another URL, and a summary run's ID with another `--key`, each answer `USAGE_INVALID` before any effect, and the counted calls prove that neither starts a program. Change the config's profile, and the remaining steps still use the recorded one. A completed run returns its verified result without another paid call. A lost completion message is reconciled only when every recorded artifact verifies.

### M15-T3 Refuse unsafe work

- **Read:** [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume), [Inspection and explicit resume](../north-star/contract.md#inspection-and-explicit-resume)
- **Proves:** `extensions/transcript/testdata/script/transcript_resume_unsafe.txtar`

A changed input or artifact, missing or mismatched evidence, a summary whose child agent run did not complete or has no record, and an unknown transcription without its complete artifact set each return `RESUME_UNSAFE`, with the available artifacts, and `--force` overrides none of them. In a batch, such a refusal fails its item while the other items continue, and a changed prompt refuses the whole run. Resumed from another folder, an unknown summary whose child completed publishes the child's saved answer, because the call passes the recorded `output_dir` as `--cwd`, and the counted stub harness runs no second time. Delete that child's record, and the summary stays `unknown`, with `next` naming `ncly transcript summary run <source-run-id> --key <item key>`, while the harness still runs no second time. A stub agent extension that answers the repeated call with `USAGE_INVALID` leaves the summary `unknown` with `RESUME_UNSAFE` too. While another process holds the child run, the summary stays `unknown`, and the item fails with `TEMPORARY`, whose `next` is `ncly run view <child-run-id>`, with no harness start; once the holder completes the run, a resume publishes its saved answer. The run that ended with a 504 after upload is refused with no replay. A failed attempt whose paid step stays pending keeps its code, message, and hint.

### M15-T4 Resume a summary run

- **Read:** [ncly transcript summary run](../../extensions/transcript/spec.md#ncly-transcript-summary-run)
- **Proves:** `extensions/transcript/testdata/script/transcript_summary_resume.txtar`

`ncly run resume` on a summary run verifies the transcript that each step's evidence names, and continues its eligible pending or failed summaries without transcribing. A rejected item keeps its saved error.

### M15-T5 Recover lost output

- **Read:** [ncly run](../north-star/core-spec.md#ncly-run)
- **Proves:** `testdata/script/run_recovery.txtar`

Lose stdout, add more than 20 unrelated runs, and recover the run through filters applied before the limit, with `more` and bounded item-key previews, as a caller without a run ID must. A caller that passed `--run-id` recovers the same run with `ncly run view <run-id>` alone. Repeat without a remembered start time. Preserve ambiguous matches, and keep the saved `run.status` apart from the lock-based `active`.
