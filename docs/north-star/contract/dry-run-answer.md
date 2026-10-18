# Dry-run answer

A recorded operation answers `--dry-run --json` with `dry_run`, set to `true`, then `plan` and `pending_checks`. The recorded commands of extensions, such as [transcript](../../../extensions/transcript/spec.md#dry-run-examples), use this shape.

| Key | Type | Meaning |
|---|---|---|
| `plan` | array | One object per item, in input order, with `key`, `output` or `output_dir`, `steps`, and `effects` |
| `plan[].steps` | array | Each step that execution would reach, with `id` and `action`, which is `run` or `reuse` |
| `plan[].effects` | object | The booleans `user_writes`, `network`, and `paid`, for the item's `run` steps |
| `pending_checks` | array | Each check that a `run` step needs and that only execution can make, with `id`, such as `auth.deepgram`, and `message`. A dry run reads no key, so a `run` step that needs a key keeps its authentication pending. The array is empty when no `run` step needs such a check |

A new item has `output`, the output folder that will hold its result folder. An item that already has a result folder, such as a verified-result replay, has `output_dir`. A repeated call that execution would refuse fails its dry run with the same error. A dry run may check local completion evidence but never commits it.
