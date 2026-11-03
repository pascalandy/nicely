# Rule index of the build loop

This part of the [rule index](../rule-index.md) holds the rules of the rounds of the `nicely-extension` skill.

| Rule | Enforced by |
|---|---|
| Round 1 records the intake, then the spec, then a failing scenario, then the handler | [Finish gate][finish] checks their commits in the [ledger][ledger] |
| The intake gives, for each fact, its source, its version, and the part read, with each experiment apart; a link alone is not evidence | The [intake record][intake] needs these keys; partial: the reviewer judges that the part read supports the fact |
| Round 1 records the Scenario commit, the scenario, the exit status, and the failed assertion of the first scenario run | The [ledger][ledger-run] checks the record |
| The first run names a scenario of its commit, and its assertion holds a line of that scenario | The [ledger][ledger-run] checks the record |
| `just signoff` runs the first scenario again at its commit, and never runs the text of the record | [Signoff replay][signoff-replay] |
| Read the source first, then run experiments only to confirm what you read | Prose on purpose: no record shows what the author understood before an experiment |
| Each settled fact of the intake links the part read, at a fixed revision; a fact without a read is marked `inference` | The [intake record][intake] gives each fact one kind, and only a `sourced` fact a source |
| Each unknown fact of the intake is an open line with an ID before the spec | The [intake record][intake] and the [ledger][ledger] check the spec ID at the Spec commit |
| The reviewer of round 1 refuses a fact that only cites a source | Prose on purpose: only a reader can judge that the cited part supports the fact |
| Before the first review, a test file holds a block that a reader needs, with a value that is not the default and private text | [Kit `reads`][reads] needs the private text, and [a case][reads-matrix] for each Output case of the matrix; partial: round 1 runs `just check` before its review. The reviewer judges the value |
| An independent reader proves that the private text is gone and the content shows correctly; a handler that keeps the raw block fails | [Kit `reads`][reads] searches the private text, compares the reference, and needs a bad variant that holds the private text |
| Each experiment uses a new scratch folder; never delete files recursively inside `sh -c` or a loop | Prose only |
| The tests and `just acceptance` need no root, no mount, and no system service | Prose only |
| Each case of the Output group that applies has a test file that the tests make, checked by an independent reader | [Kit Output cases][reads-matrix] for each case that `## Case matrix` names; partial: the reviewer checks that the matrix names each case that applies |
| The tests make their files; never search the home folder or copy a file of the user | [Kit `reads`][reads] reads only files of the folder of the scenario; partial: no isolation of the system hides the files of the user |
| Round 1 tests each generic case that applies, and the first review tests families of cases | Prose on purpose: no manifest lists every family of a domain, so only the reviewer finds a missing one |
| Before the review of round 1, `## Case matrix` gives each family its tests, and each line of Failures and Acceptance a family | [Finish matrix][finish-matrix] checks the references only, not that the matrix is complete or correct |
| Each refusal and each branch that classifies an input has a pair of tests | [Finish matrix][finish-matrix] needs a spec ID in Accepted of each family; the reviewer checks that it passes |
| The first reviewer checks that the matrix is complete and runs it, before any repair | Prose on purpose: a receipt cannot show that the matrix is complete, or that the reviewer judged it |
| A repair adds a pair of tests for its whole family, and only then is its finding `fixed` | [Ledger findings][ledger-findings] need a passing and a refused line in its family; partial: the reviewer judges that the family is whole |
| A root cause is fixed before the next review, and the last rounds verify | Prose on purpose: history orders the repairs, but cannot show that a cause is complete |
| Build each command with the `nicely-extension` skill, in rounds from its first pull request | Prose on purpose: a loaded skill does not show that the agent followed it, and the ledger checks the recorded parts |
| The first round writes every line of `spec.md`; a line without a test stays `open` | Prose on purpose: no machine finds a promise that every version of `spec.md` leaves out |
| Each round ends with an independent review and the readiness review | [Signoff][signoff] runs the [finish gate][finish], which needs a readiness review of the signed commit by a named reviewer |
| Finish reads the intake and the reviews only in `intake.toml` and `review.toml`; a key that a record does not define fails | [Build records][records] |
| `just records` writes `## Intake` of `vision.md` and `readiness.md`; a view that differs from its record fails | [Build records][records], [views][records-view] |
| A Markdown `readiness.md` of an older template is history, and the records start a new build after it | [Ledger][ledger], [ledger history][ledger-history] |
| After `just update` from a template before `just finish`, the first build starts after the update commit, and its intake reads the current `spec.md` | The [ledger][ledger] needs an Intake after the commit that added `intake.toml`; partial: the reviewer judges that the intake read the current `spec.md` |
| A failure that only Nicely can fix gets an issue in pascalandy/nicely and an accepted limit that links it, until `just update` brings the fix | Prose on purpose: only the agent can judge that the fix belongs to the SDK or the template, not to the extension |
| At the exit, filled forms, no open line, accepted limits, and a passing readiness verdict for this commit, by a named reviewer, from its findings | [Finish gate][finish] |
| A build has at most 5 rounds | The [ledger][ledger] counts them, and the [finish gate][finish] fails after 5 |
| At the limit, stop and report; an answer that lets the agent choose starts no round | Prose on purpose: the ledger refuses a sixth round, but only the agent stops, reports, and reads what the user allows |
| A pull request for review, a stop report, and the delivery are three acts; deliver only after "Done when" | [Signoff][signoff] runs the [finish gate][finish], and [release][release] needs a receipt with `just finish` |
| Release runs `just finish` again on the signed commit before the tag | [Release][release], [release receipt][release-receipt] |
| The ledger names the head, the reviewers, the commits that each review covered, and the result of `just finish` of each round | [Ledger][ledger] |
| Each finding keeps its ID and its highest severity, and gets a state each round; a `warn` or `block` closes only as `fixed` or `limit` | [Ledger findings][ledger-findings], [ledger history][ledger-history] |
| The spec ID of each finding that the ledger records as `open` is a row under `## Open lines` of `spec.md` | [Ledger findings][ledger-findings] |
| Each finding has one ID, title, severity, state, spec ID, and family; a finding without a state fails | [Readiness][readiness] |
| Finish reads a severity only in the key `severity` of a finding; the words block, warn, and advise in prose grade nothing | [Readiness][readiness], [build records][records] |
| A broken promise of `spec.md` or of the help text is at least `warn` at its first review | Prose on purpose: only a reader can decide that a behavior breaks a promise of the prose |
| Each round keeps its row and has a committed review; a build ends only at a review that passes | [Ledger history][ledger-history] |
| Accepting a limit starts no round; after a review, only **Limits.** of `spec.md` changes | [Finish gate][finish-verdict] |
| A change of `spec.md` that narrows a promise to close a `warn` or `block` is a limit that only the user accepts | Prose on purpose: a diff cannot decide that new words narrow a promise, or that the user agreed |

[finish]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/finish.go
[reads]: https://github.com/pascalandy/nicely/blob/main/sdk/extensiontest/acceptance.go
[reads-matrix]: https://github.com/pascalandy/nicely/blob/main/sdk/extensiontest/acceptance_matrix.go
[ledger]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger.go
[ledger-run]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger_run.go
[ledger-findings]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger_findings.go
[ledger-history]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/ledger_history.go
[finish-matrix]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/spec_matrix.go
[finish-verdict]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/readiness_verdict.go
[readiness]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/readiness.go
[intake]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/intake.go
[records]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/records.go
[records-view]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/records_view.go
[signoff]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/signoff.go
[signoff-replay]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/signoff_replay.go
[release]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/release.go
[release-receipt]: https://github.com/pascalandy/nicely/blob/main/sdk/extensionrepo/release_receipt.go
