# transcript extension spec

`transcript` is a first-party external extension, built as `ncly-transcript` from this folder. It transcribes YouTube videos and Zoom recordings with Deepgram, then summarizes the result through the [agent extension](../agent/spec.md), which it calls through the CLI. It embeds a Python program while that program moves to Go. It follows the agent contract of [cli-spec.md](../../docs/north-star/cli-spec.md). [M11](../../docs/milestones/M11-transcript.md) builds transcription, [M12](../../docs/milestones/M12-transcript-summary.md) the summary, [M13](../../docs/milestones/M13-resume.md) resume, and [M14](../../docs/milestones/M14-transcript-finish.md) the rest.

## ncly transcript

Transcribes YouTube videos and Zoom recordings with Deepgram, then summarizes the result through an agent profile.

```
ncly transcript run youtube --url <url> [<url>...] [--prompt <name>] [--profile <name>] [--no-summary] [--output <folder>] [--open] [--timeout <duration>] [--dry-run] [--json]
ncly transcript run zoom (--latest | --path <folder>) [--prompt <name>] [--profile <name>] [--no-summary] [--output <folder>] [--open] [--timeout <duration>] [--dry-run] [--json]
ncly transcript summary run <run-id> [--prompt <name>] [--profile <name>] [--timeout <duration>] [--dry-run] [--json]
ncly transcript prompt list [--json]
```

- `ncly transcript run` needs the `deepgram` key, which its manifest declares and a human grants to transcript. Without it, the command exits 78 with `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, or `KEY_NOT_GRANTED`, as [Keys](../../docs/north-star/cli-spec.md#keys) and [Grants](../../docs/north-star/cli-spec.md#grants) define.
- `--profile` names an [agent profile](../agent/spec.md#profiles) for the summary. Without a profile, the run saves the transcript, skips the summary, and adds `TRANSCRIPT_SUMMARY_SKIPPED`, whose hint shows a profile to add. `--no-summary` skips the summary without a warning. It cannot combine with `--profile` or `--prompt`; these are two optional exclusive flag groups
- `--dry-run` follows the [global definition](../../docs/north-star/cli-spec.md#global-flags): it validates the plan without a key, a paid request, or a change to the user's files, and answers as [Dry-run answer](../../docs/north-star/cli-spec.md#dry-run-answer) defines, with the shapes of [Dry-run examples](#dry-run-examples).
- `--timeout` covers preparation and execution, 570 seconds by default, with at most 15 additional seconds for process cleanup. An agent gives its shell a longer limit than that total, with a margin, or chooses a shorter `ncly` timeout. No universal shell limit or default is assumed. A batch may need a longer timeout
- `--open` opens each result folder as soon as it appears, through `open` on macOS and `xdg-open` on Linux. Dry run never opens it
- Exit 75 follows Retry safety, including any local writes. Never rerun an exit 1 automatically, because work may be partial or already billed
- The extension's Go code owns the summary step and runs it through `ncly agent run --json`, with the transcript as material and the recorded profile as `--harness`, `--model`, `--effort`, and `--provider`, so tools stay off, a resume keeps the recorded profile, and the [agent extension](../agent/spec.md) owns profiles, adapters, and the depth limit. Python stops at the transcript and returns its verified artifacts
- A recorded run returns `run_id`. Its normal execution answer always contains `results`, even for one item, as [Output](../../docs/north-star/cli-spec.md#output) describes. `summary run` and resume use that same item shape. This is a declared batch result; the object answers of core remain unchanged
- Each item carries `url` or the Zoom `recording` folder, `output_dir`, `transcript`, and `summary`. The last three are paths or `null` until available. An unsuccessful item adds its `errors`; a saved transcript remains referenced when its summary fails. A URL given twice is `USAGE_INVALID` before any effect, because each item key names one item

Preparation validates all arguments and local inputs before effects. When a summary is selected, it resolves the prompt and runs `ncly agent run --dry-run --json` with the profile, which resolves the profile for the record and checks depth, harness version, and tools-off support before transcription dispatch. No harness starts to check its login; that remains pending. Without the agent extension, a selected summary fails with `PREREQ_MISSING` before any effect, and its hint shows how to install `agent` or how to add `--no-summary`. `--no-summary` resolves no prompt or profile and needs no agent. A preparation refusal creates no record and bills nothing.

An attempted item's error stays in `results`. The overall verdict follows Output across all attempted items. A human-action cause after transcription therefore does not authorize repeating the original paid command. Each recorded failure's recovery hint names inspection or explicit resume; an unknown summary names the explicit new summary operation.

### Output, prompts, and cookies

Each item gets its own result folder inside the output folder. The folder holds `raw_transcript.txt`, `raw_sentences.txt`, `raw_transcript.json`, and any summary. `--output` names the output folder. Without it, the config decides, and the current folder is the default, as with `yt-dlp` and `gh run download`. In every JSON key, `output` names an output folder, and `output_dir` names a result folder.

```toml
[transcript]
browser = "arc"
prompt_paths = ["~/code/prompts"]

