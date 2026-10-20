# Test kit

This part says what the SDK's test kit runs and checks in an extension repository, from M18-T3, as [D063](../decisions/D063-sdk-test-kit.md) decides. The kit is the package `sdk/extensiontest`. `go test ./...` runs it with no Nicely checkout and no import of core's `internal/`.

## What it runs

- The real `ncly-<domain>`, built from `cmd/ncly-<domain>/`, with the sidecars that `Export` writes, installed in scratch folders
- The `ncly` that `go tool ncly` builds from the Nicely version of `go.mod`
- The scenarios of `testdata/script/`, with a scratch `HOME` and scratch XDG folders. They call no paid service and read no real keychain
- A fake child program, which the kit builds from Go source, because a Go module zip keeps no executable bit

A scratch `HOME` does not hide the OS keychain. `ncly` reads it only when the variable of a key is empty, so each granted key in a scenario gets a non-empty sentinel value. A missing-key case runs `ncly-<domain>` directly, with `NCLY_MISSING_KEYS`, as [Grants](grants.md) defines.

## What it checks

- `extension.Check`, and a skill named after the domain, which every external extension ships, as [Extensions](extensions.md) defines
- Each answer of a scenario against the answer schema that `contract.AnswerSchema` builds from the command's declaration, for a success and for a failure
- Each command that runs appears in the `ncly <domain> ...` call of at least one scenario
- The skill names no `ncly <domain>` command that the manifest lacks. Core's commands and other domains stay free to mention
- When the repository holds `.copier-answers.yml`, its Nicely version equals the one in `go.mod`. The failure says to run `just update <version>`

The two checks of commands match literal `ncly <domain> ...` text, and parse no shell or Markdown. Each failure names the file, the missing link, and its fix.

The kit proves the declared cases. Whether each promise of `spec.md` has a scenario, and whether a command has a hidden effect, stay judgments for the author and the reviewer.
