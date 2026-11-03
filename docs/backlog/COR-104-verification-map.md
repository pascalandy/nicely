---
id: COR-104
core: "verification map"
state: Active. Core and template first, then video-convert
---

# COR-104 verification map

## Demo

`just status` counts prose-only rules, the ledger names the level and owner of each correction, and video-convert releases through the kit.

## Scope

[D068][d068] holds the map, the way a correction moves, and the measures.
Each card names its owner (A, B, or C) and level, and fails first on a past mistake.
The last merge of T1 to T9 releases a minor version with release-nicely, as this plan asks.
Trial findings are evidence of gaps only.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| COR-104-T1 | Rule indexes and ledger columns | agent | - | done |
| COR-104-T2 | SDK analyzers in every check | agent | - | done |
| COR-104-T3 | Owned verification runs | agent | COR-104-T2 | todo |
| COR-104-T4 | Typed build records | agent | - | todo |
| COR-104-T5 | Acceptance evaluator in the kit | agent | - | todo |
| COR-104-T6 | Publication declaration | agent | - | todo |
| COR-104-T7 | Publication helpers | agent | COR-104-T6 | todo |
| COR-104-T8 | TLC runner in the kit | agent | - | todo |
| COR-104-T9 | Adoption and shared corrections | agent | COR-104-T4 | todo |
| COR-104-T10 | Video-convert adopts the loop | agent | COR-104-T1, COR-104-T2, COR-104-T3, COR-104-T4, COR-104-T5, COR-104-T6, COR-104-T7, COR-104-T8, COR-104-T9 | todo |

### COR-104-T1 Rule indexes and ledger columns

- **Read:** [D068][d068], [N2][n2], [QA ledger](../qa/ledger.md)
- **Proves:** `tools/cmd/plan/rules_test.go`, `tools/cmd/plan/qa_test.go`

A, level 2, measures 1 and 2.
Core gets an index of the 154 rules of N2 in `docs/north-star/rules/`.
`just test` checks both indexes: each enforcer links a file, or says `Prose only` or `Prose on purpose`.
`just status` prints the `Prose only` count of each index, and `extensionrepo check` that of an extension.
The template index marks 11 judgment rules `Prose on purpose`, and deletes one rule that signoff replay enforces.
The QA ledger gets a Corrections table, with Level and Owner columns for each correction.
It fails first at 7f3ae60e, where `ledger_history.go` enforced rules that no row named until b3dc8b48.

### COR-104-T2 SDK analyzers in every check

- **Read:** [D068](../north-star/decisions/D068-verification-map.md), [P1][p1]
- **Proves:** `sdk/analysis/runnerdiscard/runnerdiscard_test.go`, `scripts/template-test.sh`

A, level 2, measure 2.
`sdk/analysis` holds Go analyzers, and `go vet -vettool` runs them.
Core's `just lint` and `extensionrepo check` run them, at the Nicely version that the extension pins.
The first rule refuses `io.Discard` in `Stdout` or `Stderr` of a `runner.Program`.
It fails first on the parent of 3e6a4eff, and on a rendered extension with that line.

### COR-104-T3 Owned verification runs

- **Read:** [D068](../north-star/decisions/D068-verification-map.md), [N1][n1]
- **Proves:** `sdk/extensiontest/isolation_test.go`, `scripts/template-recipes-test.sh`

A, level 1, measure 2, for 26 corrections of caller state and lost notices.
One SDK constructor builds the environment and scratch folder of the kit, signoff, release, and acceptance.
One result carries the answer, notices, and exit code to the caller.
An analyzer rule refuses `os.Environ` in these packages outside that constructor. Replacing `HOME` alone is no isolation.
It fails first on the parent of 09809608, which passed `os.Environ()`, and on the lost notices of e35bd0bb.

### COR-104-T4 Typed build records

- **Read:** [D068][d068], [Finish gate](../north-star/contract/finish-gate.md), [N2][n2]
- **Proves:** `scripts/template-finish-test.sh`, `scripts/template-ledger-test.sh`

A, level 1, measures 1 and 2, for 17 corrections of review prose and five template rules.
Intake facts and review findings become typed records that the SDK reads.
A fact is sourced, an inference, or unknown, and an unknown fact opens a line.
A finding has one ID, severity, state, spec ID, and family. `fixed` needs a passing and a refused test of its family.
The SDK writes `## Intake` and `readiness.md` from them, and finish reads only the records.
An extension with Markdown records starts a new build when it updates, and keeps them as history.
It fails first on the parent of 59d33f38, where a finding with no state passed finish, and on the free-form grades of #422.

### COR-104-T5 Acceptance evaluator in the kit