[transcript.youtube]
output = "~/Documents/transcripts"
prompt = "follow_along_note"

[transcript.zoom]
recordings = "~/Documents/Zoom"
output = "~/Documents/meetings"
prompt = "short_summary"
```

| Key | Default | Meaning |
|---|---|---|
| `browser` | none | The browser whose YouTube cookies the download uses: `arc`, on macOS only, or a browser that `yt-dlp` reads, such as `chrome`, `chromium`, `brave`, `firefox`, or `safari`. Without it, the download uses anonymous access, which YouTube often refuses |
| `prompt_paths` | none | Folders searched before the bundled prompts. A prompt is `<name>.md`, or `<name>/prompt.md` |
| `youtube.output`, `zoom.output` | the current folder | The output folder of each source |
| `youtube.prompt` | `follow_along_note` | The default prompt for YouTube |
| `zoom.recordings` | `~/Documents/Zoom` | The folder where Zoom saves recordings. `--latest` picks the newest meeting folder, and `--path` takes a folder name or a full path |
| `zoom.prompt` | `short_summary` | The default prompt for Zoom |

Go reserves each new result folder, named after the video title or meeting, or `<name>-2`, `<name>-3`, and so on when taken. Under the output resource lock, it records the candidate path before creating it and rechecks any collision. Creating the directory reserves its name; two new transcriptions never share a folder or write into an existing one. A failure after reservation exits 1 for the new transcription, even if no audio was sent, because repeating the command creates another folder. Explicit resume reuses the recorded folder instead. Only effects of the current invocation decide retry safety, not paid work reused from an earlier attempt.

Repeating `transcript run` transcribes and bills Deepgram again. Resume and `summary run` reuse the recorded folder. Every summary uses `<prompt-name>-<run-id>.md`, including the initial summary, so a new operation preserves every earlier record's artifact fingerprint.

`ncly transcript prompt list --json` returns `prompts`, sorted by name, each with `name`, `path`, and `source`, which is `config` for a folder of `prompt_paths` and `bundled` otherwise. A name appears once. The first folder of `prompt_paths` that holds it wins, then the bundled prompts. Inside one folder, `<name>.md` wins over `<name>/prompt.md`, because the file comes before the folder.

Machine paths belong in the local config. A Zoom meeting folder starts with its date and time, such as `2026-05-03 14.46.55`, and the rest of its name, which Zoom writes in the user's language, becomes the title.

When the configured browser's cookies cannot be read, the download falls back to anonymous access and the answer adds `TRANSCRIPT_BROWSER_COOKIES_SKIPPED`. In a batch, the warning appears once in the top-level `warnings`, because the browser setting covers the whole run. `ncly doctor transcript` reports the same problem as `warn`. The Arc adapter checks the `yt-dlp` function that it changes, instead of an exact `yt-dlp` version. When that function is missing, the adapter reports the problem and the download falls back the same way.

The extension runs the program with `uv run --script --upgrade-package yt-dlp --no-python-downloads --cache-dir <folder>`, where the folder is `uv` inside Nicely's cache, such as `~/.cache/nicely/uv`. The flag wins over an inherited `UV_CACHE_DIR`, so uv writes only inside Nicely's cache, as a dry run requires. uv never downloads a Python. A missing Python 3.12 or later is `PREREQ_MISSING`, and `ncly doctor transcript` offers its install command. Each Python invocation uses the newest `yt-dlp` release at or above the minimum that the program names, so a fix reaches users without an `ncly` release. The program pins its other dependencies. A dry run may contact the package index to prepare that cache; it is not an offline guarantee. Failed dependency preparation returns `RUNTIME` with a setup hint; the command's own timeout follows Programs. Doctor reads installed versions without upgrading packages.

These options of the transcript CLI change:

| Transcript CLI | ncly |
|---|---|
| `--output-dir` | `--output` |
| `--debug`, `TRANSCRIPT_DEBUG` | `NCLY_DEBUG=1` |
| `--provider`, `--model`, `--effort` | A profile in the config |
| `--no-progress` | Removed, because `--json` and non-interactive mode already hide the spinner |
| `--preview` | Removed, because `ncly markdown view` renders a summary from M18 |
| `list prompts` | `ncly transcript prompt list` |
| `list profiles` | `ncly agent profile list` |
| `list models` | Removed. M16 decides whether `ncly agent model list` replaces it |
| `doctor --source` | `ncly doctor transcript` |

### Steps and resume

Each transcription item has `transcribe`, then `summarize` unless skipped. Python prepares the complete item set before emitting its initial plan. Go validates that set and creates the initial `running` record with pending steps before allowing any user-side mutation. Temporary audio preparation may overlap that validation after Python sends the plan. `started_at` is the invocation's start time and stays unchanged on resume. A step reaches `completed` only after Go independently verifies its required files, syncs them, and durably records their fingerprints.

- `transcribe` turns `unknown` before result-folder reservation and upload, as the protocol below requires. A download failure before intent leaves it `pending` with a saved problem. After acknowledgement, it turns `failed` only when transport failed before any audio was sent or Deepgram returned a 4xx refusal. Any 5xx status and any transport timeout or lost connection after sending audio leave it `unknown`. The transport disables automatic upload retries, including SDK retries; one acknowledgement permits at most one request
- `summarize` turns `unknown` before `ncly agent run` starts. It turns `failed` only when that run exits 2 or 78, because the agent extension exits with those codes only before its harness starts. Any other end without a saved summary leaves the step `unknown`, because the harness may already have billed. The step's evidence keeps the agent run's `run_id`
- The record keeps the `inputs` that the table below lists. It keeps as tools the program and `yt-dlp` versions that the `intent` reports, and the harness version that the agent run reports
- Resume takes the URL, selected Zoom folder, and absolute output folder from the record. It never selects `--latest` again. It rereads the prompt and refuses a changed fingerprint. For Zoom, it verifies the recorded audio and compares the program's `audio` in `plan` and `intent` before acknowledgement. Resume uses the recorded profile without resolving it again; `summary run --profile` changes it. Only programs needed by `run` steps require a current version check, so a summary-only resume needs no `uv` or Deepgram key. A version above its minimum does not block resume; below it returns `PREREQ_MISSING`. Unsupported record or protocol versions return `RESUME_UNSAFE`
- Resume verifies the recorded inputs and artifacts, reuses completed steps, and runs pending or failed ones. For an unknown transcription only, the complete recorded artifact set of `raw_transcript.txt`, `raw_sentences.txt`, and `raw_transcript.json` may establish local completion after a lost final message. Go independently verifies every path, size, hash, and usable transcript format. Dry run plans reuse without rewriting the record; explicit resume persists the reconciled completion without another upload
- An unanswered transcription without that complete evidence, any unknown summary, a changed input or artifact, or an unexpected file at a path that resume would write returns `RESUME_UNSAFE`, retaining the available artifacts. `--force` overrides none of these refusals. The existence of an unrecorded file alone never proves paid completion
- An `unknown` `summarize` step names `ncly transcript summary run <run-id>` in its hint. An `unknown` `transcribe` step names `ncly run view <run-id>`, because a new transcription may bill Deepgram again

The transcript domain defines these `inputs` keys:

| Key | Type | Meaning |
|---|---|---|
| `source` | string | `youtube` or `zoom` |
| `output` | string | The absolute output folder |
| `audio` | object | `transcript run zoom` only. The `path`, `size_bytes`, and `sha256` of the first `*.m4a` of the meeting folder in name order, selected once by Python. A folder without one is `NOT_FOUND` |
| `prompt` | object or null | The prompt `name` and the `sha256` of its text, or `null` when the summary is skipped |
| `profile` | object or null | The resolved profile: `name`, or `null` when it has none, `harness`, `provider`, a key omitted when there is none, `model`, and `effort`. `null` when the summary is skipped |
| `source_run` | string | `summary run` only. The run ID of the source run |

A Zoom run records this:

```json
{
  "format_version": 1,
  "run_id": "20261007-214501-3f9a2c",
  "path": "transcript run zoom",
  "status": "completed",
  "started_at": "2026-10-07T21:45:01Z",
  "updated_at": "2026-10-07T21:49:12Z",
  "ncly_version": "v0.1.0",
  "attempts": 1,
  "inputs": {
    "source": "zoom",
    "output": "/Users/me/meetings",
    "audio": {"path": "/Users/me/Documents/Zoom/2026-05-03 14.46.55 Weekly sync/audio1.m4a", "size_bytes": 48213504, "sha256": "<sha256>"},
    "prompt": {"name": "short_summary", "sha256": "<sha256>"},
    "profile": {"name": "everyday", "harness": "claude", "model": "<model id>", "effort": "medium"}
  },
  "tools": {"transcript": "4.1.0", "yt-dlp": "2026.7.4", "claude": "2.1.291"},
  "items": [
    {
      "key": "/Users/me/Documents/Zoom/2026-05-03 14.46.55 Weekly sync",
      "output_dir": "/Users/me/meetings/Weekly_sync",
      "steps": [
        {"id": "transcribe", "status": "completed", "started_at": "2026-10-07T21:45:20Z", "finished_at": "2026-10-07T21:47:58Z", "artifacts": [
          {"path": "/Users/me/meetings/Weekly_sync/raw_transcript.txt", "size_bytes": 51234, "sha256": "<sha256>"},
          {"path": "/Users/me/meetings/Weekly_sync/raw_sentences.txt", "size_bytes": 52011, "sha256": "<sha256>"},
          {"path": "/Users/me/meetings/Weekly_sync/raw_transcript.json", "size_bytes": 403877, "sha256": "<sha256>"}
        ], "error": null},
        {"id": "summarize", "status": "completed", "started_at": "2026-10-07T21:47:59Z", "finished_at": "2026-10-07T21:49:12Z", "artifacts": [{"path": "/Users/me/meetings/Weekly_sync/short_summary-20261007-214501-3f9a2c.md", "size_bytes": 4096, "sha256": "<sha256>"}], "error": null}
      ]
    }
  ]
}
```

### ncly transcript summary run

`ncly transcript summary run <run-id>` summarizes saved transcripts as a new recorded operation. It resolves its profile first: `--profile`, then the source run's recorded profile, then the current `default_profile`. If none applies, a form asks in interactive mode; otherwise it exits 2 with `USAGE_INVALID` and a hint showing `--profile`. Until a profile is known it reads no item, writes no record, and bills nothing.

Before reading items or starting a summary, execution takes the source run lock and holds it until the new operation ends. A busy source returns 75 with `TEMPORARY` and a `run view` hint, before effects. Dry run only checks ownership and never holds or reserves the source for execution.

Its prompt name comes from `--prompt`, the recorded source prompt, then the source's current default, such as `youtube.prompt`. It reads the current prompt text: a missing prompt is `NOT_FOUND` before effects; a changed hash is allowed and stored in the new record. This is a new operation, while resume preserves the recorded prompt fingerprint.

An item is eligible when its transcription is completed or locally reconciles from the complete recorded artifact set defined in Steps and resume. Any other item fails in `results` with `RESUME_UNSAFE`. If none is eligible, the failed answer includes `source_run` and `results` but no `run_id`; no record or harness starts.

A mixed operation records every source item in its original order. Eligible items copy verified transcript references and original timestamps into a completed `transcribe` step, then add a pending `summarize` step. An ineligible item records only a failed `summarize` step with its `RESUME_UNSAFE` problem and its source `output_dir`, possibly `null`. That item never becomes executable in this operation. Resume reproduces its saved error while continuing eligible pending or failed summaries; including the rejected item later requires a new operation. This is transcript-domain handling, without a new shared status.

The new record keeps `source_run` for provenance and its own resolved prompt and profile. It copies no transcript content or source audio dependency.

When a new record is created, normal execution returns `run_id`, `source_run`, and `results` with the transcript item shape. Each summary goes to the recorded result folder as `<prompt-name>-<new-run-id>.md`, preserving source artifacts. `run resume` supports this record: it verifies copied transcript evidence and continues eligible pending or failed summaries without transcribing again. Unknown summaries still refuse resume and hint `ncly transcript summary run <source_run>`. This new operation may bill and starts only on an explicit call.

### Protocol with the Python program

The extension extracts the program to `~/.cache/nicely/python/<version>/` and runs it through `uv`, with the flags that [Output, prompts, and cookies](#output-prompts-and-cookies) lists. The program reads JSON Lines on stdin and writes JSON Lines on stdout. Its stderr carries diagnostics, which Go shows only under `--verbose` or `NCLY_DEBUG`. Here, Go is the extension's Go code. Each message is one JSON object with `protocol_version`, `1`, and `type`. A message with another version or an unknown type is a protocol failure.

| Message | Sender | Keys besides `protocol_version` and `type` |
|---|---|---|
| `request` | Go | `source`: `youtube` or `zoom`; `urls`: ordered URLs, empty for Zoom; `zoom`: `latest`, `path`, and `recordings`, or `null`; `output`: absolute output parent; `temp_dir`: Go-owned temporary folder; `browser` or `null`; `dry_run`: boolean; `deadline`: UTC RFC 3339 execution deadline |
| `plan` | program | `items`: the complete resolved item set, in input order. Each has `item`, `output`, and, for Zoom, selected `audio` with `path`, `size_bytes`, and `sha256`. On preparation failure, `items` is empty and `code` and `detail` identify the program failure |
| `intent` | program | `item`, `step`, `name`: proposed result-folder basename, `tools`: actual program and `yt-dlp` versions, and, for Zoom, the unchanged `audio` from `plan`. No user-side output has been created |
| `ack` | Go | `item` and `step` from the intent, plus `output_dir`: the final reserved or recorded result folder |
| `artifact` | program | `item`, `step`, `path`, `size_bytes`, and `sha256` |
| `step` | program | `item`, `step`, `status`, `code`, which is `null` for `completed`, `warnings`, a list of program warning codes, and `detail`, a diagnostic that Go shows only under `--verbose` |

`item` is the URL as given or the absolute Zoom meeting folder selected once in `plan`. `step` is `transcribe`. The key travels in `DEEPGRAM_API_KEY`, never in a message. Python puts all temporary audio and child scratch files under `temp_dir`, so Go can remove them after killing a child. It derives phase timeouts from the remaining `deadline` and starts no upload once that deadline has expired. Go owns the overall timeout verdict and process cleanup.

1. Go sends one `request`. Python returns exactly one initial `plan` in both modes, before downloads, output reservation, or paid calls. A preparation failure reports its code and exits 0 without effects or final item steps
2. Go validates the complete expected item set, order, uniqueness, selected Zoom audio, and output parent. A new Zoom selection is fixed for this execution. Resume passes the recorded folder with `latest: false`, and verifies the stored audio. Dry run creates no record or user directory and starts no download or service call. Python exits after its plan; dependency cache preparation follows the documented exception
3. For execution, Go durably creates the initial record from the validated plan, with pending steps. After emitting its plan, Python may prepare audio only in `temp_dir` while Go validates and records that plan. This free temporary work needs no second handshake. Before any user-side write or upload, Python sends `intent` and waits for `ack`
4. Go validates the intent against the fixed plan. For a new result folder, it selects and durably records the candidate `output_dir` and unknown step before reserving the directory under the resource lock. For resume, it revalidates and uses the recorded folder. It completes the persistence barrier before sending `ack` with that final folder. Failure prevents acknowledgement. Python stops before writing outputs or uploading when stdin closes or no matching acknowledgement arrives within 10 seconds
5. After saving files only under the acknowledged folder, Python sends one `artifact` per required file, then `step` with status `completed`. Artifact paths are absolute within that folder, sizes are nonnegative byte integers, and hashes are 64 lowercase hexadecimal characters. Go independently verifies those files and the complete required set, syncs them, and durably records completion. Each artifact's verified evidence is persisted when received, so a lost final step can reconcile only a complete recorded set
6. A failed item sends `step` with its status and program code: `pending` before intent, otherwise `failed` or `unknown` as the table defines. Go keeps its public code, message, and recovery hint on the record, including pending preparation failures. Exactly one terminal step is required for each planned item. Python exits 0 whatever their outcomes; Go derives the verdict. On Ctrl-C or SIGTERM, Python stops its children within 10 seconds and reports open steps when possible

A nonzero exit, missing initial plan, unexpected or duplicate item, illegal message order, missing final step, or malformed message is a protocol failure. Go keeps verified evidence, leaves acknowledged unresolved effects unknown, and exits 1 with `RUNTIME`. Untouched pending steps stay pending with their failure cause. Cancellation and timeout retain the verdict from [Programs that ncly runs](../../docs/north-star/cli-spec.md#programs-that-ncly-runs). A dry-run plan or zero exit alone never proves that execution completed an item.

| Program code | Step status | Error code | Catalog ID |
|---|---|---|---|
| `recording_not_found` | `pending` | `NOT_FOUND` | `transcript.recording_not_found` |
| `download_temporary` | `pending` | `TEMPORARY` | `transcript.download_temporary` |
| `download_failed` | `pending` | `TRANSCRIPT_DOWNLOAD_FAILED` | `transcript.download_failed` |
| `deepgram_unreachable` | `failed` | `TEMPORARY` | `transcript.deepgram_unreachable` |
| `deepgram_key_rejected` | `failed` | `AUTH_REJECTED` | `transcript.deepgram_key_rejected` |
| `deepgram_failed` | `failed` | `RUNTIME` | `transcript.deepgram_failed` |
| `deepgram_unknown` | `unknown` | `RUNTIME` | `transcript.deepgram_unknown` |
| `transcript_unusable` | `unknown` | `RUNTIME` | `transcript.transcript_unusable` |
| `write_failed` | `pending` before the intent, `unknown` after it | `RUNTIME` | `transcript.write_failed` |

The only program warning code is `browser_cookies_skipped`, which Go maps to `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` with the catalog ID `transcript.browser_cookies_skipped`.

`deepgram_unreachable` means transport failed before audio was sent, such as a refused connection, DNS or TLS failure, or a received HTTP refusal with status 408 or 429. A received HTTP 408 response is distinct from a transport timeout without a response. `deepgram_key_rejected` means 401 or 403; `deepgram_failed` means any other 4xx refusal. Any 5xx status, including 503, or a timeout or lost connection after sending audio is `deepgram_unknown`.

The item's `TEMPORARY` describes its cause. Only the overall exit 75 permits repeating the whole invocation. After a reserved result folder, started paid request, or unknown effect, the normal answer retains that item cause in `results` and carries a top-level `RUNTIME` with exit 1. Inspect the saved run before choosing resume or a new operation.

## Dry-run examples

A new transcription reports its output parent without reserving a result folder:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output":"/Users/me/transcripts","steps":[{"id":"transcribe","action":"run"}],"effects":{"user_writes":true,"network":true,"paid":true}}],"pending_checks":[{"id":"auth.deepgram","message":"Authentication is checked only during execution."}]}
```

