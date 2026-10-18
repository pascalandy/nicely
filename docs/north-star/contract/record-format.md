# Record format

Each run has one record, `runs/<run_id>.json` in the state directory, such as `~/.local/state/nicely/runs/`. A run keeps an artifact that no user folder holds, such as the answer of `ncly agent run`, in the folder `runs/<run_id>/` beside its record. The record is one JSON object. While holding `runs/<run_id>.lock`, the writer, core or an extension through the SDK, writes a temporary file in the same directory, syncs that file, renames it, then syncs the parent directory. Newly created state directories are synced with their parents too. This persistence barrier finishes before acknowledging a non-repeatable effect. A failed barrier prevents the effect. Atomic visibility alone is not durable acknowledgement.

The platform helper uses the durable flush supported by the system. On macOS, it also uses `F_FULLFSYNC` after syncing metadata, because [Apple's fsync documentation](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/fsync.2.html) distinguishes an ordinary sync from flushing the drive cache. A failed required flush stops the operation. Persistence relies on the filesystem and device honoring these operations; Nicely cannot promise survival of arbitrary hardware failure.

The execution process takes the run lock before it reads the record, and holds it until it exits. It waits at most 250 milliseconds for a lock that another process holds, then decides every case of [Run IDs](run-ids.md), including a missing record, on a read made while it holds the lock. A record file that cannot be read or parsed counts as an unsupported version, never as an absent record, so it never starts a second paid run. Inspection and dry run try the run lock without waiting and without creating its file, and release it at once, so they never make an execution see the run as busy.

The operating system releases each lock when its owner dies. Each lock in the state folder, such as `runs/<run_id>.lock` or `grants.lock`, is its own file, which nobody deletes: once a lock file is deleted, a new process can create and lock another file under the same name while an earlier process still holds the deleted one, and both own the lock. A lock file's existence alone therefore does not establish ownership. Before recording a step as completed, the domain independently verifies its artifacts and syncs their files and containing directories.

`run_id` is the ID that the caller chose or one that the SDK generated, as [Run IDs](run-ids.md) defines, such as `20261007-214501-3f9a2c`. Readers never parse it. A run's temporary directory sits in the system temporary directory, carries the run ID in its name, and is removed when the run ends, including after Ctrl-C or SIGTERM.

| Key | Type | Meaning |
|---|---|---|
| `format_version` | integer | The record format, initially `1`. A reader rejects any other value before work |
| `run_id` | string | The run ID |
| `path` | string | The command path, such as `transcript run youtube` |
| `status` | string | `running`, `completed`, `failed`, `interrupted`, or `unknown` |
| `started_at`, `updated_at` | string | UTC times in RFC 3339, such as `2026-10-07T21:45:01Z` |
| `ncly_version` | string | The release that last wrote the record |
| `attempts` | integer | The number of executions: 1, plus one per repeated call that safely starts work again |
| `inputs` | object | What the call states, as non-secret values and fingerprints, under keys that the domain defines. A call with the same run ID compares them |
| `resolved` | object | The non-secret values that preparation picked from the config or the machine, such as the selected model and effort, under keys that the domain defines. No call compares them to decide whether it is the same call, and every later attempt reads them from the record |
| `tools` | object | The version of each program that the run used, keyed by program name |
| `items` | array | One object per input item, in input order. A command with one input has one item |

Each item has `key`, a string that names the item, such as its URL, `output_dir`, the folder selected for its results, or `null` until the domain selects it, and `steps`, an array in execution order. The domain records `output_dir` before creating or writing that folder. Each step has these keys:

| Key | Type | Meaning |
|---|---|---|
| `id` | string | The step name that the domain defines, such as `transcribe` |
| `effect` | string | The strongest effect that the step can make, which the domain declares in the initial record: `none`; `repeatable`, harmless to repeat or covered by a demonstrated idempotent operation; `non_repeatable`, which a repeat would duplicate or conflict with; or `paid`, a request that may bill. Each value is more cautious than the one before it. Before an attempt persists the step's intent, record support raises the value to the command's current declaration when that one is more cautious, and nothing ever lowers it |
| `status` | string | `pending`, `completed`, `failed`, or `unknown` |
| `started_at`, `finished_at` | string or null | UTC times, `null` until the step starts or ends |
| `artifacts` | array | One object per verified output, with `path`, `size_bytes`, and `sha256` |
| `error` | object or null | The latest unsuccessful attempt's `code`, `message`, and `hint`, as in the public problem object. A preparation failure can leave the step `pending` with an error. An untouched pending or completed step has `null` |
| `evidence` | object, optional | Data beyond `artifacts` that the domain needs to reuse the step or to answer again, with its own `format_version`. A step without such data omits the key |

A reader treats a record that lacks a key of these two tables other than `evidence`, holds a value outside the type or the values that the tables give, or holds a `run_id` other than the name of its file, as a record that it cannot parse.

The domain defines when its validated preparation creates the initial record, before its first user-side mutation or paid dispatch. It starts with pending steps and status `running`. A run's saved status follows this table:

| Status | When written |
|---|---|
| `running` | The initial record and each repeated call that safely starts work again |
| `completed` | Every planned step has verified completion |
| `failed` | Execution ends without a signal, with a known error and no unknown step |
| `interrupted` | Ctrl-C or SIGTERM stops execution and no step has an unresolved effect, preserving each step's evidence |
| `unknown` | Execution ends, with or without a signal, and at least one step has an unresolved effect |

`unknown` outranks `interrupted`, because the first question after a stop is whether an effect may have happened, such as a paid request that may have billed. A signal still exits 130 or 143, as [Exit codes](exit-codes.md) defines, whatever the saved status. Abrupt death can leave `running` with no lock owner. Inspection preserves that saved status and reports `active: false`; it never invents completion or marks untouched pending steps unknown. A repeated call decides from the step evidence.

Each extension reuses these keys where they fit and defines its own `inputs`, `resolved`, and evidence in its spec. Core reads records to list and show them without interpreting an extension's evidence. The domain owns verification and replay; there is no workflow engine. Compatibility governs any later format change.
