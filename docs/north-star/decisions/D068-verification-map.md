# D068 Map each check to an owner and a level

Decided 2026-11-02 by Pascal, for [COR-104](../../backlog/COR-104-verification-map.md).
It extends [D067](D067-share-verification-rules.md) and keeps its line: evaluators in the SDK and kit, facts in each extension, only wiring and forms in the template.

## Owners and levels

Three owners verify Nicely:

- **A, core**: the SDK, the test kit `sdk/extensiontest`, the repository gates `sdk/extensionrepo`, CI, and the `nicely` skill
- **B, the template**: `template/`, with its wiring, its forms, the `nicely-extension` skill, and its rule index
- **C, each extension**: its code, its facts, its tests, and its own skill. ncly-video-convert is the first

Four verification levels form a pyramid, with level 1 at the top:

1. Code: the architecture, types, and invariants that make a wrong state impossible to write
2. Checks: lint, analyzers, tests, scenarios, acceptance, CI, and formal models
3. Prose: rules, skills, and `AGENTS.md`
4. QA by an agent that did not write the change, and UAT by a human

These levels are not the layers of the [code map](../architecture.md#code-map).

## The map

| Level | A, core | B, template | C, each extension |
|---|---|---|---|
| 1. Code | SDK types with one owner for each invariant, such as `sdk/runner`, `sdk/record`, `sdk/wrap`, and `sdk/contract` | Rendered entry points that only call the SDK | Domain code: shared preparation, input refusal, and its own publication |
| 2. Checks | Kit checks and helpers, repository gates, Go analyzers, the TLC runner, `just check`, and core CI | One-line recipes, the CI workflow, and the forms that the gates read | Scenarios, Go tests, `acceptance.sh`, declared facts, and its TLA+ model with its mapping |
| 3. Prose | `AGENTS.md`, the decisions, the [core rule index](../rules/README.md), and the `nicely` skill | Its `AGENTS.md`, the rule index, and the `nicely-extension` skill | Its `AGENTS.md`, its skill, and **Limits.** in `spec.md` |
| 4. QA and UAT | verify-nicely by another agent, the UAT protocol, and the [UAT ledger](../../qa/ledger.md) | The readiness review by another agent, and the `qa.md` form | `just acceptance` on real input, its UAT, and the acceptance of Pascal |

Each extension keeps its publication logic: staging, link, rename, and replace.
A and B give the checks, helpers, fixtures, forms, and skill lines that prove it.
The TLC runner is a kit tool.
Each extension keeps its TLA+ model, its mapping to its code, its probes, and its expected results.

A and B get each practice first. C adopts it last, through `just update`.
A card that adds an evaluator ships its template form in the same pull request.

## Move a correction

1. Choose the owner with D067.
2. Choose the highest level that can prevent the failure:
   - Level 1 when one type, one owner, or one constructor makes the wrong state impossible to write
   - Level 2 when a machine can decide from the code, a declaration, or an observed run
   - Level 3 only for a judgment that no machine can make
   - Level 4 for authority, perception, and consent
3. Prove the correction with a fixture that reproduces the past failure. The fixture fails without the correction.
4. Update the rule index of the owner, and add a row with its level and owner to the correction ledger.

A rule index gives each rule one enforcer: a link to its check, `Prose only`, or `Prose on purpose` with the judgment that keeps it in prose.
`Prose only` marks a rule that waits for a check.
A card that adds the check moves the rule up.
A correction can also land below the level of step 2, such as one scenario for a class that a type could prevent.
Its ledger row shows that gap, and a later card moves the class up.

When an extension finds a correction that belongs to A or B:

1. Its agent opens an issue in pascalandy/nicely with the input, the host, and the failure.
2. The extension records a limit under **Limits.** of `spec.md` that links the issue. The user accepts it like any other limit.
3. A Nicely release carries the fix. After `just update`, a proven line replaces the limit.

The `nicely-extension` skill teaches this step.

## Measures

These three measures replace the score of the 36 trial cases.

| Measure | What it counts | Where to read it | Start |
|---|---|---|---|
| 1. Pyramid shape | The rules whose enforcer is `Prose only`, for each owner | `just status` for the core and template indexes, `extensionrepo check` for an extension | A: 79 of 154 rules. B: 24 of 57 rules ([N2][n2]). C, video-convert: 3 of 15 ([S4][s4]) |
| 2. Correction ledger | The level where each correction landed, and its owner | The Level and Owner columns of the Corrections table in the [QA ledger](../../qa/ledger.md) | 66 of 237 past corrections landed below their highest working level ([N1][n1]) |
| 3. Real use | Video-convert adopts the loop, and its corrections feed measure 2 | COR-104-T10, and the rows of video-convert in the ledger | Not adopted at the 2026-11-02 start. Adopted on 2026-11-04 in [COR-104-T10](../../backlog/COR-104-verification-map.md#cor-104-t10-video-convert-adopts-the-loop) |

**Why.** Since 2026-10-01, 232 of 237 corrections reached level 1 or 2 ([N1][n1]).
But 66 landed below the highest level that works, and four classes hold 50 of them.
These classes are review prose read as state, caller state in checks, lost notices, and process lifetime.
The trial score measured a fixed set of cases. A full score still hid gaps in private data, preparation, and review.
A map tells each correction its owner and level before anyone writes it.
An SDK analyzer reached an extension at its pinned version through the unchanged template recipe ([P1][p1]).
The TLC runner moved into the kit, and video-convert kept the same 12 outcomes ([P2][p2]).

**Rejected.** A fifth level for formal models, because a model is a check of level 2.
A new count of trial cases, because a count does not say where a correction belongs.
A publication service in the SDK, because it takes publication away from the extension.
A copy of the TLC runner or of analyzer rules in each extension, for the reasons of D067.
One count of all prose, because a judgment belongs in prose and would hide the rules that wait for a check.

[n1]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/analysis/N1-correct.md
[s4]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/scouts/S4-inventory.md
[n2]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/analysis/N2-prose-triage.md
[p1]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/proto/P1/REPORT.md
[p2]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/proto/P2/REPORT.md
