# Run IDs

A recorded command accepts `--run-id <id>`, so that its caller knows the run's ID before the run starts, as [D052](../decisions/D052-caller-run-ids.md) decides. Without it, the SDK generates one. An ID holds 1 to 64 lowercase letters, digits, and hyphens, and starts with a letter or a digit. Any other value is `USAGE_INVALID`.

A call takes the run lock of its ID before it reads the record, as [Record format](record-format.md) defines. When another process still holds the run after that short wait, the call exits 75 with `TEMPORARY` and a `run view` hint, and starts nothing. Otherwise it decides on a read of the record made while it holds the lock. A call whose ID no record holds starts a new run under that ID. A call whose ID a record holds starts no new run. Before any effect, it takes the first case that applies:

1. The record's format version is unsupported, or its file cannot be read or parsed: `RESUME_UNSAFE`
2. The record's command path, `inputs`, or ordered item keys differ from what the call states: `USAGE_INVALID`, because the ID names another call. An item key that preparation picked comes from the record and is not compared
3. The run completed: its verified saved result, as [Records and evidence](records-and-evidence.md) builds it, without the rest of preparation and without running a step again
4. The domain can establish completed work from the saved record and complete local artifacts: it verifies and records completion, as [Inspection and explicit resume](inspection-and-explicit-resume.md) defines, then returns the saved result without dispatch
5. Every step whose intent the run persisted declares `none` or `repeatable`, both in the record and in the command's current declaration: the call runs every step again under the same ID. It increments `attempts` and writes `running` before the first step
6. Otherwise: `RESUME_UNSAFE`, with the artifacts still available, including when the current declaration of such a step is missing or more cautious than the recorded one

A repeated call with the same ID therefore never repeats an effect that the record does not prove safe, and a dry run with it describes what the call would do without changing the record. After lost output, a caller that chose the ID reads the run with `ncly run view <id>`, or makes the same call again, and needs no search with `ncly run list`.

Deleting a record forgets its run: a later direct call with its ID starts a new run and may bill again. Removing records is not a recovery mechanism. Nested recorded calls and their missing-record recovery rules are deferred with the first actual consumer in [COR-621](../../backlog/COR-621-resume-sdk.md).