A summary alone, from `ncly transcript summary run <run-id> --dry-run --json`, reuses the transcript and waits for the harness to start:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output_dir":"/Users/me/transcripts/Video_title","steps":[{"id":"transcribe","action":"reuse"},{"id":"summarize","action":"run"}],"effects":{"user_writes":true,"network":true,"paid":true}}],"pending_checks":[{"id":"agent.claude","message":"The harness login is checked when the summary starts."}]}
```

A resume that reuses every step, from `ncly run resume <run-id> --dry-run --json`, has nothing pending:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output_dir":"/Users/me/transcripts/Video_title","steps":[{"id":"transcribe","action":"reuse"},{"id":"summarize","action":"reuse"}],"effects":{"user_writes":false,"network":false,"paid":false}}],"pending_checks":[]}
```

## Codes

Transcript answers with core's codes and these of its own. For exit 1, `RESUME_UNSAFE` leads, then `TRANSCRIPT_DOWNLOAD_FAILED`, then `RUNTIME`.

| Code | Exit | When | Since |
|---|---|---|---|
| `TRANSCRIPT_DOWNLOAD_FAILED` | 1 | `yt-dlp` could not fetch the audio of a video, for a reason other than the network | M11 |

| Warning | When | Since |
|---|---|---|
| `TRANSCRIPT_SUMMARY_SKIPPED` | No profile applies, so a transcript run saved the transcript without a summary | M12 |
| `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` | The configured browser's cookies could not be read, so the download used anonymous access | M14 |

## Examples

```bash
# Plan a transcription without a key, a paid request, or a user-file write
ncly transcript run youtube --url "https://www.youtube.com/watch?v=VIDEO_ID" --dry-run --json

# Keep the full successful answer and run ID before extracting paths
ncly transcript run youtube --url "$URL" --json > answer.json 2> diagnostics.txt
jq -r '.run_id, .results[].output_dir' answer.json

# A missing key, seen by a script
ncly transcript run youtube --url "$URL" --json
# stderr: {"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
# exit code: 78
```
