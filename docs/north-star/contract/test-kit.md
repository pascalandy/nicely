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
- The scenarios of `testdata/script/`, in an [owned](owned-runs.md) scratch run with a scratch `HOME` and scratch XDG folders. They call no paid service and read no real keychain
- A `PATH` without any folder of the caller that holds an `ncly-*` file, so discovery finds only the extension under test
- A fake child program, which the kit builds from Go source, because a Go module zip keeps no executable bit
- On Linux, the guard, which the kit builds from Go source too. It supervises the calls of [Publication helpers](publication-helpers.md)

A scratch `HOME` does not hide the OS keychain. `ncly` reads it only when the variable of a key is empty. So each service that the manifest declares gets a non-empty sentinel key, `sentinel-<service>`, in its variable. On Linux, the kit also points D-Bus at no bus, so a read that a scenario forces finds no keychain instead of the user's. A missing-key case runs `ncly-<slug>` directly, with `NCLY_MISSING_KEYS`, as [Grants](grants.md) defines.

## Scenario commands

[Scenario commands](scenario-commands.md) lists the commands of a scenario, and says what each command of the kit checks.

## What it checks

- `extension.Check`, and a skill named after the slug, which every external extension ships, as [Extensions](extensions.md) defines
- Each answer of a call through `ncly` or `ncly-<slug>`, for a success and for a failure. Its schema is the one that `contract.AnswerSchema` builds from the command's declaration
- No scenario runs `ncly` or `ncly-<slug>` through `exec` in the foreground, by name or by path, which would skip the answer check. The kit splits each line into words as testscript does, so quotes and comments hide no call. A background call, which ends with `&`, keeps `exec`, and the kit does not check its answer
- Each runnable command has a recorded call through `ncly`, `ncly-<slug>`, `exits`, `signal`, `appears`, `refuses`, `predicts`, `keeps`, or `holds` after its answer passes validation
- A command whose declared result kind is `text` counts a completed success, since that kind has no JSON answer
- Coverage runs after all scenarios finish. Comments, assertions, skipped conditions, help, and shell or background `exec` calls earn no coverage
- `just scenario <name>` runs one scenario. It sets `NCLY_TEST_SCENARIO` to that name, and selects only that scenario with `-run`. The kit then skips coverage, which needs every scenario (#422)
- Human text earns no coverage for a command with a structured result
- A path-qualified Nicely call in `exits` is refused with the registered command spelling
- A failed report on stdout is checked as one complete object, including its envelope and declared result fields
- Each `reads` case reads the output with its reader, compares the reference, searches each private text, and requires its bad variant to fail. After all scenarios, each Output case of `## Case matrix` needs a case, as [Acceptance cases](acceptance-cases.md) says
- `## Publication` of `spec.md` declares each mode of each command that may write a user file, and the manifest agrees, as [Publication declaration](publication.md) says. After all scenarios, a check proves each obligation, or a reason covers it
- Each line of Failures and of Acceptance in `spec.md` starts with its own spec ID, and a test claims each ID, as [Spec IDs](#spec-ids) says. Each claim names an ID of `spec.md`
- The skill names no `ncly <root>` command that the manifest lacks. Core's commands and commands outside that root stay free to mention, including sibling extensions
- When the repository holds `.copier-answers.yml`, its `_commit`, the Nicely version of the template, equals the one in `go.mod`. For a development build, both name one commit. The failure says to run `just update <version>`

The skill check matches literal `ncly <root> ...` text and parses no shell or Markdown.
Each failure names the file, the missing link, and its fix.
The fixture extension repository in `testdata/extensions/standalone/` proves each check.
`testdata/extensions/standalone-named/` proves the named-extension checks, and Nicely's `just test`, `just lint`, `just fmt-check`, and `just tidy-check` cover both modules.

The kit proves the declared cases. Whether a claimed test proves its line, and whether a command has a hidden effect, stay judgments for the author and the reviewer.

`acceptance.sh` checks the primary output of the real input with a reader that is not the wrapped program.
A check that reads only a header or the file type is not sufficient.
The check compares the content that this reader gives with an independent reference, such as a file of expected text that the author writes.
An acceptance case makes the same check on a made input, and the kit runs it, as [Acceptance cases](acceptance-cases.md) says.

### Acceptance test files

Each case of the Output group of [Generic cases](generic-cases.md) that applies to the command needs a test file (#420).
A scenario makes that file in the extension's own domain.
`reads` proves its safe outcome, or the declared outcome of an accepted privacy limit, with an independent reader.
[Acceptance cases](acceptance-cases.md) defines this check. Two cases need more:

- **Keeps data that a reader needs.** The test file holds a valid block that a reader needs, and that block holds private text. The block uses a value that is not the default. An independent reader proves the declared privacy behavior and that the output still shows the content correctly. Without an [accepted privacy limit](acceptance-cases.md#accepted-privacy-limits), private text is absent, and the bad variant keeps the raw block and fails (#423)
- **Loses data that changes the meaning.** Some declarations change how a reader shows the content. The test file gives each of them a value that is not the default. The independent reference applies that declaration before it compares the content, so a lost declaration fails the check

The tests make each file that they use. They never copy a file of the user, and the author never searches the home folder of the user for sample files.
The tests make the file of each case above before the first review.
The tests and `just acceptance` need no root, no mount, and no system service.

Other scenario runners use `NewCoverage`, `Coverage.Observe`, and `Coverage.Missing` for the same answer and coverage checks.
`Observe` receives completed calls, including their exit code and both streams, without changing process execution or output.
Nicely's runner registers these checks for `agent` and `transcript`, for successful and failed calls.
It checks declared codes for `hello`, whose signal commands run in the background.
Its path-qualified direct calls pass the same answer check.
Its `exits` command checks successful and failed calls, including `exits 0` for successes.
`Coverage.ScenarioCommands` takes the registered host's absolute path and supplies `signal`, `survives`, `appears`, `protect`, and `refuses` to other scenario runners.

## Manifest examples

[Manifest examples](manifest-examples.md) says how the kit runs each example of the manifest, and how it reports a skipped example.

## Declared codes

Each declared error and warning must occur in a schema-valid JSON answer for each command and mode of its scope.
The kit reads top-level `errors` and `warnings`, and batch item `errors`.
Comments, assertions, arbitrary result fields, report findings, and human text earn no code coverage.
An invalid answer earns no code coverage.

The manifest's `commands` and `modes` select the scope, as [Manifest](manifest.md) defines.
Without a scope, a declaration selects every running command, in run and each supported dry run.
Narrow the scope to the commands and modes where the code can occur.
Use `untested` when a text command's warning cannot appear in a JSON answer.
The check leaves runtime prefix admission unchanged.

An `untested` declaration defers its code coverage and gives a reason with an issue URL.
The reviewer approves each deferral.
After all scenarios finish, the kit prints each deferral through `extensionrepo.Notice`, including its declaration position and reason.
The template's plain `just check` prints these reservations for the [UAT report](../../qa/protocol.md#report).
Other runners use `Coverage.Reservations` after their scenarios.
Each missing observation names its declaration position, kind, code, command, mode, and fix.

## Spec IDs

[Spec IDs](spec-ids.md) defines the links between spec promises and their tests.
