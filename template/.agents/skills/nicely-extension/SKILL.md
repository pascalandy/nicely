---
name: nicely-extension
description: Use for each command of this Nicely extension, from the first one, when wrapping a script or a program, when changing a command, or when the user asks whether the extension is ready.
---

# Nicely extension

This skill takes each command from intake to a verified command, in rounds.
`AGENTS.md` gives the reading order, recipes, and extension root.
The domain names a family of work, such as `video`.
An optional name identifies one job, such as `convert`.
Neither names the wrapped tool.
Read "Extension identity" in the contract for the slug and command ownership.

The contract holds every rule. `just contract` prints its folder at the Nicely version that `go.mod` pins. Read "What a great extension does" in its `extensions.md` first. Then read only the parts that it links.

## Intake

Read [intake.md](intake.md) before you write code. It gives the facts to get from a script or from a program.

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
A case name is not a test. Each case of the Output group that applies gets a test file in `acceptance.sh`, as "Acceptance test files" of `test-kit.md` in the contract says.

## Test files

The tests of the extension use files that the tests make, such as a text file that a scenario or `acceptance.sh` writes.
Do not search the home folder of the user for sample files. Never copy a file of the user into a test.

## Rounds

A build of a command starts at the intake and ends at "Done when" in `AGENTS.md`.
It has at most 5 rounds, and the ledger in `readiness.md` counts them.
A round is new commits on the open pull request, or a new pull request.
Each round ends with the independent review, the readiness review, and `just finish`.

### Round 1

Do these steps in this order. Commit each of the first four steps on its own:

1. Record the intake facts under `## Intake` in `vision.md`, as "Evidence of the intake" in `intake.md` says. The facts include the effects, the exit codes, the resource use, and the known failures. A fact that no source gives stays an open line of `spec.md`.
2. Write every line of the command in `spec.md`, from the generic cases of the contract and the intake. Each line that has no test stays `open`, under `## Open lines`. Write the table under `## Case matrix` too, as "Case matrix" says.
3. Write the first scenario at the public command boundary, with its claim `# spec: <id>`. Commit it while it fails. Run it at that commit with `just scenario <name>`. Keep that command, its exit status, and the line of the scenario that failed. A compile error is no such failure. `just signoff` runs it again. When the scenario cannot run without the command, commit a handler that does no work in this step. Then the run reaches the assertion.
4. Change the handler until that scenario passes, with "Change a command" in `AGENTS.md`.
5. When no line stays open, write `None.` under `## Open lines`, with a sentence that cites the generic cases.
6. Replace every `Fill:` line of `vision.md`, `spec.md`, and `qa.md`.
7. Fill in `acceptance.sh` from `## Acceptance` in `spec.md`, with a test file for each generic case of the Output group that applies. For a paid command, limit it to the dry run.
8. End the round.

Give the commits of steps 1 to 4 to the reviewer, who records them under `## Order of round 1` in the ledger.
Give the reviewer the record of the run of step 3 too. The reviewer checks it, and records it under `## First failing run`.
The first round is incomplete by design. The open lines start the next round.

### Case matrix

Before the review of round 1, `## Case matrix` of `spec.md` lists each family of cases of each command (#423):

- One row for each generic case that applies, and for each family that the intake found
- A family groups the inputs that one rule reads, such as each way that the input declares how to read it
- Each refusal and each branch that classifies an input has a pair of tests: one input passes, and one input is refused
- A repair of a finding adds a pair of tests for its whole family, not only a test of the reported example

`just finish` fails when a family names no test, or when a line of Failures or Acceptance is in no family.

### Each next round

1. Select open lines, and the findings that the ledger records as `open`.
2. Move each line from `## Open lines` into Failures or Acceptance.
3. For each line, write a scenario with the claim `# spec: <id>`. It fails first, for the missing behavior. For a finding, test its whole family with a pair, as "Case matrix" says.
4. Change the handler until the scenario passes. The line is then `proven`.
5. End the round.

### End of a round

1. Run `just check` until every check passes.
2. For a command that needs a key, relay the grant hint to the human and wait.
3. Run `just acceptance` on the real input until its independent checks pass.
4. Ask an agent that did not write the round to review the changes of the round. Give it the commits of the round.
5. Ask an agent that did not write the code to run the readiness review. It records the review and the ledger in `readiness.md`. One independent pass can do both reviews.
6. Write each finding that a test can prove as an open line, which names the finding by its ID. Each finding that the ledger records as `open` keeps its line under `## Open lines`, so `spec.md` and the ledger agree. A finding of this round with no state is `open` too.
7. Run `just finish`. Stop as "Done when" in `AGENTS.md` says. `just finish` fails until the loop is done.
8. When `just finish` fails after round 5, stop, as "Round limit" says.
9. Otherwise, start the next round from the open lines and the open findings. Add its commits to the open pull request, or open a new one. Base a new pull request on the branch of the previous one when that one is not merged.

### Reach "Done when" in 5 rounds

- Round 1 includes a test file for each generic case that applies
- Ask the first reviewer to check that `## Case matrix` is complete and to run it, before it reviews any repair
- Ask the first reviewer to test families of cases, such as all the declarations that change the meaning, not one counterexample
- Fix a root cause, such as a parser or a conversion rule, before the next review
- Keep the last rounds for verification, not for one patch per warning

### Round limit

At the limit of 5 rounds, stop. Do not start round 6.
Report to the user each open line of `spec.md`, and each finding that the ledger records as `open`, with its ID.
Then wait for the user.
An answer that lets you choose, such as "decide yourself", accepts no limit and starts no round.
The user can accept each open finding as a limit. That needs no new round, as `readiness-review.md` says.

### Review, stop, and delivery

These are three different acts:

- A pull request opens for review. It can open in round 1, before the loop is done
- A stop report goes to the user at the round limit. It delivers nothing
- The delivery, such as a merge or a release, starts only after `just finish` passes on the head

A delivery skill or a merge gate, such as a "full stop" or "full send" request, is a delivery.
It never replaces "Done when", and it never ends the rounds early.
The release receipt and the human items of `qa.md` come after the merge, as "Release" in `AGENTS.md` says.

### Commands that replace a file

For a command that replaces a user file, read [replace-a-file.md](replace-a-file.md). It lists the invariants to prove, and the commands of the test kit that prove them.

## Readiness review

An agent that did not write the code runs the readiness review at the end of each round, and when the user asks for it.
Read [readiness-review.md](readiness-review.md) for the review and the form of `readiness.md`.
