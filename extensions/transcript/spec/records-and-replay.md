# Records and replay

Each run has one item and one `transcribe` step, which declares the `paid` effect.
The manifest supports recorded execution and dry run; domain resume remains unimplemented.
Use the SDK for records, locks, and publication, as [Operations](../../../docs/north-star/contract/operations.md) defines.

Python emits the single-item plan before user-side effects.
Go validates it and durably creates the initial record with a pending step.
The step becomes unknown before result-folder reservation or upload.
Go records completion only after it verifies and syncs the required artifacts, as [Verified artifacts](#verified-artifacts) says.

- A download failure before intent leaves the step pending with its code, message, and hint.
- After acknowledgement, a known pre-upload transport failure or a received 4xx refusal makes the step failed.
  A 5xx response or lost connection after sending audio leaves it unknown.
  Upload retries, including SDK retries, are disabled.
  One acknowledgement permits at most one upload request.
- The record keeps the program and `yt-dlp` versions reported by the intent, in `tools` under `transcript` and `yt-dlp`.
  Its item key is the URL exactly as supplied.
- A repeated call uses the run lock and input checks of [Run IDs](../../../docs/north-star/contract/run-ids.md).
  A different command, URL, or explicit output argument is `USAGE_INVALID` before effects.
  A busy run returns `TEMPORARY` with a `run view` hint and starts no program.
- A completed run returns its verified saved result without Python, a dependency upgrade, or a Deepgram request.
  It needs no key: the extension reads the key only for a new attempt, after the run lock and this decision.
  Core may still read a granted key first, as [Keys](../../../docs/north-star/contract/keys.md) defines, and a replay answers even when that read fails.
  A dry run with the same ID reads the record without the lock and decides alike: it plans reuse or refuses.
  Before it returns success, it applies [Verified artifacts](#verified-artifacts) and compares each size and hash with the record.
  Missing, changed, or unsupported evidence returns `RESUME_UNSAFE`.
- A lost final protocol message leaves the step unknown, so a repeat refuses with `RESUME_UNSAFE`.
  Pascal cut its local recovery on 2026-10-18, as [D038](../../../docs/north-star/decisions/D038-python-protocol.md) records.
  [XT-102](../../../docs/backlog/XT-102-transcript.md#later) parks it.
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
A new attempt of a pending run reads `output` and `browser` from `resolved`, without the configuration.
Its output parent must still exist, or the call answers `RESUME_UNSAFE` and starts nothing.
The recorded output folder and verified artifacts determine replay.
A saved run keeps its original `started_at`.

## Verified artifacts

After a `completed` step, Go checks each required name before it records completion:

1. The name is a regular file directly inside `output_dir`, not a symlink.
2. Go computes its size and SHA-256, and checks that its content is usable.
3. Go syncs the three files and the folder, then calls `Run.Complete` with the three artifacts.

Usable content:

- `raw_transcript.txt` is valid UTF-8 and not blank.
- `raw_sentences.txt` is valid UTF-8. An empty sentence list stays valid, as in the source.
- `raw_transcript.json` is valid UTF-8 and holds one JSON array, with nothing after it.

A file that fails a check leaves the step unknown, with the overall `RUNTIME` of a protocol failure.
Completion lists the three artifacts with `path`, `size_bytes`, and `sha256`.
Its `evidence` holds `format_version: 1` and the program's `warnings`.
