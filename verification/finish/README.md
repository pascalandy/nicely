# Finish gate model

`Finish.tla` models the finish gate of an extension repository, as [Finish gate](../../docs/north-star/contract/finish-gate.md) defines it.
It is the first model of core, as COR-104-T8 asks.
It asks one question: can `just finish` pass while a finding stays open?

`just verify-models` checks two configurations, as [verification/tlc.json](../tlc.json) lists them:

- `Current` must explore every state with no violation
- `PreFix` is the gate at the parent of 59d33f38, and it must violate `NoOpenPass`

At the parent of 59d33f38, a finding of the last round with no state added no problem.
So a review could add such a finding, write the verdict `Ready`, and `just finish` passed.
TLC finds that trace in four states: the initial state, the report of an `advise` with no state, the verdict `Ready`, and the pass.

## Code mapping

| Model | Code of `sdk/extensionrepo` |
|---|---|
| `Report`, `Present`, `born` | `findingProblems` of `readiness.go`, and the round of a finding ID in `stateProblems` of `ledger_findings.go` |
| `Restate`, `Closed`, `StateOK` | `findingProblems` refuses a finding with no state. `stateProblems` checks the evidence of each state |
| `Prove` | A fixed finding names its commit, and a family of `## Case matrix` with a test that passes and a test that refuses: `pairProblems` |
| `Accept` | A limit names a row under **Limits.** of `spec.md` that the user accepted: `acceptedLimit` of `readiness_verdict.go` |
| `Reason` | A declined finding is an `advise` with a reason |
| `Regrade` | The history check of `ledger_history.go`: the severity of a finding never goes down |
| `Review`, `VerdictOK` | `verdictProblems` of `readiness_verdict.go` |
| `Finish` | `finishRepository` of `finish.go` passes only with no problem |
| `NextRound`, `round` | The rounds of the ledger in `ledger.go` |
| `PreFix` | `stateProblems` at the parent of 59d33f38 |

## Bounds and limits

The model has two findings and two rounds, so a finding can come from this round or from an earlier one.
Each finding has a state, a severity, and the evidence of its state.

- `NoOpenPass`: when finish passes, each finding holds a state and its evidence
- `NoUnfixedBlockPass`: when finish passes, each `block` is fixed, even one that the user accepted as a limit
- `NoWorkPass`: the verdict `Needs work` never passes
- `TypeOK`: the domains of the variables

The model proves bounded safety only.
It assumes valid TOML, IDs, commits, and ancestry, and a `spec.md` with no open line.
It leaves out the parser, edited history, reset builds, several builds, and the limit of five rounds.
The record binds the model and its configurations, not the Go code, so a change of the gate needs a review of this mapping.
