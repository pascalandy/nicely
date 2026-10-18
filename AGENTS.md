# AGENTS.md

This file tells an agent how to change Nicely. [README.md](README.md) describes Nicely for its users.

## Read in this order

1. [vision.md](docs/north-star/vision.md), [principles.md](docs/north-star/principles.md), [architecture.md](docs/north-star/architecture.md), and [dev-preferences.md](docs/north-star/dev-preferences.md), at the start of every session.
2. Run `just next` on a checkout of the latest `origin/main`, or of the top of the stack you are extending, because it reads the plan of the checkout. It prints the one card to do, or the step that unblocks it. Read what the card links under **Read**: sections of [contract.md](docs/north-star/contract.md) and [core-spec.md](docs/north-star/core-spec.md) for core, and of `extensions/<name>/spec.md` for an extension.
3. A decision in [decisions/](docs/north-star/decisions/README.md), only when a rule blocks your card.

Take rules only from these files and the specs that your card links. `docs/archived/` holds history, such as the design chat. `just status` shows every milestone and the cards of the current one.

## Find what you need

Each file answers one question. Read a file when its row says so, and add to it when its row says so.

| Question | File | Read when | Add to it when |
|---|---|---|---|
| What is Nicely, and how do agents use it? | [vision.md](docs/north-star/vision.md) | Every session | The purpose or the audience changes |
| Which rules shape every design choice? | [principles.md](docs/north-star/principles.md) | Every session | Pascal adds or changes a principle, with its decision |
| Where does each part live, and who owns it? | [architecture.md](docs/north-star/architecture.md) | Every session | A layer, a part, a package, or a domain appears |
| What does a word mean, and which words to avoid? | [glossary.md](docs/north-star/glossary.md) | A word is unclear | Before the first use of a new term |
| Which rules does every command follow? | [contract.md](docs/north-star/contract.md) | Your card links a section | A card adds a primitive, a flag, a code, or a format |
| What do core's commands do? | [core-spec.md](docs/north-star/core-spec.md) | Your card links a section | A card adds or changes a command of core |
| What do an extension's commands do? | `extensions/<name>/spec.md` | Your card links a section | A card adds or changes a command of that extension |
| Why does a rule exist? | [decisions/](docs/north-star/decisions/README.md) | A rule blocks your card | A rule is chosen, changed, or replaced |
| How does Pascal want changes made? | [dev-preferences.md](docs/north-star/dev-preferences.md) | Every session | Pascal states a new preference |
| What is the plan? | [docs/milestones/](docs/milestones/) | `just next` names a milestone | Planning work. Any other idea goes to M99 |
| How does an agent prove a card on the real `ncly`? | [verify-nicely](.agents/skills/verify-nicely/SKILL.md) | The Verify step of your card | A card changes the steps or the end state of a feature, or adds one |

## Work on a card

A card is one pull request. Do the card that `just next` prints, and only that card. When `just next` lists a card waiting on Pascal, prepare what that card asks if nobody has yet, such as a mockup, tell Pascal, and stop once no agent card can start.

Copy this checklist into the pull request and tick each step when its condition holds. Mark a step `n/a` with its reason when the card has nothing for it, such as a card that adds no user-facing text.

```
- [ ] 1. Spec: the sections that the card reads cover every command, flag, JSON key, and error code you add or change
- [ ] 2. Scenario: each scenario that the card proves fails for the missing behavior
- [ ] 3. Code: those scenarios pass
- [ ] 4. Text: each new user-facing string is a catalog entry with a description
- [ ] 5. Checks: `just check` passes
- [ ] 6. Verify: an agent that did not write the change follows `.agents/skills/verify-nicely/SKILL.md` on each feature that the card affects, and the pull request quotes its evidence
- [ ] 7. Docs: the card is done in its milestone, README.md shows any change a user sees, and a skill teaches any command you add: the extension's skill, or the `nicely` skill for a core command
- [ ] 8. Sign off: `just signoff` passes
```

