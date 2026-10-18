# Scenarios

This folder holds core's testscript scenarios, as `.txtar` files. An extension keeps its own in `extensions/<name>/testdata/script/`, and the test binary of `cmd/ncly` runs both. Prefer a scenario, as D017 decides, and write a unit test only for logic that a scenario cannot reach.

## Write a scenario

1. Name each scenario as the card's **Proves** line does.
2. Reuse the contract coverage of M00, in `docs/milestones/M00-foundation.md`, and add the command's distinct risks.
3. For writing or paid work, check the filesystem effects, the partial results, and retry safety, not just the JSON text.
4. For a recorded operation, add a boundary scenario. The boundary hook kills the process that runs the operation at every step boundary. After each kill, the scenario checks the answer, the record, the files, and the counted calls. Once the operation supports resume, it also checks that a resume never repeats an unknown effect.
5. Run the scenarios by name with `just scenario <name>... [--count N]`. It fails unless each scenario passes N times. A raw `go test -run` pattern can select nothing and still pass.

Done when the scenario fails for the missing behavior. A test of a shared component never replaces the end-to-end proof of the first real command.

## Isolate a scenario

- Point `HOME` and the XDG folders at `$WORK`.
- Put stub executables in `PATH` for harnesses, `uv`, and paid services.
- Use the test keychain and the boundary hook, which live only in the test binary.
- Run each command's scenarios once under the pseudo-locale, so text outside the catalog fails a check.

## Scenario commands

The test binary adds these commands, beside others. The comment above each one, in `cmd/ncly/script_test.go` or `cmd/ncly/scriptcmds_test.go`, gives its usage.

| Command | Does |
|---|---|
| `onlyncly` | Keeps on `PATH` only `ncly` and the listed commands, so the other commands of the test binary are no extensions |
| `pseudo` | Fails on a word of a stream that the pseudo-locale does not mark |
| `exits` | Asserts an exact exit code, which `! exec` cannot |
| `answer` | Decodes a stream as one whole JSON answer, or as a failed report |
| `snapshot`, `unchanged` | Prove that a dry run changed nothing outside Nicely's cache |
| `agree` | Checks that a help page agrees with the `ncly describe` answer of its command |
| `schema` | Validates an answer against the schema that describe gives for it |
| `mentions` | Checks that a skill names only the commands that describe gives and the codes that exist |
