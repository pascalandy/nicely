---
name: nicely-extension
description: Use when wrapping a script or a program as a command of this Nicely extension, when changing a command, or when the user asks whether the extension is ready.
---

# Nicely extension

This skill takes a script from intake to a verified command.
`AGENTS.md` gives the reading order, recipes, and extension root.
The domain names a family of work, such as `video`.
An optional name identifies one job, such as `convert`.
Neither names the wrapped tool.
Read "Extension identity" in the contract for the slug and command ownership.

The contract holds every rule. `just contract` prints its folder at the Nicely version that `go.mod` pins. Read "What a great extension does" in its `extensions.md` first. Then read only the parts that it links.

## Intake

Get these facts before you write code. Read the source of the script, and ask its author what the source does not show. Do not run the script to learn what it does, because even its `--help` can write or bill.

- The program and its entry point, such as `python3 tool.py`, with its source and revision, for `vision.md`
- Its inputs: the arguments, stdin, files, environment variables, and the folder that it reads
- Its result: what proves that the work happened, on which stream, and in which format
- Its effects: each write, network call, and paid request, also in a dry run or in `--help`
- Its keys and logins, and where it reads them
- How long it runs, what it starts in the background, and whether it can ask a question
- Its exit codes, and what a repeat of the same call does

Write each fact that nobody can give as an open question in `spec.md`.

## Choice of operations

- Give each command one bounded operation beneath the extension root. Never relay arbitrary arguments to the script
- A named root can take arguments and run, such as `ncly video convert <file>`
- Give a script that does several jobs one command per job, and only for the jobs that the user needs
- Put the script in `worker/` when the extension can embed it. `Call.Worker` gives its path, and `sdk/wrap` runs it
- Use `sdk/wrap` for one program per call. Use `sdk/runner` for several programs, or for a timeout that can exit 75 safely
- Declare each effect, program, and key in `manifest.toml`. When you cannot tell whether an effect occurs, declare it
- Give `--dry-run` to a command that writes or bills. The dry run shares the preparation and starts no effect

## Error mapping

`sdk/wrap` answers each failure that "Wrapped programs" in the contract lists. The handler maps the rest:

- Check each input before the program starts. Answer `USAGE_INVALID`, with a hint that shows a call that works
- Map each exit code that the spec names to its outcome. Return `Ended.Failed` for every other code, so no code passes through
- Decode the output into a typed value. Return `Ended.Malformed` when a field is missing or invalid, so a failure never becomes a success
- Never give exit 75 for the program's own 75. Only the handler can prove that a repeat is safe, as "Retry safety" in the contract says
- Put each message and hint in `locales/en.toml`, never in code

## Verification

1. Follow "Change a command" in `AGENTS.md`. Give each case of the command's spec a scenario
2. Run `just check`. Its test kit installs the extension into scratch folders and runs the scenarios
3. Fill in `acceptance.sh` from `## Acceptance` in `spec.md`. Then run `just acceptance` with the real input. For a paid command, run only its dry run
4. For a command that needs a key, relay the grant hint to the human and wait. Never grant a key yourself

## Readiness review

Run this review only when the user asks for it. The checks of the contract apply in all cases: `just check` runs them, and declining the review never waives one.

Review the extension against "What a great extension does" and against the facts of the intake. Write each finding with these six parts:

- **Evidence**: the file and the line, the scenario, or the output that shows the problem
- **Command**: the command that the finding affects
- **Consequence**: what goes wrong for a user or an agent
- **Severity**: `block` when the command breaks the contract, `warn` when the adapter works around a loss, or `advise` for a better practice
- **Owner**: who can fix it, the author of the extension or the author of the script
- **Fix**: the smallest change that removes the problem

End the review with one verdict:

- **Ready**: no finding has the severity `block`
- **Ready with stated limits**: no finding blocks, and the user accepts each `warn` as a limit. Write each accepted limit in the "Limits" part of its command in `spec.md`
- **Needs work**: at least one finding blocks

Never reject a script because it is small, young, or private. Judge what the command does, not where the script comes from.
