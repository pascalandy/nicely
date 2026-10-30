# AGENTS.md

This file routes an agent to the rules for changing Nicely.

## Start a session

1. Read each file whose row in [Find what you need](#find-what-you-need) says Every session.
2. Run `just next` on a checkout of the latest `origin/main`, or of the top of the stack you extend. It prints the one card to do, or the step that unblocks it.
3. Read what the card links under **Read**. Take rules only from the files that this one names and those that your card links, never from `docs/archived/`.

Run `lefthook install` once per clone.

## Find what you need

| Question | File | Read when | Add to it when |
|---|---|---|---|
| What is Nicely, and how do agents use it? | [vision.md](docs/north-star/vision.md) | Every session | The purpose or the audience changes |
| Which rules shape every design choice? | [principles.md](docs/north-star/principles.md) | Every session | Pascal changes a principle, with its decision |
| Where does each part live, and who owns it? | [architecture.md](docs/north-star/architecture.md) | Every session | A layer, a part, a package, or a domain appears |
| How does Pascal want changes made? | [dev-preferences.md](docs/north-star/dev-preferences.md) | Every session | Pascal states a new preference |
| What does a word mean? | [glossary.md](docs/north-star/glossary.md) | A word is unclear | Before the first use of a new term |
| Which rules does every command follow? | [contract.md](docs/north-star/contract.md) | Your card links a section | A card adds a primitive, a verb, a flag, a code, or a format |
| What do core's commands do? | [core-spec.md](docs/north-star/core-spec.md) | Your card links a section | A card adds or changes a command of core |
| What do an extension's commands do? | `extensions/<name>/spec.md` | Your card links a section | A card adds or changes a command of that extension |
| Why does a rule exist? | [decisions/](docs/north-star/decisions/README.md) | A rule blocks your card | A rule is chosen, changed, or replaced |
| How is the plan made? | [docs/backlog/](docs/backlog/README.md) | `just next` names an item | Planning work. Any other idea becomes a 6xx item |
| How is a scenario written? | [testdata/script/](testdata/script/README.md) | Step 2 of a card | A card changes how a scenario is written |
| How is a card proven on the real `ncly`? | [verify-nicely](.agents/skills/verify-nicely/SKILL.md) | Step 6 of a card | A card changes or adds a feature |
| How is a release made? | [release-nicely](.agents/skills/release-nicely/SKILL.md) | Your merge ships an item or fixes a released behavior | The release steps change |
| Accept a UAT? | [docs/qa/](docs/qa/README.md) | A UAT is assigned | A UAT ends |
| Where does a new rule go? | The file whose question it answers | Before you write a rule | Never in this file |
| Where does a new check go? | [D067](docs/north-star/decisions/D067-share-verification-rules.md) | Before you add a check | The placement rule changes |

## Work on a card

A card is one pull request. Do the card that `just next` prints, and only that card. When Pascal assigns you an issue outside the plan, that issue is your card, and its pull request says it is outside the plan. When `just next` lists a card waiting on Pascal, prepare what it asks if nobody has yet, such as a mockup, tell Pascal, and stop once no agent card can start.

Copy this checklist into the pull request. Tick each step when its condition holds, or mark it `n/a` with its reason.

```
- [ ] 1. Spec: the sections that the card reads cover every command, flag, JSON key, and error code you add or change
- [ ] 2. Scenario: each scenario that the card proves fails for the missing behavior
- [ ] 3. Code: those scenarios pass
- [ ] 4. Text: each new user-facing string is a catalog entry with a description
- [ ] 5. Checks: `just check` passes
- [ ] 6. Verify: an agent that did not write the change runs verify-nicely on each feature that the card affects, and the pull request quotes its evidence
- [ ] 7. Docs: the card is done in its backlog item, README.md shows any change a user sees, and a skill teaches any command you add: the extension's, or `nicely` for core
- [ ] 8. Sign off: `just signoff` passes, or `just signoff-layers` for a stack
```

Write the spec first, and fix it first when a source it cites changes, such as a harness version. When `just check` fails three times on one check, stop and report what you tried. Sign off only through a recipe, because a bare `gh signoff` runs nothing.

## Code

The [code map](docs/north-star/architecture.md#code-map) names the package and the spec section of each part. Each package comment in `sdk/` holds the rules of its code, such as cleanup in `sdk/runner` and the look in `sdk/tui`.

## Rules

Read [RULES.md](RULES.md) every session for writing, commits, and external pull request reviews.
