---
id: COR-103
core: "verification loop"
state: "Active. 15 cards across nicely and ncly-video-convert"
---

# COR-103 verification loop

## Demo

The kit checks executed calls and declared codes, and video-convert proves its error cases, release, and publication by machine.

## Scope

Share the rule, keep the facts, copy only the wiring.
A check's evaluator lives in the SDK or the kit.
The extension supplies its facts: inputs, fixtures, expected values, and independent oracles.
The template copies only the lines that run them, and the forms an author fills.
Each check fails first on one named past mistake.
[The plan](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md) holds the details, the backtest of 164 defects, and the prototypes.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| COR-103-T1 | Video-convert error cases and still images | agent | - | todo |
| COR-103-T2 | Declared codes with scopes | agent | - | done |
| COR-103-T3 | Count executed calls | agent | - | done |
| COR-103-T4 | Run manifest examples | agent | COR-103-T3 | todo |
| COR-103-T5 | Observe declared codes | agent | COR-103-T2, COR-103-T3 | todo |
| COR-103-T6 | Signal and refuses commands | agent | COR-103-T3 | done |
| COR-103-T7 | One install and uninstall path | agent | - | done |
| COR-103-T8 | Update through the target updater | agent | COR-103-T7 | todo |
| COR-103-T9 | Sign off an isolated checkout | agent | COR-103-T7 | todo |
| COR-103-T10 | Release with a receipt | agent | COR-103-T7, COR-103-T9 | todo |
| COR-103-T11 | macOS job in the template | agent | - | done |
| COR-103-T12 | Author path from the published tag | agent | - | done |
| COR-103-T13 | Rules and prose | agent | COR-103-T4, COR-103-T5, COR-103-T6, COR-103-T8, COR-103-T10, COR-103-T11, COR-103-T12 | todo |
| COR-103-T14 | Video-convert rollout | agent | COR-103-T1, COR-103-T13 | todo |
| COR-103-T15 | Video-convert publication model | agent | COR-103-T14 | todo |

### COR-103-T1 Video-convert error cases and still images

