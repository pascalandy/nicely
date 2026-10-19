# M13 Transcript

Status: done
Version: v0.13.0

## Demo

One explicit YouTube URL produces verified transcript files and an inspectable run through real `uv`, the Python engine, and a test transport.

## Scope

`transcript` is a first-party external extension in `extensions/transcript/`, built as `ncly-transcript`.
It installs with the `just install` of [M09](M09-agent-profiles.md).
The [transcript spec](../../extensions/transcript/spec.md) defines one YouTube video and one `transcribe` step.
A human grants its required Deepgram key.

Copy only the YouTube engine of `authoring/andy/transcript/` in `pascalandy/skills` at `0c9c4a2`, with the tests of the code it keeps.
Leave out summary, Zoom, queues, prompts, discovery, previews, progress, and the legacy CLI.
Each new file stays under 10,000 bytes, as `just limits` requires.
Embed the engine with `go:embed`, extract it into Nicely's cache, and run it through `uv`.

The command records intent before paid dispatch, verifies artifacts, locks runs, handles cancellation, and replays verified results.
An incomplete run never causes a second upload under the same run ID.
Summary, resume, and extra transcript features stay in
[M99](M99-parking-lot/transcript-summary.md), [M99](M99-parking-lot/resume.md), and
[M99](M99-parking-lot/transcript-extras.md).

Scenarios use the test keychain, stubs, and counted calls.
One path runs real Go, real `uv`, and the engine with a test transport that only the test binary installs.
No scenario sends a paid request or adds a production endpoint override.
Each card grows the transcript skill with what it ships.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M13-T1 | Python engine | agent | — | done |
| M13-T2 | Protocol and dry run | agent | M13-T1 | done |
| M13-T3 | Transcription with a record | agent | M13-T2 | done |
| M13-T4 | Failures and replay | agent | M13-T3 | done |
| M13-T5 | Interruption and contenders | agent | M13-T3, M13-T4 | done |

### M13-T1 Python engine

- **Read:** [Output and cookies](../../extensions/transcript/spec/output-and-cookies.md), [Protocol with the Python program](../../extensions/transcript/spec/protocol.md)
- **Proves:** `extensions/transcript/python/`, whose checks run in just check

Copy the engine and the tests of the code it keeps.
Run ruff, pyright, and pytest through `uv` in a `python` recipe that `check` runs.
Fix the nine pyright errors of the source rather than suppress them.
Goldens made by the source program at `0c9c4a2` prove the three files for one recorded Deepgram response.
`PROVENANCE.md` names the source SHA, the MIT notice, the removed features, and each intentional change.

### M13-T2 Protocol and dry run

- **Read:** [Protocol with the Python program](../../extensions/transcript/spec/protocol.md), [Output and cookies](../../extensions/transcript/spec/output-and-cookies.md), [Dry-run answer](../north-star/contract/dry-run-answer.md)
- **Proves:** `extensions/transcript/testdata/script/transcript_protocol.txtar`, `extensions/transcript/testdata/script/transcript_dry_run.txtar`, `extensions/transcript/python/`

Add the JSON Lines protocol to the engine, then embed, extract, and run it from Go.
Python tests prove each sequence rule and each row of the program-code table.
Unlike the source, a 5xx after the upload is `deepgram_unknown`, and no retry repeats an upload.
yt-dlp gets `--no-playlist` and `--ignore-config`, and a cookie flag only when `browser` is `arc`.
The script header asks for yt-dlp at its minimum or newer and pins every other dependency.
A stub `uv` proves that Go refuses a missing plan, a plan for another item, and an unknown version.
Real `uv` and Python prove the dry run reads no key and creates no record or result folder.
Add the transcript feature to verify-nicely, and create the transcript skill.

### M13-T3 Transcription with a record

- **Read:** [Records and replay](../../extensions/transcript/spec/records-and-replay.md), [Protocol with the Python program](../../extensions/transcript/spec/protocol.md), [Record format](../north-star/contract/record-format.md), [Run IDs](../north-star/contract/run-ids.md), [ncly run](../north-star/core-spec.md#ncly-run)
- **Proves:** `extensions/transcript/testdata/script/transcript_run.txtar`, `extensions/transcript/testdata/script/transcript_barrier.txtar`, `extensions/transcript/testdata/script/transcript_boundaries.txtar`

Record the plan, record each candidate folder, then create it with an exclusive `mkdir`.
Complete the persistence barrier before `ack`; a failed barrier sends no `ack` and starts no upload.
After the completed step, Go checks the three files, computes sizes and hashes, and syncs them before it records completion.
Prove one real-uv run with one counted upload, then read it with `ncly run view` in a new process.
The boundary scenario kills Go before and after intent and completion, then checks the record, files, uploads, and a repeat.
A crash after a name collision leaves the existing folder untouched and starts no upload.
Register `AUTH_REJECTED`; a 401 gives that item code under an overall `RUNTIME`.

### M13-T4 Failures and replay

- **Read:** [Records and replay](../../extensions/transcript/spec/records-and-replay.md), [Protocol with the Python program](../../extensions/transcript/spec/protocol.md), [Retry safety](../north-star/contract/retry-safety.md), [Keys](../north-star/contract/keys.md)
- **Proves:** `extensions/transcript/testdata/script/transcript_failures.txtar`, `extensions/transcript/testdata/script/transcript_protocol_errors.txtar`, `extensions/transcript/testdata/script/transcript_replay.txtar`, `extensions/transcript/testdata/script/transcript_keys.txtar`

A 504 after the upload counts one upload and leaves the step unknown; a repeat refuses with `RESUME_UNSAFE`.
A stub `uv` proves each protocol failure that the spec lists.
A completed run replays after a config change without starting `uv` or uploading.
A changed or missing artifact refuses, and another URL or `--output` fails before any effect.
Prove missing keys, missing grants, an unavailable keychain, and a granted environment key.

### M13-T5 Interruption and contenders

- **Read:** [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md), [Records and evidence](../north-star/contract/records-and-evidence.md), [Records and replay](../../extensions/transcript/spec/records-and-replay.md)
- **Proves:** `extensions/transcript/testdata/script/transcript_cancel.txtar`, `extensions/transcript/testdata/script/transcript_contenders.txtar`

Ctrl-C after `ack` exits 130, keeps `output_dir` in the answer, and leaves no child process.
SIGTERM exits 143, and a signal after durable completion keeps the completion.
A second call under a busy run ID exits 75 and starts nothing.
Two new runs with one title write two folders.
