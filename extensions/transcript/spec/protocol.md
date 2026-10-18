# Protocol with the Python program

The Go wrapper runs the embedded program through `uv`, as [Output and cookies](output-and-cookies.md) defines.
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
   Dependency preparation follows the cache exception of [Output and cookies](output-and-cookies.md).
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
[Programs that ncly runs](../../../docs/north-star/contract/programs-that-ncly-runs.md).
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
