# Readiness review

This file says how an agent that did not write the code reviews the extension, and how it writes `readiness.md`.

An agent that did not write the code runs this review at the end of each round, and when the user asks for it.
The author of the round never writes it. The checks of the contract apply in all cases, and `just check` runs them.

Review the extension against "What a great extension does" and against the facts of the intake.
Write the review in `readiness.md`, which starts with these three lines:

```
Commit: <the full SHA of the reviewed commit>
Reviewer: <the agent that wrote this review, which wrote none of the code>
Verdict: <Ready, Ready with stated limits, or Needs work>
```

Then commit `readiness.md` alone. `just finish` accepts the verdict while only `readiness.md` and the part **Limits.** of `spec.md` differ from that commit.
`finish-gate.md` in the contract lists each check of `just finish`.

## Findings

Put the findings under `## Findings`, or write `None.` when no round found any.
Give each finding its own `###` heading, which starts with its ID, such as `### R2-F3 Long lines are cut`.
`R2` is the round that first reported it, and `F3` its number in that round.
Keep each finding of an earlier round, with the same ID and with a state for this round.
Write each part of a finding as a bullet that starts with its bold name, such as ``- **Severity**: `warn` ``:

- **Evidence**: the file and the line, the scenario, or the output that shows the problem
- **Command**: the command that the finding affects
- **Consequence**: what goes wrong for a user or an agent
- **Severity**: `block` when the command breaks the contract, `warn` when the adapter works around a loss, or `advise` for a better practice
- **State**: `open`, `fixed`, `limit`, or `declined`. A finding of this round can have no state yet
- **Spec**: the spec ID of its line: the open line, the line that the test of the fix proves, or the accepted limit
- **Commit**: for a `fixed` finding, the commit that fixed it
- **Reason**: for a `declined` finding, the reason of the author
- **Owner**: who can fix it, the author of the extension or the author of the script
- **Fix**: the smallest change that removes the problem

Name a severity only in the bullet **Severity** of its finding.
After the first three lines, `just finish` fails on the words block, warn, or advise anywhere else.
Examples are a table, fenced code, and prose.

A loss that a user can see is `warn` or `block`, never `advise`.
Examples are data that the output keeps or changes, and a promise of `spec.md` that does not hold.
When in doubt between `warn` and `advise`, use `warn`.
Never lower the severity of a finding in a later round.

A `warn` or `block` finding closes only as `fixed` or `limit`. Only an `advise` can be `declined`.
A change of `spec.md` that narrows a promise to close a `warn` or `block` finding is a `limit`. Only the user can accept it.
`None.` under `## Open lines` does not close a finding that the ledger records as `open`.

End the review with one verdict. The verdict follows from the findings, and `just finish` checks it:

- **Ready**: each `block` or `warn` finding is `fixed`, and `spec.md` states no limit
- **Ready with stated limits**: each `block` finding is `fixed`, and each `warn` is `fixed` or a `limit` that the user accepted
- **Needs work**: any other case

Only the user accepts a limit. Ask the user, and wait for the answer.
Then write the limit as a row under **Limits.** of its command, with the columns ID, State, Limit, and Accepted.
This change starts no round: the reviewer gives the finding the state `limit` in the review of the same commit.
Commit the limit in `spec.md`, then commit `readiness.md` alone.
Its State is `limit`. Accepted records who accepted it, when, and where, such as a review comment.
A loss that you report but that the user did not accept is a `warn` finding, not a limit.

## Ledger

After the findings, keep the ledger of the build. Add one row for this round, and keep the rows of earlier rounds:

```
## Rounds

| Round | Head | Change review | Change commits | Readiness review | Finish |
|---|---|---|---|---|---|
| 1 | <full SHA of the head of round 1> | <reviewer> | <first>..<last> | <reviewer> | `fail` |
| 2 | <full SHA of the head of this round> | <reviewer> | <first>..<last> | <reviewer> | `pending` |

## Order of round 1

| Step | Commit |
|---|---|
| Intake | <sha> |
| Spec | <sha> |
| Scenario | <sha> |
| Handler | <sha> |
```

Change commits names the commits that the change review covered. The readiness review covers Head.
When one pass did both reviews, name the same reviewer in both cells.
Finish holds the result of `just finish` for each earlier round, and `pending` for this round.
Keep the row of each earlier round as it was. Only its Finish changes, from `pending` to its result.
Commit `readiness.md` at the end of each round, so each round has a committed review of its head.
A build has at most 5 rounds.
A round whose Finish is `pass` ends a build. A new build, such as of another command, starts at a new intake after it.
The round numbers go on, and the ledger keeps each finding of each build.
After `just update` from a template without a ledger, keep each finding of the older reviews, with an ID and a state.

Never reject a script because it is small, young, or private. Judge what the command does, not where the script comes from.
