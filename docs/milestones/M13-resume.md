# M13 Resume

Status: planned
Version: v0.13.0

## Demo

`ncly run resume` continues a saved transcript run without transcribing again, and refuses to repeat uncertain paid work.

## Scope

Core finds the extension that owns a record from its command path and hands it the resume. The extension decides what to reuse and what to execute, as [Inspection and explicit resume](../north-star/cli-spec.md#inspection-and-explicit-resume) and transcript's [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume) define.

## Open questions

Settle each one in [ncly run](../north-star/cli-spec.md#ncly-run) and [Extensions](../north-star/cli-spec.md#extensions), then delete this section.

- How core hands a resume to the extension that owns a record: the manifest's `resume` mode, and the argument or variable that names the run
- How core answers `RESUME_UNSAFE` for a command whose manifest declares no resume, before starting the extension

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M13-T1 | Resume dispatch | agent | — | todo |
| M13-T2 | Reuse saved work | agent | M13-T1 | todo |
| M13-T3 | Refuse unsafe work | agent | M13-T2 | todo |
| M13-T4 | Resume a summary run | agent | M13-T2 | todo |
| M13-T5 | Recover lost output | agent | M13-T2 | todo |

### M13-T1 Resume dispatch

- **Read:** [ncly run](../north-star/cli-spec.md#ncly-run), [Extensions](../north-star/cli-spec.md#extensions)
- **Proves:** `testdata/script/run_resume.txtar`

`ncly run resume <run-id>` hands the run to the extension that owns it. An unknown run is `NOT_FOUND`, a run held by another process exits 75 with `TEMPORARY` and a `run view` hint before any effect, and an unsupported command or record version is `RESUME_UNSAFE`. Dry run describes the remaining steps only.

### M13-T2 Reuse saved work

- **Read:** [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume)
- **Proves:** `testdata/script/transcript_resume.txtar`

Resume a saved transcript whose summary never started: counted calls prove that transcription is not repeated, and a summary-only resume needs no `uv` and no Deepgram key. Change the config's profile, and the remaining steps still use the recorded one. A completed run returns its verified result without another paid call. A lost completion message is reconciled only when every recorded artifact verifies.

### M13-T3 Refuse unsafe work

- **Read:** [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume), [Inspection and explicit resume](../north-star/cli-spec.md#inspection-and-explicit-resume)
- **Proves:** `testdata/script/transcript_resume_unsafe.txtar`

A changed input or artifact, missing or mismatched evidence, an unknown summary, and an unknown transcription without its complete artifact set each return `RESUME_UNSAFE`, with the available artifacts, and `--force` overrides none of them. The run that ended with a 504 after upload is refused with no replay. A failed attempt whose paid step stays pending keeps its code, message, and hint.

### M13-T4 Resume a summary run

- **Read:** [ncly transcript summary run](../../extensions/transcript/spec.md#ncly-transcript-summary-run)
- **Proves:** `testdata/script/transcript_summary_resume.txtar`

`ncly run resume` on a summary run verifies the copied transcript evidence and continues its eligible pending or failed summaries without transcribing. A rejected item keeps its saved error.

### M13-T5 Recover lost output

- **Read:** [ncly run](../north-star/cli-spec.md#ncly-run)
- **Proves:** `testdata/script/run_recovery.txtar`

Lose stdout, add more than 20 unrelated runs, and recover the run through filters applied before the limit, with `more` and bounded item-key previews. Repeat without a remembered start time. Preserve ambiguous matches, and keep the saved `run.status` apart from the lock-based `active`.
