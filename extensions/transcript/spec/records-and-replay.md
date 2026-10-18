# Records and replay

Each run has one item and one `transcribe` step, which declares the `paid` effect.
The manifest supports recorded execution and dry run; domain resume remains unimplemented.
Use the SDK for records, locks, and publication, as [Operations](../../../docs/north-star/contract/operations.md) defines.

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
- A repeated call uses the run lock and input checks of [Run IDs](../../../docs/north-star/contract/run-ids.md).
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

The domain records these `inputs`, as [Records and evidence](../../../docs/north-star/contract/records-and-evidence.md) defines:

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
