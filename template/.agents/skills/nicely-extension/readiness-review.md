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

Then commit `readiness.md` alone. `just finish` accepts the verdict while only `readiness.md` differs from that commit.
Put the findings under `## Findings`. Give each finding its own `###` heading, or write `None.` when the review finds none.
Write each part of a finding as a bullet that starts with its bold name, such as ``- **Severity**: `warn` ``:

- **Evidence**: the file and the line, the scenario, or the output that shows the problem
- **Command**: the command that the finding affects
- **Consequence**: what goes wrong for a user or an agent
- **Severity**: `block` when the command breaks the contract, `warn` when the adapter works around a loss, or `advise` for a better practice
- **Limit**: for a `warn` that the user accepted, the spec ID of its row under **Limits.**
- **Owner**: who can fix it, the author of the extension or the author of the script
- **Fix**: the smallest change that removes the problem

Name a severity only in the bullet **Severity** of its finding.
`just finish` fails on a severity in a table, in fenced code, or in prose under `## Findings`.

A loss that a user can see is `warn` or `block`, never `advise`.
Examples are data that the output keeps or changes, and a promise of `spec.md` that does not hold.
When in doubt between `warn` and `advise`, use `warn`.

End the review with one verdict. The verdict follows from the findings, and `just finish` checks it:

- **Ready**: no finding has the severity `block` or `warn`, and `spec.md` states no limit
- **Ready with stated limits**: no finding blocks, and each `warn` names in **Limit** a limit that the user accepted
- **Needs work**: at least one finding blocks, or a `warn` names no limit that the user accepted

Only the user accepts a limit. Ask the user, and wait for the answer.
Then write the limit as a row under **Limits.** of its command, with the columns ID, State, Limit, and Accepted.
Its State is `limit`. Accepted records who accepted it, when, and where, such as a review comment.
A loss that you report but that the user did not accept is a `warn` finding, not a limit.

Never reject a script because it is small, young, or private. Judge what the command does, not where the script comes from.
