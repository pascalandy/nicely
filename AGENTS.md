# AGENTS.md

This file tells an agent how to change Nicely. [README.md](README.md) describes Nicely for its users.

## Read in this order

1. [docs/north-star/guide.md](docs/north-star/guide.md), at the start of every session.
2. The current milestone in `docs/milestones/`: the lowest-numbered file whose status is not `done`.
3. The sections of [docs/north-star/cli-spec.md](docs/north-star/cli-spec.md) that your task touches.
4. [docs/north-star/decision-records.md](docs/north-star/decision-records.md), only when a rule blocks your task.

Take rules only from these files. `docs/archived/` holds history, such as the design chat.

## Work on a milestone task

Pick the next unticked task of the current milestone, and set the milestone to `active` when its first task starts. When the milestone has no task list yet, write one first, as M0 does. Copy this checklist and tick each step when its condition holds. Mark a step `n/a` with its reason when the task has nothing for it, such as a task that adds no user-facing text.

```
- [ ] 1. Spec: cli-spec.md covers every command, flag, JSON key, and error code you add or change
- [ ] 2. Scenario: a testscript scenario fails for the missing behavior
- [ ] 3. Code: the scenario passes
- [ ] 4. Text: each new user-facing string is a catalog entry with a description
- [ ] 5. Checks: `just check` passes
- [ ] 6. Docs: the milestone boxes you finished are ticked, and README.md shows any change a user sees
- [ ] 7. Sign off: `just signoff` passes
```

1. **Spec first.** Write the section in cli-spec.md before code, tagged with the milestone. When the milestone file holds a draft of it, move the draft and leave a link in its place. Done when a reader can call the command from the spec alone.
2. **Scenario before code.** Scenarios live in `testdata/script/` as `.txtar` files. Reuse the [M0 contract coverage](docs/milestones/M0-foundation.md#contract-coverage) and add the command's distinct risks. For writing or paid work, verify filesystem effects, partial results, and retry safety, not just the JSON text. Done when the scenario fails for the missing behavior. A shared-component test does not replace the first real command's end-to-end proof.
3. **Code.** Follow [Code](#code). Done when the scenario passes.
4. **Text.** Give each catalog entry an ID and a description of where the text appears, so a translator picks the right sense.
5. **Checks.** When `just check` fails, fix the cause and rerun. After three failed attempts on the same check, stop and report what you tried.
6. **Docs.** Tick the boxes of the milestone that your work completes, and update README.md when a user would see the change. When every box of the "Done when" list is ticked, set the milestone status to `done`.
7. **Sign off.** `just signoff` reruns the checks, then calls `gh signoff`. Always sign off through the recipe, because a bare `gh signoff` posts a green status without running anything.

## Rules

- Where cli-spec.md is silent on CLI design, follow what serious CLIs agree on, such as `gh`, `kubectl`, Terraform, Docker, and cargo, and the guidelines of clig.dev. The `coding-standard` skill condenses clig.dev. When the choice becomes a rule, record the comparison in decision-records.md.
- Prefer end-to-end testscript scenarios. Write a unit test only for logic that a scenario cannot reach.
- Write code, comments, docs, and commit messages in English.
- Stay inside the current milestone. Write any other idea as one line in [M99](docs/milestones/M99-parking-lot.md), then continue.
- To change a principle or the agent contract, add or fix an entry in decision-records.md. Ask Pascal before merging, unless Pascal asked for the change.
- Use placeholders such as `/Users/me`, `host-a`, and `example.com` in code, tests, and docs. The pre-commit hook runs gitleaks with generic rules for home paths and private IP ranges. Rules that name Pascal's hosts live in `lefthook-local.yml`, which git ignores.

## Code

```
cmd/ncly/           entry point, a few lines
internal/cli/       one package per domain or standalone command
internal/contract/  output, error registry, exit codes, modes, and the streams every command writes to
internal/i18n/      catalogs in locales/, lookup, plurals, and number, size, and date formats
internal/config/    shared and local config files, environment, flags, and XDG paths
internal/operation/ run records and result publication, first used by transcript in M1
internal/agent/     harness adapters and profile execution, first used by summaries in M1
internal/platform/  OS differences: open, clipboard, trash, keychain, and file locks
internal/run/       runs other programs: Python programs, harnesses, and extensions
internal/tui/       the shared interactive parts
python/             Python programs while they move to Go
testdata/script/    testscript scenarios
tools/              a separate module that pins the dev tools, run through `just`
```

Run `lefthook install` once per clone, so the hooks format, lint, and scan each commit and test each push.

- Build interactive screens from `internal/tui` only, so every command looks the same. Add a component there when the first command needs it.
- Keep one declaration per command for help, completion, discovery, and validation, as [Command descriptions](docs/north-star/cli-spec.md#command-descriptions-m0) defines. Domain code owns the steps and resume evidence. Shared operation support follows [Operations](docs/north-star/cli-spec.md#operations-m0-contract-m1-execution).
- Run every other program through `internal/run`, which implements [Programs that ncly runs](docs/north-star/cli-spec.md#programs-that-ncly-runs-m1). Start each program in its own process group, so a signal reaches its descendants, apply the timeout, and keep keys out of every log.
- Write a shared file to a temporary file and rename it under a lock from `internal/platform`. Revalidate the destination while holding the lock. Give each run its own temporary files. A lock conflict follows [Retry safety](docs/north-star/cli-spec.md#retry-safety-m0), including effects already produced by this invocation.
- Before changing a machine-readable value or stored format, check [Compatibility](docs/north-star/cli-spec.md#compatibility-m0) and add a consumer case that proves the change preserves the promised meaning.
- Move a deleted user file to the trash through `internal/platform`.
- Send every string a human reads through the catalog. Name catalog IDs `<domain>.<thing>`, and never reuse an ID for a new meaning. Give every message that holds a number its plural forms. Format numbers, sizes, durations, and dates through `internal/i18n`. Let translated text set its own width.
- In scenarios, point `HOME` and the XDG folders at `$WORK`, put stub executables in `PATH` for harnesses, `uv`, and paid services, and use the test keychain that lives only in the test binary. Run each command's scenarios once under the pseudo-locale, so text outside the catalog fails a check. `exits <code> <command>` asserts an exact exit code, because the built-in `! exec` only asserts a failure.

## Commit

Write each commit message as `<emoji> <type>: <scope>: <summary>`, such as `📚 docs: guide: add the report command term`. The `commit` skill lists the types and the body format.

## Merge

"Merge" means ship: commit, push, sign off, open a pull request, merge it into origin/main, and sync local main. Don't ask for confirmation at any step, except the one that the contract rule in [Rules](#rules) requires. Stop only on a real blocker, and report it.

Keep the history linear. Rebase the branch onto origin/main, sign off its head, then push that head to `main` with `git push origin <branch>:main`. GitHub marks the pull request merged, and `main` keeps the commit that was checked. Then fast-forward the local `main` checkout.

## Release

Release only when Pascal asks.

1. Run `just release-check vX.Y.Z` and fix anything it reports.
2. Tag and push: `git tag vX.Y.Z && git push origin vX.Y.Z`.
3. Start the release: `gh workflow run release.yml -f tag=vX.Y.Z`. The workflow publishes the archives and the AUR package, then updates the formula in the Homebrew tap.

GitHub Actions runs only when started this way.

## Review an external pull request

1. Run `gh pr checkout <number>`.
2. Compare the diff with the current milestone and cli-spec.md.
3. Run `just signoff`. The signoff status is a required check on `main`, so the pull request merges only after this step.
