# M16 Transcript finish

Status: planned
Version: v0.16.0

## Demo

A fresh agent session discovers, checks, and runs the one-video transcript command, then recovers its verified result by run ID.

## Scope

Finish the YouTube path of [M13](M13-transcript.md) with doctor checks, the transcript skill, and fresh-session evidence.
The [transcript spec](../../extensions/transcript/spec.md) owns its behavior.
Zoom, prompt discovery, `--open`, progress displays, and additional browser support remain in
[M99](M99-parking-lot.md#transcript-extras).

## Open questions

Settle each one in [extensions/transcript/spec.md](../../extensions/transcript/spec.md), then delete this section.

- The flags, check IDs, and codes of `ncly transcript status`, limited to the selected YouTube path and its proven dependencies
- Which free Deepgram endpoint can verify authentication under `--live`, with no upload or bill
- The transcript skill's minimal guidance on result files, chosen run IDs, inspection, and refusal of uncertain paid work

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M16-T1 | Doctor for YouTube | agent | — | todo |
| M16-T2 | Transcript skill | agent | — | todo |
| M16-T3 | Fresh-session check | agent | M16-T1, M16-T2 | todo |

### M16-T1 Doctor for YouTube

- **Read:** [Output and cookies](../../extensions/transcript/spec.md#output-and-cookies), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Status command](../north-star/contract.md#status-command)
- **Proves:** `extensions/transcript/testdata/script/transcript_doctor.txtar`, `extensions/transcript/testdata/script/transcript_cookies.txtar`

Add `ncly transcript status` so `ncly doctor transcript` reports the dependencies established in M13 and the selected cookie path.
Without `--live`, use installed-version probes only, including a local Python probe.
Never start `uv run` to check installation because it may contact the package index.
Under `--live`, check authentication only through the chosen free Deepgram endpoint.
Unreadable configured cookies produce the specified warning.

### M16-T2 Transcript skill

- **Read:** [ncly transcript](../../extensions/transcript/spec.md#ncly-transcript), [Records and replay](../../extensions/transcript/spec.md#records-and-replay)
- **Proves:** `extensions/transcript/SKILL.md`

Teach the one-video command and its result files.
Choose a run ID before execution, preserve the full answer, and inspect a failed run before another action.
Explain verified-result replay and why an unknown paid outcome cannot be retried.
Use only commands that have shipped.

### M16-T3 Fresh-session check

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Records and replay](../../extensions/transcript/spec.md#records-and-replay)
- **Proves:** the session's transcript, attached to the pull request

Give a fresh agent only the line that points to `ncly skill view nicely`.
It discovers the transcript command, plans one URL, and runs it under a chosen run ID.
It inspects a recorded failure without reading full transcripts or logs.
It retrieves a verified result after lost stdout and refuses an unknown paid outcome.
Use the established test transport for counted calls; claim live service proof only when that separate check actually ran.
Fix the skills until the session succeeds.

## After this milestone

These steps wait on Pascal and block no card. Pascal decides when.

- Pascal moves and grants his Deepgram key on each machine, then runs the paid check against the README's test video locally on macOS.
  Omarchy's live check waits for [#23](https://github.com/pascalandy/nicely/issues/23).
- Keep the existing Python CLI and its skills for every workflow Nicely does not yet cover.
  Retire or redirect them only after Pascal confirms that Nicely covers his actually required uses.
  The narrower YouTube milestone alone does not establish migration parity.
