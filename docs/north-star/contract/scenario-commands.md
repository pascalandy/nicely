# Scenario commands

This part says which commands a scenario of an extension repository uses, and what each command of the kit checks.
[Test kit](test-kit.md) says how the kit runs the scenarios.

A scenario uses the commands of [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript), and these ones of the kit:

| Command | What it does |
|---|---|
| `ncly <arg>...` | Runs `ncly`, which exits 0, or another code after `!`, and checks the answer |
| `ncly-<slug> <arg>...` | Runs the extension directly, the same way |
| `exits <code> <command> <arg>...` | Runs a command, these two included, and checks its exact exit code |
| `signal INT\|TERM <marker> <command> <arg>...` | Waits for a new marker, signals the host, and checks its exit and answer |
| `survives <file>...` | Requires these finished files at the next signal's marker and unchanged after its exit |
| `appears <marker> <path> <command> <arg>...` | Creates a new file at the marker, and requires a failed run that keeps that file |
| `protect <path>...` | Selects files, folders, or absent paths that the next `signal` or `appears` call and subsequent `refuses` calls must keep unchanged |
| `refuses <code> <command> <arg>...` | Requires the same refusal code and exit in a dry run and a run |
| `grant <service>` | Grants the key of a declared service to the extension at its scratch path |
| `fake <name>` | Puts the fake child in `$WORK/bin`, first in `PATH`, under the name `name` |

The fake child runs the actions of `$WORK/bin/<name>.fake`, one per line: `stdout <text>`, `stderr <text>`, `stdin`, `env`, `sleep <duration>`, and `exit <code>`. `env` writes the names of its variables, never a value. The package comment of `sdk/extensiontest/fake` defines each action.

`signal`, `appears`, and `refuses` accept only registered `ncly` and `ncly-<slug>` calls of runnable commands.
They enable JSON and non-interactive mode, and validate each answer through `Coverage.Observe`.
Explicit flags that disable JSON or non-interactive mode are refused.
These commands leave the last call's streams available to `stdout` and `stderr` assertions.
Their `ncly` calls use the registered host's absolute path, even when a scenario replaces `PATH`.

`signal` runs the host in the foreground, without a shell cleanup command supplying its result.
Its marker must not exist before the call.
It waits up to 10 seconds for a regular marker file, then sends the signal to the host process.
The exit deadline allows `runner.Cleanup`, `runner.DrainLimit`, and a three-second scheduling margin after the signal.
The host starts in its own process group, which the kit kills on timeout or failure.
Stream draining stops within `runner.DrainLimit` after the host exits, even when a child keeps a stream open.
INT requires exit 130 and `INTERRUPTED`; TERM requires exit 143 and `TERMINATED`.
`survives` selects files for one signal call, including files that the call finishes before its marker.
Their bytes and permissions must remain unchanged after the host exits.

`refuses` requires a command with a declared dry run and arguments without dry-run flags.
It adds `--dry-run` for the first call and uses the original arguments for the second call.
Both calls receive the same stdin bytes.
Both calls must fail with the requested leading error code and the same nonzero exit code.
Each call must preserve the protected paths, including absent paths and each folder's contents.
The comparison records symbolic links without following them.
Nicely's cache stays unprotected unless a scenario selects its path.
Select inputs and possible outputs with `protect` before each refusal group.

A `protect` selection applies to each later `refuses` call, until the next `protect` call replaces it.
It also applies to the next `signal` or `appears` call, which ends it for these two commands.
Give a new `protect` call before each `signal` or `appears` call that must keep paths.
`signal` compares the protected paths before the host starts and after it exits, except its marker.
So an interrupted run must keep its input, leave an absent output absent, and keep an old output.

To prove that a command never writes through a link, make its output a symbolic link to its input.
Protect both paths, then call `refuses`.
A write through the link changes the input, and a replacement of the link changes the link.

`appears` proves that a run without `--force` never replaces a file that appears during the run.
It refuses `--force` among the arguments.
Its marker and its path must not exist before the call.
It runs the host as `signal` does, and waits up to 10 seconds for a regular marker file.
Then it creates the path with an exclusive create, and writes bytes that the kit chooses.
So the host must wait for the path after it writes the marker, for example in a stub of the wrapped program.
After the path appears, the host has 10 seconds to exit, and `runner.DrainLimit` to close its streams.
The host must fail with a valid answer, and its exit must not be 0.
The path must stay the same regular file, with the same bytes and permissions.
So a run that renames over the file, writes through it, or puts a link in its place fails.
The protected paths must stay unchanged, except the marker and the path.

These calls prove the invariants of a command that replaces a user file:

- The input stays unchanged: `protect` the input before `refuses`, `signal`, and `appears`.
- A failure leaves no partial output and keeps the old output: `protect` the output path or its folder before `refuses` and `signal`.
- The command never writes through a link: `refuses`, with the output as a symbolic link to the input.
- Without `--force`, the command never replaces a file that appears during the run: `appears`.
