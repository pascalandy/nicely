# Finish gate

This part says what `just finish` checks in an extension repository, as [Extension repository](extension-repository.md) defines it.
`just finish` is the exit of the loop of the `nicely-extension` skill. `just check` stays green for an incomplete round.

`just finish` fails on each of these:

- A missing `vision.md`, `spec.md`, or `qa.md`
- A line of `vision.md`, `spec.md`, or `qa.md` that starts with `Fill:` after its indent, also in fenced code
- A line under `## Open lines`
- A line under **Limits.** that is no table row with a spec ID, the state `limit`, and the user's acceptance
- A `readiness.md` whose first three lines hold no `Commit: <sha>`, no `Reviewer: <who>`, or no verdict `Ready` or `Ready with stated limits`. A byte order mark before them is ignored
- A reviewed commit that differs from the working tree outside `readiness.md`, and outside the part **Limits.** of `spec.md`. So the user accepts a limit after a review, with no new round
- A finding of `readiness.md` without its own `###` heading under `## Findings`, or without one severity: `block`, `warn`, or `advise`
- A severity after the first three lines, outside the line `- **Severity**:` of a finding, such as in a table, in fenced code, or in prose. Under `## Findings`, the word severity fails too. The message names the word or the table that caused it
- `Ready` with a `block` or `warn` finding that is not `fixed`, or with a limit in `spec.md`
- `Ready with stated limits` with a `block` finding that is not `fixed`, or with a `warn` that is neither `fixed` nor an accepted `limit`

## The ledger

`readiness.md` also holds the ledger of the repository (#417).
A build of a command goes from its intake to "Done when".
The ledger keeps each round and each finding of each build.
The reviewer of each round updates it, and commits `readiness.md` alone.

Under `## Rounds`, a table has the columns Round, Head, Change review, Change commits, Readiness review, and Finish.
Each row is one round, numbered from 1. The numbers go on from one build to the next.
Head is the full SHA of the reviewed commit.
Change commits names the commits that the change review covered, as `<first>..<last>`.
The readiness review covers Head.
When one pass did both reviews, both cells name the same reviewer.
Finish is `pass` or `fail` for an earlier round, and `pending` for the last round, whose result is this run.
A round before the last whose Finish is `pass` ends a build. The next round starts the next build.

Under `## Order of round 1`, a table with the columns Step and Commit names the commit of each step of the current build: Intake, Spec, Scenario, and Handler.

Each finding under `## Findings` starts its heading with a stable ID, such as `### R2-F3`.
`R2` is the round that first reported it, and the ID stays in later rounds.
The reviewer of a later round gives each earlier finding one state, in a line `- **State**:`:

| State | Meaning | Parts |
|---|---|---|
| `open` | Work remains | `- **Spec**:` with its spec ID |
| `fixed` | A test proves the fix | `- **Commit**:` with the fixing commit, and `- **Spec**:` with a line of Failures or Acceptance |
| `limit` | The user accepted the loss | `- **Spec**:` with the ID of an accepted row under **Limits.** |
| `declined` | The author does not fix it | `- **Reason**:` with the reason of the author. Only an `advise` can be declined |

A finding of the last round can have no state yet.

`just finish` fails on each of these:

- No round under `## Rounds`, or a row without its six cells, its number, a full Head, or both reviewers
- A line under `## Rounds` or `## Order of round 1` that is no row of its table, such as a label that would hide rows
- A build of more than 5 rounds, also a build that ended. The message tells the agent to stop and to report to the user
- A round with the Finish `pass` whose newest committed review does not pass. That review needs a passing verdict that follows from its findings, valid states, and each finding of the reviews before it
- An Intake that is not after the Head of the round that ended the build before
- A commit of a round or of a fixed finding that is not HEAD or an ancestor of HEAD
- A last round whose Head is not the `Commit:` of `readiness.md`
- A step of round 1 that is missing, out of order, or in the commit of the step before it
- A finding without an ID, a finding of an earlier round without a state, or a state without its parts
- A finding with the state `open`, even when `## Open lines` holds `None.`
- A `warn` or `block` finding with the state `declined`
- A finding of an earlier committed `readiness.md` that is missing, or whose severity went down. A finding without an ID keeps its title
- A committed review whose `Commit:` is the Head of no round, or a round before the last without a committed review of its Head
- A round of an earlier committed ledger that is missing or changed. Only a Finish that was `pending` changes

The history is every committed `readiness.md` of HEAD, so a moved Intake hides no finding and resets no round.
Finish checks a review that ends a build with the limits of the current `spec.md`. A finding with the state `limit` keeps its limit row, so the check holds after a limit that the user accepted later.
A review without a ledger, from a template before the ledger, needs a passing verdict only to end a build.
Its form of #409, a line `- **Limit**:` with a spec ID, reads as the state `limit`.

At 5 rounds, a failing `just finish` also tells the agent to stop and to report each open line and each open finding.
