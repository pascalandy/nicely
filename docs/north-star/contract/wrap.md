# Wrapped programs

This part says how `sdk/wrap` runs one program for a handler, such as a worker. A handler that needs more than one program uses `sdk/runner` directly.

A wrapped program runs through the program runner of [Programs that ncly runs](programs-that-ncly-runs.md), in every mode:

- It is a program that the command declares in `requires.bins`, as [Manifest](manifest.md) defines. Any other program fails with `RUNTIME`
- It gets an empty stdin. The runner keeps its stdout and its stderr up to a size limit each, 1 MB unless the handler sets another. With `DiscardStdout`, the runner keeps none of stdout
- It runs without the terminal, in a new session, as Programs that ncly runs defines. A program that needs the terminal uses a plain SDK handler
- Its environment comes from `runner.Environ`, plus only the keys that the command declares and that `Call.Key` returns
- It starts from the caller's folder, so a relative path such as `./input.md` keeps the caller's meaning

`sdk/wrap` turns each failure of a wrapped program into the result that the contract defines for it:

| Case | Result |
|---|---|
| The program is missing or too old, or its version probe fails or times out | `PREREQ_MISSING` before any work, as [Required programs](../core-spec.md#required-programs) defines. The probe gets at most 10 seconds, within the command's timeout. `ProgramCheck.Failed` leaves a timeout to its caller, so `sdk/wrap` maps it |
| The input is invalid | `USAGE_INVALID`, which the handler answers before it calls `sdk/wrap` |
| The program exits with a code other than 0, or writes on stderr | The outcome that the handler maps for that exit, and `RUNTIME` for any other exit. No exit code of the program passes through unmapped, and its own 75 never makes a repeat safe |
| The output is missing, malformed, or over the size limit | `RUNTIME`. A decoder never turns a failure into a success. Only stdout counts for the limit, and stderr keeps its first bytes. Discarded stdout never counts |
| The command's timeout expires | The process tree stops, and `sdk/wrap` answers `RUNTIME` with exit 1. Exit 75 needs [Retry safety](retry-safety.md), which only the handler can show, so such a handler uses `sdk/runner` |
| Ctrl-C or SIGTERM arrives | The runner signals the process tree, and the exit is 130 or 143 |

A program that prints its progress on stdout, such as HandBrakeCLI, can pass the size limit in a long run. The program can succeed, but `sdk/wrap` still answers `RUNTIME`, and the decoder never runs. A handler whose decoder reads no stdout sets `DiscardStdout`. A handler that reads it sets an `OutputLimit` that its longest run cannot pass.

A command that declares more keeps the rules of each declaration:

- A writing command shares preparation with its dry run, which starts no effect, and needs consent to overwrite, as [Global flags](global-flags.md) defines
- A command that needs a key fails before work when its grant is missing or denied, as [Grants](grants.md) defines
- A command with paid or reusable partial work writes a record before its first effect, as [Operations](operations.md) defines

`runner.Environ` removes only the variables named `*_API_KEY`. Another secret in the caller's environment, such as a token, still reaches the program. The extension's spec states this limit when its program could read such a secret.
