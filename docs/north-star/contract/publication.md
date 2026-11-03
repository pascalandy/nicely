# Publication declaration

This part says how `spec.md` declares each command that creates or replaces a user file, and what the test kit requires of it.
It applies [Generic cases](generic-cases.md) to the publication of the output, as COR-104-T6 of [D068](../decisions/D068-verification-map.md) decides.
The extension keeps its publication code. The kit checks the declaration and the runs.

## The form

`spec.md` holds a part `## Publication`. Its first table has one row for each mode of each command that creates or replaces a user file:

| Command | Mode | Destination | Overwrite | Protected | Operation | Dry run |
|---|---|---|---|---|---|---|
| `ncly notes export` | create | `output` | `overwrite` | `file` | `link` | checks |
| `ncly notes export` | replace | `output` | `overwrite` | `file` | `rename` | checks |

- **Command**: a running command of the manifest, in backticks. It declares a dry run, because every command that writes accepts `--dry-run`
- **Mode**: `create` for a run without `--force`, or `replace` for a run with it. A command with `replace` declares the switch `force`
- **Destination**: the key that holds the path of the published file, in the dry-run answer and in the answer of the run. Both schemas of the manifest declare it
- **Overwrite**: the boolean key of the dry-run answer that says whether the run replaces a file. The dry-run schema declares it
- **Protected**: the arguments, such as `file`, and the flags with one value, such as `--input`, whose paths the run keeps unchanged. Write `None` when no path is protected
- **Operation**: each operation that publishes the file in this mode. Each operation is also an ability that the file system of the destination needs
- **Dry run**: `checks` when the dry run checks each operation as the run does, and refuses as the run refuses. `pending` when the dry run cannot check them, so it predicts no success

These are the operations:

| Operation | What it does |
|---|---|
| `link` | Makes a hard link, which fails on an existing file |
| `rename-noreplace` | Renames, and refuses to replace a file, such as `renameat2` with `RENAME_NOREPLACE` on Linux, or `renamex_np` with `RENAME_EXCL` on macOS |
| `rename` | Renames, and replaces a file. The mode `create` cannot use it, because Generic cases forbid a check and then a rename that replaces |

With `pending`, the dry-run answer lists `publication.<operation>` in `pending_checks` for each operation, such as `publication.link`.
An item is that string, or an object whose `id` is that string.

A command whose `effects` set `user_writes`, as [Manifest](manifest.md) says, runs the mode `create`.
It also runs the mode `replace` when it declares the switch `force`.
Each mode that it runs has a row, or a reason for the obligation `publication`.
A row of a command whose `effects` do not set `user_writes` fails.
When no command creates or replaces a user file, the part says `None`.

## Obligations

Each row gives obligations. A scenario proves each one with a check of the kit, or the second table gives its reason:

| Obligation | Of | The check that proves it |
|---|---|---|
| `predicts` | Each row | A `predicts` call of the command in that mode passes |
| `appears` | A row `create` | An `appears` call of the command passes |
| Each operation, such as `link` | Each row | A check on a destination without that ability. The kit has no such check yet, so the obligation needs a reason, as COR-104-T7 plans |
| `publication` | A mode with no row | None: the mode writes no user file, though the `effects` of its command say so |

The second table has the columns Command, Mode, Obligation, and Reason:

| Command | Mode | Obligation | Reason |
|---|---|---|---|
| `ncly notes export` | create | `link` | The refusal of `refuses` covers a destination without hard link, and the source review checks that the run never falls back |

The Mode of the obligation `publication` is the mode with no row, or `-` for a command with no row.
Each reason names an obligation of its row, and says why no scenario proves it and what proves it instead.
The reviewer of round 1 accepts each reason, as for an `untested` code.

## What the kit checks

Before the scenarios, the kit checks the form against the manifest, as [Test kit](test-kit.md#what-it-checks) lists.
After all scenarios finish, it reports each obligation without a passing check and without a reason.
It also reports a reason for an obligation that a check proved, so that no reason stays after its check.
`just scenario` skips this report, as it skips coverage.
[Scenario commands](scenario-commands.md) says what `predicts` checks.