- **Read:** [qa-scenarios](https://github.com/pascalandy/nicely-research/tree/main/synthesis/verification-loop/prototypes/qa-scenarios), [plan step 1](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** testdata/script and acceptance.sh, in ncly-video-convert

In ncly-video-convert, commit the 14 prototype scenarios and refuse a still image, as `spec.md` decides (#351).
In the same pull request, `acceptance.sh` requires that refusal, and `spec.md` and `qa.md` stop accepting still images.
Acceptance fully decodes the H.265 output.

### COR-103-T2 Declared codes with scopes

- **Read:** [Manifest](../north-star/contract/manifest.md), [Valid manifest](../north-star/contract/valid-manifest.md), [plan step 2](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `sdk/extension/extension_test.go`

The manifest declares warnings, and each code takes optional commands, modes, and an `untested` reason with an issue.
Migrate transcript, agent, and the fixtures first.
`modes = ["typo"]` must fail.

### COR-103-T3 Count executed calls

- **Read:** [Test kit](../north-star/contract/test-kit.md), [plan step 3](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `sdk/extensiontest/extensiontest_test.go`

Coverage counts recorded calls after their answer check, never a text mention.
A path-qualified call is checked or refused, and a failed report is checked whole.
Both cases of #335 fail first.

### COR-103-T4 Run manifest examples

- **Read:** [Test kit](../north-star/contract/test-kit.md), [agent manifest](../../extensions/agent/manifest.toml), [plan step 4](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `sdk/extensiontest/extensiontest_test.go`

The kit runs each example with fake programs, as a dry run when the command offers one.
Otherwise it runs an example only when the command declares no paid effect.
The agent example without its model fails first.

### COR-103-T5 Observe declared codes

- **Read:** [declared-codes](https://github.com/pascalandy/nicely-research/tree/main/synthesis/verification-loop/prototypes/declared-codes), [plan step 5](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `sdk/extensiontest/extensiontest_test.go`

Each declared code is observed in each command and mode of its scope, or carries `untested`.
The kit lists each exemption, and the UAT report shows them.
A mention in a comment never counts.

### COR-103-T6 Signal and refuses commands

- **Read:** [Test kit](../north-star/contract/test-kit.md), [Global flags](../north-star/contract/global-flags.md), [plan step 6](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `sdk/extensiontest/extensiontest_test.go`

`signal` requires 130 and `INTERRUPTED` for INT, or 143 and `TERMINATED` for TERM, and checks what survives.
`refuses` runs one input as a dry run and a run, requires the same code, and keeps the protected files.
A fake file that passes the dry run fails first, as in v0.1.1.

### COR-103-T7 One install and uninstall path

- **Read:** [Extension repository](../north-star/contract/extension-repository.md), [template justfile](../../template/justfile), [plan step 7](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `sdk/extensiontest/extensiontest_test.go`, `scripts/template-recipes-test.sh`

An SDK program installs with the executable last and uninstalls it first.
The kit's scratch install uses it, and each suite ends with `NOT_FOUND`.
The recipes become one line.

### COR-103-T8 Update through the target updater

- **Read:** [template justfile](../../template/justfile), [plan step 8](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `scripts/template-update-test.sh`, `scripts/template-test.sh`

`just update` runs the updater of the target version.
An old repository uses a bootstrap through `go run`.
An untouched v0.23.2 render must update and recover from a conflict.

### COR-103-T9 Sign off an isolated checkout

- **Read:** [signoff.sh](../../scripts/signoff.sh), [check-layers.sh](../../scripts/check-layers.sh), [plan step 9](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `scripts/signoff-test.sh`

The SDK program's `signoff` checks a frozen SHA in a detached worktree, writes a receipt, and signs that SHA.
The template and core's `signoff.sh` call it.
A dirty fix that passes in the working tree must fail.

### COR-103-T10 Release with a receipt

- **Read:** [QA north star](../qa/north-star.md), [template AGENTS.md](../../template/AGENTS.md.jinja), [plan step 10](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `scripts/template-release-test.sh`

`just release` requires the receipt of the pushed SHA, runs acceptance, and puts the receipt in the tag.
It then installs through the README from the published remote, and requires the tagged SHA.
It measures Economy and uninstalls.
A README that installs the previous tag fails first.

### COR-103-T11 macOS job in the template

- **Read:** [D016](../north-star/decisions/D016-local-ci.md), [plan step 11](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `template/.github/workflows/check.yml`

Add a `macos-15` job for public repositories.
Core keeps its Mini check at release until #382 is fixed.

### COR-103-T12 Author path from the published tag

- **Read:** [release-nicely](../../.agents/skills/release-nicely/SKILL.md), [plan step 12](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `.agents/skills/release-nicely/scripts/author-path.sh`

At release, render the published tag through the module proxy, then follow the documented first steps.
v0.18.3 must reproduce the findings of #323 and #324, and the current tag must pass.

### COR-103-T13 Rules and prose

- **Read:** [Decisions](../north-star/decisions/README.md), [template AGENTS.md](../../template/AGENTS.md.jinja), [plan step 13, sections 5 and 6](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** `scripts/template-test.sh`

A decision records the placement rule, and core's AGENTS.md routes to it.
The template keeps one index of rules and their enforcers.
The skills say where a `--json` error goes (#352).
The template test requires `## Acceptance` in the reading order.

### COR-103-T14 Video-convert rollout

- **Read:** [plan step 14](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** testdata/script, in ncly-video-convert

In ncly-video-convert, after the release of T2 to T13, update through the T8 bootstrap and declare the two warnings.
Move T1's scenarios onto `signal` and `refuses`.
`GPU_RUNTIME_MISSING` must fail in both modes until a scenario observes it.

### COR-103-T15 Video-convert publication model

- **Read:** [publication](https://github.com/pascalandy/nicely-research/tree/main/synthesis/verification-loop/prototypes/publication), [plan step 15](https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-loop/plan.md)
- **Proves:** the record test and just verify-publication, in ncly-video-convert

In ncly-video-convert, commit the TLA+ model, its probe, and its skill behind `just verify-publication`.
A record test in `just check` fails when the record is absent, or older than the code, model, probe, or tool pins.
The probe requires that a signal after the commit still names the output.
