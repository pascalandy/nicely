# transcript extension spec

`transcript` is a first-party external extension, built as `ncly-transcript` from this folder.
It transcribes one YouTube video with Deepgram through the existing Python program.
It follows the [contract](../../docs/north-star/contract.md).
[M13](../../docs/milestones/M13-transcript.md) builds this command.
[M16](../../docs/milestones/M16-transcript-finish.md) adds its doctor checks, skill, and fresh-session proof.

The program comes from `authoring/andy/transcript/` in `pascalandy/skills`.
Copy its transcription implementation; do not rewrite it in Go for this milestone.
Other source features remain unexposed.
Summary awaits a verified tools-off path under [D028](../../docs/north-star/decisions/D028-outside-text.md).
It is parked in [M99](../../docs/milestones/M99-parking-lot.md#transcript-summary), alongside
[resume](../../docs/milestones/M99-parking-lot.md#resume) and
[transcript extras](../../docs/milestones/M99-parking-lot.md#transcript-extras).

This specification describes planned behavior.
The open questions in M13 and M16 must be resolved before their implementation is ready.

## ncly transcript

Transcribes one explicitly selected YouTube video with Deepgram.

```
ncly transcript run youtube --url <url> [--output <folder>] [--run-id <id>] [--timeout <duration>] [--dry-run] [--json]
```

- `--url` selects one video using the normal scalar-flag rules.
  Use the download program's single-video mode.
  A playlist-only URL is invalid; a watch URL with playlist parameters still selects only that video.
- The manifest declares `deepgram` in the command's `requires.keys`.
  A human must grant the key to transcript.
  Missing credentials or grants use `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, or `KEY_NOT_GRANTED`, as
  [Keys](../../docs/north-star/contract/keys.md) and [Grants](../../docs/north-star/contract/grants.md) define.
- `--dry-run` uses the same local preparation as execution.
  It reads no key and starts no audio download or paid request.
  It creates no user file, result folder, or run record.
  Only the dependency cache preparation described below may write or use the network.
- `--timeout` covers preparation and execution, with a default of 570 seconds and up to 15 additional seconds for cleanup.
  A caller gives its shell enough time for that total or chooses a shorter command timeout.
- The command uses the existing plain-text output style.
  It declares no progress display.
- An attempted transcription returns `run_id` and `results` with one item, including on execution failure.
  The existing `batch` result envelope is retained; its use does not enable multiple inputs.
  The item holds `ok`, `url`, `output_dir`, and `transcript`.
  The last two are paths or `null` until available.
  An unsuccessful item adds `errors`, and verified files remain referenced.
- Invocation, preparation, and replay refusals attempt no item and omit `results`.
  Their top-level code is the actual refusal, such as `USAGE_INVALID`, `KEY_NOT_GRANTED`, or `RESUME_UNSAFE`.
  Include `run_id` when the refusal concerns an existing record; earlier preparation creates none.
  A successful replay returns the one verified item.
  The schema permits absent `results` on a refusal and requires one item on success.
  A failed attempted item instead uses the batch verdict: top-level `RUNTIME`, `TEMPORARY`, or the signal code,
  with the specific cause inside its item. Do not change the shared answer contract.
- Overall exit 75 alone permits an automatic repeat.
  A paid request or result-folder reservation excludes 75 for that invocation.
  Preserve the full answer and inspect its run before choosing another operation.

Preparation validates the selected URL, output parent, configuration, and required programs before user-side effects.
A preparation refusal creates no record and bills nothing.
Execution checks the granted key before download or output reservation.
The command needs no agent extension, profile, or summary prompt.

### Output and cookies

The result folder holds `raw_transcript.txt`, `raw_sentences.txt`, and `raw_transcript.json`.
These are the required transcript artifacts.
M13 must establish any other output files of the copied program before implementation is ready.

`--output` selects the output parent.
Without it, `transcript.youtube.output` selects the parent, or the current folder when no value is configured.
In JSON, `output` names the parent and `output_dir` names the result folder.

```toml
[transcript]
browser = "arc"

[transcript.youtube]
output = "~/Documents/transcripts"
```

Machine-specific paths belong in local config.
The retained cookie integration is the source program's Arc path on macOS.
`browser = "arc"` selects it.
Without a browser setting, use anonymous access.
The supported path and its dependencies need the evidence listed in M13.
Additional browser integrations are not required for this release.

Unreadable configured cookies fall back to anonymous access with `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` in the answer's top-level `warnings`.
Anonymous access can still fail to download the video; the warning never means transcription succeeded.
The Arc adapter checks the `yt-dlp` function it adapts.
A missing function produces the same fallback warning.
M16 makes `ncly doctor transcript` report the cookie problem as `warn`.

Go selects a result-folder name from the video title, adding `-2`, `-3`, and so on when taken.
Under the output resource lock, it records the candidate path and intent, then creates the folder while rechecking collisions.
Two new runs never share a result folder or write into an existing one.
After an interrupted reservation, another invocation does not write into that folder.
Verified-result replay only reads its existing artifacts.
A new run selects a new folder and may bill Deepgram again.

Extract the embedded Python program through a temporary folder and atomic rename into `~/.cache/nicely/python/<version>/`, following XDG overrides.
Run it with `uv run --script --upgrade-package yt-dlp --no-python-downloads --cache-dir <folder>`.
Here, `<folder>` is `uv` inside Nicely's cache, such as `~/.cache/nicely/uv`.
The explicit flag overrides an inherited `UV_CACHE_DIR`.
Python 3.12 or later must already be installed; `uv` never downloads it.

The program uses the newest `yt-dlp` release at or above its declared minimum and pins its other Python dependencies.
Declare every actual download dependency in the manifest; M13 must prove that set from an empty cache.
Do not assume that `uv` and Python alone suffice.
A dry run may contact the package index to prepare Nicely's cache.
Dependency preparation failure returns `RUNTIME` with a setup hint.
Doctor probes installed tools without upgrading them.
The source CLI's `--output-dir` becomes `--output`; debug output uses `NCLY_DEBUG`.

### Records and replay

Each run has one item and one `transcribe` step, which declares the `paid` effect.
The manifest supports recorded execution and dry run; domain resume remains unimplemented.
Use the SDK for records, locks, and publication, as [Operations](../../docs/north-star/contract/operations.md) defines.

Python emits the single-item plan before user-side effects.
Go validates it and durably creates the initial record with a pending step.
The step becomes unknown before result-folder reservation or upload.
Go records completion only after independently verifying and syncing the required artifacts.

- A download failure before intent leaves the step pending with its code, message, and hint.
- After acknowledgement, a known pre-upload transport failure or a received 4xx refusal makes the step failed.
  A 5xx response or lost connection after sending audio leaves it unknown.
  Upload retries, including SDK retries, are disabled.
  One acknowledgement permits at most one upload request.
- The record keeps the program and `yt-dlp` versions reported by the intent.
  Its item key is the URL exactly as supplied.
- A repeated call uses the run lock and input checks of [Run IDs](../../docs/north-star/contract/run-ids.md).
  A different command, URL, or explicit output argument is `USAGE_INVALID` before effects.
  A busy run returns `TEMPORARY` with a `run view` hint and starts no program.
- A completed run returns its verified saved result without Python, a dependency upgrade, or a Deepgram request.
  Check every required path, size, hash, and usable transcript format before returning success.
  Missing, changed, or unsupported evidence returns `RESUME_UNSAFE`.
- Complete, already recorded artifact evidence may resolve a lost final protocol message locally.
  Under the run lock, Go verifies that complete set and persists completion before returning success.
  This never writes transcript artifacts or starts another upload.
  Dry run reports reuse without changing the record.
  Unrecorded files alone prove nothing.
- Before any intent, the shared run-ID rules may safely retry preparation.
  After a paid intent, incomplete or uncertain work is refused with `RESUME_UNSAFE`.
  A missing response never proves the request was free.
  No flag authorizes replay of an uncertain upload.
- A refusal preserves available artifacts and names `ncly run view <run-id>` in its hint and `next`.
  Starting a new run is an explicit choice that may incur another charge.

The domain records these `inputs`, as [Records and evidence](../../docs/north-star/contract/records-and-evidence.md) defines:

| Key | Type | Meaning |
|---|---|---|
| `source` | string | Always `youtube` |
| `output` | string or null | The absolute parent explicitly supplied through `--output`, or `null` when configuration selects it |

The item key records the one URL.
The domain records these `resolved` values, without using them to compare a repeated call:

| Key | Type | Meaning |
|---|---|---|
| `output` | string | The absolute output parent selected during preparation |
| `browser` | string or null | `arc` when selected, otherwise `null` |

A configuration change never redirects a saved result.
The recorded output folder and verified artifacts determine replay.
A saved run keeps its original `started_at`; local completion reconciliation starts no new execution attempt.

### Protocol with the Python program

The Go wrapper runs the embedded program through `uv`, as [Output and cookies](#output-and-cookies) defines.
The program reads JSON Lines on stdin and writes JSON Lines on stdout.
Its stderr carries diagnostics, shown only under `--verbose` or `NCLY_DEBUG`.
Each message has `protocol_version: 1` and `type`.
An unsupported version or unknown type is a protocol failure.

The wrapper retains the existing message shapes with one selected item.
The unused `zoom` request field stays `null`; it exposes no Zoom command.

| Message | Sender | Keys besides `protocol_version` and `type` |
|---|---|---|
| `request` | Go | `source`: `youtube`; `urls`: an array containing the one URL; `zoom`: `null`; `output`: absolute output parent; `temp_dir`: Go-owned temporary folder; `browser`: `arc` or `null`; `dry_run`: boolean; `deadline`: UTC RFC 3339 execution deadline |
| `plan` | program | `items`: one object with `item` and `output`. On preparation failure, `items` is empty, and `code` and `detail` identify the failure |
| `intent` | program | `item`, `step`, `name`: proposed result-folder basename, `tools`: actual program and `yt-dlp` versions. No user-side output has been created |
| `ack` | Go | `item` and `step` from the intent, plus `output_dir`: the final reserved result folder |
| `artifact` | program | `item`, `step`, `path`, `size_bytes`, and `sha256` |
| `step` | program | `item`, `step`, `status`, `code`, which is `null` for `completed`, `warnings`: program warning codes, and `detail`: a verbose diagnostic |

`item` is the URL as supplied, and `step` is `transcribe`.
The key travels in `DEEPGRAM_API_KEY`, never in a message.
Python puts temporary audio and child scratch files under `temp_dir`.
It derives phase timeouts from the remaining `deadline` and starts no upload after that deadline.
Go owns the overall timeout verdict and process cleanup.

1. Go sends one request.
   Python returns one initial plan before downloads, output reservation, or paid calls.
   Preparation failure reports its code and exits 0 without effects or a terminal item step.
2. Go verifies the one expected item and its output parent.
   Dry run creates no record or user directory and starts no download or service call.
   Python exits after its plan.
   Dependency preparation follows the cache exception above.
3. For execution, Go durably records the validated plan with its pending step.
   After sending the plan, Python may prepare audio only in `temp_dir`.
   Before user-side writes or upload, Python sends `intent` and waits for `ack`.
4. Go validates the intent against the fixed plan.
   Under the output resource lock, it durably records the candidate `output_dir` and unknown step before reserving the folder.
   Complete the persistence barrier before sending `ack` with the final folder.
   A failed barrier prevents acknowledgement and upload.
   Python stops before effects if stdin closes or a matching acknowledgement does not arrive within 10 seconds.
5. Python saves files only inside the acknowledged folder.
   It sends one artifact message per required file, then a completed step.
   Paths are absolute within that folder; sizes are nonnegative byte integers; hashes are 64 lowercase hexadecimal characters.
   Go independently verifies each file and persists its evidence as received.
   Go verifies the complete required set and syncs it before durably recording completion.
6. A failure sends the step's status and program code.
   Go preserves its public code, message, and recovery hint, including for pending preparation failures.
   Exactly one terminal step message is required after a successful plan.
   Python exits 0 for a reported item outcome; Go derives the command verdict.
   On Ctrl-C or SIGTERM, Python stops its children within 10 seconds and reports the open step when possible.

A nonzero exit, missing plan, wrong item, illegal ordering, duplicate message, missing terminal step, or malformed message is a protocol failure.
Go preserves verified evidence, leaves acknowledged unresolved effects unknown, and exits 1 with `RUNTIME`.
An untouched step stays pending with its failure cause.
Cancellation and timeout keep the verdict required by
[Programs that ncly runs](../../docs/north-star/contract/programs-that-ncly-runs.md).
A dry-run plan or zero exit alone never proves execution succeeded.

| Program code | Step status | Error code | Catalog ID |
|---|---|---|---|
| `download_temporary` | `pending` | `TEMPORARY` | `transcript.download_temporary` |
| `download_failed` | `pending` | `TRANSCRIPT_DOWNLOAD_FAILED` | `transcript.download_failed` |
| `deepgram_unreachable` | `failed` | `TEMPORARY` | `transcript.deepgram_unreachable` |
| `deepgram_key_rejected` | `failed` | `AUTH_REJECTED` | `transcript.deepgram_key_rejected` |
| `deepgram_failed` | `failed` | `RUNTIME` | `transcript.deepgram_failed` |
| `deepgram_unknown` | `unknown` | `RUNTIME` | `transcript.deepgram_unknown` |
| `transcript_unusable` | `unknown` | `RUNTIME` | `transcript.transcript_unusable` |
| `write_failed` | `pending` before intent, `unknown` after it | `RUNTIME` | `transcript.write_failed` |

`browser_cookies_skipped` is the only program warning.
Go maps it to `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` with catalog ID `transcript.browser_cookies_skipped`.

`deepgram_unreachable` covers transport failure before audio was sent, such as refused connections, DNS or TLS failures, and received HTTP 408 or 429 refusals.
A received HTTP 408 differs from a transport timeout without a response.
`deepgram_key_rejected` means 401 or 403; other 4xx refusals use `deepgram_failed`.
Any 5xx status, timeout, or lost connection after sending audio uses `deepgram_unknown`.

The item's `TEMPORARY` describes its cause.
After result-folder reservation, paid dispatch, or an unknown effect, the overall runtime failure exits 1 with `RUNTIME`.
The item retains its cause in `results`.
Only the overall exit 75 permits repeating the invocation.

## Dry-run examples

A new transcription reports its output parent without reserving a result folder:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output":"/Users/me/transcripts","steps":[{"id":"transcribe","action":"run"}],"effects":{"user_writes":true,"network":true,"paid":true}}],"pending_checks":[{"id":"auth.deepgram","message":"Authentication is checked only during execution."}]}
```

The same call with an existing completed run ID verifies its saved artifacts and plans reuse only:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output_dir":"/Users/me/transcripts/Video_title","steps":[{"id":"transcribe","action":"reuse"}],"effects":{"user_writes":false,"network":false,"paid":false}}],"pending_checks":[]}
```

A refusal during replay fails its dry run with the same `RESUME_UNSAFE`, without changing the record.

## Codes

Transcript uses core's codes and the following domain code.
`RESUME_UNSAFE` is a top-level replay refusal with no attempted item. `TRANSCRIPT_DOWNLOAD_FAILED` is an item cause; a failed attempted transcription uses the shared batch verdict, never that cause as its top-level lead.

| Code | Exit | When | Since |
|---|---|---|---|
| `TRANSCRIPT_DOWNLOAD_FAILED` | 1 | `yt-dlp` could not fetch the video's audio for a reason other than a temporary network failure | M13 |

| Warning | When | Since |
|---|---|---|
| `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` | The configured cookies could not be read, so download attempted anonymous access | M13 |

## Examples

```bash
# Plan one video without a key, paid request, or user-file write
ncly transcript run youtube --url "https://www.youtube.com/watch?v=VIDEO_ID" --dry-run --json

# Choose the ID before execution and preserve the complete answer
ncly transcript run youtube --url "$URL" --run-id video-notes-1 --json > answer.json 2> diagnostics.txt
jq -r '.run_id, .results[].output_dir' answer.json

# Inspect after a lost answer or an uncertain outcome
ncly run view video-notes-1 --json

# The same call returns verified completion or refuses uncertain work
ncly transcript run youtube --url "$URL" --run-id video-notes-1 --json
```
