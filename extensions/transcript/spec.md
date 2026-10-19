# transcript extension spec

`transcript` is a first-party external extension, built as `ncly-transcript` from this folder.
It transcribes one YouTube video with Deepgram through the existing Python program.
It follows the [contract](../../docs/north-star/contract.md).
[M13](../../docs/milestones/M13-transcript.md) builds this command.
[M16](../../docs/milestones/M16-transcript-finish.md) adds its doctor checks and fresh-session proof.

The program comes from `authoring/andy/transcript/` in `pascalandy/skills` at `0c9c4a2`, as [The engine](spec/protocol.md#the-engine) says.
Copy only its YouTube transcription path; do not rewrite it in Go for this milestone.
Other source features remain unexposed.
Summary awaits a verified tools-off path under [D028](../../docs/north-star/decisions/D028-outside-text.md).
It is parked in [M99](../../docs/milestones/M99-parking-lot/transcript-summary.md), alongside
[resume](../../docs/milestones/M99-parking-lot/resume.md) and
[transcript extras](../../docs/milestones/M99-parking-lot/transcript-extras.md).

This specification describes planned behavior.
The readiness of M13 settled its open questions in these files.
The open questions of M16 must be resolved before its implementation is ready.

## ncly transcript

Transcribes one explicitly selected YouTube video with Deepgram.

```
ncly transcript run youtube --url <url> [--output <folder>] [--run-id <id>] [--timeout <duration>] [--dry-run] [--json]
```

- `--url` selects one video using the normal scalar-flag rules.
  Use the download program's single-video mode.
  A playlist-only URL is invalid; a watch URL with playlist parameters still selects only that video.
  Go checks the URL before it starts `uv`, with the three forms of the source program.
  They are `youtube.com/watch?v=<id>`, `youtu.be/<id>`, and `youtube.com/shorts/<id>`, over HTTP or HTTPS.
  `youtube.com` may start with `www.`. Any other URL is `USAGE_INVALID`, and the engine trusts the URL of its request.
- The manifest declares `deepgram` in the command's `requires.keys`.
  A human must grant the key to transcript.
  Missing credentials or grants use `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, or `KEY_NOT_GRANTED`, as
  [Keys](../../docs/north-star/contract/keys.md) and [Grants](../../docs/north-star/contract/grants.md) define.
- `--dry-run` uses the same local preparation as execution.
  It reads no key and starts no audio download or paid request.
  It creates no user file, result folder, or run record.
  Only the dependency cache preparation of [Output and cookies](spec/output-and-cookies.md) may write or use the network.
- `--timeout` covers preparation and execution, with a default of 570 seconds and up to 15 additional seconds for cleanup.
  A caller gives its shell enough time for that total or chooses a shorter command timeout.
- The command uses the existing plain-text output style.
  It declares no progress display.
- An attempted transcription returns `run_id` and `results` with one item, including on execution failure.
  The existing `batch` result envelope is retained; its use does not enable multiple inputs.
  The item holds `ok`, `url`, `output_dir`, and `transcript`.
  The last two are paths or `null` until available; `transcript` is the path of `raw_transcript.txt`.
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
A key outside visible ASCII, such as one with a line break, answers `AUTH_REJECTED` before any record, and the answer never quotes it.
The command needs no agent extension, profile, or summary prompt.

Three parts of this spec, each in its own file, define the rest of the command:

- [Output and cookies](spec/output-and-cookies.md)
- [Records and replay](spec/records-and-replay.md)
- [Protocol with the Python program](spec/protocol.md)

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
