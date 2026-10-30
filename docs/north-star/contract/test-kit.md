# Test kit

[Extension identity](extension-identity.md) defines the slug and root used below.
`<root>` stands for one or two command words, not a slug word.
The kit matches the complete root when it checks scenario coverage and skill commands.

This part says what the SDK's test kit runs and checks in an extension repository, as [D063](../decisions/D063-sdk-test-kit.md) decides. The kit is the package `sdk/extensiontest`. `go test ./...` runs it with no Nicely checkout and no import of core's `internal/`. One test of the repository calls it:

```go
func TestExtension(t *testing.T) { extensiontest.Run(t, Extension) }
```

## What it runs

- The real `ncly-<slug>`, built from `cmd/ncly-<slug>/`, with the sidecars that `Export` writes, installed in scratch folders
- The `ncly` that `go tool ncly` builds from the Nicely version of `go.mod`, whose `tool` line names `github.com/pascalandy/nicely/cmd/ncly`
- The scenarios of `testdata/script/`, with a scratch `HOME` and scratch XDG folders. They call no paid service and read no real keychain
- A `PATH` without any folder of the caller that holds an `ncly-*` file, so discovery finds only the extension under test
- A fake child program, which the kit builds from Go source, because a Go module zip keeps no executable bit

A scratch `HOME` does not hide the OS keychain. `ncly` reads it only when the variable of a key is empty. So each service that the manifest declares gets a non-empty sentinel key, `sentinel-<service>`, in its variable. On Linux, the kit also points D-Bus at no bus, so a read that a scenario forces finds no keychain instead of the user's. A missing-key case runs `ncly-<slug>` directly, with `NCLY_MISSING_KEYS`, as [Grants](grants.md) defines.

## Scenario commands

A scenario uses the commands of [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript), and these ones of the kit:

| Command | What it does |
|---|---|
| `ncly <arg>...` | Runs `ncly`, which exits 0, or another code after `!`, and checks the answer |
| `ncly-<slug> <arg>...` | Runs the extension directly, the same way |
| `exits <code> <command> <arg>...` | Runs a command, these two included, and checks its exact exit code |
| `signal INT\|TERM <marker> <command> <arg>...` | Waits for a new marker, signals the host, and checks its exit and answer |
| `survives <file>...` | Requires these finished files at the next signal's marker and unchanged after its exit |
| `protect <path>...` | Selects files, folders, or absent paths that subsequent `refuses` calls must keep unchanged |
| `refuses <code> <command> <arg>...` | Requires the same refusal code and exit in a dry run and a run |
| `grant <service>` | Grants the key of a declared service to the extension at its scratch path |
| `fake <name>` | Puts the fake child in `$WORK/bin`, first in `PATH`, under the name `name` |

The fake child runs the actions of `$WORK/bin/<name>.fake`, one per line: `stdout <text>`, `stderr <text>`, `stdin`, `env`, `sleep <duration>`, and `exit <code>`. `env` writes the names of its variables, never a value. The package comment of `sdk/extensiontest/fake` defines each action.

`signal` and `refuses` accept only registered `ncly` and `ncly-<slug>` calls of runnable commands.
They enable JSON and non-interactive mode, and validate each answer through `Coverage.Observe`.
Explicit flags that disable JSON or non-interactive mode are refused.
Both commands leave the last call's streams available to `stdout` and `stderr` assertions.

`signal` runs the host in the foreground, without a shell cleanup command supplying its result.
Its marker must not exist before the call.
It waits up to 10 seconds for a regular marker file, then sends the signal to the host process.
The host must exit within 15 seconds after the signal.
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

## What it checks

- `extension.Check`, and a skill named after the slug, which every external extension ships, as [Extensions](extensions.md) defines
- Each answer of a call through `ncly` or `ncly-<slug>`, for a success and for a failure. Its schema is the one that `contract.AnswerSchema` builds from the command's declaration
- No scenario runs `ncly` or `ncly-<slug>` through `exec` in the foreground, by name or by path, which would skip the answer check. The kit splits each line into words as testscript does, so quotes and comments hide no call. A background call, which ends with `&`, keeps `exec`, and the kit does not check its answer
- Each runnable command has a recorded call through `ncly`, `ncly-<slug>`, `exits`, `signal`, or `refuses` after its answer passes validation
- A command whose declared result kind is `text` counts a completed success, since that kind has no JSON answer
- Coverage runs after all scenarios finish. Comments, assertions, skipped conditions, help, and shell or background `exec` calls earn no coverage
- Human text earns no coverage for a command with a structured result
- A path-qualified Nicely call in `exits` is refused with the registered command spelling
- A failed report on stdout is checked as one complete object, including its envelope and declared result fields
- Each line of Failures and of Acceptance in `spec.md` starts with its own spec ID, and a test claims each ID, as [Spec IDs](#spec-ids) says. Each claim names an ID of `spec.md`
- The skill names no `ncly <root>` command that the manifest lacks. Core's commands and commands outside that root stay free to mention, including sibling extensions
- When the repository holds `.copier-answers.yml`, its `_commit`, the Nicely version of the template, equals the one in `go.mod`. The failure says to run `just update <version>`

The skill check matches literal `ncly <root> ...` text and parses no shell or Markdown.
Each failure names the file, the missing link, and its fix.
The fixture extension repository in `testdata/extensions/standalone/` proves each check.
`testdata/extensions/standalone-named/` proves the named-extension checks, and Nicely's `just test`, `just lint`, `just fmt-check`, and `just tidy-check` cover both modules.

The kit proves the declared cases. Whether a claimed test proves its line, and whether a command has a hidden effect, stay judgments for the author and the reviewer.

Other scenario runners use `NewCoverage`, `Coverage.Observe`, and `Coverage.Missing` for the same answer and coverage checks.
`Observe` receives completed calls, including their exit code and both streams, without changing process execution or output.
Nicely's runner registers these checks for `agent` and `transcript`, for successful and failed calls.
Its `exits` command checks successful and failed calls, including `exits 0` for successes.
`Coverage.ScenarioCommands` supplies `signal`, `survives`, `protect`, and `refuses` to other scenario runners.

## Spec IDs

The kit reads each part of `spec.md` named `Failures` or `Acceptance`. A part starts at a heading, or at a line that starts with a bold label, such as `**Failures.**`. It ends at the next one. Each row of a table and each bullet of these parts is one promise, and starts with its spec ID in backticks:

| ID | Condition | Code |
|---|---|---|
| `missing-file` | The file is missing or is no regular file | `USAGE_INVALID` |

- `duration`: The duration of the output, within 0.1 s

A spec ID is lowercase words that hyphens join. No two lines share one. A bullet takes a colon after its ID, so a bullet that only starts with code holds no ID. A prose paragraph holds no promise.

A test claims IDs with a line of its own, `# spec: <id>...`, above the commands that prove them. The line goes in the script of a scenario, or in `acceptance.sh`, which proves the real input of `## Acceptance`.
