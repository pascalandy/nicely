# M13 Transcript

Status: planned
Version: v0.13.0

## Demo

One explicit YouTube URL produces verified transcript files and an inspectable run through real `uv`, the Python program, and a test transport.

## Scope

`transcript` is a first-party external extension in `extensions/transcript/`, built as `ncly-transcript`.
It installs with the `just install` of [M09](M09-agent-profiles.md).
The [transcript spec](../../extensions/transcript/spec.md) defines one YouTube video and one `transcribe` step.
A human grants its required Deepgram key.

Copy the existing program from `authoring/andy/transcript/` in `pascalandy/skills`, with its tests and license material.
Embed it with `go:embed`, extract it atomically into Nicely's cache, and run it through `uv`.
Keep its existing transcription and Arc cookie path.
Leave its other source features unexposed; this milestone does not port them to Go.

The first command retains durable intent before paid dispatch, verified artifacts, run locks, cancellation, and replay of verified results.
An incomplete run never causes a second upload under the same run ID.
Summary, multi-step resume, and extra transcript features are in
[M99](M99-parking-lot.md#transcript-summary), [M99](M99-parking-lot.md#resume), and
[M99](M99-parking-lot.md#transcript-extras).

Scenarios use the test keychain and stubs for failure injection and counted calls.
One path uses actual Go, real `uv`, and the actual Python program with a test-only Deepgram transport.
No scenario sends a paid request or adds a production endpoint override.
These are implementation requirements; this planned milestone has no runtime proof yet.

## Open questions

Settle each one in [extensions/transcript/spec.md](../../extensions/transcript/spec.md), then delete this section.

- Which source revision to copy, and how `just check` runs its ruff, pyright, and pytest checks through `uv`
- How a test-only transport reaches the actual Python program without a production endpoint override
- Which files the program writes besides the three transcripts, and when, so recorded completion covers its actual outputs
- Which download dependencies an empty-cache execution needs, including any `ffmpeg` or JavaScript runtime requirement, and their manifest checks
- Which source cookie behavior proves one useful YouTube path, including Arc on macOS, without requiring a browser matrix
- How Go verifies usable transcript content and complete artifact evidence after a lost final protocol message

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M13-T1 | Python program | agent | — | todo |
| M13-T2 | Protocol and dry run | agent | M13-T1 | todo |
| M13-T3 | Transcription with a record | agent | M13-T2 | todo |
| M13-T4 | Failures and replay | agent | M13-T3 | todo |
| M13-T5 | Interruption and contenders | agent | M13-T3 | todo |

### M13-T1 Python program

- **Read:** [Protocol with the Python program](../../extensions/transcript/spec.md#protocol-with-the-python-program)
- **Proves:** `extensions/transcript/python/`, whose tests run in just check

Copy the program and wire its checks through `uv`.
Its tests save a transcript through the test-only Deepgram transport.
Preserve the existing transcription path while adding the JSON protocol that the wrapper needs.

### M13-T2 Protocol and dry run

- **Read:** [Protocol with the Python program](../../extensions/transcript/spec.md#protocol-with-the-python-program), [Dry-run answer](../north-star/contract.md#dry-run-answer), [Output and cookies](../../extensions/transcript/spec.md#output-and-cookies)
- **Proves:** `extensions/transcript/testdata/script/transcript_protocol.txtar`, `extensions/transcript/testdata/script/transcript_dry_run.txtar`

Embed and extract the program, then exchange one `request` and `plan` through `uv`.
A stub proves Go's protocol handling; real `uv` and Python prove an empty-cache dry run.
Compare user files, config, keychain, and state before and after.
Only documented cache preparation is allowed: no key read, download, paid request, record, or result folder.
Check empty result folders explicitly because `snapshot` ignores folders.

### M13-T3 Transcription with a record

- **Read:** [Records and replay](../../extensions/transcript/spec.md#records-and-replay), [Protocol with the Python program](../../extensions/transcript/spec.md#protocol-with-the-python-program), [Record format](../north-star/contract.md#record-format)
- **Proves:** `extensions/transcript/testdata/script/transcript_run.txtar`, `extensions/transcript/testdata/script/transcript_barrier.txtar`, `extensions/transcript/testdata/script/transcript_keys.txtar`

Record the one validated item before user-side writes or paid dispatch.
Persist its destination and intent before reservation, then complete the persistence barrier before sending `ack`.
A failed barrier sends no acknowledgement and starts no upload.
Prove one run through actual Go, real `uv`, Python, the test transport, and independently verified artifacts.
Inspect its record in a new process with `ncly run view`.
The answer contains one `results` item.
Prove missing keys and grants, and a granted environment key when the keychain is unavailable.

### M13-T4 Failures and replay

- **Read:** [Records and replay](../../extensions/transcript/spec.md#records-and-replay), [Retry safety](../north-star/contract.md#retry-safety), [Output](../north-star/contract.md#output)
- **Proves:** `extensions/transcript/testdata/script/transcript_failures.txtar`, `extensions/transcript/testdata/script/transcript_protocol_errors.txtar`, `extensions/transcript/testdata/script/transcript_replay.txtar`

Stop before paid dispatch and after dispatch but before the response.
A 504 after upload counts exactly one request and leaves transcription unknown with exit 1.
Keep pending failure causes, and handle malformed messages and timeouts as specified.
After a paid intent, a repeated call with the same run ID returns only verified completed work.
Complete recorded artifacts may resolve a lost final message locally, without another upload.
Changed or incomplete evidence is refused; no override bypasses that refusal.
A changed URL or explicit output argument fails before effects.
Prove that invocation and replay refusals omit attempted results and keep their top-level code, while a failed transcription item uses the shared batch verdict.

### M13-T5 Interruption and contenders

- **Read:** [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs), [Records and evidence](../north-star/contract.md#records-and-evidence)
- **Proves:** `extensions/transcript/testdata/script/transcript_cancel.txtar`, `extensions/transcript/testdata/script/transcript_contenders.txtar`

Ctrl-C stops the process tree, preserves completed files, and returns `INTERRUPTED` with the available result paths.
Two callers using one run ID never start two uploads.
Two new runs choosing the same output parent never share a result folder.
Lock failures use the overall retry verdict required by the effects already made.
