# M16 Transcript finish

Status: done
Version: v0.16.0

## Demo

A fresh agent session discovers, checks, and runs the one-video transcript command, then recovers its verified result by run ID.

## Scope

Finish the YouTube path of [M13](M13-transcript.md) with doctor checks and fresh-session evidence.
The transcript skill ships with M13, and this milestone adds the doctor step to it.
[ncly transcript status](../../extensions/transcript/spec.md#ncly-transcript-status) and [Fresh session](../../extensions/transcript/spec/test-transport.md#fresh-session) own its behavior.
Zoom, prompt discovery, `--open`, progress displays, additional browser support, and the `--live` Deepgram check remain in
[Transcript extras](../backlog/XT-102-transcript.md#transcript-extras).

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M16-T1 | Doctor for YouTube | agent | — | done |
| M16-T2 | Fresh-session check | agent | M16-T1 | done |

### M16-T1 Doctor for YouTube

- **Read:** [ncly transcript status](../../extensions/transcript/spec.md#ncly-transcript-status), [Output and cookies](../../extensions/transcript/spec/output-and-cookies.md), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Status command](../north-star/contract/status-command.md)
- **Proves:** `extensions/transcript/testdata/script/transcript_doctor.txtar`, `extensions/transcript/testdata/script/transcript_cookies.txtar`

Add `ncly transcript status`, following the status command of the skill extension.
Its findings follow the host probes in the doctor component and never repeat them.
Nothing reads a key or starts `uv`: a stub `uv` logs no call, and a stub `python3` logs only `--version`.
A missing `uv` fails with `PREREQ_MISSING` and exit 78.
Unreadable configured cookies produce the specified warning.
The transcript skill tells an agent to run `ncly doctor transcript` first.

### M16-T2 Fresh-session check

- **Read:** [Fresh session](../../extensions/transcript/spec/test-transport.md#fresh-session), [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Records and replay](../../extensions/transcript/spec/records-and-replay.md)
- **Proves:** `scripts/transcript-session.sh`, with the session's transcript and its upload log, attached to the pull request

Give a fresh agent only the line that points to `ncly skill view nicely`.
Its `ncly` carries the test transport, in a folder outside any checkout, as [Fresh session](../../extensions/transcript/spec/test-transport.md#fresh-session) says.
It discovers the transcript command, dry-runs one fixture URL, and runs it under a chosen run ID.
It inspects a recorded failure without reading full transcripts or logs.
It retrieves a verified result after lost stdout and refuses an unknown paid outcome.
Label the evidence as fixture-backed, and claim live service proof only from Pascal's check.
Fix the skills until the session succeeds.

## After this milestone

These steps wait on Pascal and block no card. Pascal decides when.

- Pascal moves and grants his Deepgram key on each machine, then runs the paid check against the README's test video locally on macOS.
  Omarchy's live check waits for [#23](https://github.com/pascalandy/nicely/issues/23).
- Keep the existing Python CLI and its skills for every workflow Nicely does not yet cover.
  Retire or redirect them only after Pascal confirms that Nicely covers his actually required uses.
  The narrower YouTube milestone alone does not establish migration parity.
