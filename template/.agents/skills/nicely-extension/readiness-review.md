# Readiness review

This file says how an agent that did not write the code reviews the extension, and how it writes `review.toml`.

An agent that did not write the code runs this review at the end of each round, and when the user asks for it.
The author of the round never writes it. The checks of the contract apply in all cases, and `just check` runs them.

Review the extension against "What a great extension does" and against the facts of the intake.
Write the review in `review.toml`, a typed record with these first keys:

```toml
commit = "<the full SHA of the reviewed commit>"
reviewer = "<the agent that wrote this review, which wrote none of the code>"
verdict = "<Ready, Ready with stated limits, or Needs work>"
```

Then run `just records`, which writes `readiness.md` from the record. Never edit `readiness.md`.
Commit `review.toml` and `readiness.md` alone. `just finish` accepts the verdict while only they and the part **Limits.** of `spec.md` differ from that commit.
`build-records.md` in the contract gives each key, and `finish-gate.md` each check of `just finish`.

## Findings

Give each finding its own table `[[findings]]`, or none when no round found any.
Its `id` is such as `R2-F3`: `R2` is the round that first reported it, and `F3` its number in that round.
Keep each finding of an earlier round, with the same ID, and update its state for this round.
Give each finding a value for each of these keys:

- `title`: what goes wrong, in a few words
- `severity`: `block` when the command breaks the contract, `warn` when the adapter works around a loss, or `advise` for a better practice
- `state`: `open`, `fixed`, `limit`, or `declined`. A finding of this round is `open`
- `spec`: the spec ID of its line: the open line, the line that the test of the fix proves, or the accepted limit
- `family`: the family of `## Case matrix` of that line
- `commit`: for a `fixed` finding, the commit that fixed it
- `reason`: for a `declined` finding, the reason of the author

These keys hold prose: `evidence`, the file, line, scenario, or output that shows the problem. `command`, the command that it affects. `consequence`, what goes wrong for a user or an agent. `owner`, who can fix it. `fix`, the smallest change that removes the problem.

`just finish` reads a grade only in `severity`, and a state only in `state`. It fails on a key that the record does not define, such as `grade`.
A word in prose grades nothing, so "a block for me" in `evidence` is no grade. Give each grade in `severity`, before the verdict.

A loss that a user can see is `warn` or `block`, never `advise`.
Examples are data that the output keeps or changes, and a promise of `spec.md` that does not hold.
When in doubt between `warn` and `advise`, use `warn`.
A broken promise of `spec.md` or of the help text is at least `warn` at its first review. This applies also when no data is lost.
Grade it so in the round that first reports it. These examples set the grade:

- **A message that names a wrong cause.** `inspect` answers that `notes.md` is missing, but the file exists and is not readable. At least `warn`, because the agent fixes the wrong thing
- **A dry run that predicts a result that the run does not give.** The dry run says that the output is `notes.txt`, and the run writes `notes.md.txt`. At least `warn`. `block` when the run replaces a file that the dry run said it keeps
- **An output whose meaning changes.** `notes.txt` drops the line `encoding: latin-1` and keeps the bytes, so a reader shows other letters. At least `warn`. `block` when `spec.md` promises the same text

Never lower the severity of a finding in a later round.
Give a finding the state `fixed` only when its repair adds a pair of tests for its whole family, as `## Case matrix` lists it (#423).
`just finish` needs a line that passes and a line that refuses in that family.

A `warn` or `block` finding closes only as `fixed` or `limit`. Only an `advise` can be `declined`.
A change of `spec.md` that narrows a promise to close a `warn` or `block` finding is a `limit`. Only the user can accept it.
`None.` under `## Open lines` does not close a finding that the ledger records as `open`.
The `spec` of each `open` finding is a row under `## Open lines` of `spec.md`. `just finish` fails without it.

End the review with one verdict. The verdict follows from the findings, and `just finish` checks it:

- **Ready**: each `block` or `warn` finding is `fixed`, and `spec.md` states no limit
- **Ready with stated limits**: each `block` finding is `fixed`, and each `warn` is `fixed` or a `limit` that the user accepted
- **Needs work**: any other case

Only the user accepts a limit. Ask the user, and wait for the answer.
Then write the limit as a row under **Limits.** of its command, with the columns ID, State, Limit, and Accepted.
An open finding moves its row from `## Open lines` to that table, with the same ID.
This change starts no round: the reviewer gives the finding the state `limit` in the review of the same commit.
Commit the limit in `spec.md`, then commit `review.toml` and `readiness.md` alone.
Its State is `limit`. Accepted records who accepted it, when, and where, such as a review comment.
A loss that you report but that the user did not accept is a `warn` finding, not a limit.

## Round 1

In round 1, review the order of the work, not only the code.
First, before you review any repair, check `## Case matrix` of `spec.md` (#423):

- Each generic case that applies, and each family that the intake found, has its row
- Each refusal and each branch that classifies an input has a pair of tests: one input passes, and one input is refused
- Run the tests of the matrix. A family that the matrix lacks is a finding
- `just finish` checks only that each family names a test. It does not check that the matrix is complete, or that each test proves its family

Then review the rest of round 1:

- The intake commit comes before the spec commit. `just finish` checks the keys of `intake.toml`: each `sourced` fact gives its source, its revision, and the part read
- Refuse a fact whose part read does not support it, such as a link or a file name. Such a fact is an `inference`, or it is a finding
- Each `unknown` fact is an open line of `spec.md` with its ID, in the spec commit
- The first failing run gives the scenario, its exit status, and the assertion that failed at the scenario commit. Run `just scenario <name>` again at that commit when in doubt, and `just signoff` runs it again too. Refuse a failure that is only a compile error, because it does not prove the missing behavior
- Each case of the Output group of `generic-cases.md` that applies has a test file that the tests make, and an independent check. A case name alone is not a test
- Test families of cases, such as all the declarations that change the meaning, not one counterexample

## Ledger

After the findings, keep the ledger of the build in `review.toml`. Add one table `[[rounds]]` for this round, and keep the tables of earlier rounds:

```toml
[[rounds]]
head = "<full SHA of the head of round 1>"
change_review = "<reviewer>"
change_commits = "<first>..<last>"
readiness_review = "<reviewer>"
finish = "fail"

[[rounds]]
head = "<full SHA of the head of this round>"
change_review = "<reviewer>"
change_commits = "<first>..<last>"
readiness_review = "<reviewer>"
finish = "pending"

[order]
intake = "<sha>"
spec = "<sha>"
scenario = "<sha>"
handler = "<sha>"

[first_run]
commit = "<the Scenario commit of [order]>"
scenario = "<the name of the first scenario, which ran at its commit>"
exit = 1
assertion = "<the line of the scenario that failed, as the output of the run names it>"
```

`change_commits` names the commits that the change review covered. The readiness review covers `head`.
When one pass did both reviews, name the same reviewer in both keys.
`finish` holds the result of `just finish` for each earlier round, and `pending` for this round.
Keep each earlier round as it was. Only its `finish` changes, from `pending` to its result.
Commit `review.toml` and `readiness.md` at the end of each round, so each round has a committed review of its head.
A build has at most 5 rounds.
A round whose `finish` is `pass` ends a build. A new build, such as of another command, starts at a new intake after it.
The round numbers go on, and the ledger keeps each finding of each build.
After `just update` from a template with a Markdown `readiness.md`, start a new build at a new intake commit. Git keeps the Markdown reviews as history.

Never reject a script because it is small, young, or private. Judge what the command does, not where the script comes from.