1. **Spec first.** Write the section before code: in contract.md or core-spec.md for core, or in `extensions/<name>/spec.md` for an extension. When a source that the section cites has changed, such as a harness version, fix the spec first. Done when a reader can call the command from the spec alone.
2. **Scenario before code.** Scenarios live in `testdata/script/`, or in `extensions/<name>/testdata/script/` for an extension, as `.txtar` files, under the names that the card's **Proves** line gives. Reuse the [M00 contract coverage](docs/milestones/M00-foundation.md#contract-coverage) and add the command's distinct risks. For writing or paid work, verify filesystem effects, partial results, and retry safety, not just the JSON text. For a recorded operation, add a boundary scenario. The boundary hook kills the process that runs the operation at every step boundary, and the scenario checks the answer, the record, the files, and the counted calls after each kill. Once the operation supports resume, it also checks that a resume never repeats an unknown effect. Run scenarios by name with `just scenario <name>... [--count N]`. It fails unless each scenario passes N times. A raw `go test -run` pattern can select nothing and still pass. Done when the scenario fails for the missing behavior. A shared-component test does not replace the first real command's end-to-end proof.
3. **Code.** Follow [Code](#code). Done when the scenario passes.
4. **Text.** Give each catalog entry an ID and a description of where the text appears, so a translator picks the right sense.
5. **Checks.** When `just check` fails, fix the cause and rerun. After three failed attempts on the same check, stop and report what you tried. For a stack, `just check-layers` runs `just check` on each layer as origin holds it, bottom to top. It fails a layer that does not contain the layer below. Its run's folder keeps one log per layer, and `summary.txt` with the full SHA and the verdict of each layer. Pass the branches bottom to top when the checkout is not on a branch of the stack.
6. **Verify.** When the card changes the steps or the end state of a feature in `.agents/skills/verify-nicely/features/`, or adds one, the implementer first updates or adds its file. Then an agent that did not write the change follows the skill on each feature that the card affects. It runs on a small model of the other family than the implementer's, at effort `xhigh`, in the same checkout, so it sees uncommitted changes. The pull request quotes the lines of its evidence that prove each end state, with home paths replaced by placeholders. It names the commit that `evidence/commit.txt` records, followed by `dirty` when the file says so. A dirty run proves the checkout as it was, not that commit alone. A fix that changes a verified behavior needs a run that covers it. A run at the top of the stack covers it when it drives the features that the fix affects. A restack keeps the recorded commit as it is. The pull request names the new head beside it, and the run that covers each later fix. A card that reaches nothing the scenarios replace, such as the real keychain, a real terminal, a harness, `uv`, or a paid service, marks the step `n/a`. Done when the quoted evidence names its commit and proves the end state of each feature that the card affects.
7. **Docs.** In the milestone's cards table, set the card's status to `done`. The milestone becomes `active` with its first done card and `done` with its last, and `just test` rejects any other combination. Update README.md when a user would see the change. Add a line for a new command to the skill that teaches its domain, or to the `nicely` skill for a core command.
8. **Sign off.** `just signoff` reruns the checks, then calls `gh signoff`. Always sign off through the recipe, because a bare `gh signoff` posts a green status without running anything.

## Make a milestone ready

When `just next` names a planned milestone with open questions, make it ready in one pull request:

1. Settle each open question in the spec that it names, then delete the **Open questions** section. Ask Pascal only what is his to decide, such as a look, a cost, or a principle.
2. Finish the cards. Trace each card through the [code map](docs/north-star/architecture.md#code-map): name the package that it changes, and check that this package may reach what the card needs. A part that the package may not import, such as the host's discovery for an extension, is reached through the CLI, after the milestone that adds its command. A card changes one capability, in one pull request that a reviewer reads in one sitting. It links what to read and names the scenarios that prove it. A milestone has at most five agent cards: split a bigger one into two milestones. A decision that only Pascal can make, such as a pick, is a card whose owner is `Pascal`, and the cards that wait for it depend on it. A step that waits on Pascal while no card waits on it, such as a check on his Mac, goes under **After this milestone** instead, so it never blocks the next milestone.
3. Keep the demo to one sentence that the cards prove together.
4. Set the status to `ready`. `just next` then prints the first card.

A planned milestone without open questions needs no separate pull request: `just next` prints its first card, and that card's pull request also sets the milestone to `active`.

A milestone file holds `Status`, `Version`, and the sections **Demo**, an optional **Scope**, **Open questions** while it is planned, **Cards**, and an optional **After this milestone**. **Cards** holds a table with the columns Card, Title, Owner, Depends on, and Status, then one `### <card> <title>` section per card. An agent card's section starts with a **Read** line that links what to read and a **Proves** line that names its proof. `just test` checks this format, including the links. No milestone file holds a checkbox: work that matters is a card. A new milestone takes the number after the last one. To change the order, renumber only milestones without a done card, in a pull request that only changes the plan.

## Rules

- Where the specs are silent on CLI design, follow what serious CLIs agree on, such as `gh`, `kubectl`, Terraform, Docker, and cargo, and the guidelines of clig.dev. The `coding-standard` skill condenses clig.dev. When the choice becomes a rule, record the comparison in a new decision in [decisions/](docs/north-star/decisions/README.md).
- Prefer end-to-end testscript scenarios. Write a unit test only for logic that a scenario cannot reach.
- Write code, comments, docs, and commit messages in English.
- Stay inside your card. Write any other idea as one line in [M99](docs/milestones/M99-parking-lot.md), then continue.
- To change a principle or the contract, add or fix a decision in [decisions/](docs/north-star/decisions/README.md).
- Use placeholders such as `/Users/me`, `host-a`, and `example.com` in code, tests, and docs. The pre-commit hook runs gitleaks with generic rules for home paths and private IP ranges. Rules that name Pascal's hosts live in `lefthook-local.yml`, which git ignores.

## Code

The [code map](docs/north-star/architecture.md#code-map) names the package and the spec section of each part. The rules below name each package by its role.

Run `lefthook install` once per clone, so the hooks format, lint, and scan each commit and test each push. On Pascal's machines, a fetch or a push that moves `origin/main` also fast-forwards the main clone, unless it is off `main`, has uncommitted changes to tracked files, or has commits origin lacks.

- Keep the code of `sdk/` and `extensions/` free of imports from `internal/`, as `just lint` checks. The one exception is a test that imports `internal/specdoc` to compare the code with the docs. An extension reaches core through its environment and the CLI, and another extension only through the CLI, as [Extensions](docs/north-star/contract.md#extensions) says. Adding a domain never edits core.
- Build interactive screens from the shared tui package only, so every command looks the same. Add a component there when the first command needs it, after Pascal picks its look from a published mockup, as [dev-preferences.md](docs/north-star/dev-preferences.md#decide-the-look-with-a-published-mockup) asks. The spec of each command holds the look that Pascal picked in [M07](docs/milestones/M07-forms.md). A command whose spec gives no look reuses the help and error styles of M00 and prints the rest as plain text, so it waits on no mockup.
- Keep one declaration per command for help, completion, discovery, and validation, as [Command descriptions](docs/north-star/contract.md#command-descriptions) defines. An extension's manifest holds its declarations. The extension owns its steps and resume evidence. Record support follows [Operations](docs/north-star/contract.md#operations).
- Follow [Programs that ncly runs](docs/north-star/contract.md#programs-that-ncly-runs) for process groups, timeouts, cleanup, and output streams. That section names the macOS `security` exception.
- Keep keys out of every log.
- When the runner reads a program's stdout or stderr, use pipes it owns. A writer copied by `exec.Cmd` can keep `Wait` blocked until descendants close the pipe.
- Send cleanup signals through the runner to the whole process tree. `exec.CommandContext` kills only the direct child by default. Keep the signal handler registered until cleanup ends. Stopping `signal.NotifyContext` first lets a second Ctrl-C end `ncly` immediately.
- Shorten the 15-second cleanup limit in the test binary to prove the SIGKILL path without waiting.
- Write a shared file to a temporary file and rename it under a lock from the platform helpers. Revalidate the destination while holding the lock. Give each run its own temporary files. Write the record during cleanup with a context that cancellation does not end, such as `context.WithoutCancel`, within the 15 seconds of cleanup. A lock conflict follows [Retry safety](docs/north-star/contract.md#retry-safety), including effects already produced by this invocation.
- Before changing a machine-readable value or stored format, check [Compatibility](docs/north-star/contract.md#compatibility) and add a consumer case that proves the change preserves the promised meaning.
- Move a deleted user file to the trash through the platform helpers.
- When a hint names a command that runs as is, put the same command in the problem's `next`, as [Output](docs/north-star/contract.md#output) defines.
- Send every string a human reads through the catalog. Name catalog IDs `<domain>.<thing>`, and never reuse an ID for a new meaning. Give every message that holds a number its plural forms. Format numbers, sizes, durations, and dates through the i18n package. Let translated text set its own width.
- In scenarios, point `HOME` and the XDG folders at `$WORK`, put stub executables in `PATH` for harnesses, `uv`, and paid services, and use the test keychain and the boundary hook, which live only in the test binary. `onlyncly [command...]` keeps in `PATH` only `ncly`, the listed commands of the test binary, and `$WORK/bin` when it exists. The other `ncly-*` commands of the test binary are then no extensions, and a system program needs its absolute path, such as `/usr/bin/sed`. Run each command's scenarios once under the pseudo-locale, so text outside the catalog fails a check: `pseudo <stdout|stderr> [word...]` fails on any unmarked word other than flags, codes, placeholders, example lines, and the listed command names. `exits <code> <command>` asserts an exact exit code, because the built-in `! exec` only asserts a failure. `answer <stdout|stderr> <ok|CODE>` decodes the whole JSON answer, `answer stdout report` decodes a failed report, which has `ok: false` and no `errors` key, `snapshot` with `unchanged` proves that a dry run changed nothing outside Nicely's cache, `agree <help-file> <describe-file>` proves that a help page agrees with the `ncly describe` answer of its command, `schema <answer-file> <describe-file> <schema|dry_run_schema>` validates an answer against the schema that describe gives for it, and `mentions <skill-file> <describe-file> <codes-file>` proves that a skill names only the commands that describe gives and the codes that the `codes` program of the test binary lists.

## Commit

Write each commit message as `<emoji> <type>: <scope>: <summary>`, such as `📚 docs: glossary: add the report command term`. The `commit` skill lists the types and the body format.

## Release

Release only when Pascal asks. A release is never a card, so it never blocks a milestone. Milestone `M<n>` ships in `v0.<n>.0`, M00 in v0.0.1, and a fix between two milestones takes the next patch number, as [D043](docs/north-star/decisions/D043-milestone-versions.md) decides.

1. Run `just release-check vX.Y.Z` and fix anything it reports.
2. Tag and push: `git tag vX.Y.Z && git push origin vX.Y.Z`.
3. Start the release: `gh workflow run release.yml -f tag=vX.Y.Z`. The workflow publishes the archives, then updates the formula in the Homebrew tap.

GitHub Actions runs only when started this way. After the first release, Pascal checks on his Mac that `brew install pascalandy/tap/ncly` builds `ncly`, and that `ncly <Tab>` completes in zsh.

## Review an external pull request

1. Run `gh pr checkout <number>`.
2. Compare the diff with its card and the specs that the card reads.
3. Run `just signoff`.