- **Read:** [D068][d068], [Test kit](../north-star/contract/test-kit.md), [N2][n2]
- **Proves:** `sdk/extensiontest/acceptance_test.go`, `scripts/template-test.sh`

A, level 2, measures 1 and 2, for five template rules.
The extension declares each acceptance case: a made input, a reader other than the wrapped program, a reference, and any private marker.
The kit reads the output with that reader, compares it with the reference, and searches the marker.
It requires the declared bad variant to fail, and a case for each Output case that applies by the first review.
It fails first on the parent of 86d24e83, where a prefix match accepted a larger number, and on the private text of #425.

### COR-104-T6 Publication declaration

- **Read:** [D068][d068], [Generic cases](../north-star/contract/generic-cases.md), [N3][n3]
- **Proves:** `sdk/extensiontest/publication_test.go`, `scripts/template-test.sh`

A, level 2, measure 2.
The `spec.md` form declares each command that creates or replaces a user file.
It gives the destination, protected paths, operation of each mode, needed file system abilities, and dry-run treatment.
The kit requires an executed check for each declared obligation, or a reason.
A successful dry run changes no protected path, and predicts the destination and overwrite choice of the run.
It fails first on #425: a dry run passed, and the run failed on a destination with no hard link.

### COR-104-T7 Publication helpers

- **Read:** [D068][d068], [Scenario commands](../north-star/contract/scenario-commands.md), [N3][n3]
- **Proves:** `sdk/extensiontest/paired_test.go`

A, level 2, measure 2.
A helper protects the declared paths around any call, and checks children before its own cleanup.
Another holds a worker after a partial staged write, then observes the public path and the commit.
A Linux fixture refuses link, no-replace rename, or replacing rename for the child only, and names a platform that cannot.
`replace-a-file.md` covers every writer, private exclusive staging, the permission policy, and failure before and after commit.
It fails first on a fallback that checks the output, then renames over it, which generic cases forbid ([N3][n3]).

### COR-104-T8 TLC runner in the kit

- **Read:** [D068](../north-star/decisions/D068-verification-map.md), [P2][p2]
- **Proves:** `sdk/verification/verification_test.go`, `verification/finish/Finish.tla`

A, level 2, measure 2.
`sdk/verification` gets TLC, checks its SHA-256 before Java starts, runs model suites, and records only matching outcomes.
It accepts a local jar with that digest. Changed bytes, as on 2026-11-02, never prove a model, and a new pin needs a review.
The template gets one optional `just verify-models` recipe, and `just check` reads only the record.
Core's first model, its finish gate, fails first on the behavior before 59d33f38.

### COR-104-T9 Adoption and shared corrections

- **Read:** [D068][d068], [Update an extension](../north-star/contract/update-an-extension.md)
- **Proves:** `scripts/template-update-test.sh`, `scripts/template-test.sh`

B, level 2, measure 3.
As #427 asks, the contract and the template skill say how an existing extension adopts `just finish`.
Its first build starts at the update commit, and its intake reads its current `spec.md`.
The update test renders an extension at v0.23.4, updates it, runs that first build, and requires `just signoff` to pass.
The skill also teaches the step of D068 for a correction of A or B, as prose on purpose.
It fails first on #427, and the step on #407, where an extension could not fix answers of `sdk/wrap`.

### COR-104-T10 Video-convert adopts the loop

- **Read:** [D068](../north-star/decisions/D068-verification-map.md), [P2][p2]
- **Proves:** readiness.md and the release receipt, in ncly-video-convert

C, level 2, measures 2 and 3.
In ncly-video-convert, update to that release and run the first build of #427.
Its publication proof runs through `sdk/verification`, and keeps its model, mapping, probes, and expected results.
It declares its publications and acceptance cases.
Each correction gets a ledger row, and a correction of A or B gets the step of D068.
It fails first on its record, which rejects the new pin ([P1][p1]), and on #427. Done when `just signoff` and `just release` pass.

## After

Pascal runs a UAT of video-convert at the release of COR-104-T10, and its corrections enter the ledger.

## Later

### Unprivileged scratch runs

Tests, acceptance, and reviews run in a new scratch folder, with no root, mount, or system service.
Two template rules stay `Prose only` until then.

Revisit condition: a review stalls on a recursive delete or mounts a file system again, as in #425.

### Process lifetime in the runner

`sdk/runner` makes the output of each program an explicit choice, and the COR-104-T2 rule on `io.Discard` goes.

Revisit condition: a new correction of exit, signal, drain, or cleanup, class K04 of [N1][n1].

[d068]: ../north-star/decisions/D068-verification-map.md
[n1]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/analysis/N1-correct.md
[n2]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/analysis/N2-prose-triage.md
[n3]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/analysis/N3-publication.md
[p1]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/proto/P1/REPORT.md
[p2]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/proto/P2/REPORT.md
