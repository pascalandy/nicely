# Nicely verification map

This folder is the maintained source for verifying what a user of `ncly` sees. Read this index before driving `ncly`, then use the file of each feature as its recipe.

## Baseline preconditions

- A run from `just verify`, whose folder `$run` holds `scratch/bin/ncly`
- The Doctor step of [SKILL.md](../SKILL.md#doctor) passed for this run
- Each command runs through `just verify step "$run" <feature> <ncly args...>`, from the root of the checkout

## Driving conventions

- Start each feature from a new run, unless its preconditions say otherwise
- Copy each command as it is written, flags and quotes included
- Read machine values from `--json` answers, and human text from the run without `--json`
- Check the exit code of every step against the one the feature names: a report that finds a problem exits 78, and that is its expected result
- Read stdout, stderr, and the exit code of a step from its transcript. On a nonzero exit, `just` adds its own `error: Recipe` line on the terminal, which is not the stderr of `ncly`

## Proof and skip reporting

- A proof quotes the call, its stdout, its stderr, and its exit code, from `just verify quote`, which holds only the `ncly` calls of `step`. Report the output of `prepare` and of a check outside `ncly`, such as `grep`, beside the quote
- A write is proved by a second command that reads its result
- Report an unreachable step with the command that failed and the precondition that it lacked
- Never report a step as verified through another entry point than the one it names

## Feature entry contract

Each feature file starts with an H1 title and one paragraph about what the user sees. Four H2 sections follow, in this order:

1. `Sub-features` lists short IDs, one line each
2. `How to get to it (user POV)` lists every command that reaches the feature
3. `Driving it with verify.sh` starts with `Preconditions:`, then pairs each user action with its exact command and its observable result
4. `Gotchas` lists the traps that waste or invalidate a run

## Features

- `hello.md` covers the Go fixture installed as an external extension: discovery, help, answers, errors, reports, describe, its skill, and doctor
- `auth.md` covers `ncly auth` against the real keychain, with a test service and a fake key: status, login, logout, and the cleanup of the fake key
