# Finish gate

This part says what `just finish` checks in an extension repository, as [Extension repository](extension-repository.md) defines it.
`just finish` is the exit of the loop of the `nicely-extension` skill. `just check` stays green for an incomplete round.
It reads the intake and the reviews only in their typed records, which [Build records](build-records.md) defines.

`just finish` fails on each of these:

- A missing `vision.md`, `spec.md`, `qa.md`, or `intake.toml`
- A line of `vision.md`, `spec.md`, or `qa.md` that starts with `Fill:` after its indent, also in fenced code. Also a line of `intake.toml` that starts with `# Fill:`
- A record that is not TOML, or that holds a key that the record does not define. A view that differs from what `just records` writes
- A line under `## Open lines`
- A line under **Limits.** that is no table row with a spec ID, the state `limit`, and the user's acceptance
- No table under `## Case matrix` of `spec.md` with the columns Family, Source, Accepted, and Refused (#423). Also a family that names no spec ID in Accepted, or a Refused cell with no spec ID and not `None`
- A spec ID of `## Case matrix` that is no line of Failures or Acceptance and no accepted limit, so no test proves it. Also a line of Failures or Acceptance that no family names
- A fact with no statement, or a kind other than `sourced`, `inference`, or `unknown`. A sourced fact without its source, revision, and read part. An inference or an unknown with a source
- An unknown fact without a spec ID, or whose spec ID is no open line, no line of Failures or Acceptance, and no accepted limit
- An experiment without its command or its result
- No `review.toml`, or one without a full `commit`, a `reviewer`, or the verdict `Ready` or `Ready with stated limits`
- A reviewed commit that differs from the working tree outside `readiness.md` and `review.toml`, and outside the part **Limits.** of `spec.md`. So the user accepts a limit after a review, with no new round. A row of `## Open lines` whose ID became a row of **Limits.** may leave too
- `Ready` with a `block` or `warn` finding that is not `fixed`, or with a limit in `spec.md`
- `Ready with stated limits` with a `block` finding that is not `fixed`, or with a `warn` that is neither `fixed` nor an accepted `limit`

The check of `## Case matrix` reads the references only. It cannot prove that the matrix is complete or correct.
Finish reads a severity only in the field `severity` of a finding. The words block, warn, and advise in prose grade nothing (#419).

## The ledger

`review.toml` also holds the ledger of the repository (#417).
A build of a command goes from its intake to "Done when".
The ledger keeps each round and each finding of each build.
The reviewer of each round updates it, runs `just records`, and commits `review.toml` and `readiness.md` alone.

Each `[[rounds]]` table is one round. Its place gives its number, from 1. The numbers go on from one build to the next.
`head` is the full SHA of the reviewed commit.
`change_commits` names the commits that the change review covered, as `<first>..<last>`.
The readiness review covers `head`.
When one pass did both reviews, both reviewers are the same.
`finish` is `pass` or `fail` for an earlier round, and `pending` for the last round, whose result is this run.
A round before the last whose `finish` is `pass` ends a build. The next round starts the next build.

`[order]` names the commit of each step of round 1 of the current build: `intake`, `spec`, `scenario`, and `handler`.

`[first_run]` records the run of the first scenario of the current build at its commit, before the handler (#420):

- `commit`: the Scenario commit of `[order]`, so a record of an earlier build never stands in
- `scenario`: the name of a scenario `testdata/script/<name>.txtar` that the Scenario commit holds, which `just scenario <name>` ran (#422)
- `exit`: its exit status, which is not 0
- `assertion`: the assertion of the scenario that failed. It holds the text of that line of the scenario

The reviewer refuses a failure that is only a compile error, because it does not prove the missing behavior.

### The first failing run again

`just signoff` runs `just finish`, then the first failing run of `review.toml` again (#422):

1. It reads the name of the scenario from `scenario`, and refuses a name that is no file name. It never runs the text of the record
2. It checks out the Scenario commit in a new scratch worktree, outside the repository
3. It runs that scenario alone with `go test`, as `just scenario` does
4. The run must fail on a line of the scenario, and the recorded assertion must hold the text of that line

Signoff signs nothing when the run passes, or fails before any line of the scenario, such as on a compile error.
The receipt of signoff records the run under `first_run`.

### Findings

The reviewer of a later round gives each earlier finding a new state when it changes:

| State | Meaning | Needs |
|---|---|---|
| `open` | Work remains | `spec`: the ID of a row under `## Open lines` |
| `fixed` | A pair of tests proves the fix | `commit`: the fixing commit. `spec`: a line of Failures or Acceptance in its `family`. That family names a line that passes and a line that refuses |
| `limit` | The user accepted the loss | `spec`: the ID of an accepted row under **Limits.** |
| `declined` | The author does not fix it | `reason`. Only an `advise` can be declined |

`just finish` fails on each of these:

- No round, or a round without a full `head`, both reviewers, or its `finish`
- A build of more than 5 rounds, also a build that ended. The message tells the agent to stop and to report to the user
- A round with the `finish` `pass` whose newest committed review would fail `just finish` as the repository stood then. Its forms hold no `Fill:` line and no open line, and only the records, their views, and **Limits.** of `spec.md` differ from its head. Its verdict and states follow from its findings, which keep each finding of the reviews before it
- An Intake that is not after the head of the round that ended the build before
- An Intake that is not after the last commit of a Markdown `readiness.md`, as [Build records](build-records.md#markdown-records) says
- An Intake that is not after the newest commit that added `intake.toml`, whose facts are the example of the template, as [Update an extension](update-an-extension.md#adopt-just-finish) says
- A commit of a round or of a fixed finding that is not HEAD or an ancestor of HEAD
- A last round whose `head` is not the `commit` of `review.toml`
- A step of round 1 that is missing, out of order, or in the commit of the step before it
- An unknown fact of the intake at the Intake commit whose spec ID is no row under `## Open lines` at the Spec commit
- No `[first_run]`, or one without the Scenario commit, or without a scenario of that commit. Also an exit status of 0, or an assertion that holds no line of that scenario
- A finding without its ID, title, severity, state, spec, or family, or a state without what it needs. Two findings with one ID
- A finding with the state `open`. Its `spec` must also be a row under `## Open lines` of `spec.md`, so `spec.md` and the ledger agree on the open work (#420, #423)
- A `warn` or `block` finding with the state `declined`
- A finding of an earlier committed `review.toml` that is missing, or whose severity went down
- A committed review whose `commit` is the head of no round, or a round before the last without a committed review of its head
- A round of an earlier committed ledger that is missing or changed. Only a `finish` that was `pending` changes

The history is every committed `review.toml` of HEAD, so a moved Intake hides no finding and resets no round.
Finish checks a review that ends a build with the `spec.md` of its own commit. A later `spec.md` before the next round stands in when only **Limits.** differs, so a limit committed after the review counts.
The rows of `## Open lines` whose IDs became rows of **Limits.** may leave too, so the user can accept an open finding at the limit of rounds.
A limit or a spec ID that a later build changes does not change that check.

At 5 rounds, a failing `just finish` also tells the agent to stop and to report each open line and each open finding.
