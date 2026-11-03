# Analyzers

This part says which Go analyzers the SDK holds, and how core and an extension repository run them.
[D068](../decisions/D068-verification-map.md) puts these checks at level 2, with core as their owner.

## How they run

The package `sdk/analysis` holds the rules, and its command `nicelyvet` serves them to `go vet -vettool`.
`nicelyvet` only reads the packages that `go vet` gives it, so `go run` of it on a package path fails.

- Core's `just lint` runs them on every package of core and of its fixture extensions, through the `tool` line of `go.mod`
- `extensionrepo check`, which `just check` of an extension repository calls, runs them after `go vet`

`extensionrepo check` builds `nicelyvet` into a scratch folder from the Nicely module that the repository pins.
An extension therefore gets the rules of its pinned version, with no copy of a rule and no change to its recipe.
A new rule reaches an extension after `just update` to the Nicely release that adds it.

A rule reads production files only.
It skips each `_test.go` file, because a test can do on purpose what the rule refuses.

## Rules

| Rule | Refuses | Use instead | Past mistake |
|---|---|---|---|
| `callerenv` | `os.Environ` or `syscall.Environ` in `sdk/extensionrepo` and `sdk/extensiontest`, outside `extensionrepo.NewOwnedRun` | An [owned run](owned-runs.md) | The parent of 09809608 |
| `runnerdiscard` | `io.Discard` as the `Stdout` or `Stderr` of a `runner.Program`, in a struct literal or a field assignment | `nil`, or no field | 3e6a4eff |

**callerenv.** Each check that copies the environment of its caller filters it in its own way.
Each new caller setting then needs a fix in each copy, as 17 past corrections did.
The rule reads a call and a function value.
It reads only the two packages of the kit and the repository gates, so an extension never meets it.
A program that starts with no `Env` inherits the environment of the caller too, and the rule cannot see that.

**runnerdiscard.** The runner reads a writer that is not a file through a pipe.
A descendant that keeps that pipe open holds the call until the drain limit of [Programs that ncly runs](programs-that-ncly-runs.md).
A `nil` writer gives the program the null device, with no pipe.
The rule follows no variable, so `w := io.Discard` then `Stdout: w` passes.
It reads keyed fields only, because `go vet` already refuses unkeyed fields of an imported struct.

## Add a rule

1. Write cases that repeat the past mistake, and see them fail.
2. Add the rule as a package beside `runnerdiscard`.
3. List the rule in `Analyzers` of `sdk/analysis`.
4. Add a row to the table of [Rules](#rules).
