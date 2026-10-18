# M16 Transcript finish

Status: planned
Version: v0.16.0

## Demo

A fresh agent session discovers, checks, and runs the one-video transcript command, then recovers its verified result by run ID.

## Scope

Finish the YouTube path of [M13](M13-transcript.md) with doctor checks and fresh-session evidence.
The transcript skill ships with M13, and this milestone adds the doctor step to it.
The [transcript spec](../../extensions/transcript/spec.md) owns its behavior.
Zoom, prompt discovery, `--open`, progress displays, and additional browser support remain in
[M99](M99-parking-lot/transcript-extras.md).

## Open questions

Settle each one in the [transcript spec](../../extensions/transcript/spec.md), then delete this section.

- The flags and check IDs of `ncly transcript status`. Doctor already probes the programs of the manifest, so status reports only the cookie check and, under `--live`, the Deepgram check
- Documented proof that `GET /v1/auth/token` bills nothing. Without that proof, ship the local checks and park `--live` in M99
- How `--live` reports a 401, a timeout, and a 5xx. Only a 401 claims that Deepgram rejected the key
- Whether the test binary of `cmd/ncly`, linked as `ncly`, gives the fresh session its test transport

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M16-T1 | Doctor for YouTube | agent | — | todo |
| M16-T2 | Fresh-session check | agent | M16-T1 | todo |

### M16-T1 Doctor for YouTube

- **Read:** [Output and cookies](../../extensions/transcript/spec/output-and-cookies.md), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Status command](../north-star/contract/status-command.md)
- **Proves:** `extensions/transcript/testdata/script/transcript_doctor.txtar`, `extensions/transcript/testdata/script/transcript_cookies.txtar`

Add `ncly transcript status`, following the status command of the skill extension.
Its findings follow the host probes in the doctor component and never repeat them.
Without `--live`, nothing reads a key or starts `uv run`.
Unreadable configured cookies produce the specified warning.
Under `--live`, a server of the test binary sees exactly one request.
The transcript skill tells an agent to run `ncly doctor transcript` first.

### M16-T2 Fresh-session check

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Records and replay](../../extensions/transcript/spec/records-and-replay.md)
- **Proves:** the session's transcript, attached to the pull request

Give a fresh agent only the line that points to `ncly skill view nicely`.
Its `ncly` carries the test transport, as the readiness answer says.
It discovers the transcript command, plans one URL, and runs it under a chosen run ID.
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
