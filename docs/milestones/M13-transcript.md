# M13 Transcript

Status: planned
Version: v0.13.0

## Demo

`ncly transcript run youtube --no-summary` saves a transcript through the real Python program and a test transport, after a dry run that changed nothing.

## Scope

`transcript` is a first-party external extension in `extensions/transcript/`, built as `ncly-transcript`, as [extensions/transcript/spec.md](../../extensions/transcript/spec.md) defines. It needs the `deepgram` key, which a human grants to it, and it ships through the official tap of [M12](M12-taps.md), like `agent`. This milestone brings the transcribe step. The summary arrives in [M14](M14-transcript-summary.md), resume in [M15](M15-resume.md), and Zoom, prompts, and the spinner in [M16](M16-transcript-finish.md).

Copy the Python transcript CLI from `authoring/andy/transcript/` in `pascalandy/skills` into `extensions/transcript/python/`, with its prompts, tests, and README. From then on, fixes land in Nicely first. Embed it with `go:embed`. On first run, extract it to `~/.cache/nicely/python/<version>/` through a temporary folder and a rename, then run it with `uv`.

Scenarios use the test keychain and stubs for failure injection and counted calls. One path runs the actual Go, the real `uv`, and the actual Python with a fake Deepgram transport that test code supplies. No scenario starts a real paid service, and production gains no endpoint override or fixture CLI. Stubs exercise failures, while the real Python program must separately prove that it honors the protocol.

## Open questions

Settle each one in [extensions/transcript/spec.md](../../extensions/transcript/spec.md), then delete this section.

- How `just check` runs ruff, pyright, and pytest on `extensions/transcript/python/` through `uv`
- How the test transport reaches the real Python program in a scenario without a production endpoint override

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M13-T1 | Python program | agent | — | todo |
| M13-T2 | Protocol and dry run | agent | M13-T1 | todo |
| M13-T3 | Transcription with a record | agent | M13-T2 | todo |
| M13-T4 | Failures | agent | M13-T3 | todo |
| M13-T5 | Interruption and contenders | agent | M13-T3 | todo |

### M13-T1 Python program

- **Read:** [Protocol with the Python program](../../extensions/transcript/spec.md#protocol-with-the-python-program)
- **Proves:** `extensions/transcript/python/`, whose tests run in `just check`

Copy the program, and make `just check` run ruff, pyright, and pytest on it through `uv`. Python's tests save a transcript through a fake Deepgram transport supplied in test code only.

### M13-T2 Protocol and dry run

- **Read:** [Protocol with the Python program](../../extensions/transcript/spec.md#protocol-with-the-python-program), [Dry-run answer](../north-star/cli-spec.md#dry-run-answer), [Output, prompts, and cookies](../../extensions/transcript/spec.md#output-prompts-and-cookies)
- **Proves:** `testdata/script/transcript_protocol.txtar`, `testdata/script/transcript_dry_run.txtar`

Embed and extract the program, run it with the `uv` flags of the spec, and exchange `request` and `plan`. A stub `uv` proves that Go handles the protocol. The real program and the real `uv` prove a dry run from an empty cache. Compare user files, config, keychain, and state before and after the dry run: only documented cache writes, no key read, paid call, harness start, or new user folder. Detect a new empty result folder explicitly, because `snapshot` ignores folders.

### M13-T3 Transcription with a record

- **Read:** [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume), [Protocol with the Python program](../../extensions/transcript/spec.md#protocol-with-the-python-program), [Record format](../north-star/cli-spec.md#record-format)
- **Proves:** `testdata/script/transcript_run.txtar`, `testdata/script/transcript_barrier.txtar`, `testdata/script/transcript_keys.txtar`

Record the validated items before any user-side write or paid dispatch. Go records each destination before reserving it and completes the persistence barrier before sending `ack`; a failed barrier sends no `ack` and starts no upload. Complete one run through the actual Go, the real `uv`, the actual Python, the test transport, verified artifacts, and a record inspected in a new process with `ncly run view`. Answers use `results` for one item and for many. A missing key exits 78 with `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, or `KEY_NOT_GRANTED`, and a granted key in the environment works when the keychain is unavailable.

### M13-T4 Failures

- **Read:** [Steps and resume](../../extensions/transcript/spec.md#steps-and-resume), [Retry safety](../north-star/cli-spec.md#retry-safety), [Output](../north-star/cli-spec.md#output)
- **Proves:** `testdata/script/transcript_failures.txtar`, `testdata/script/transcript_protocol_errors.txtar`

Stop before paid dispatch, and after dispatch but before the response. A failure before the intent keeps its cause as the pending step's code, message, and hint. A 504 after upload counts exactly one request, so transport retries are off, and ends with exit 1 and an unknown transcription. Malformed Python messages and timeouts fail as the spec defines. Item codes are causes; only the overall exit can permit a retry.

### M13-T5 Interruption and contenders

- **Read:** [Programs that ncly runs](../north-star/cli-spec.md#programs-that-ncly-runs), [Records and evidence](../north-star/cli-spec.md#records-and-evidence)
- **Proves:** `testdata/script/transcript_cancel.txtar`, `testdata/script/transcript_contenders.txtar`

Ctrl-C stops the whole tree, keeps the result folder and its finished file, and answers `INTERRUPTED` with `output_dir`. A signal takes precedence over partial batch errors. Two contenders against one destination and one record keep every artifact, and the locks and the overall exit follow Retry safety after earlier effects.
