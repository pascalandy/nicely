---
name: nicely-extension
description: Use for each command of this Nicely extension, from the first one, when wrapping a script or a program, when changing a command, or when the user asks whether the extension is ready.
---

# Nicely extension

This skill takes each command from intake to a verified command, in rounds of pull requests.
`AGENTS.md` gives the reading order, recipes, and extension root.
The domain names a family of work, such as `video`.
An optional name identifies one job, such as `convert`.
Neither names the wrapped tool.
Read "Extension identity" in the contract for the slug and command ownership.

The contract holds every rule. `just contract` prints its folder at the Nicely version that `go.mod` pins. Read "What a great extension does" in its `extensions.md` first. Then read only the parts that it links.

## Intake of a script

Get these facts before you write code. Read the source of the script, and ask its author what the source does not show. Do not run the script to learn what it does, because even its `--help` can write or bill.

- The program and its entry point, such as `python3 tool.py`, with its source and revision, for `vision.md`
- Its inputs: the arguments, stdin, files, environment variables, and the folder that it reads
- Its result: what proves that the work happened, on which stream, and in which format
- Its effects: each write, network call, and paid request, also in a dry run or in `--help`
- Its keys and logins, and where it reads them
- How long it runs, what it starts in the background, and whether it can ask a question
- Its exit codes, and what a repeat of the same call does

Write each fact that nobody can give as an open question in `spec.md`.

## Intake of a program

A third-party program, such as an encoder, has no author to ask.
Get the facts of the script intake from its documentation and its manual page.
Also read its exit codes, its limits, and its known failure modes, such as the open issues of its tracker.
Do not run a command that writes or bills to learn its behavior.
Write each fact that the documentation does not give as an open question in `spec.md`.

## Choice of operations

- A named root can take arguments and run, such as `ncly video convert <file>`
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

## Acceptance lines

The lines of Failures and Acceptance in `spec.md` are the acceptance criteria of a command.

1. For a command that reads or writes user files, start from the cases of `generic-cases.md` in the contract.
2. Add one line for each case of the program that the intake found, such as an exit code, a limit, or a known failure.
3. Give each line a spec ID and a state, as the opening of `spec.md` says: `open`, `proven`, or `limit`.

A case that cannot occur for the command needs no line.

## Rounds

Each round is one pull request.

### Round 1

1. Implement the main path of the command, with "Change a command" in `AGENTS.md`.
2. Write every line of the command in `spec.md`. Each line that has no test stays `open`, under `## Open lines`.
3. Fill in `acceptance.sh` from `## Acceptance` in `spec.md`. For a paid command, limit it to the dry run.
4. End the round.

The first implementation is incomplete by design. The open lines start the next round.

### Each next round

1. Select open lines.
2. Move each line from `## Open lines` into Failures or Acceptance.
3. For each line, write a scenario with the claim `# spec: <id>`. It fails first, for the missing behavior.
4. Change the handler until the scenario passes. The line is then `proven`.
5. End the round.

### End of a round

1. Run `just check` until every check passes.
2. For a command that needs a key, relay the grant hint to the human and wait.
3. Run `just acceptance` on the real input until its independent checks pass.
4. Ask an agent that did not write the round to review its pull request.
5. Run the readiness review, and record it in `readiness.md`.
6. Write each finding that a test can prove as an open line.
7. Stop as "Done when" in `AGENTS.md` says. Its `just finish` fails until the loop is done.
8. Otherwise, open the next pull request from the open lines and the other findings. When the previous pull request is not merged, base it on that branch.

A delivery skill or a merge gate, such as a "full stop" or "full send" request, runs after "Done when" holds.
It never replaces "Done when", and it never ends the rounds early.
The release receipt and the human items of `qa.md` come after the merge, as "Release" in `AGENTS.md` says.

For a command that replaces a user file or replays paid work, run `create-formal-verification-skill`.

## Readiness review

Run this review at the end of each round, and when the user asks for it. The checks of the contract apply in all cases, and `just check` runs them.

Review the extension against "What a great extension does" and against the facts of the intake.
Write the review in `readiness.md`, which starts with these two lines:

```
Commit: <the full SHA of the reviewed commit>
Verdict: <Ready, Ready with stated limits, or Needs work>
```

Then commit `readiness.md` alone. `just finish` accepts the verdict while only `readiness.md` differs from that commit.
Write each finding with these six parts:

- **Evidence**: the file and the line, the scenario, or the output that shows the problem
- **Command**: the command that the finding affects
- **Consequence**: what goes wrong for a user or an agent
- **Severity**: `block` when the command breaks the contract, `warn` when the adapter works around a loss, or `advise` for a better practice
- **Owner**: who can fix it, the author of the extension or the author of the script
- **Fix**: the smallest change that removes the problem

End the review with one verdict:

- **Ready**: no finding has the severity `block` or `warn`
- **Ready with stated limits**: no finding blocks, and the user accepted each `warn` as a `limit` line
- **Needs work**: at least one finding blocks, or a `warn` has no `limit` line that the user accepted

Only the user accepts a limit. Ask the user, and wait for the answer.
Then write the limit as a row under **Limits.** of its command, with the columns ID, State, Limit, and Accepted.
Its State is `limit`. Accepted records who accepted it, when, and where, such as a review comment.
A loss that you report but that the user did not accept is a `warn` finding, not a limit.

Never reject a script because it is small, young, or private. Judge what the command does, not where the script comes from.
