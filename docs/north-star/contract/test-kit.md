# Test kit

This part says what the SDK's test kit runs and checks in an extension repository, as [D063](../decisions/D063-sdk-test-kit.md) decides. The kit is the package `sdk/extensiontest`. `go test ./...` runs it with no Nicely checkout and no import of core's `internal/`. One test of the repository calls it:

```go
func TestExtension(t *testing.T) { extensiontest.Run(t, Extension) }
```

## What it runs

- The real `ncly-<domain>`, built from `cmd/ncly-<domain>/`, with the sidecars that `Export` writes, installed in scratch folders
- The `ncly` that `go tool ncly` builds from the Nicely version of `go.mod`, whose `tool` line names `github.com/pascalandy/nicely/cmd/ncly`
- The scenarios of `testdata/script/`, with a scratch `HOME` and scratch XDG folders. They call no paid service and read no real keychain
- A `PATH` without any folder of the caller that holds an `ncly-*` file, so discovery finds only the extension under test
- A fake child program, which the kit builds from Go source, because a Go module zip keeps no executable bit

A scratch `HOME` does not hide the OS keychain. `ncly` reads it only when the variable of a key is empty. So each service that the manifest declares gets a non-empty sentinel key, `sentinel-<service>`, in its variable. On Linux, the kit also points D-Bus at no bus, so a read that a scenario forces finds no keychain instead of the user's. A missing-key case runs `ncly-<domain>` directly, with `NCLY_MISSING_KEYS`, as [Grants](grants.md) defines.

## Scenario commands

A scenario uses the commands of [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript), and these ones of the kit:

| Command | What it does |
|---|---|
| `ncly <arg>...` | Runs `ncly`, which exits 0, or another code after `!`, and checks the answer |
| `ncly-<domain> <arg>...` | Runs the extension directly, the same way |
| `exits <code> <command> <arg>...` | Runs a command, these two included, and checks its exact exit code |
| `grant <service>` | Grants the key of a declared service to the extension at its scratch path |
| `fake <name>` | Puts the fake child in `$WORK/bin`, first in `PATH`, under the name `name` |

The fake child runs the actions of `$WORK/bin/<name>.fake`, one per line: `stdout <text>`, `stderr <text>`, `stdin`, `env`, `sleep <duration>`, and `exit <code>`. `env` writes the names of its variables, never a value. The package comment of `sdk/extensiontest/fake` defines each action.

## What it checks

- `extension.Check`, and a skill named after the domain, which every external extension ships, as [Extensions](extensions.md) defines
- Each answer of a call through `ncly` or `ncly-<domain>`, for a success and for a failure. Its schema is the one that `contract.AnswerSchema` builds from the command's declaration
- No scenario runs `ncly` or `ncly-<domain>` through `exec` in the foreground, by name or by path, which would skip the answer check. The kit splits each line into words as testscript does, so quotes and comments hide no call. A background call, which ends with `&`, keeps `exec`, and the kit does not check its answer
- Each command that runs appears in the `ncly <domain> ...` call of at least one scenario
- The skill names no `ncly <domain>` command that the manifest lacks. Core's commands and other domains stay free to mention
- When the repository holds `.copier-answers.yml`, its `_commit`, the Nicely version of the template, equals the one in `go.mod`. The failure says to run `just update <version>`

The two checks of commands match literal `ncly <domain> ...` text, and parse no shell or Markdown. Each failure names the file, the missing link, and its fix. The fixture extension repository in `testdata/extensions/standalone/` proves each check.

The kit proves the declared cases. Whether each promise of `spec.md` has a scenario, and whether a command has a hidden effect, stay judgments for the author and the reviewer.
